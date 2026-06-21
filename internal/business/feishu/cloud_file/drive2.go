package cloudfile

import (
	"context"
	"fmt"
	"io"
	"strings"

	libraryFeishu "github.com/armylong/go-library/service/feishu"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
)

type Drive2UploadType string

const (
	Drive2UploadTypeAll       Drive2UploadType = "UploadAll"
	Drive2UploadTypeMultipart Drive2UploadType = "UploadMultipart"
)

type Drive2UploadRequest struct {
	Type       Drive2UploadType
	TableToken string
	ParentType string
	FileName   string
	Size       int64
	File       io.Reader
	UploadID   string
	Seq        int
	BlockNum   int
}

type Drive2UploadResponse struct {
	Done      bool
	FileToken string
	Seq       int
	BlockNum  int
}

type drive2Business struct{}

var Drive2Business = &drive2Business{}

func (b *drive2Business) Upload(ctx context.Context, req *Drive2UploadRequest) (*Drive2UploadResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("上传参数不能为空")
	}

	switch req.Type {
	case Drive2UploadTypeAll:
		return b.UploadAll(ctx, req)
	case Drive2UploadTypeMultipart:
		return b.UploadMultipart(ctx, req)
	default:
		return nil, fmt.Errorf("不支持的上传类型: %s", req.Type)
	}
}

func (b *drive2Business) UploadAll(ctx context.Context, req *Drive2UploadRequest) (*Drive2UploadResponse, error) {
	if err := validateDrive2UploadAllRequest(req); err != nil {
		return nil, err
	}
	if req.Size > uploadAllMaxSize {
		return nil, fmt.Errorf("文件大小超过20MB，请改用分片上传")
	}

	parentType := strings.TrimSpace(req.ParentType)
	if parentType == "" {
		parentType = `bitable_file`
	}

	client, userAccessToken := newDrive2Client()
	uploadReq := larkdrive.NewUploadAllMediaReqBuilder().
		Body(larkdrive.NewUploadAllMediaReqBodyBuilder().
			FileName(req.FileName).
			ParentType(parentType).
			ParentNode(req.TableToken).
			Size(int(req.Size)).
			Extra(fmt.Sprintf(`{\"drive_route_token\":\"%s\"}`, req.TableToken)).
			File(req.File).
			Build()).
		Build()

	resp, err := client.Drive.V1.Media.UploadAll(ctx, uploadReq, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}
	if !resp.Success() {
		return nil, resp.CodeError
	}
	if resp.Data == nil || resp.Data.FileToken == nil {
		return nil, fmt.Errorf("上传文件失败: 缺少 file_token")
	}

	return &Drive2UploadResponse{
		Done:      true,
		FileToken: *resp.Data.FileToken,
	}, nil
}

func (b *drive2Business) UploadMultipart(ctx context.Context, req *Drive2UploadRequest) (*Drive2UploadResponse, error) {
	if err := validateDrive2UploadMultipartRequest(req); err != nil {
		return nil, err
	}

	client, userAccessToken := newDrive2Client()
	partReq := larkdrive.NewUploadPartMediaReqBuilder().
		Body(larkdrive.NewUploadPartMediaReqBodyBuilder().
			UploadId(req.UploadID).
			Seq(req.Seq).
			Size(int(req.Size)).
			File(req.File).
			Build()).
		Build()

	partResp, err := client.Drive.V1.Media.UploadPart(ctx, partReq, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return nil, fmt.Errorf("上传分片%d失败: %w", req.Seq, err)
	}
	if !partResp.Success() {
		return nil, fmt.Errorf("上传分片%d失败: %s", req.Seq, larkcore.Prettify(partResp.CodeError))
	}

	result := &Drive2UploadResponse{
		Done:     false,
		Seq:      req.Seq,
		BlockNum: req.BlockNum,
	}
	if req.Seq != req.BlockNum-1 {
		return result, nil
	}

	finishReq := larkdrive.NewUploadFinishMediaReqBuilder().
		Body(larkdrive.NewUploadFinishMediaReqBodyBuilder().
			UploadId(req.UploadID).
			BlockNum(req.BlockNum).
			Build()).
		Build()

	finishResp, err := client.Drive.V1.Media.UploadFinish(ctx, finishReq, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return nil, fmt.Errorf("完成上传失败: %w", err)
	}
	if !finishResp.Success() {
		return nil, fmt.Errorf("完成上传失败: %s", larkcore.Prettify(finishResp.CodeError))
	}
	if finishResp.Data == nil || finishResp.Data.FileToken == nil {
		return nil, fmt.Errorf("完成上传失败: 缺少 file_token")
	}

	result.Done = true
	result.FileToken = *finishResp.Data.FileToken
	return result, nil
}

func validateDrive2UploadAllRequest(req *Drive2UploadRequest) error {
	if req == nil {
		return fmt.Errorf("上传参数不能为空")
	}
	if strings.TrimSpace(req.TableToken) == "" {
		return fmt.Errorf("tableToken不能为空")
	}
	if strings.TrimSpace(req.FileName) == "" {
		return fmt.Errorf("fileName不能为空")
	}
	if req.Size <= 0 {
		return fmt.Errorf("文件大小必须大于0")
	}
	if req.File == nil {
		return fmt.Errorf("文件内容不能为空")
	}
	return nil
}

func validateDrive2UploadMultipartRequest(req *Drive2UploadRequest) error {
	if req == nil {
		return fmt.Errorf("上传参数不能为空")
	}
	if strings.TrimSpace(req.UploadID) == "" {
		return fmt.Errorf("uploadId不能为空")
	}
	if req.BlockNum <= 0 {
		return fmt.Errorf("blockNum必须大于0")
	}
	if req.Seq < 0 || req.Seq >= req.BlockNum {
		return fmt.Errorf("seq超出范围")
	}
	if req.Size <= 0 {
		return fmt.Errorf("分片大小必须大于0")
	}
	if req.File == nil {
		return fmt.Errorf("分片内容不能为空")
	}
	return nil
}

func newDrive2Client() (*lark.Client, string) {
	fsConfig := libraryFeishu.GetFsConfig()
	userAccessToken := libraryFeishu.GetUserAccessToken(nil)
	client := lark.NewClient(fsConfig.AppId, fsConfig.AppSecret)
	return client, userAccessToken
}
