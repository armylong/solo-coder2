package sqlite_long

import (
	"context"
	"database/sql"
	"errors"
	"os"

	sqliteLong "github.com/armylong/armylong-go/internal/cs/sqlite_long"
	"github.com/armylong/go-library/service/sqlite"
)

type sqliteLongBusiness struct{}

var SqliteLongBusiness = &sqliteLongBusiness{}

// 获取数据库概览信息
func (b *sqliteLongBusiness) Overview(ctx context.Context, req *sqliteLong.OverviewRequest) (*sqliteLong.OverviewResponse, error) {
	db := sqlite.DB.DB()

	var tables []sqliteLong.TableInfo
	rows, err := db.Query(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name NOT LIKE 'sqlite_%' 
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		var rowCount int64
		db.QueryRow("SELECT COUNT(*) FROM " + tableName).Scan(&rowCount)

		columnCount := getTableColumnCount(db, tableName)

		tables = append(tables, sqliteLong.TableInfo{
			Name:        tableName,
			RowCount:    rowCount,
			ColumnCount: columnCount,
		})
	}

	dbPath := getDatabasePath()
	var dbSize int64
	if info, err := os.Stat(dbPath); err == nil {
		dbSize = info.Size()
	}

	return &sqliteLong.OverviewResponse{
		DatabasePath: dbPath,
		DatabaseSize: dbSize,
		TableCount:   len(tables),
		Tables:       tables,
	}, nil
}

// 获取表列表(分页)
func (b *sqliteLongBusiness) TableList(ctx context.Context, req *sqliteLong.TableListRequest) (*sqliteLong.TableListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	db := sqlite.DB.DB()

	var allTables []sqliteLong.TableInfo
	rows, err := db.Query(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name NOT LIKE 'sqlite_%' 
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		var rowCount int64
		db.QueryRow("SELECT COUNT(*) FROM " + tableName).Scan(&rowCount)

		tables := sqliteLong.TableInfo{
			Name:     tableName,
			RowCount: rowCount,
		}
		allTables = append(allTables, tables)
	}

	total := len(allTables)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return &sqliteLong.TableListResponse{
		Tables:   allTables[start:end],
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 校验SQL标识符是否合法，防注入
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	first := s[0]
	if first >= '0' && first <= '9' {
		return false
	}
	return true
}

// 获取表数据，支持筛选、排序和分页
func (b *sqliteLongBusiness) TableData(ctx context.Context, req *sqliteLong.TableDataRequest) (*sqliteLong.TableDataResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	db := sqlite.DB.DB()
	tableName := req.TableName

	whereClause := ""
	whereValues := make([]any, 0)
	if len(req.Filters) > 0 {
		first := true
		for col, val := range req.Filters {
			if !isValidIdentifier(col) {
				continue
			}
			if !first {
				whereClause += " AND "
			}
			whereClause += col + " LIKE ?"
			whereValues = append(whereValues, "%"+val+"%")
			first = false
		}
		if whereClause != "" {
			whereClause = " WHERE " + whereClause
		}
	}

	orderClause := ""
	if req.OrderBy != "" && isValidIdentifier(req.OrderBy) {
		orderDir := "ASC"
		if req.OrderDir == "DESC" || req.OrderDir == "desc" {
			orderDir = "DESC"
		}
		orderClause = " ORDER BY " + req.OrderBy + " " + orderDir
	}

	countSql := "SELECT COUNT(*) FROM " + tableName + whereClause
	var total int64
	if len(whereValues) > 0 {
		db.QueryRow(countSql, whereValues...).Scan(&total)
	} else {
		db.QueryRow(countSql).Scan(&total)
	}

	offset := (req.Page - 1) * req.PageSize
	selectSql := "SELECT * FROM " + tableName + whereClause + orderClause + " LIMIT ? OFFSET ?"

	var dataRows *sql.Rows
	var err error
	queryValues := make([]any, len(whereValues))
	copy(queryValues, whereValues)
	queryValues = append(queryValues, req.PageSize, offset)

	dataRows, err = db.Query(selectSql, queryValues...)
	if err != nil {
		return nil, err
	}
	defer dataRows.Close()

	columns, err := dataRows.Columns()
	if err != nil {
		return nil, err
	}

	var rows []map[string]any
	for dataRows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := dataRows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		rows = append(rows, row)
	}

	return &sqliteLong.TableDataResponse{
		TableName: tableName,
		Columns:   columns,
		Rows:      rows,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}

// 获取表结构信息
func (b *sqliteLongBusiness) TableSchema(ctx context.Context, req *sqliteLong.TableSchemaRequest) (*sqliteLong.TableSchemaResponse, error) {
	db := sqlite.DB.DB()

	rows, err := db.Query("PRAGMA table_info(" + req.TableName + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []sqliteLong.ColumnInfo
	for rows.Next() {
		var col sqliteLong.ColumnInfo
		var defaultVal *string
		if err := rows.Scan(&col.CID, &col.Name, &col.Type, &col.NotNull, &defaultVal, &col.PrimaryKey); err != nil {
			continue
		}
		if defaultVal != nil {
			col.DefaultVal = *defaultVal
		}
		columns = append(columns, col)
	}

	return &sqliteLong.TableSchemaResponse{
		TableName: req.TableName,
		Columns:   columns,
	}, nil
}

// 获取数据库文件路径
func getDatabasePath() string {
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		return homeDir + "/sqlite/database.db"
	}
	return "/tmp/sqlite/database.db"
}

// 获取表的列数
func getTableColumnCount(db *sql.DB, tableName string) int {
	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		return 0
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	return count
}

// 清空表数据
func (b *sqliteLongBusiness) ClearTable(ctx context.Context, req *sqliteLong.ClearTableRequest) error {
	db := sqlite.DB.DB()

	_, err := db.Exec("DELETE FROM " + req.TableName)
	if err != nil {
		return errors.New("清空表失败: " + err.Error())
	}

	return nil
}

// 删除单行
func (b *sqliteLongBusiness) DeleteRow(ctx context.Context, req *sqliteLong.DeleteRowRequest) error {
	db := sqlite.DB.DB()

	pkColumn := req.PrimaryKey
	if pkColumn == "" {
		pkColumn = "id"
	}

	result, err := db.Exec("DELETE FROM "+req.TableName+" WHERE "+pkColumn+" = ?", req.RowID)
	if err != nil {
		return errors.New("删除行失败: " + err.Error())
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("未找到要删除的行")
	}

	return nil
}

// 执行SQL语句，自动判断查询还是修改
func (b *sqliteLongBusiness) ExecuteSql(ctx context.Context, req *sqliteLong.ExecuteSqlRequest) (*sqliteLong.ExecuteSqlResponse, error) {
	db := sqlite.DB.DB()
	sqlStr := req.Sql

	sqlUpper := trimLeftSpace(sqlStr)
	isQuery := len(sqlUpper) >= 6 && (sqlUpper[:6] == "SELECT" || sqlUpper[:6] == "PRAGMA")

	if isQuery {
		dataRows, err := db.Query(sqlStr)
		if err != nil {
			return nil, errors.New("执行SQL失败: " + err.Error())
		}
		defer dataRows.Close()

		columns, err := dataRows.Columns()
		if err != nil {
			return nil, errors.New("获取列信息失败: " + err.Error())
		}

		var rows []map[string]any
		for dataRows.Next() {
			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := dataRows.Scan(valuePtrs...); err != nil {
				continue
			}

			row := make(map[string]any)
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			}
			rows = append(rows, row)
		}

		return &sqliteLong.ExecuteSqlResponse{
			IsQuery: true,
			Columns: columns,
			Rows:    rows,
			Total:   int64(len(rows)),
		}, nil
	} else {
		result, err := db.Exec(sqlStr)
		if err != nil {
			return nil, errors.New("执行SQL失败: " + err.Error())
		}

		affected, _ := result.RowsAffected()
		return &sqliteLong.ExecuteSqlResponse{
			IsQuery:  false,
			Affected: affected,
		}, nil
	}
}

// 去除左空白并转大写，用于判断SQL类型
func trimLeftSpace(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
			return toUpper(s[i:])
		}
	}
	return ""
}

func toUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// 更新行数据
func (b *sqliteLongBusiness) UpdateRow(ctx context.Context, req *sqliteLong.UpdateRowRequest) error {
	db := sqlite.DB.DB()

	pkColumn := req.PrimaryKey
	if pkColumn == "" {
		pkColumn = "id"
	}

	if len(req.Updates) == 0 {
		return errors.New("没有要更新的字段")
	}

	setClause := ""
	values := make([]any, 0)
	first := true
	for col, val := range req.Updates {
		if col == pkColumn {
			continue
		}
		if !first {
			setClause += ", "
		}
		setClause += col + " = ?"
		values = append(values, val)
		first = false
	}

	if setClause == "" {
		return errors.New("没有有效的更新字段")
	}

	values = append(values, req.RowID)
	sqlStr := "UPDATE " + req.TableName + " SET " + setClause + " WHERE " + pkColumn + " = ?"

	result, err := db.Exec(sqlStr, values...)
	if err != nil {
		return errors.New("更新失败: " + err.Error())
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("未找到要更新的行")
	}

	return nil
}

// 批量删除行
func (b *sqliteLongBusiness) DeleteRows(ctx context.Context, req *sqliteLong.DeleteRowsRequest) (*sqliteLong.DeleteRowsResponse, error) {
	db := sqlite.DB.DB()

	pkColumn := req.PrimaryKey
	if pkColumn == "" {
		pkColumn = "id"
	}

	if len(req.RowIDs) == 0 {
		return nil, errors.New("没有选择要删除的行")
	}

	placeholders := ""
	values := make([]any, len(req.RowIDs))
	for i, id := range req.RowIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
		values[i] = id
	}

	sqlStr := "DELETE FROM " + req.TableName + " WHERE " + pkColumn + " IN (" + placeholders + ")"

	result, err := db.Exec(sqlStr, values...)
	if err != nil {
		return nil, errors.New("批量删除失败: " + err.Error())
	}

	affected, _ := result.RowsAffected()
	return &sqliteLong.DeleteRowsResponse{
		Deleted: affected,
	}, nil
}
