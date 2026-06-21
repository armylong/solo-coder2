package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"encoding/json"

	feishuCloudDocBusiness "github.com/armylong/armylong-go/internal/business/feishu/cloud_doc"
	configWork "github.com/armylong/armylong-go/internal/common/config"
	libraryUtils "github.com/armylong/go-library/utils"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/urfave/cli/v2"
)

var SoloCoderWorkHome string

const (
	SoloCoderFeishuDocAppToken = `HO1Kb3JQpa9e0WsHhQ0c2dpLnYb`
	SoloCoderFeishuDocTableId  = `tblfVRXx3qotQ4mO`
	SoloCoderFeishuDocViewId   = `vewKdKoVia`
)

var SoloCoderTaskRoundList = []string{
	"第一轮",
	"第二轮",
	"第三轮",
	"第四轮",
	"第五轮",
}

var SoloCoderTaskTypeList = []string{
	"Bug修复",
	"0-1代码生成",
	"Feature迭代",
	"代码重构",
	"工程化",
	"代码测试",
}

var SoloCoderTaskBusinessDomainList = []string{
	"纯后端API服务",
	"Web前端",
	"全栈Web应用",
	"游戏开发",
	"数据分析与可视化",
	"3D/交互可视化",
	"AI/ML应用",
	"科学计算",
	"命令行工具",
}

var SoloCoderTaskModifyScopeList = []string{
	"单文件",
	"模块内多文件",
	"跨模块多文件",
	"跨系统多模块",
}

var SoloCoderTaskIsCompletedList = []string{
	"完成了任务",
	"未完成任务",
}

var SoloCoderTaskIsSatisfiedList = []string{
	"满意",
	"不满意",
}

var SoloCoderTaskGithubURL = "https://github.com/armylong/solo-coder"

type SoloCoderTask struct {
	TraeSessionID    string `json:"Trae Session ID"`
	UserPrompt       string `json:"User Prompt"`
	Round            string `json:"轮次"`
	TaskType         string `json:"任务类型"`
	BusinessDomain   string `json:"业务领域"`
	ModifyScope      string `json:"修改范围"`
	IsCompleted      string `json:"任务是否完成"`
	IsSatisfied      string `json:"产物及过程是否满意"`
	DissatisfyReason string `json:"不满意原因"`
	GithubURL        string `json:"github地址"`
	BranchOrFolder   string `json:"分支/文件夹"`
	LogTrace         string `json:"日志轨迹"`
	// Screenshots      string `json:"截图(userprompt附件/产物/运行结果/对话)"`
}

type SoloCoderTaskList []SoloCoderTask

func init() {
	userHomeDir, _ := os.UserHomeDir()
	SoloCoderWorkHome = fmt.Sprintf("%s/works/solo-coder/works", userHomeDir)
}

// solo coder会话命令（开发中）
func SoloCoderSessionHandler(c *cli.Context) error {
	id := c.String("id")
	if id == "" {
		return fmt.Errorf("题目ID不能为空")
	}

	// /Users/zhangzelong/Library/Application Support/Trae CN/User/workspaceStorage/

	// _, err := sqlite.DB.DB().Exec(sql)
	// if err != nil {
	// 	return fmt.Errorf("sqlite exec error: %v", err)
	// }

	return nil
}

func SoloCoderUploadFeishuHandler(c *cli.Context) error {
	// 循环WorkHome下所有目录, 判断是否为未上传题目目录, 判断目录中是否有: configWork.WorkDoneFileName
	workHome := SoloCoderWorkHome
	dirs, err := os.ReadDir(workHome)
	if err != nil {
		return fmt.Errorf("读取目录 %s 失败: %v", workHome, err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		// 判断目录中是否有上传完成标记文件, 有则跳过
		uploadDoneFile := filepath.Join(workHome, dir.Name(), configWork.UploadDoneFileName)
		if _, err := os.Stat(uploadDoneFile); err == nil {
			continue
		}
		// 判断目录中是否有完成标记文件, 没有则跳过
		doneFile := filepath.Join(workHome, dir.Name(), configWork.WorkDoneFileName)
		if _, err := os.Stat(doneFile); os.IsNotExist(err) {
			continue
		}
		// 没有完成标记的目录, 如果没有工作文件的话, 忽略
		workFile := filepath.Join(workHome, dir.Name(), configWork.WorkFileName)
		if _, err := os.Stat(workFile); os.IsNotExist(err) {
			continue
		}
		fmt.Println(dir.Name())
		// 解析工作文件, SoloCoderTaskList
		data, err := os.ReadFile(workFile)
		if err != nil {
			return fmt.Errorf("读取工作文件 %s 失败: %v", workFile, err)
		}
		var tasksList SoloCoderTaskList
		if err := json.Unmarshal(data, &tasksList); err != nil {
			return fmt.Errorf("解析工作文件 %s 失败: %v", workFile, err)
		}

		// 判断任务列表是否可用
		taskReadyList, err := isTaskListValid(tasksList)
		if err != nil {
			return err
		}
		if taskReadyList == nil {
			continue
		}

		// 循环上传任务
		for _, task := range taskReadyList {
			// 上传任务到飞书
			if err := uploadTaskToFeishu(task); err != nil {
				return fmt.Errorf("上传任务到飞书失败: %v", err)
			} else {
				fmt.Printf("上传任务到飞书成功: %v\n", task.TraeSessionID)
			}
		}
		// 上传完成后, 写入上传完成标记文件
		if err := os.WriteFile(uploadDoneFile, []byte("1"), 0644); err != nil {
			return fmt.Errorf("写入上传完成标记文件 %s 失败: %v", uploadDoneFile, err)
		}
	}

	return nil
}

// 判断任务列表是否可用
func isTaskListValid(tasksList SoloCoderTaskList) (taskReadyList SoloCoderTaskList, err error) {
	// 任务列表长度必须为5条, 报警提示
	if len(tasksList) != 5 {
		return nil, fmt.Errorf("任务列表长度必须为5条")
	}

	beforeReadyYn := false
	for _i, task := range tasksList {
		taskDataSuccess, err := isTaskDataValid(&task)
		if err != nil {
			return nil, fmt.Errorf("任务第 %d 个数据校验失败: %v", _i+1, err)
		}
		// 第一条如果没填完, 代表数据出错了, 报警提示
		if _i == 0 && !taskDataSuccess {
			return nil, fmt.Errorf("有完成标识, 但是任务列表中第1条数据没写完")
		}
		if !taskDataSuccess {
			beforeReadyYn = false
			continue
		} else {
			if _i > 0 && !beforeReadyYn {
				return nil, fmt.Errorf("上个任务未填写完成, 但是下个任务 %d 填写完成, 这是不正常的", _i+1)
			}
			beforeReadyYn = true
			taskReadyList = append(taskReadyList, task)
		}
	}

	// 一条也没完成代表数据出错了, 报警提示
	if len(taskReadyList) == 0 {
		return nil, fmt.Errorf("任务列表中没有完成的任务")
	}

	// 只完成了一条, 则跳过
	if len(taskReadyList) == 1 {
		return nil, fmt.Errorf("任务列表中只完成了一条任务, 不能上传")
	}

	return taskReadyList, nil
}

// 判断任务数据是否可用
func isTaskDataValid(task *SoloCoderTask) (ret bool, err error) {
	if task.TraeSessionID != "" && (task.UserPrompt == "" ||
		task.Round == "" ||
		task.TaskType == "" ||
		task.BusinessDomain == "" ||
		task.ModifyScope == "" ||
		task.IsCompleted == "" ||
		task.IsSatisfied == "" ||
		task.GithubURL == "" ||
		task.BranchOrFolder == "" ||
		task.LogTrace == "") {
		return false, fmt.Errorf("任务数据里个别字段缺失, 请检查任务数据")
	}
	// 所有字段如果有为空, 则跳过
	if task.TraeSessionID == "" ||
		task.UserPrompt == "" ||
		task.Round == "" ||
		task.TaskType == "" ||
		task.BusinessDomain == "" ||
		task.ModifyScope == "" ||
		task.IsCompleted == "" ||
		task.IsSatisfied == "" ||
		task.GithubURL == "" ||
		task.BranchOrFolder == "" ||
		task.LogTrace == "" {
		return false, nil
	}
	if task.IsCompleted == "完成了任务" && (task.IsSatisfied != "满意" || task.DissatisfyReason != "") {
		return false, fmt.Errorf("完成了任务后, 必须为满意+不满意原因必须为空")
	}
	if task.IsCompleted == "未完成任务" && (task.IsSatisfied != "不满意" || task.DissatisfyReason == "") {
		return false, fmt.Errorf("未完成任务的, 必须为不满意+不满意原因不能为空")
	}

	// 轮次
	if !libraryUtils.InArray(task.Round, SoloCoderTaskRoundList) {
		return false, fmt.Errorf("轮次 %s 选项不正确", task.Round)
	}
	// 任务类型
	if !libraryUtils.InArray(task.TaskType, SoloCoderTaskTypeList) {
		return false, fmt.Errorf("任务类型 %s 选项不正确", task.TaskType)
	}
	// 业务领域
	if !libraryUtils.InArray(task.BusinessDomain, SoloCoderTaskBusinessDomainList) {
		return false, fmt.Errorf("业务领域 %s 选项不正确", task.BusinessDomain)
	}
	// 修改范围
	if !libraryUtils.InArray(task.ModifyScope, SoloCoderTaskModifyScopeList) {
		return false, fmt.Errorf("修改范围 %s 选项不正确", task.ModifyScope)
	}
	// 任务是否完成
	if !libraryUtils.InArray(task.IsCompleted, SoloCoderTaskIsCompletedList) {
		return false, fmt.Errorf("任务是否完成 %s 选项不正确", task.IsCompleted)
	}
	// 产物及过程是否满意
	if !libraryUtils.InArray(task.IsSatisfied, SoloCoderTaskIsSatisfiedList) {
		return false, fmt.Errorf("产物及过程是否满意 %s 选项不正确", task.IsSatisfied)
	}
	return true, nil
}

// 上传任务
func uploadTaskToFeishu(task SoloCoderTask) error {
	// 先获取一个飞书未填写的任务行
	var recordId string
	// 读取是不是有已存在的
	req1 := larkbitable.NewSearchAppTableRecordReqBuilder().
		AppToken(SoloCoderFeishuDocAppToken).
		TableId(SoloCoderFeishuDocTableId).
		PageSize(1).
		Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
			ViewId(SoloCoderFeishuDocViewId).
			Filter(&larkbitable.FilterInfo{
				Conjunction: larkcore.StringPtr(`and`),
				Conditions: []*larkbitable.Condition{
					{
						FieldName: larkcore.StringPtr(`Trae Session ID`),
						Operator:  larkcore.StringPtr(`is`),
						Value:     []string{task.TraeSessionID},
					},
				},
			}).
			// AutomaticFields().
			Build()).
		Build()
	ctx := context.Background()
	feishuResp1, err := feishuCloudDocBusiness.BaseTablesBusiness.SearchBaseTables(ctx, req1)
	if err != nil {
		return err
	}

	if feishuResp1 != nil && feishuResp1.Data != nil && len(feishuResp1.Data.Items) > 0 {
		recordId = *feishuResp1.Data.Items[0].RecordId
	}

	if recordId == "" {
		// 读取飞书多维表格中未完成的工作
		req := larkbitable.NewSearchAppTableRecordReqBuilder().
			AppToken(SoloCoderFeishuDocAppToken).
			TableId(SoloCoderFeishuDocTableId).
			PageSize(1).
			Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
				ViewId(SoloCoderFeishuDocViewId).
				Filter(&larkbitable.FilterInfo{
					Conjunction: larkcore.StringPtr(`and`),
					Conditions: []*larkbitable.Condition{
						{
							FieldName: larkcore.StringPtr(`Trae Session ID`),
							Operator:  larkcore.StringPtr(`isEmpty`),
							Value:     []string{},
						},
						{
							FieldName: larkcore.StringPtr(`User Prompt`),
							Operator:  larkcore.StringPtr(`isEmpty`),
							Value:     []string{},
						},
					},
				}).
				// AutomaticFields().
				Build()).
			Build()

		feishuResp, err := feishuCloudDocBusiness.BaseTablesBusiness.SearchBaseTables(ctx, req)
		if err != nil {
			return err
		}
		if feishuResp == nil || feishuResp.Data == nil {
			fmt.Println("返回数据为空")
			return fmt.Errorf("返回数据为空")
		}
		recordId = *feishuResp.Data.Items[0].RecordId
	}

	if recordId == "" {
		return fmt.Errorf("未找到未完成的工作行")
	}

	// 提交接口, 填写任务数据
	data := map[string]any{
		`Trae Session ID`: task.TraeSessionID,
		`User Prompt`:     task.UserPrompt,
		`轮次`:              task.Round,
		`任务类型`:            task.TaskType,
		`业务领域`:            task.BusinessDomain,
		`修改范围`:            task.ModifyScope,
		`任务是否完成`:          task.IsCompleted,
		`产物及过程是否满意`:       task.IsSatisfied,
		`不满意原因`:           task.DissatisfyReason,
		`github地址`:        task.GithubURL,
		`分支/文件夹`:          task.BranchOrFolder,
		`日志轨迹`:            task.LogTrace,
	}
	// 更新飞书多维表格记录字段
	resp, err := soloCoderUploadFeishuDoc(ctx, recordId, data)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("更新飞书多维表格记录字段失败")
	}

	return nil
}

// 更新飞书多维表格记录字段
func soloCoderUploadFeishuDoc(ctx context.Context, recordId string, Fields map[string]any) (resp *larkbitable.UpdateAppTableRecordResp, err error) {
	// 创建请求对象
	req := larkbitable.NewUpdateAppTableRecordReqBuilder().
		AppToken(SoloCoderFeishuDocAppToken).
		TableId(SoloCoderFeishuDocTableId).
		RecordId(recordId).
		// UserIdType(`open_id`).
		// IgnoreConsistencyCheck(true).
		AppTableRecord(larkbitable.NewAppTableRecordBuilder().
			Fields(Fields).
			Build()).
		Build()
	resp, err = feishuCloudDocBusiness.BaseTablesBusiness.UpdateBaseTables(ctx, req)

	// fmt.Println(larkcore.Prettify(resp))
	return
}
