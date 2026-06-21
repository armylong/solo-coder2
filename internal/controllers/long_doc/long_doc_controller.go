package long_doc

import (
	"context"

	longDocBiz "github.com/armylong/armylong-go/internal/business/long_doc"
	"github.com/armylong/armylong-go/internal/middlewares"
	longDocCs "github.com/armylong/armylong-go/internal/cs/long_doc"
)

// LongDocController 长文档
type LongDocController struct{}

// 获取文档树
func (c *LongDocController) ActionList(ctx context.Context, req *longDocCs.DocListRequest) (*longDocCs.DocListResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil, nil
	}
	return longDocBiz.LongDocBusiness.GetDocList(uid)
}

// 创建文档
func (c *LongDocController) ActionCreate(ctx context.Context, req *longDocCs.CreateDocRequest) (*longDocCs.CreateDocResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil, nil
	}
	return longDocBiz.LongDocBusiness.CreateDoc(uid, req)
}

// 删除文档
func (c *LongDocController) ActionDelete(ctx context.Context, req *longDocCs.DeleteDocRequest) error {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil
	}
	return longDocBiz.LongDocBusiness.DeleteDoc(uid, req)
}

// 获取文档详情
func (c *LongDocController) ActionGet(ctx context.Context, req *longDocCs.GetDocRequest) (*longDocCs.GetDocResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil, nil
	}
	return longDocBiz.LongDocBusiness.GetDoc(uid, req)
}

// 保存文档内容
func (c *LongDocController) ActionSave(ctx context.Context, req *longDocCs.SaveDocRequest) error {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil
	}
	return longDocBiz.LongDocBusiness.SaveDoc(uid, req)
}

// 重命名文档
func (c *LongDocController) ActionRename(ctx context.Context, req *longDocCs.RenameDocRequest) error {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil
	}
	return longDocBiz.LongDocBusiness.RenameDoc(uid, req)
}

// 移动文档（拖拽排序）
func (c *LongDocController) ActionMove(ctx context.Context, req *longDocCs.MoveDocRequest) error {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil
	}
	return longDocBiz.LongDocBusiness.MoveDoc(uid, req)
}
