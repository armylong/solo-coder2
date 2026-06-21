package ppz

import (
	"context"
	"errors"
	"fmt"

	ppzCs "github.com/armylong/armylong-go/internal/cs/ppz"
	ppzModel "github.com/armylong/armylong-go/internal/model/ppz"
)

type businessRouteBusiness struct{}

var BusinessRouteBusiness = &businessRouteBusiness{}

// 路线列表
func (b *businessRouteBusiness) List(ctx context.Context, req *ppzCs.BusinessRouteListRequest) (*ppzCs.BusinessRouteListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	total, err := ppzModel.TbPpzBusinessRouteModel.Count(req.Status, req.Keyword)
	if err != nil {
		return nil, fmt.Errorf("获取运营路线数量失败: %w", err)
	}

	routes, err := ppzModel.TbPpzBusinessRouteModel.List(req.Status, req.Keyword, req.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("获取运营路线列表失败: %w", err)
	}

	areaNameMap, err := b.buildAreaNameMap(routes)
	if err != nil {
		return nil, fmt.Errorf("获取区域名称失败: %w", err)
	}

	list := make([]*ppzCs.BusinessRouteItem, 0, len(routes))
	for _, route := range routes {
		item := &ppzCs.BusinessRouteItem{
			RouteId:   route.RouteId,
			RouteName: route.RouteName,
			AAreaId:   route.AAreaId,
			AAreaName: areaNameMap[route.AAreaId],
			BAreaId:   route.BAreaId,
			BAreaName: areaNameMap[route.BAreaId],
			Status:    route.Status,
			CreatedAt: route.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: route.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if item.AAreaName == "" {
			item.AAreaName = "未知区域"
		}
		if item.BAreaName == "" {
			item.BAreaName = "未知区域"
		}
		list = append(list, item)
	}

	return &ppzCs.BusinessRouteListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 批量获取区域名称
func (b *businessRouteBusiness) buildAreaNameMap(routes []*ppzModel.TbPpzBusinessRoute) (map[int64]string, error) {
	areaIdSet := make(map[int64]bool)
	for _, route := range routes {
		if route.AAreaId > 0 {
			areaIdSet[route.AAreaId] = true
		}
		if route.BAreaId > 0 {
			areaIdSet[route.BAreaId] = true
		}
	}

	areaNameMap := make(map[int64]string)
	for areaId := range areaIdSet {
		area, err := ppzModel.TbPpzBusinessAreaModel.GetById(areaId)
		if err != nil {
			continue
		}
		areaNameMap[areaId] = area.AreaName
	}
	return areaNameMap, nil
}

// 创建路线
func (b *businessRouteBusiness) Create(ctx context.Context, req *ppzCs.BusinessRouteCreateRequest) (*ppzCs.BusinessRouteCreateResponse, error) {
	if req.RouteName == "" {
		return nil, errors.New("路线名称不能为空")
	}
	if req.AAreaId == 0 {
		return nil, errors.New("请选择A点区域")
	}
	if req.BAreaId == 0 {
		return nil, errors.New("请选择B点区域")
	}
	if req.AAreaId == req.BAreaId {
		return nil, errors.New("A点和B点不能相同")
	}

	if _, err := ppzModel.TbPpzBusinessAreaModel.GetById(req.AAreaId); err != nil {
		return nil, errors.New("A点区域不存在")
	}
	if _, err := ppzModel.TbPpzBusinessAreaModel.GetById(req.BAreaId); err != nil {
		return nil, errors.New("B点区域不存在")
	}

	status := req.Status
	if status == 0 {
		status = ppzModel.BusinessRouteStatusNormal
	}

	route := &ppzModel.TbPpzBusinessRoute{
		RouteName: req.RouteName,
		AAreaId:   req.AAreaId,
		BAreaId:   req.BAreaId,
		Status:    status,
	}

	routeId, err := ppzModel.TbPpzBusinessRouteModel.Create(route)
	if err != nil {
		return nil, fmt.Errorf("创建运营路线失败: %w", err)
	}

	return &ppzCs.BusinessRouteCreateResponse{
		RouteId: routeId,
	}, nil
}

// 更新路线
func (b *businessRouteBusiness) Update(ctx context.Context, req *ppzCs.BusinessRouteUpdateRequest) (*ppzCs.BusinessRouteUpdateResponse, error) {
	if req.RouteId == 0 {
		return nil, errors.New("路线ID不能为空")
	}
	if req.RouteName == "" {
		return nil, errors.New("路线名称不能为空")
	}
	if req.AAreaId == 0 {
		return nil, errors.New("请选择A点区域")
	}
	if req.BAreaId == 0 {
		return nil, errors.New("请选择B点区域")
	}
	if req.AAreaId == req.BAreaId {
		return nil, errors.New("A点和B点不能相同")
	}

	route, err := ppzModel.TbPpzBusinessRouteModel.GetById(req.RouteId)
	if err != nil {
		return nil, errors.New("运营路线不存在")
	}

	if route.Status == ppzModel.BusinessRouteStatusDeleted {
		return nil, errors.New("该运营路线已删除")
	}

	if _, err := ppzModel.TbPpzBusinessAreaModel.GetById(req.AAreaId); err != nil {
		return nil, errors.New("A点区域不存在")
	}
	if _, err := ppzModel.TbPpzBusinessAreaModel.GetById(req.BAreaId); err != nil {
		return nil, errors.New("B点区域不存在")
	}

	route.RouteName = req.RouteName
	route.AAreaId = req.AAreaId
	route.BAreaId = req.BAreaId
	route.Status = req.Status

	err = ppzModel.TbPpzBusinessRouteModel.Update(route)
	if err != nil {
		return nil, fmt.Errorf("更新运营路线失败: %w", err)
	}

	return &ppzCs.BusinessRouteUpdateResponse{}, nil
}

// 路线详情
func (b *businessRouteBusiness) Get(ctx context.Context, req *ppzCs.BusinessRouteGetRequest) (*ppzCs.BusinessRouteGetResponse, error) {
	if req.RouteId == 0 {
		return nil, errors.New("路线ID不能为空")
	}

	route, err := ppzModel.TbPpzBusinessRouteModel.GetById(req.RouteId)
	if err != nil {
		return nil, errors.New("运营路线不存在")
	}

	if route.Status == ppzModel.BusinessRouteStatusDeleted {
		return nil, errors.New("该运营路线已删除")
	}

	areaNameMap, err := b.buildAreaNameMap([]*ppzModel.TbPpzBusinessRoute{route})
	if err != nil {
		return nil, fmt.Errorf("获取区域名称失败: %w", err)
	}

	item := &ppzCs.BusinessRouteItem{
		RouteId:   route.RouteId,
		RouteName: route.RouteName,
		AAreaId:   route.AAreaId,
		AAreaName: areaNameMap[route.AAreaId],
		BAreaId:   route.BAreaId,
		BAreaName: areaNameMap[route.BAreaId],
		Status:    route.Status,
		CreatedAt: route.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: route.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if item.AAreaName == "" {
		item.AAreaName = "未知区域"
	}
	if item.BAreaName == "" {
		item.BAreaName = "未知区域"
	}

	return &ppzCs.BusinessRouteGetResponse{
		Route: item,
	}, nil
}

// 停用路线
func (b *businessRouteBusiness) Disable(ctx context.Context, req *ppzCs.BusinessRouteDisableRequest) (*ppzCs.BusinessRouteDisableResponse, error) {
	if req.RouteId == 0 {
		return nil, errors.New("路线ID不能为空")
	}

	route, err := ppzModel.TbPpzBusinessRouteModel.GetById(req.RouteId)
	if err != nil {
		return nil, errors.New("运营路线不存在")
	}

	if route.Status == ppzModel.BusinessRouteStatusDeleted {
		return nil, errors.New("该运营路线已删除")
	}

	if route.Status == ppzModel.BusinessRouteStatusDisabled {
		return nil, errors.New("该运营路线已停用")
	}

	err = ppzModel.TbPpzBusinessRouteModel.Disable(req.RouteId)
	if err != nil {
		return nil, fmt.Errorf("停用运营路线失败: %w", err)
	}

	return &ppzCs.BusinessRouteDisableResponse{}, nil
}

// 启用路线
func (b *businessRouteBusiness) Enable(ctx context.Context, req *ppzCs.BusinessRouteEnableRequest) (*ppzCs.BusinessRouteEnableResponse, error) {
	if req.RouteId == 0 {
		return nil, errors.New("路线ID不能为空")
	}

	route, err := ppzModel.TbPpzBusinessRouteModel.GetById(req.RouteId)
	if err != nil {
		return nil, errors.New("运营路线不存在")
	}

	if route.Status == ppzModel.BusinessRouteStatusDeleted {
		return nil, errors.New("该运营路线已删除")
	}

	if route.Status == ppzModel.BusinessRouteStatusNormal {
		return nil, errors.New("该运营路线已启用")
	}

	err = ppzModel.TbPpzBusinessRouteModel.Enable(req.RouteId)
	if err != nil {
		return nil, fmt.Errorf("启用运营路线失败: %w", err)
	}

	return &ppzCs.BusinessRouteEnableResponse{}, nil
}

// 删除路线
func (b *businessRouteBusiness) Delete(ctx context.Context, req *ppzCs.BusinessRouteDeleteRequest) (*ppzCs.BusinessRouteDeleteResponse, error) {
	if req.RouteId == 0 {
		return nil, errors.New("路线ID不能为空")
	}

	route, err := ppzModel.TbPpzBusinessRouteModel.GetById(req.RouteId)
	if err != nil {
		return nil, errors.New("运营路线不存在")
	}

	if route.Status == ppzModel.BusinessRouteStatusDeleted {
		return nil, errors.New("该运营路线已删除")
	}

	err = ppzModel.TbPpzBusinessRouteModel.Delete(req.RouteId)
	if err != nil {
		return nil, fmt.Errorf("删除运营路线失败: %w", err)
	}

	return &ppzCs.BusinessRouteDeleteResponse{}, nil
}
