package api_catcher

import (
	"context"
	"fmt"
	"testing"

	apiCatcherCs "github.com/armylong/armylong-go/internal/cs/api_catcher"
)

func TestUpload(t *testing.T) {
	testData := []struct {
		filterList []string
		apiData    map[string]interface{}
	}{
		{
			filterList: []string{"api/user", "api/order"},
			apiData: map[string]interface{}{
				"id":            "1712345678901_abc123",
				"url":           "https://example.com/api/user/info",
				"method":        "GET",
				"headers":       map[string]interface{}{"Content-Type": "application/json"},
				"params":        map[string]interface{}{"page": 1, "size": 10},
				"request_body":  nil,
				"response_body": map[string]interface{}{"code": 0, "data": map[string]interface{}{"name": "test"}},
				"status":        200,
				"capture_time":  1712345678901,
				"duration":      120,
			},
		},
		{
			filterList: []string{"api/product"},
			apiData: map[string]interface{}{
				"id":            "1712345678902_def456",
				"url":           "https://example.com/api/product/list",
				"method":        "POST",
				"headers":       map[string]interface{}{"Content-Type": "application/json", "Authorization": "Bearer token123"},
				"params":        nil,
				"request_body":  map[string]interface{}{"category": "electronics"},
				"response_body": map[string]interface{}{"code": 0, "data": []interface{}{map[string]interface{}{"id": 1, "name": "Product 1"}}},
				"status":        200,
				"capture_time":  1712345678902,
				"duration":      250,
			},
		},
		{
			filterList: []string{"api/order", "api/payment"},
			apiData: map[string]interface{}{
				"id":            "1712345678903_ghi789",
				"url":           "https://example.com/api/order/create",
				"method":        "POST",
				"headers":       map[string]interface{}{"Content-Type": "application/json"},
				"params":        nil,
				"request_body":  map[string]interface{}{"product_id": 123, "quantity": 2},
				"response_body": map[string]interface{}{"code": 0, "data": map[string]interface{}{"order_id": "ORD123456"}},
				"status":        201,
				"capture_time":  1712345678903,
				"duration":      380,
			},
		},
		{
			filterList: []string{"api/user"},
			apiData: map[string]interface{}{
				"id":            "1712345678904_jkl012",
				"url":           "https://example.com/api/user/profile",
				"method":        "PUT",
				"headers":       map[string]interface{}{"Content-Type": "application/json"},
				"params":        nil,
				"request_body":  map[string]interface{}{"nickname": "新昵称"},
				"response_body": map[string]interface{}{"code": 0, "message": "success"},
				"status":        200,
				"capture_time":  1712345678904,
				"duration":      95,
			},
		},
		{
			filterList: []string{"api/admin"},
			apiData: map[string]interface{}{
				"id":            "1712345678905_mno345",
				"url":           "https://example.com/api/admin/users",
				"method":        "DELETE",
				"headers":       map[string]interface{}{"Authorization": "Bearer admin_token"},
				"params":        map[string]interface{}{"user_id": 999},
				"request_body":  nil,
				"response_body": map[string]interface{}{"code": 0, "message": "deleted"},
				"status":        204,
				"capture_time":  1712345678905,
				"duration":      60,
			},
		},
	}

	for i, td := range testData {
		req := &apiCatcherCs.UploadRequest{
			FilterList: td.filterList,
			ApiData:    td.apiData,
		}

		id, err := ApiCatcherBusiness.Upload(context.Background(), req)
		if err != nil {
			t.Fatalf("上传数据失败 [%d]: %v", i, err)
		}

		fmt.Printf("插入成功 [%d]: id=%d, url=%s\n", i+1, id, td.apiData["url"])
	}

	fmt.Printf("\n成功插入 %d 条测试数据\n", len(testData))
}

func TestList(t *testing.T) {
	list, total, err := ApiCatcherBusiness.List(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("查询列表失败: %v", err)
	}

	fmt.Printf("总数: %d\n", total)
	for i, item := range list {
		fmt.Printf("[%d] ID: %d, CreatedAt: %d\n", i+1, item.ID, item.CreatedAt)
	}
}
