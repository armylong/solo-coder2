package api_catcher

// 基础请求
type BaseRequest struct {
	Limit  int `json:"limit" form:"limit"`   // 每页数量
	Offset int `json:"offset" form:"offset"` // 偏移量
}

// 按ID查询-请求
type GetRequest struct {
	ID int64 `json:"id" form:"id"` // 记录ID
}

// 上传-请求
type UploadRequest struct {
	FilterList []string               `json:"filter_list"` // 过滤列表
	ApiData    map[string]interface{} `json:"api_data"`    // 抓包数据
}

// 上传-响应
type UploadResponse struct {
	ID int64 `json:"id,omitempty"` // 记录ID
}

// 抓包记录
type ApiCatcherRecord struct {
	ID        int64  `json:"id"`         // 记录ID
	Data      string `json:"data"`       // 抓包数据(JSON)
	CreatedAt int64  `json:"created_at"` // 创建时间
}

// 列表-响应
type ListResponse struct {
	List  []*ApiCatcherRecord `json:"list"`  // 记录列表
	Total int                 `json:"total"` // 总数
}

// 删除-请求
type DeleteRequest struct {
	ID int64 `json:"id" form:"id"` // 记录ID
}

// 下载-请求
type DownloadRequest struct {
	ID    int64  `json:"id" form:"id"`       // 续传起始ID
	Limit int    `json:"limit" form:"limit"` // 每次拉取数量
	Date  string `json:"date" form:"date"`   // 日期(YYYY-MM-DD)
}
