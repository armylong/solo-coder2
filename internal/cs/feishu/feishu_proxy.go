package feishu

import "encoding/json"

// 飞书代理-请求
// 前端传完整api_path，后端直接透传HTTP请求到飞书开放平台
type FeishuProxyRequest struct {
	ApiPath    string            `json:"api_path" form:"api_path"`       // 完整API路径，如 /open-apis/bitable/v1/apps/xxx/tables/yyy/fields
	HttpMethod string            `json:"http_method" form:"http_method"` // HTTP方法(GET/POST/PUT/DELETE)
	Header     map[string]string `json:"header" form:"header"`           // 请求头(后端自动加Authorization)
	Query      map[string]string `json:"query" form:"query"`             // URL查询参数
	Body       json.RawMessage   `json:"body" form:"body"`               // 请求体(任意JSON，POST/PUT时用)
}

// 飞书通用代理-响应（原始JSON字节）
type FeishuProxyResponse = json.RawMessage
