package long_doc

import (
	"errors"

	longDocCs "github.com/armylong/armylong-go/internal/cs/long_doc"
	longDocModel "github.com/armylong/armylong-go/internal/model/long_doc"
)

type longDocBusiness struct{}

var LongDocBusiness = &longDocBusiness{}

// 将扁平文档列表转成树形结构
func buildDocTree(docs []*longDocModel.TbLongDoc, parentDocId int64) []*longDocCs.DocTreeNode {
	var nodes []*longDocCs.DocTreeNode
	
	for _, doc := range docs {
		if doc.ParentDocId == parentDocId {
			node := &longDocCs.DocTreeNode{
				DocId:       doc.DocId,
				ParentDocId: doc.ParentDocId,
				DocName:     doc.DocName,
				SortOrder:   doc.SortOrder,
			}
			node.Children = buildDocTree(docs, doc.DocId)
			nodes = append(nodes, node)
		}
	}
	
	return nodes
}

// 检查descendantId是否是ancestorId的后代（防止循环移动）
func isDescendant(uid, ancestorId, descendantId int64) bool {
	if ancestorId == 0 {
		return false
	}
	childDocs, _ := longDocModel.TbLongDocModel.ListByParentId(uid, ancestorId)
	for _, child := range childDocs {
		if child.DocId == descendantId {
			return true
		}
		if isDescendant(uid, child.DocId, descendantId) {
			return true
		}
	}
	return false
}

// 获取文档树
func (b *longDocBusiness) GetDocList(uid int64) (*longDocCs.DocListResponse, error) {
	if uid <= 0 {
		return nil, errors.New("用户ID不能为空")
	}

	docs, err := longDocModel.TbLongDocModel.ListByUid(uid)
	if err != nil {
		return &longDocCs.DocListResponse{Documents: []*longDocCs.DocTreeNode{}}, nil
	}

	tree := buildDocTree(docs, 0)

	return &longDocCs.DocListResponse{Documents: tree}, nil
}

// 创建文档
func (b *longDocBusiness) CreateDoc(uid int64, req *longDocCs.CreateDocRequest) (*longDocCs.CreateDocResponse, error) {
	if uid <= 0 {
		return nil, errors.New("用户ID不能为空")
	}
	if req.DocName == "" {
		return nil, errors.New("文档名称不能为空")
	}

	if req.ParentDocId > 0 {
		parentDoc, err := longDocModel.TbLongDocModel.GetByIdAndUid(req.ParentDocId, uid)
		if err != nil || parentDoc == nil {
			return nil, errors.New("父文档不存在")
		}
	}

	doc := &longDocModel.TbLongDoc{
		Uid:         uid,
		ParentDocId: req.ParentDocId,
		DocName:     req.DocName,
		DocValue:    "",
	}

	docId, err := longDocModel.TbLongDocModel.Create(doc)
	if err != nil {
		return nil, errors.New("创建文档失败")
	}

	return &longDocCs.CreateDocResponse{
		DocId:       docId,
		DocName:     req.DocName,
		ParentDocId: req.ParentDocId,
	}, nil
}

// 删除文档（含子文档）
func (b *longDocBusiness) DeleteDoc(uid int64, req *longDocCs.DeleteDocRequest) error {
	if uid <= 0 {
		return errors.New("用户ID不能为空")
	}
	if req.DocId <= 0 {
		return errors.New("文档ID不能为空")
	}

	doc, err := longDocModel.TbLongDocModel.GetByIdAndUid(req.DocId, uid)
	if err != nil || doc == nil {
		return errors.New("文档不存在")
	}

	err = longDocModel.TbLongDocModel.Delete(req.DocId, uid)
	if err != nil {
		return errors.New("删除文档失败")
	}

	return nil
}

// 获取文档详情
func (b *longDocBusiness) GetDoc(uid int64, req *longDocCs.GetDocRequest) (*longDocCs.GetDocResponse, error) {
	if uid <= 0 {
		return nil, errors.New("用户ID不能为空")
	}
	if req.DocId <= 0 {
		return nil, errors.New("文档ID不能为空")
	}

	doc, err := longDocModel.TbLongDocModel.GetByIdAndUid(req.DocId, uid)
	if err != nil || doc == nil {
		return nil, errors.New("文档不存在")
	}

	return &longDocCs.GetDocResponse{
		DocId:       doc.DocId,
		ParentDocId: doc.ParentDocId,
		DocName:     doc.DocName,
		DocValue:    doc.DocValue,
	}, nil
}

// 保存文档内容
func (b *longDocBusiness) SaveDoc(uid int64, req *longDocCs.SaveDocRequest) error {
	if uid <= 0 {
		return errors.New("用户ID不能为空")
	}
	if req.DocId <= 0 {
		return errors.New("文档ID不能为空")
	}

	doc, err := longDocModel.TbLongDocModel.GetByIdAndUid(req.DocId, uid)
	if err != nil || doc == nil {
		return errors.New("文档不存在")
	}

	err = longDocModel.TbLongDocModel.UpdateDocValue(req.DocId, uid, req.DocValue)
	if err != nil {
		return errors.New("保存文档失败")
	}

	return nil
}

// 重命名文档
func (b *longDocBusiness) RenameDoc(uid int64, req *longDocCs.RenameDocRequest) error {
	if uid <= 0 {
		return errors.New("用户ID不能为空")
	}
	if req.DocId <= 0 {
		return errors.New("文档ID不能为空")
	}
	if req.DocName == "" {
		return errors.New("文档名称不能为空")
	}

	doc, err := longDocModel.TbLongDocModel.GetByIdAndUid(req.DocId, uid)
	if err != nil || doc == nil {
		return errors.New("文档不存在")
	}

	doc.DocName = req.DocName
	err = longDocModel.TbLongDocModel.Update(doc)
	if err != nil {
		return errors.New("重命名失败")
	}

	return nil
}

// 移动文档（拖拽排序）
func (b *longDocBusiness) MoveDoc(uid int64, req *longDocCs.MoveDocRequest) error {
	if uid <= 0 {
		return errors.New("用户ID不能为空")
	}
	if req.DocId <= 0 {
		return errors.New("文档ID不能为空")
	}

	doc, err := longDocModel.TbLongDocModel.GetByIdAndUid(req.DocId, uid)
	if err != nil || doc == nil {
		return errors.New("文档不存在")
	}

	// 不能移动到自己的子文档下
	if isDescendant(uid, req.DocId, req.ParentDocId) {
		return errors.New("不能将文档移动到其子文档下")
	}

	siblings, err := longDocModel.TbLongDocModel.ListByParentId(uid, req.ParentDocId)
	if err != nil {
		return errors.New("获取同级文档失败")
	}

	var newSortOrder int
	if req.Position < 0 {
		req.Position = 0
	}
	if req.Position >= len(siblings) {
		maxOrder, _ := longDocModel.TbLongDocModel.GetMaxSortOrder(uid, req.ParentDocId)
		newSortOrder = maxOrder + 1
	} else {
		for i, sibling := range siblings {
			if sibling.DocId == req.DocId {
				continue
			}
			if i == req.Position {
				newSortOrder = sibling.SortOrder
				break
			}
		}
	}

	err = longDocModel.TbLongDocModel.UpdateParentAndSort(req.DocId, uid, req.ParentDocId, newSortOrder)
	if err != nil {
		return errors.New("移动文档失败")
	}

	return nil
}
