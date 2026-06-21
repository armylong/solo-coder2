package ppz

import "encoding/json"

// 上传地址-请求
type UploadAddressRequest struct {
	AddressId int64           `json:"address_id" form:"address_id"` // 地址ID（编辑时传）
	Remark    string          `json:"remark" form:"remark"`         // 地址备注
	GaodeData json.RawMessage `json:"gaode_data" form:"gaode_data"` // 高德地址数据
}

// 上传地址-响应
type UploadAddressResponse struct {
	AddressId int64 `json:"address_id"` // 地址ID
}

// 地址列表-请求
type AddressListRequest struct{}

// 地址列表项
type AddressListItem struct {
	AddressId int64           `json:"address_id"` // 地址ID
	Remark    string          `json:"remark"`     // 地址备注
	GaodeData json.RawMessage `json:"gaode_data"` // 高德地址数据
	Sort      int64           `json:"sort"`       // 排序值
}

// 地址列表-响应
type AddressListResponse struct {
	List []*AddressListItem `json:"list"` // 地址列表
}

// 更新地址排序-请求
type UpdateAddressSortRequest struct {
	AddressId int64 `json:"address_id" form:"address_id"` // 地址ID
	NewSort   int64 `json:"new_sort" form:"new_sort"`     // 新排序值
}

// 更新地址排序-响应
type UpdateAddressSortResponse struct{}

// 删除地址-请求
type DeleteAddressRequest struct {
	AddressId int64 `json:"address_id" form:"address_id"` // 地址ID
}

// 删除地址-响应
type DeleteAddressResponse struct{}

// 地址详情-请求
type GetAddressDetailRequest struct {
	AddressId int64 `json:"address_id" form:"address_id"` // 地址ID
}

// 地址详情-响应
type GetAddressDetailResponse struct {
	AddressId int64           `json:"address_id"` // 地址ID
	Remark    string          `json:"remark"`     // 地址备注
	GaodeData json.RawMessage `json:"gaode_data"` // 高德地址数据
	Sort      int64           `json:"sort"`       // 排序值
}
