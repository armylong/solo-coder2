package gaode

import (
	"context"
	"fmt"

	gaodeCs "github.com/armylong/armylong-go/internal/cs/gaode"
)

// 高德API路径映射
var gaodeApiPaths = map[string]string{
	"regeo":          "/v3/geocode/regeo",
	"searchPoi":      "/v3/place/text",
	"searchDistrict": "/v3/config/district",
}

// 按API名代理转发
func (b *GaodeBusiness) ProxyGaode(ctx context.Context, req *gaodeCs.GaodeProxyRequest) (*gaodeCs.GaodeProxyResponse, error) {
	apiPath, ok := gaodeApiPaths[req.Api]
	if !ok {
		return nil, fmt.Errorf("不支持的高德API: %s", req.Api)
	}
	return b.Proxy(ctx, apiPath, req)
}
