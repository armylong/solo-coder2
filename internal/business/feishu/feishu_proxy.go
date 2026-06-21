package feishu

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	feishuCs "github.com/armylong/armylong-go/internal/cs/feishu"
	libraryFeishu "github.com/armylong/go-library/service/feishu"
	"github.com/armylong/go-library/service/httpx"
)

type feishuProxyBusiness struct{}

var FeishuProxyBusiness = &feishuProxyBusiness{}

// 飞书开放平台基础URL
const feishuBaseUrl = "https://open.feishu.cn"

// 飞书通用代理
// 前端传api_path+http_method+参数，后端直接发起HTTP请求透传到飞书开放平台
// 不依赖飞书SDK，纯HTTP透传
func (b *feishuProxyBusiness) ProxyFeishu(ctx context.Context, req *feishuCs.FeishuProxyRequest) (*feishuCs.FeishuProxyResponse, error) {

	// 构造完整URL
	fullUrl := feishuBaseUrl + req.ApiPath
	if len(req.Query) > 0 {
		params := url.Values{}
		for k, v := range req.Query {
			params.Set(k, v)
		}
		fullUrl = fullUrl + "?" + params.Encode()
	}

	// 构造请求头
	headers := map[string]string{
		"Authorization": libraryFeishu.GetUserAccessTokenHeader(nil),
		"Content-Type":  "application/json; charset=utf-8",
	}
	// 合并自定义header
	for k, v := range req.Header {
		headers[k] = v
	}

	// POST/PUT时，如果body为空，默认补{}，避免飞书报参数错误
	method := strings.ToUpper(req.HttpMethod)
	body := req.Body
	if (method == "POST" || method == "PUT") && len(body) == 0 {
		body = []byte("{}")
	}

	// 根据http_method调不同的httpx方法
	var data []byte
	var err error
	switch method {
	case "GET":
		data, err = httpx.GetWithHeader(fullUrl, headers)
	case "POST":
		data, err = httpx.PostWithHeader(fullUrl, body, headers)
	case "PUT":
		data, err = httpx.PutWithHeader(fullUrl, body, headers)
	case "DELETE":
		data, err = httpx.DeleteWithHeader(fullUrl, headers)
	default:
		return nil, fmt.Errorf("不支持的HTTP方法: %s", req.HttpMethod)
	}

	if err != nil {
		return nil, fmt.Errorf("飞书API请求失败: %w", err)
	}

	// 透传原始响应
	raw := feishuCs.FeishuProxyResponse(data)
	return &raw, nil
}
