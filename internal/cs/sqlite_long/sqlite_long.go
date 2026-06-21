package sqlite_long

// 数据库概览-请求
type OverviewRequest struct {
}

// 数据库概览-响应
type OverviewResponse struct {
	DatabasePath string      `json:"database_path"` // 数据库文件路径
	DatabaseSize int64       `json:"database_size"` // 数据库文件大小(字节)
	TableCount   int         `json:"table_count"`   // 表数量
	Tables       []TableInfo `json:"tables"`        // 表信息列表
}

// 表基本信息
type TableInfo struct {
	Name        string `json:"name"`         // 表名
	RowCount    int64  `json:"row_count"`    // 行数
	ColumnCount int    `json:"column_count"` // 列数
}

// 表列表-请求
type TableListRequest struct {
	Page     int `json:"page" form:"page"`           // 页码
	PageSize int `json:"page_size" form:"page_size"` // 每页数量
}

// 表列表-响应
type TableListResponse struct {
	Tables   []TableInfo `json:"tables"`    // 表信息列表
	Total    int         `json:"total"`     // 总数
	Page     int         `json:"page"`      // 当前页
	PageSize int         `json:"page_size"` // 每页数量
}

// 表数据-响应
type TableDataResponse struct {
	TableName string           `json:"table_name"` // 表名
	Columns   []string         `json:"columns"`    // 列名列表
	Rows      []map[string]any `json:"rows"`       // 数据行
	Total     int64            `json:"total"`      // 总行数
	Page      int              `json:"page"`       // 当前页
	PageSize  int              `json:"page_size"`  // 每页数量
}

// 表结构-请求
type TableSchemaRequest struct {
	TableName string `json:"table_name" form:"table_name"` // 表名
}

// 表结构-响应
type TableSchemaResponse struct {
	TableName string       `json:"table_name"` // 表名
	Columns   []ColumnInfo `json:"columns"`    // 列信息列表
}

// 列信息
type ColumnInfo struct {
	CID        int    `json:"cid"`         // 列序号
	Name       string `json:"name"`        // 列名
	Type       string `json:"type"`        // 类型
	NotNull    int    `json:"not_null"`    // 是否非空
	DefaultVal any    `json:"default_val"` // 默认值
	PrimaryKey int    `json:"primary_key"` // 是否主键
}

// 清空表-请求
type ClearTableRequest struct {
	TableName string `json:"table_name" form:"table_name"` // 表名
}

// 删除行-请求
type DeleteRowRequest struct {
	TableName  string `json:"table_name" form:"table_name"`    // 表名
	RowID      int64  `json:"row_id" form:"row_id"`            // 行ID
	PrimaryKey string `json:"primary_key" form:"primary_key"`  // 主键列名
}

// 执行SQL-请求
type ExecuteSqlRequest struct {
	Sql       string `json:"sql" form:"sql"`                     // SQL语句
	TableName string `json:"table_name" form:"table_name"`       // 关联表名
}

// 执行SQL-响应
type ExecuteSqlResponse struct {
	IsQuery  bool             `json:"is_query"`            // 是否查询语句
	Columns  []string         `json:"columns,omitempty"`   // 列名(查询时)
	Rows     []map[string]any `json:"rows,omitempty"`      // 数据行(查询时)
	Total    int64            `json:"total,omitempty"`     // 总行数(查询时)
	Affected int64            `json:"affected,omitempty"`  // 影响行数(非查询时)
}

// 更新行-请求
type UpdateRowRequest struct {
	TableName  string         `json:"table_name" form:"table_name"`    // 表名
	RowID      int64          `json:"row_id" form:"row_id"`            // 行ID
	PrimaryKey string         `json:"primary_key" form:"primary_key"`  // 主键列名
	Updates    map[string]any `json:"updates" form:"updates"`          // 要更新的字段和值
}

// 批量删除行-请求
type DeleteRowsRequest struct {
	TableName  string  `json:"table_name" form:"table_name"`    // 表名
	RowIDs     []int64 `json:"row_ids" form:"row_ids"`          // 行ID列表
	PrimaryKey string  `json:"primary_key" form:"primary_key"`  // 主键列名
}

// 批量删除行-响应
type DeleteRowsResponse struct {
	Deleted int64 `json:"deleted"` // 已删除行数
}

// 表数据-请求
type TableDataRequest struct {
	TableName string            `json:"table_name" form:"table_name"`   // 表名
	Page      int               `json:"page" form:"page"`               // 页码
	PageSize  int               `json:"page_size" form:"page_size"`     // 每页数量
	OrderBy   string            `json:"order_by" form:"order_by"`       // 排序列
	OrderDir  string            `json:"order_dir" form:"order_dir"`     // 排序方向(asc/desc)
	Filters   map[string]string `json:"filters" form:"filters"`         // 筛选条件
}
