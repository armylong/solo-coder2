package feishu

// 飞书上传-请求
// 通过 multipart/form-data 提交文件和上传参数
// type 支持 UploadAll 和 UploadMultipart
// UploadMultipart 场景下需要额外传 upload_id、seq、block_num
type FeishuUploadRequest struct {
	Type       string `json:"type" form:"type"`
	TableToken string `json:"table_token" form:"table_token"`
	ParentType string `json:"parent_type" form:"parent_type"`
	UploadID   string `json:"upload_id" form:"upload_id"`
	Seq        int    `json:"seq" form:"seq"`
	BlockNum   int    `json:"block_num" form:"block_num"`
}

// 飞书上传-响应
// done 表示当前上传流程是否完成，完成后返回最终 file_token
type FeishuUploadResponse struct {
	Done      bool   `json:"done"`
	FileToken string `json:"file_token"`
	Seq       int    `json:"seq"`
	BlockNum  int    `json:"block_num"`
}
