package cmd

import (
	"context"
	"fmt"

	workBusiness "github.com/armylong/armylong-go/internal/business/work"
	configWork "github.com/armylong/armylong-go/internal/common/config"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/urfave/cli/v2"
)

// 豆包Pro评测命令
type doubaoProTestingCmd struct {
	feishuDocAppToken string
	feishuDocTableId  string
	feishuDocViewId   string
	feishuOpenId      string

	workHome string

	tableFieldNameOpenId string

	tableFieldNameScore1        string
	tableFieldNameScore2        string
	tableFieldNameScore3        string
	tableFieldNameScore4        string
	tableFieldNameScoreRemark   string
	tableFieldNameScoreOptimize string

	tableFieldNameSpMemory string
	tableFieldNameContext  string
	tableFieldNameLevel1   string
	tableFieldNameLevel2   string
	tableFieldNamePrompt   string
	tableFieldNameResponse string

	downloadFields       []string
	uploadFields         []string
	uploadRequiredFields []string
}

var DoubaoProTestingCmd = &doubaoProTestingCmd{}

func init() {
	DoubaoProTestingCmd.feishuDocAppToken = `CE3BwYISBiEG4KkG04UcTfr6nRh`
	DoubaoProTestingCmd.feishuDocTableId = `tbliWHNKeW9dcQnw`
	DoubaoProTestingCmd.feishuDocViewId = `vewGNON9rb`
	DoubaoProTestingCmd.feishuOpenId = `ou_8ba15f1ac045cca7d993b572471ca996`

	DoubaoProTestingCmd.workHome = `/root/works/doubao_testing`

	DoubaoProTestingCmd.tableFieldNameOpenId = `作业人飞书账号`

	DoubaoProTestingCmd.tableFieldNameScore1 = `需求理解`
	DoubaoProTestingCmd.tableFieldNameScore2 = `正确性`
	DoubaoProTestingCmd.tableFieldNameScore3 = `完整性`
	DoubaoProTestingCmd.tableFieldNameScore4 = `体验性`
	DoubaoProTestingCmd.tableFieldNameScoreRemark = `备注`
	DoubaoProTestingCmd.tableFieldNameScoreOptimize = `优化建议`

	DoubaoProTestingCmd.tableFieldNameSpMemory = `SP_memory`
	DoubaoProTestingCmd.tableFieldNameContext = `context`
	DoubaoProTestingCmd.tableFieldNameLevel1 = `message_intention_v4_offline_level1`
	DoubaoProTestingCmd.tableFieldNameLevel2 = `message_intention_v4_offline_level2`
	DoubaoProTestingCmd.tableFieldNamePrompt = `prompt`
	DoubaoProTestingCmd.tableFieldNameResponse = `response`

	DoubaoProTestingCmd.downloadFields = []string{
		DoubaoProTestingCmd.tableFieldNameSpMemory,
		DoubaoProTestingCmd.tableFieldNameContext,
		DoubaoProTestingCmd.tableFieldNameLevel1,
		DoubaoProTestingCmd.tableFieldNameLevel2,
		DoubaoProTestingCmd.tableFieldNamePrompt,
		DoubaoProTestingCmd.tableFieldNameResponse,
	}

	DoubaoProTestingCmd.uploadFields = []string{
		DoubaoProTestingCmd.tableFieldNameScore1,
		DoubaoProTestingCmd.tableFieldNameScore2,
		DoubaoProTestingCmd.tableFieldNameScore3,
		DoubaoProTestingCmd.tableFieldNameScore4,
		DoubaoProTestingCmd.tableFieldNameScoreRemark,
		DoubaoProTestingCmd.tableFieldNameScoreOptimize,
	}

	DoubaoProTestingCmd.uploadRequiredFields = []string{
		DoubaoProTestingCmd.tableFieldNameScore1,
		DoubaoProTestingCmd.tableFieldNameScore2,
		DoubaoProTestingCmd.tableFieldNameScore3,
		DoubaoProTestingCmd.tableFieldNameScore4,
		DoubaoProTestingCmd.tableFieldNameScoreRemark,
	}
}

// 豆包Pro评测入口，按action分发
func (d *doubaoProTestingCmd) DoubaoProTestingHandler(c *cli.Context) error {
	ctx := c.Context
	action := ""
	if c.NArg() > 0 {
		action = c.Args().Get(0)
	}

	switch action {
	case "download":
		d.downloadWorks(ctx)
	case "upload":
		d.uploadWorks(ctx)
	default:
		fmt.Printf("未知命令: %s\n", action)
		fmt.Println("可用命令: download, upload")
	}
	return nil
}

// 从飞书下载未完成的评测任务
func (d *doubaoProTestingCmd) downloadWorks(ctx context.Context) {
	workBusiness.DownloadBusiness.WorkHome = d.workHome
	workBusiness.DownloadBusiness.DownloadFields = d.downloadFields
	workBusiness.DownloadBusiness.FeishuDocAppToken = d.feishuDocAppToken
	workBusiness.DownloadBusiness.FeishuDocTableId = d.feishuDocTableId
	workBusiness.DownloadBusiness.FeishuDocViewId = d.feishuDocViewId
	workBusiness.DownloadBusiness.FilterConditions = d.getUncompletedWorksFilter()

	err := workBusiness.DownloadBusiness.DownloadWorks(ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println()
}

// 上传评测结果到飞书
func (d *doubaoProTestingCmd) uploadWorks(ctx context.Context) {
	workBusiness.UploadBusiness.WorkHome = d.workHome
	workBusiness.UploadBusiness.UploadFields = d.uploadFields
	workBusiness.UploadBusiness.UploadRequiredFields = d.uploadRequiredFields
	workBusiness.UploadBusiness.FeishuDocAppToken = d.feishuDocAppToken
	workBusiness.UploadBusiness.FeishuDocTableId = d.feishuDocTableId
	workBusiness.UploadBusiness.FeishuDocViewId = d.feishuDocViewId
	workBusiness.UploadBusiness.FilterConditions = d.getUncompletedWorksFilter()

	err := workBusiness.UploadBusiness.UploadWorks(ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println()
}

// 构建未完成任务的筛选条件
func (d *doubaoProTestingCmd) getUncompletedWorksFilter() []*larkbitable.Condition {
	return []*larkbitable.Condition{
		{
			FieldName: &d.tableFieldNameOpenId,
			Operator:  &configWork.OperatorIs,
			Value:     []string{d.feishuOpenId},
		},
		{
			FieldName: &d.tableFieldNameScore1,
			Operator:  &configWork.OperatorIsEmpty,
			Value:     []string{},
		},
		{
			FieldName: &d.tableFieldNameScore2,
			Operator:  &configWork.OperatorIsEmpty,
			Value:     []string{},
		},
		{
			FieldName: &d.tableFieldNameScore3,
			Operator:  &configWork.OperatorIsEmpty,
			Value:     []string{},
		},
		{
			FieldName: &d.tableFieldNameScore4,
			Operator:  &configWork.OperatorIsEmpty,
			Value:     []string{},
		},
		{
			FieldName: &d.tableFieldNameScoreRemark,
			Operator:  &configWork.OperatorIsEmpty,
			Value:     []string{},
		},
	}
}
