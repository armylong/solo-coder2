package gaode

import "encoding/json"

// 获取高德地图Key-请求
type GetGaodeMapKeyRequest struct{}

// 获取高德地图Key-响应
type GetGaodeMapKeyResponse struct {
	Key string `json:"key"` // 高德地图JS Key
}

// 高德代理-请求
type GaodeProxyRequest struct {
	Api    string            `json:"api" form:"api"`       // API名称(regeo/searchPoi/searchDistrict)
	Header map[string]string `json:"header" form:"header"` // 自定义请求头
	Query  map[string]string `json:"query" form:"query"`   // URL参数
	Body   map[string]string `json:"body" form:"body"`     // POST参数
}

// 高德代理-响应（原始JSON）
type GaodeProxyResponse = json.RawMessage
