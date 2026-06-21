package ppz

// 路线列表-请求
type BusinessRouteListRequest struct {
	Page     int    `json:"page" form:"page"`           // 页码，从1开始
	PageSize int    `json:"page_size" form:"page_size"` // 每页条数
	Status   int    `json:"status" form:"status"`       // 状态筛选
	Keyword  string `json:"keyword" form:"keyword"`     // 关键词搜索
}

// 路线列表项
type BusinessRouteItem struct {
	RouteId   int64  `json:"route_id"`   // 路线ID
	RouteName string `json:"route_name"` // 路线名称
	AAreaId   int64  `json:"a_area_id"`  // A端区域ID
	AAreaName string `json:"a_area_name"` // A端区域名称
	BAreaId   int64  `json:"b_area_id"`  // B端区域ID
	BAreaName string `json:"b_area_name"` // B端区域名称
	Status    int    `json:"status"`     // 状态
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// 路线列表-响应
type BusinessRouteListResponse struct {
	List     []*BusinessRouteItem `json:"list"`      // 路线列表
	Total    int64                `json:"total"`     // 总数
	Page     int                  `json:"page"`      // 当前页码
	PageSize int                  `json:"page_size"` // 每页条数
}

// 创建路线-请求
type BusinessRouteCreateRequest struct {
	RouteName string `json:"route_name" form:"route_name"` // 路线名称
	AAreaId   int64  `json:"a_area_id" form:"a_area_id"`   // A端区域ID
	BAreaId   int64  `json:"b_area_id" form:"b_area_id"`   // B端区域ID
	Status    int    `json:"status" form:"status"`         // 状态
}

// 创建路线-响应
type BusinessRouteCreateResponse struct {
	RouteId int64 `json:"route_id"` // 路线ID
}

// 更新路线-请求
type BusinessRouteUpdateRequest struct {
	RouteId   int64  `json:"route_id" form:"route_id"`     // 路线ID
	RouteName string `json:"route_name" form:"route_name"` // 路线名称
	AAreaId   int64  `json:"a_area_id" form:"a_area_id"`   // A端区域ID
	BAreaId   int64  `json:"b_area_id" form:"b_area_id"`   // B端区域ID
	Status    int    `json:"status" form:"status"`         // 状态
}

// 更新路线-响应
type BusinessRouteUpdateResponse struct{}

// 路线详情-请求
type BusinessRouteGetRequest struct {
	RouteId int64 `json:"route_id" form:"route_id"` // 路线ID
}

// 路线详情-响应
type BusinessRouteGetResponse struct {
	Route *BusinessRouteItem `json:"route"` // 路线详情
}

// 停用路线-请求
type BusinessRouteDisableRequest struct {
	RouteId int64 `json:"route_id" form:"route_id"` // 路线ID
}

// 停用路线-响应
type BusinessRouteDisableResponse struct{}

// 启用路线-请求
type BusinessRouteEnableRequest struct {
	RouteId int64 `json:"route_id" form:"route_id"` // 路线ID
}

// 启用路线-响应
type BusinessRouteEnableResponse struct{}

// 删除路线-请求
type BusinessRouteDeleteRequest struct {
	RouteId int64 `json:"route_id" form:"route_id"` // 路线ID
}

// 删除路线-响应
type BusinessRouteDeleteResponse struct{}

// 启用的区域简要信息
type BusinessAreaActiveItem struct {
	AreaId   int64  `json:"area_id"`   // 区域ID
	AreaName string `json:"area_name"` // 区域名称
}

// 获取启用区域列表-请求
type BusinessAreaListActiveRequest struct{}

// 获取启用区域列表-响应
type BusinessAreaListActiveResponse struct {
	List []*BusinessAreaActiveItem `json:"list"` // 区域列表
}
