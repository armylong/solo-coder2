package sqlite_long

import (
	"context"

	sqliteLongBiz "github.com/armylong/armylong-go/internal/business/sqlite_long"
	sqliteLongCs "github.com/armylong/armylong-go/internal/cs/sqlite_long"
)

// SQLite数据库管理控制器
type SqliteLongController struct{}

// 获取数据库概览
func (c *SqliteLongController) ActionOverview(ctx context.Context, req *sqliteLongCs.OverviewRequest) (*sqliteLongCs.OverviewResponse, error) {
	return sqliteLongBiz.SqliteLongBusiness.Overview(ctx, req)
}

// 获取表列表
func (c *SqliteLongController) ActionTableList(ctx context.Context, req *sqliteLongCs.TableListRequest) (*sqliteLongCs.TableListResponse, error) {
	return sqliteLongBiz.SqliteLongBusiness.TableList(ctx, req)
}

// 获取表数据
func (c *SqliteLongController) ActionTableData(ctx context.Context, req *sqliteLongCs.TableDataRequest) (*sqliteLongCs.TableDataResponse, error) {
	return sqliteLongBiz.SqliteLongBusiness.TableData(ctx, req)
}

// 获取表结构
func (c *SqliteLongController) ActionTableSchema(ctx context.Context, req *sqliteLongCs.TableSchemaRequest) (*sqliteLongCs.TableSchemaResponse, error) {
	return sqliteLongBiz.SqliteLongBusiness.TableSchema(ctx, req)
}

// 清空表数据
func (c *SqliteLongController) ActionClearTable(ctx context.Context, req *sqliteLongCs.ClearTableRequest) error {
	return sqliteLongBiz.SqliteLongBusiness.ClearTable(ctx, req)
}

// 删除单行
func (c *SqliteLongController) ActionDeleteRow(ctx context.Context, req *sqliteLongCs.DeleteRowRequest) error {
	return sqliteLongBiz.SqliteLongBusiness.DeleteRow(ctx, req)
}

// 执行SQL语句
func (c *SqliteLongController) ActionExecuteSql(ctx context.Context, req *sqliteLongCs.ExecuteSqlRequest) (*sqliteLongCs.ExecuteSqlResponse, error) {
	return sqliteLongBiz.SqliteLongBusiness.ExecuteSql(ctx, req)
}

// 更新行数据
func (c *SqliteLongController) ActionUpdateRow(ctx context.Context, req *sqliteLongCs.UpdateRowRequest) error {
	return sqliteLongBiz.SqliteLongBusiness.UpdateRow(ctx, req)
}

// 批量删除行
func (c *SqliteLongController) ActionDeleteRows(ctx context.Context, req *sqliteLongCs.DeleteRowsRequest) (*sqliteLongCs.DeleteRowsResponse, error) {
	return sqliteLongBiz.SqliteLongBusiness.DeleteRows(ctx, req)
}
