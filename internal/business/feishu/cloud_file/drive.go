package cloudfile

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	libraryFeishu "github.com/armylong/go-library/service/feishu"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
)

const uploadAllMaxSize = 20 * 1024 * 1024

type driveBusiness struct{}

var DriveBusiness = &driveBusiness{}

func (b *driveBusiness) Upload(ctx context.Context, filePath string, tableToken string) (fileToken string, err error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}
	if stat.Size() <= uploadAllMaxSize {
		return b.UploadAll(ctx, filePath, tableToken)
	}
	return b.UploadMultipart(ctx, filePath, tableToken)
}

func (b *driveBusiness) UploadAll(ctx context.Context, filePath string, tableToken string) (fileToken string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	fsConfig := libraryFeishu.GetFsConfig()
	userAccessToken := libraryFeishu.GetUserAccessToken(nil)

	client := lark.NewClient(fsConfig.AppId, fsConfig.AppSecret)

	req := larkdrive.NewUploadAllMediaReqBuilder().
		Body(larkdrive.NewUploadAllMediaReqBodyBuilder().
			FileName(filepath.Base(filePath)).
			ParentType(`bitable_file`).
			ParentNode(tableToken).
			Size(int(stat.Size())).
			Extra(fmt.Sprintf(`{\"drive_route_token\":\"%s\"}`, tableToken)).
			File(f).
			Build()).
		Build()

	// 发起请求
	resp, err := client.Drive.V1.Media.UploadAll(ctx, req, larkcore.WithUserAccessToken(userAccessToken))

	// 处理错误
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Printf("logId: %s, error response: \n%s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		return "", resp.CodeError
	}

	// 业务处理
	// fmt.Println(larkcore.Prettify(resp))

	return *resp.Data.FileToken, nil
}

func (b *driveBusiness) UploadMultipart(ctx context.Context, filePath string, tableToken string) (fileToken string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	fsConfig := libraryFeishu.GetFsConfig()
	userAccessToken := libraryFeishu.GetUserAccessToken(nil)
	client := lark.NewClient(fsConfig.AppId, fsConfig.AppSecret)

	prepareReq := larkdrive.NewUploadPrepareMediaReqBuilder().
		MediaUploadInfo(larkdrive.NewMediaUploadInfoBuilder().
			FileName(filepath.Base(filePath)).
			ParentType(`bitable_file`).
			ParentNode(tableToken).
			Size(int(stat.Size())).
			Extra(fmt.Sprintf(`{\"drive_route_token\":\"%s\"}`, tableToken)).
			Build()).
		Build()

	prepareResp, err := client.Drive.V1.Media.UploadPrepare(ctx, prepareReq, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return "", fmt.Errorf("预上传失败: %w", err)
	}
	if !prepareResp.Success() {
		return "", fmt.Errorf("预上传失败: %s", larkcore.Prettify(prepareResp.CodeError))
	}

	uploadId := *prepareResp.Data.UploadId
	blockSize := *prepareResp.Data.BlockSize
	blockNum := *prepareResp.Data.BlockNum

	for i := 0; i < blockNum; i++ {
		offset := int64(i) * int64(blockSize)
		remaining := stat.Size() - offset
		chunkSize := int64(blockSize)
		if remaining < chunkSize {
			chunkSize = remaining
		}

		chunk := make([]byte, chunkSize)
		_, err := f.ReadAt(chunk, offset)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("读取分片%d失败: %w", i, err)
		}

		partReq := larkdrive.NewUploadPartMediaReqBuilder().
			Body(larkdrive.NewUploadPartMediaReqBodyBuilder().
				UploadId(uploadId).
				Seq(i).
				Size(int(chunkSize)).
				File(bytes.NewReader(chunk)).
				Build()).
			Build()

		partResp, err := client.Drive.V1.Media.UploadPart(ctx, partReq, larkcore.WithUserAccessToken(userAccessToken))
		if err != nil {
			return "", fmt.Errorf("上传分片%d失败: %w", i, err)
		}
		if !partResp.Success() {
			return "", fmt.Errorf("上传分片%d失败: %s", i, larkcore.Prettify(partResp.CodeError))
		}
		fmt.Printf("分片上传进度: %d/%d\n", i+1, blockNum)
	}

	finishReq := larkdrive.NewUploadFinishMediaReqBuilder().
		Body(larkdrive.NewUploadFinishMediaReqBodyBuilder().
			UploadId(uploadId).
			BlockNum(blockNum).
			Build()).
		Build()

	finishResp, err := client.Drive.V1.Media.UploadFinish(ctx, finishReq, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return "", fmt.Errorf("完成上传失败: %w", err)
	}
	if !finishResp.Success() {
		return "", fmt.Errorf("完成上传失败: %s", larkcore.Prettify(finishResp.CodeError))
	}

	return *finishResp.Data.FileToken, nil
}
