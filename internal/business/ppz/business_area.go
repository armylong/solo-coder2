package ppz

import (
	"context"
	"errors"
	"fmt"

	ppzCs "github.com/armylong/armylong-go/internal/cs/ppz"
	ppzModel "github.com/armylong/armylong-go/internal/model/ppz"
)

type businessAreaBusiness struct{}

var BusinessAreaBusiness = &businessAreaBusiness{}

// 区域列表
func (b *businessAreaBusiness) List(ctx context.Context, req *ppzCs.BusinessAreaListRequest) (*ppzCs.BusinessAreaListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	total, err := ppzModel.TbPpzBusinessAreaModel.Count(req.Status, req.Keyword)
	if err != nil {
		return nil, fmt.Errorf("获取运营区域数量失败: %w", err)
	}

	areas, err := ppzModel.TbPpzBusinessAreaModel.List(req.Status, req.Keyword, req.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("获取运营区域列表失败: %w", err)
	}

	list := make([]*ppzCs.BusinessAreaItem, 0, len(areas))
	for _, area := range areas {
		list = append(list, &ppzCs.BusinessAreaItem{
			AreaId:    area.AreaId,
			AreaName:  area.AreaName,
			AreaFence: area.AreaFence,
			Status:    area.Status,
			CreatedAt: area.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: area.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &ppzCs.BusinessAreaListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 创建区域
func (b *businessAreaBusiness) Create(ctx context.Context, req *ppzCs.BusinessAreaCreateRequest) (*ppzCs.BusinessAreaCreateResponse, error) {
	if req.AreaName == "" {
		return nil, errors.New("区域名称不能为空")
	}
	if req.AreaFence == nil || len(*req.AreaFence) == 0 {
		return nil, errors.New("区域围栏数据不能为空")
	}

	area := &ppzModel.TbPpzBusinessArea{
		AreaName:  req.AreaName,
		AreaFence: req.AreaFence,
		Status:    ppzModel.BusinessAreaStatusNormal,
	}

	areaId, err := ppzModel.TbPpzBusinessAreaModel.Create(area)
	if err != nil {
		return nil, fmt.Errorf("创建运营区域失败: %w", err)
	}

	return &ppzCs.BusinessAreaCreateResponse{
		AreaId: areaId,
	}, nil
}

// 更新区域
func (b *businessAreaBusiness) Update(ctx context.Context, req *ppzCs.BusinessAreaUpdateRequest) (*ppzCs.BusinessAreaUpdateResponse, error) {
	if req.AreaId == 0 {
		return nil, errors.New("区域ID不能为空")
	}
	if req.AreaName == "" {
		return nil, errors.New("区域名称不能为空")
	}
	if req.AreaFence == nil || len(*req.AreaFence) == 0 {
		return nil, errors.New("区域围栏数据不能为空")
	}

	area, err := ppzModel.TbPpzBusinessAreaModel.GetById(req.AreaId)
	if err != nil {
		return nil, errors.New("运营区域不存在")
	}

	if area.Status == ppzModel.BusinessAreaStatusDeleted {
		return nil, errors.New("该运营区域已删除")
	}

	area.AreaName = req.AreaName
	area.AreaFence = req.AreaFence

	err = ppzModel.TbPpzBusinessAreaModel.Update(area)
	if err != nil {
		return nil, fmt.Errorf("更新运营区域失败: %w", err)
	}

	return &ppzCs.BusinessAreaUpdateResponse{}, nil
}

// 区域详情
func (b *businessAreaBusiness) Get(ctx context.Context, req *ppzCs.BusinessAreaGetRequest) (*ppzCs.BusinessAreaGetResponse, error) {
	if req.AreaId == 0 {
		return nil, errors.New("区域ID不能为空")
	}

	area, err := ppzModel.TbPpzBusinessAreaModel.GetById(req.AreaId)
	if err != nil {
		return nil, errors.New("运营区域不存在")
	}

	if area.Status == ppzModel.BusinessAreaStatusDeleted {
		return nil, errors.New("该运营区域已删除")
	}

	return &ppzCs.BusinessAreaGetResponse{
		Area: &ppzCs.BusinessAreaItem{
			AreaId:    area.AreaId,
			AreaName:  area.AreaName,
			AreaFence: area.AreaFence,
			Status:    area.Status,
			CreatedAt: area.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: area.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// 停用区域
func (b *businessAreaBusiness) Disable(ctx context.Context, req *ppzCs.BusinessAreaDisableRequest) (*ppzCs.BusinessAreaDisableResponse, error) {
	if req.AreaId == 0 {
		return nil, errors.New("区域ID不能为空")
	}

	area, err := ppzModel.TbPpzBusinessAreaModel.GetById(req.AreaId)
	if err != nil {
		return nil, errors.New("运营区域不存在")
	}

	if area.Status == ppzModel.BusinessAreaStatusDeleted {
		return nil, errors.New("该运营区域已删除")
	}

	if area.Status == ppzModel.BusinessAreaStatusDisabled {
		return nil, errors.New("该运营区域已停用")
	}

	routeCount, _ := ppzModel.TbPpzBusinessRouteModel.CountByAreaId(req.AreaId)
	if routeCount > 0 {
		return nil, fmt.Errorf("该区域下存在 %d 条运营路线，请先处理相关路线", routeCount)
	}

	err = ppzModel.TbPpzBusinessAreaModel.Disable(req.AreaId)
	if err != nil {
		return nil, fmt.Errorf("停用运营区域失败: %w", err)
	}

	return &ppzCs.BusinessAreaDisableResponse{}, nil
}

// 启用区域
func (b *businessAreaBusiness) Enable(ctx context.Context, req *ppzCs.BusinessAreaEnableRequest) (*ppzCs.BusinessAreaEnableResponse, error) {
	if req.AreaId == 0 {
		return nil, errors.New("区域ID不能为空")
	}

	area, err := ppzModel.TbPpzBusinessAreaModel.GetById(req.AreaId)
	if err != nil {
		return nil, errors.New("运营区域不存在")
	}

	if area.Status == ppzModel.BusinessAreaStatusDeleted {
		return nil, errors.New("该运营区域已删除")
	}

	if area.Status == ppzModel.BusinessAreaStatusNormal {
		return nil, errors.New("该运营区域已启用")
	}

	err = ppzModel.TbPpzBusinessAreaModel.Enable(req.AreaId)
	if err != nil {
		return nil, fmt.Errorf("启用运营区域失败: %w", err)
	}

	return &ppzCs.BusinessAreaEnableResponse{}, nil
}

// 删除区域
func (b *businessAreaBusiness) Delete(ctx context.Context, req *ppzCs.BusinessAreaDeleteRequest) (*ppzCs.BusinessAreaDeleteResponse, error) {
	if req.AreaId == 0 {
		return nil, errors.New("区域ID不能为空")
	}

	area, err := ppzModel.TbPpzBusinessAreaModel.GetById(req.AreaId)
	if err != nil {
		return nil, errors.New("运营区域不存在")
	}

	if area.Status == ppzModel.BusinessAreaStatusDeleted {
		return nil, errors.New("该运营区域已删除")
	}

	routeCount, _ := ppzModel.TbPpzBusinessRouteModel.CountByAreaId(req.AreaId)
	if routeCount > 0 {
		return nil, fmt.Errorf("该区域下存在 %d 条运营路线，请先处理相关路线", routeCount)
	}

	err = ppzModel.TbPpzBusinessAreaModel.Delete(req.AreaId)
	if err != nil {
		return nil, fmt.Errorf("删除运营区域失败: %w", err)
	}

	return &ppzCs.BusinessAreaDeleteResponse{}, nil
}

// 获取启用的区域列表
func (b *businessAreaBusiness) ListActive(ctx context.Context) (*ppzCs.BusinessAreaListActiveResponse, error) {
	areas, err := ppzModel.TbPpzBusinessAreaModel.ListActive()
	if err != nil {
		return nil, fmt.Errorf("获取运营区域列表失败: %w", err)
	}

	list := make([]*ppzCs.BusinessAreaActiveItem, 0, len(areas))
	for _, area := range areas {
		list = append(list, &ppzCs.BusinessAreaActiveItem{
			AreaId:   area.AreaId,
			AreaName: area.AreaName,
		})
	}

	return &ppzCs.BusinessAreaListActiveResponse{
		List: list,
	}, nil
}
