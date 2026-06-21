package feishu

// 表格概览信息
// 用于飞书配置总览页展示已接入的表格列表
type BitableInfo struct {
	Id       int64  `json:"id"`
	AppToken string `json:"app_token"`
	Alias    string `json:"alias"`
}

// 表格配置列表-请求
type BitableConfigListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Status   int    `json:"status" form:"status"`
	Keyword  string `json:"keyword" form:"keyword"`
}

// 表格配置列表项
type BitableConfigItem struct {
	Id         int64  `json:"id"`
	AppToken   string `json:"app_token"`
	Alias      string `json:"alias"`
	Status     int    `json:"status"`
	CreatedUid int64  `json:"created_uid"`
	UpdatedUid int64  `json:"updated_uid"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// 表格配置列表-响应
type BitableConfigListResponse struct {
	List     []*BitableConfigItem `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// 创建表格配置-请求
type BitableConfigCreateRequest struct {
	AppToken string `json:"app_token" form:"app_token"`
	Alias    string `json:"alias" form:"alias"`
}

// 创建表格配置-响应
type BitableConfigCreateResponse struct {
	Id int64 `json:"id"`
}

// 更新表格配置-请求
type BitableConfigUpdateRequest struct {
	Id       int64  `json:"id" form:"id"`
	AppToken string `json:"app_token" form:"app_token"`
	Alias    string `json:"alias" form:"alias"`
	Status   int    `json:"status" form:"status"`
}

// 更新表格配置-响应
type BitableConfigUpdateResponse struct{}

// 获取表格配置详情-请求
type BitableConfigGetRequest struct {
	Id int64 `json:"id" form:"id"`
}

// 获取表格配置详情-响应
type BitableConfigGetResponse struct {
	Bitable *BitableConfigItem `json:"bitable"`
}

// 删除表格配置-请求
type BitableConfigDeleteRequest struct {
	Id int64 `json:"id" form:"id"`
}

// 删除表格配置-响应
type BitableConfigDeleteResponse struct{}
