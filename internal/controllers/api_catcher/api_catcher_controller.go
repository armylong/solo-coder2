package api_catcher

import (
	"errors"

	apiCatcherBusiness "github.com/armylong/armylong-go/internal/business/api_catcher"
	apiCatcherCs "github.com/armylong/armylong-go/internal/cs/api_catcher"
	"github.com/gin-gonic/gin"
)

// ApiCatcherController API抓包
type ApiCatcherController struct {
}

// 上传抓包数据
func (c *ApiCatcherController) ActionUpload(ctx *gin.Context, req *apiCatcherCs.UploadRequest) (*apiCatcherCs.UploadResponse, error) {
	id, err := apiCatcherBusiness.ApiCatcherBusiness.Upload(ctx, req)
	if err != nil {
		return nil, err
	}

	return &apiCatcherCs.UploadResponse{
		ID: id,
	}, nil
}

// 列表
func (c *ApiCatcherController) ActionList(ctx *gin.Context, req *apiCatcherCs.BaseRequest) (*apiCatcherCs.ListResponse, error) {
	list, total, err := apiCatcherBusiness.ApiCatcherBusiness.List(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	return &apiCatcherCs.ListResponse{
		List:  list,
		Total: total,
	}, nil
}

// 查详情
func (c *ApiCatcherController) ActionGet(ctx *gin.Context, req *apiCatcherCs.GetRequest) (*apiCatcherCs.ApiCatcherRecord, error) {
	record, err := apiCatcherBusiness.ApiCatcherBusiness.GetById(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// 删除
func (c *ApiCatcherController) ActionDelete(ctx *gin.Context, req *apiCatcherCs.DeleteRequest) error {
	return apiCatcherBusiness.ApiCatcherBusiness.Delete(ctx, req.ID)
}

// 按日期下载
func (c *ApiCatcherController) ActionDownload(ctx *gin.Context, req *apiCatcherCs.DownloadRequest) (*apiCatcherCs.ListResponse, error) {
	if req.Date == "" {
		return nil, errors.New("date参数不能为空")
	}

	list, err := apiCatcherBusiness.ApiCatcherBusiness.Download(ctx, req.ID, req.Limit, req.Date)
	if err != nil {
		return nil, err
	}

	return &apiCatcherCs.ListResponse{
		List:  list,
		Total: len(list),
	}, nil
}
