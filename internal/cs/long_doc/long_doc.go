package long_doc

// 文档列表-请求
type DocListRequest struct{}

// 文档树节点
type DocTreeNode struct {
	DocId       int64          `json:"doc_id"`       // 文档ID
	ParentDocId int64          `json:"parent_doc_id"` // 父文档ID
	DocName     string         `json:"doc_name"`     // 文档名称
	SortOrder   int            `json:"sort_order"`   // 排序值
	Children    []*DocTreeNode `json:"children,omitempty"` // 子文档
}

// 文档列表-响应
type DocListResponse struct {
	Documents []*DocTreeNode `json:"documents"` // 文档树
}

// 创建文档-请求
type CreateDocRequest struct {
	DocName     string `json:"doc_name" form:"doc_name"`         // 文档名称
	ParentDocId int64  `json:"parent_doc_id" form:"parent_doc_id"` // 父文档ID
}

// 创建文档-响应
type CreateDocResponse struct {
	DocId       int64  `json:"doc_id"`       // 文档ID
	DocName     string `json:"doc_name"`     // 文档名称
	ParentDocId int64  `json:"parent_doc_id"` // 父文档ID
}

// 删除文档-请求
type DeleteDocRequest struct {
	DocId int64 `json:"doc_id" form:"doc_id"` // 文档ID
}

// 获取文档-请求
type GetDocRequest struct {
	DocId int64 `json:"doc_id" form:"doc_id"` // 文档ID
}

// 获取文档-响应
type GetDocResponse struct {
	DocId       int64  `json:"doc_id"`       // 文档ID
	ParentDocId int64  `json:"parent_doc_id"` // 父文档ID
	DocName     string `json:"doc_name"`     // 文档名称
	DocValue    string `json:"doc_value"`    // 文档内容
}

// 保存文档-请求
type SaveDocRequest struct {
	DocId    int64  `json:"doc_id" form:"doc_id"`       // 文档ID
	DocValue string `json:"doc_value" form:"doc_value"` // 文档内容
}

// 重命名文档-请求
type RenameDocRequest struct {
	DocId   int64  `json:"doc_id" form:"doc_id"`       // 文档ID
	DocName string `json:"doc_name" form:"doc_name"`   // 新名称
}

// 移动文档-请求
type MoveDocRequest struct {
	DocId       int64 `json:"doc_id" form:"doc_id"`               // 文档ID
	ParentDocId int64 `json:"parent_doc_id" form:"parent_doc_id"` // 目标父文档ID
	Position    int   `json:"position" form:"position"`           // 目标位置索引
}
