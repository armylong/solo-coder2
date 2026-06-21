package api_catcher

import (
	"context"
	"encoding/json"

	apiCatcherCs "github.com/armylong/armylong-go/internal/cs/api_catcher"
	apiCatcherModel "github.com/armylong/armylong-go/internal/model/api_catcher"
)

type apiCatcherBusiness struct{}

var ApiCatcherBusiness = &apiCatcherBusiness{}

// 上传抓包数据
func (b *apiCatcherBusiness) Upload(ctx context.Context, req *apiCatcherCs.UploadRequest) (int64, error) {
	dataMap := map[string]interface{}{
		"filter_list": req.FilterList,
		"api_data":    req.ApiData,
	}

	jsonBytes, err := json.Marshal(dataMap)
	if err != nil {
		return 0, err
	}

	data := &apiCatcherModel.TbApiCatcher{
		Data: string(jsonBytes),
	}

	id, err := apiCatcherModel.TbApiCatcherModel.Create(data)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// 分页列表
func (b *apiCatcherBusiness) List(ctx context.Context, limit, offset int) ([]*apiCatcherCs.ApiCatcherRecord, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	list, err := apiCatcherModel.TbApiCatcherModel.List(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, _ := apiCatcherModel.TbApiCatcherModel.Count()

	result := make([]*apiCatcherCs.ApiCatcherRecord, 0, len(list))
	for _, item := range list {
		record := &apiCatcherCs.ApiCatcherRecord{
			ID:        item.ID,
			Data:      item.Data,
			CreatedAt: item.CreatedAt.Unix(),
		}
		result = append(result, record)
	}

	return result, int(total), nil
}

// 按ID查详情
func (b *apiCatcherBusiness) GetById(ctx context.Context, id int64) (*apiCatcherCs.ApiCatcherRecord, error) {
	item, err := apiCatcherModel.TbApiCatcherModel.GetById(id)
	if err != nil {
		return nil, err
	}

	return &apiCatcherCs.ApiCatcherRecord{
		ID:        item.ID,
		Data:      item.Data,
		CreatedAt: item.CreatedAt.Unix(),
	}, nil
}

// 删除
func (b *apiCatcherBusiness) Delete(ctx context.Context, id int64) error {
	return apiCatcherModel.TbApiCatcherModel.Delete(id)
}

// 按日期下载，支持从指定ID续传
func (b *apiCatcherBusiness) Download(ctx context.Context, id int64, limit int, date string) ([]*apiCatcherCs.ApiCatcherRecord, error) {
	list, err := apiCatcherModel.TbApiCatcherModel.Download(id, limit, date)
	if err != nil {
		return nil, err
	}

	result := make([]*apiCatcherCs.ApiCatcherRecord, 0, len(list))
	for _, item := range list {
		record := &apiCatcherCs.ApiCatcherRecord{
			ID:        item.ID,
			Data:      item.Data,
			CreatedAt: item.CreatedAt.Unix(),
		}
		result = append(result, record)
	}

	return result, nil
}
