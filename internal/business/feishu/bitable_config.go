package feishu

import (
	"context"
	"errors"
	"fmt"
	"strings"

	feishuCs "github.com/armylong/armylong-go/internal/cs/feishu"
	"github.com/armylong/armylong-go/internal/middlewares"
	feishuModel "github.com/armylong/armylong-go/internal/model/feishu"
)

type fsBitableConfigBusiness struct{}

var FsBitableConfigBusiness = &fsBitableConfigBusiness{}

// 表格配置列表
func (b *fsBitableConfigBusiness) List(ctx context.Context, req *feishuCs.BitableConfigListRequest) (*feishuCs.BitableConfigListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	offset := (req.Page - 1) * req.PageSize

	total, err := feishuModel.TbFeishuBitableConfigModel.Count(req.Status, req.Keyword)
	if err != nil {
		return nil, fmt.Errorf("获取飞书表格配置数量失败: %w", err)
	}

	list, err := feishuModel.TbFeishuBitableConfigModel.List(req.Status, req.Keyword, req.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("获取飞书表格配置列表失败: %w", err)
	}

	return &feishuCs.BitableConfigListResponse{
		List:     b.convertItems(list),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 创建表格配置
func (b *fsBitableConfigBusiness) Create(ctx context.Context, req *feishuCs.BitableConfigCreateRequest) (*feishuCs.BitableConfigCreateResponse, error) {
	appToken := strings.TrimSpace(req.AppToken)
	alias := strings.TrimSpace(req.Alias)
	if appToken == "" {
		return nil, errors.New("App Token不能为空")
	}

	uid := middlewares.GetLoginUIDFromContext(ctx)
	existing, err := feishuModel.TbFeishuBitableConfigModel.GetByAppToken(appToken)
	if err == nil && existing != nil {
		if existing.Status == feishuModel.FeishuBitableConfigStatusDeleted {
			existing.Alias = alias
			existing.Status = feishuModel.FeishuBitableConfigStatusNormal
			existing.UpdatedUid = uid
			if err = feishuModel.TbFeishuBitableConfigModel.Update(existing); err != nil {
				return nil, fmt.Errorf("恢复飞书表格配置失败: %w", err)
			}
			return &feishuCs.BitableConfigCreateResponse{Id: existing.Id}, nil
		}
		return nil, errors.New("该表格已存在")
	}

	data := &feishuModel.TbFeishuBitableConfig{
		AppToken:   appToken,
		Alias:      alias,
		Status:     feishuModel.FeishuBitableConfigStatusNormal,
		CreatedUid: uid,
		UpdatedUid: uid,
	}

	id, err := feishuModel.TbFeishuBitableConfigModel.Create(data)
	if err != nil {
		return nil, fmt.Errorf("创建飞书表格配置失败: %w", err)
	}

	return &feishuCs.BitableConfigCreateResponse{Id: id}, nil
}

// 更新表格配置
func (b *fsBitableConfigBusiness) Update(ctx context.Context, req *feishuCs.BitableConfigUpdateRequest) (*feishuCs.BitableConfigUpdateResponse, error) {
	if req.Id <= 0 {
		return nil, errors.New("配置ID不能为空")
	}

	appToken := strings.TrimSpace(req.AppToken)
	alias := strings.TrimSpace(req.Alias)
	if appToken == "" {
		return nil, errors.New("App Token不能为空")
	}

	data, err := feishuModel.TbFeishuBitableConfigModel.GetById(req.Id)
	if err != nil || data == nil {
		return nil, errors.New("飞书表格配置不存在")
	}
	if data.Status == feishuModel.FeishuBitableConfigStatusDeleted {
		return nil, errors.New("该飞书表格配置已删除")
	}

	duplicate, err := feishuModel.TbFeishuBitableConfigModel.GetByAppToken(appToken)
	if err == nil && duplicate != nil && duplicate.Id != data.Id && duplicate.Status != feishuModel.FeishuBitableConfigStatusDeleted {
		return nil, errors.New("App Token已存在")
	}

	uid := middlewares.GetLoginUIDFromContext(ctx)
	data.AppToken = appToken
	data.Alias = alias
	if req.Status > 0 {
		data.Status = req.Status
		if data.Status == feishuModel.FeishuBitableConfigStatusDeleted {
			data.Status = feishuModel.FeishuBitableConfigStatusNormal
		}
	}
	data.UpdatedUid = uid

	if err = feishuModel.TbFeishuBitableConfigModel.Update(data); err != nil {
		return nil, fmt.Errorf("更新飞书表格配置失败: %w", err)
	}

	return &feishuCs.BitableConfigUpdateResponse{}, nil
}

// 获取表格配置详情
func (b *fsBitableConfigBusiness) Get(ctx context.Context, req *feishuCs.BitableConfigGetRequest) (*feishuCs.BitableConfigGetResponse, error) {
	if req.Id <= 0 {
		return nil, errors.New("配置ID不能为空")
	}

	data, err := feishuModel.TbFeishuBitableConfigModel.GetById(req.Id)
	if err != nil || data == nil {
		return nil, errors.New("飞书表格配置不存在")
	}
	if data.Status == feishuModel.FeishuBitableConfigStatusDeleted {
		return nil, errors.New("该飞书表格配置已删除")
	}

	return &feishuCs.BitableConfigGetResponse{
		Bitable: b.convertItem(data),
	}, nil
}

// 删除表格配置
func (b *fsBitableConfigBusiness) Delete(ctx context.Context, req *feishuCs.BitableConfigDeleteRequest) (*feishuCs.BitableConfigDeleteResponse, error) {
	if req.Id <= 0 {
		return nil, errors.New("配置ID不能为空")
	}

	data, err := feishuModel.TbFeishuBitableConfigModel.GetById(req.Id)
	if err != nil || data == nil {
		return nil, errors.New("飞书表格配置不存在")
	}
	if data.Status == feishuModel.FeishuBitableConfigStatusDeleted {
		return nil, errors.New("该飞书表格配置已删除")
	}

	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err = feishuModel.TbFeishuBitableConfigModel.Delete(req.Id, uid); err != nil {
		return nil, fmt.Errorf("删除飞书表格配置失败: %w", err)
	}

	return &feishuCs.BitableConfigDeleteResponse{}, nil
}

// 获取概览页使用的表格列表
func (b *fsBitableConfigBusiness) GetBitables() []feishuCs.BitableInfo {
	list, err := feishuModel.TbFeishuBitableConfigModel.ListActive()
	if err != nil {
		return []feishuCs.BitableInfo{}
	}

	result := make([]feishuCs.BitableInfo, 0, len(list))
	for _, item := range list {
		result = append(result, feishuCs.BitableInfo{
			Id:       item.Id,
			AppToken: strings.TrimSpace(item.AppToken),
			Alias:    strings.TrimSpace(item.Alias),
		})
	}
	return result
}

// 清空全部表格配置
func (b *fsBitableConfigBusiness) DeleteAll() error {
	return feishuModel.TbFeishuBitableConfigModel.DeleteAll()
}

// 转换单条数据
func (b *fsBitableConfigBusiness) convertItem(item *feishuModel.TbFeishuBitableConfig) *feishuCs.BitableConfigItem {
	if item == nil {
		return nil
	}
	return &feishuCs.BitableConfigItem{
		Id:         item.Id,
		AppToken:   item.AppToken,
		Alias:      item.Alias,
		Status:     item.Status,
		CreatedUid: item.CreatedUid,
		UpdatedUid: item.UpdatedUid,
		CreatedAt:  item.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// 批量转换数据
func (b *fsBitableConfigBusiness) convertItems(list []*feishuModel.TbFeishuBitableConfig) []*feishuCs.BitableConfigItem {
	result := make([]*feishuCs.BitableConfigItem, 0, len(list))
	for _, item := range list {
		result = append(result, b.convertItem(item))
	}
	return result
}
