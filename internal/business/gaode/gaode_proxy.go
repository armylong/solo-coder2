package gaode

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	gaodeCs "github.com/armylong/armylong-go/internal/cs/gaode"
	confLibrary "github.com/armylong/go-library/service/conf"
)

type GaodeBusiness struct{}

var Business = &GaodeBusiness{}

// 代理转发高德API请求
func (b *GaodeBusiness) Proxy(ctx context.Context, apiPath string, req *gaodeCs.GaodeProxyRequest) (*gaodeCs.GaodeProxyResponse, error) {
	key := confLibrary.GetString("gaode-map-ppz-server.key")
	if key == "" {
		return nil, fmt.Errorf("高德地图Key未配置")
	}

	params := url.Values{}
	params.Set("key", key)

	for k, v := range req.Query {
		if k == "key" || v == "" {
			continue
		}
		params.Set(k, v)
	}
	for k, v := range req.Body {
		if k == "key" || v == "" {
			continue
		}
		params.Set(k, v)
	}

	apiURL := fmt.Sprintf("https://restapi.amap.com%s?%s", apiPath, params.Encode())

	var httpReq *http.Request
	var err error

	if len(req.Body) > 0 {
		form := url.Values{}
		for k, v := range req.Body {
			if v != "" {
				form.Set(k, v)
			}
		}
		httpReq, err = http.NewRequest("POST", apiURL, bytes.NewBufferString(form.Encode()))
		if err != nil {
			return nil, fmt.Errorf("创建高德API请求失败: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		httpReq, err = http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建高德API请求失败: %w", err)
		}
	}

	for k, v := range req.Header {
		if v != "" {
			httpReq.Header.Set(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求高德API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取高德API响应失败: %w", err)
	}

	return (*gaodeCs.GaodeProxyResponse)(&body), nil
}
