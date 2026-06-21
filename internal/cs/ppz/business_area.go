package ppz

import (
	ppzModel "github.com/armylong/armylong-go/internal/model/ppz"
)

// 区域列表-请求
type BusinessAreaListRequest struct {
	Page     int    `json:"page" form:"page"`           // 页码，从1开始
	PageSize int    `json:"page_size" form:"page_size"` // 每页条数
	Status   int    `json:"status" form:"status"`       // 状态筛选
	Keyword  string `json:"keyword" form:"keyword"`     // 关键词搜索
}

// 区域列表项
type BusinessAreaItem struct {
	AreaId    int64               `json:"area_id"`    // 区域ID
	AreaName  string              `json:"area_name"`  // 区域名称
	AreaFence *ppzModel.AreaFence `json:"area_fence"` // 围栏数据
	Status    int                 `json:"status"`     // 状态
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
}

// 区域列表-响应
type BusinessAreaListResponse struct {
	List     []*BusinessAreaItem `json:"list"`      // 区域列表
	Total    int64               `json:"total"`     // 总数
	Page     int                 `json:"page"`      // 当前页码
	PageSize int                 `json:"page_size"` // 每页条数
}

// 创建区域-请求
type BusinessAreaCreateRequest struct {
	AreaName  string              `json:"area_name" form:"area_name"`   // 区域名称
	AreaFence *ppzModel.AreaFence `json:"area_fence" form:"area_fence"` // 围栏数据
}

// 创建区域-响应
type BusinessAreaCreateResponse struct {
	AreaId int64 `json:"area_id"` // 区域ID
}

// 更新区域-请求
type BusinessAreaUpdateRequest struct {
	AreaId    int64               `json:"area_id" form:"area_id"`       // 区域ID
	AreaName  string              `json:"area_name" form:"area_name"`   // 区域名称
	AreaFence *ppzModel.AreaFence `json:"area_fence" form:"area_fence"` // 围栏数据
}

// 更新区域-响应
type BusinessAreaUpdateResponse struct{}

// 区域详情-请求
type BusinessAreaGetRequest struct {
	AreaId int64 `json:"area_id" form:"area_id"` // 区域ID
}

// 区域详情-响应
type BusinessAreaGetResponse struct {
	Area *BusinessAreaItem `json:"area"` // 区域详情
}

// 停用区域-请求
type BusinessAreaDisableRequest struct {
	AreaId int64 `json:"area_id" form:"area_id"` // 区域ID
}

// 停用区域-响应
type BusinessAreaDisableResponse struct{}

// 启用区域-请求
type BusinessAreaEnableRequest struct {
	AreaId int64 `json:"area_id" form:"area_id"` // 区域ID
}

// 启用区域-响应
type BusinessAreaEnableResponse struct{}

// 删除区域-请求
type BusinessAreaDeleteRequest struct {
	AreaId int64 `json:"area_id" form:"area_id"` // 区域ID
}

// 删除区域-响应
type BusinessAreaDeleteResponse struct{}
