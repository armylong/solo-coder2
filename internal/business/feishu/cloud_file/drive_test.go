package cloudfile

import (
	"context"
	"fmt"
	"testing"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

func TestDriveBusiness_UploadAll(t *testing.T) {
	ret, err := DriveBusiness.UploadAll(context.Background(), "/Users/zhangzelong/code/go-code/armylong-go/docker-file.md", "ULuObnT5cajntxsxZiGc4YyOnDh")
	if err != nil {
		t.Errorf("上传文件失败: %v", err)
	}
	fmt.Println(ret, larkcore.Prettify(ret))
}
