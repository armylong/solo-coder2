package cmd

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	feishuCloudDocBusiness "github.com/armylong/armylong-go/internal/business/feishu/cloud_doc"
	configWork "github.com/armylong/armylong-go/internal/common/config"
	libraryUtils "github.com/armylong/go-library/utils"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/spf13/cast"
	"github.com/urfave/cli/v2"
)

const (
	FeishuDocAppToken = `MtTqbmrgja4nSqsRLwycbSycn9b`
	FeishuDocTableId  = `tblSlD8LJRTJynDI`
	FeishuDocViewId   = `vewbVY9BGn`
)

var UploadFields = []string{
	"评分",
	"备注",
}

var UploadRequiredFields = []string{
	"评分",
	"备注",
}

type qaDataInit struct {
	Uid string           `json:"uid"`
	Qa  []qaDataInitItem `json:"qa"`
}

type qaDataInitItem struct {
	PageId int                    `json:"page_id"`
	Result []qaDataInitItemResult `json:"result"`
}

type qaDataInitItemResult struct {
	Pass    bool   `json:"passed"`
	Comment string `json:"comment"`
}

func getTextFieldValue(field interface{}) string {
	if field == nil {
		return ""
	}
	arr, ok := field.([]interface{})
	if !ok || len(arr) == 0 {
		return ""
	}
	var result string
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if m["type"] == "text" {
			if text, ok := m["text"].(string); ok {
				result += text
			}
		}
	}
	return result
}

func getUrlFieldValues(field interface{}) []string {
	if field == nil {
		return nil
	}
	arr, ok := field.([]interface{})
	if !ok {
		return nil
	}
	var urls []string
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if m["type"] == "url" {
			if link, ok := m["link"].(string); ok && link != "" {
				urls = append(urls, link)
			}
		}
	}
	return urls
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func RubricsHandler(c *cli.Context) error {
	fmt.Println("download | upload")
	return nil
}

func RubricsDownloadWhileHandler(c *cli.Context) error {
	for {
		RubricsDownloadHandler(c)
		time.Sleep(time.Second * 3)
	}
}

func RubricsDownloadHandler(c *cli.Context) error {
	ctx := c.Context

	workHome := c.String("work_home")
	// fmt.Println(workHome)
	workSpace := workHome + `/works`
	// fmt.Println(workSpace)

	// 创建请求对象
	req := larkbitable.NewSearchAppTableRecordReqBuilder().
		AppToken(FeishuDocAppToken).
		TableId(FeishuDocTableId).
		// UserIdType(``).
		// PageToken(``).
		PageSize(10).
		Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
			ViewId(FeishuDocViewId).
			// FieldNames([]string{`字段1`, `字段2`}).
			// Sort([]*larkbitable.Sort{
			// 	larkbitable.NewSortBuilder().
			// 		FieldName(`多行文本`).
			// 		Desc(true).
			// 		Build(),
			// }).
			Filter(&larkbitable.FilterInfo{
				Conjunction: larkcore.StringPtr(`and`),
				Conditions: []*larkbitable.Condition{
					{
						FieldName: larkcore.StringPtr(`评分`),
						Operator:  larkcore.StringPtr(`isEmpty`),
						Value:     []string{},
					},
					{
						FieldName: larkcore.StringPtr(`备注`),
						Operator:  larkcore.StringPtr(`isEmpty`),
						Value:     []string{},
					},
				},
			}).
			// AutomaticFields().
			Build()).
		Build()
	resp, err := feishuCloudDocBusiness.BaseTablesBusiness.SearchBaseTables(ctx, req)
	if err != nil {
		return err
	}
	// fmt.Println(larkcore.Prettify(resp))

	for _, item := range resp.Data.Items {

		uid := getTextFieldValue(item.Fields["uid"])
		// fmt.Println(uid)
		workDir := filepath.Join(workSpace, uid)

		prompt := getTextFieldValue(item.Fields["prompt"])
		// fmt.Println(prompt)
		repos := getUrlFieldValues(item.Fields["repo"])
		// fmt.Println(repos)
		rubrics := getTextFieldValue(item.Fields["rubrics"])
		// fmt.Println(rubrics)
		// 如果以上字段异常, 抛异常
		if uid == "" || prompt == "" || len(repos) == 0 || rubrics == "" {
			return fmt.Errorf("uid, prompt, repos, or rubrics is empty")
		}

		// 查看工作目录下是否存在该记录的工作目录
		if _, err = os.Stat(workDir); os.IsNotExist(err) {
			// 不存在则创建
			err = os.MkdirAll(workDir, 0755)
			if err != nil {
				fmt.Printf("创建工作目录 %s 失败: %v\n", workDir, err)
				continue
			}
		} else {
			// 已存在则跳过
			// fmt.Printf("工作目录已存在, 跳过: %v\n", workDir)
			continue
		}

		fmt.Println(workDir)

		if item.Fields == nil {
			fmt.Printf("字段为空: %v\n", item)
			return fmt.Errorf("fields is empty")
		}

		// 把prompt写入: 原始要求.md
		err = os.WriteFile(filepath.Join(workDir, `原始要求.md`), []byte(prompt), 0644)
		if err != nil {
			fmt.Printf("写入原始要求.md 失败: %v\n", err)
			return err
		}
		fmt.Printf("写入原始要求.md 成功: %s\n", filepath.Join(workDir, `原始要求.md`))

		// 把rubrics写入: 检查点.md
		err = os.WriteFile(filepath.Join(workDir, `检查点.md`), []byte(rubrics), 0644)
		if err != nil {
			fmt.Printf("写入检查点.md 失败: %v\n", err)
			return err
		}
		fmt.Printf("写入检查点.md 成功: %s\n", filepath.Join(workDir, `检查点.md`))

		// 循环repo的url, 下载到工作目录
		for i, repo := range repos {
			zipUrl := strings.ReplaceAll(repo, `index.html`, `code_files.zip`)
			zipPath := filepath.Join(workDir, "temp.zip")
			extractDir := filepath.Join(workDir, strconv.Itoa(i+1))

			// fmt.Printf("下载 %s\n", zipUrl)
			if err := downloadFile(zipUrl, zipPath); err != nil {
				fmt.Printf("下载失败: %v\n", err)
				return err
			}

			// fmt.Printf("解压到 %s\n", extractDir)
			if err := unzip(zipPath, extractDir); err != nil {
				fmt.Printf("解压失败: %v\n", err)
				os.Remove(zipPath)
				return err
			}

			os.Remove(zipPath)
		}
		fmt.Printf("压缩包解压完成\n")

		// 写入work.done文件
		err = os.WriteFile(filepath.Join(workDir, `work.done`), []byte(``), 0644)
		if err != nil {
			fmt.Printf("写入work.done 失败: %v\n", err)
			return err
		}
		fmt.Printf("写入work.done 成功: %s\n\n", filepath.Join(workDir, `work.done`))
	}

	// data := resp.Data

	return nil
}

func RubricsUploadWhileHandler(c *cli.Context) error {
	for {
		if err := RubricsUploadHandler(c); err != nil {
			// fmt.Printf("上传rubrics评分失败: %v\n", err)
			return err
		}
		time.Sleep(time.Second * 3)
	}
}

func RubricsUploadHandler(c *cli.Context) error {
	ctx := c.Context
	workHome := c.String("work_home")
	// fmt.Println(workHome)
	workSpace := workHome + `/works`
	// fmt.Println(workSpace)

	// 读取飞书多维表格中未完成的工作
	req := larkbitable.NewSearchAppTableRecordReqBuilder().
		AppToken(FeishuDocAppToken).
		TableId(FeishuDocTableId).
		// UserIdType(``).
		// PageToken(``).
		PageSize(10).
		Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
			ViewId(FeishuDocViewId).
			// FieldNames([]string{`字段1`, `字段2`}).
			// Sort([]*larkbitable.Sort{
			// 	larkbitable.NewSortBuilder().
			// 		FieldName(`多行文本`).
			// 		Desc(true).
			// 		Build(),
			// }).
			Filter(&larkbitable.FilterInfo{
				Conjunction: larkcore.StringPtr(`and`),
				Conditions: []*larkbitable.Condition{
					{
						FieldName: larkcore.StringPtr(`评分`),
						Operator:  larkcore.StringPtr(`isEmpty`),
						Value:     []string{},
					},
					{
						FieldName: larkcore.StringPtr(`备注`),
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

	// 创建工作数据到工作目录(如果已存在工作目录则跳过)
	// 上传已完成的工作数据至飞书多维表格
	for _, item := range feishuResp.Data.Items {
		uid := getTextFieldValue(item.Fields["uid"])
		// fmt.Println(uid)
		workDir := filepath.Join(workSpace, uid)

		// 查看工作目录下是否存在该记录的工作目录
		if _, err := os.Stat(workDir); os.IsNotExist(err) {
			// 不存在则跳过
			// fmt.Printf("工作目录不存在, 跳过: %v\n", workDir)
			continue
		}

		if item.RecordId == nil {
			fmt.Printf("记录ID为空: %v\n", item)
			return fmt.Errorf("记录ID为空: %v", item)
		}
		recordId := uid

		// 查看工作目录下是否存在上传完成标记文件
		uploadDoneFile := workDir + "/" + configWork.UploadDoneFileName
		if _, err := os.Stat(uploadDoneFile); err == nil {
			// 存在则跳过
			// fmt.Printf("上传完成标记文件已存在, 跳过: %v\n", uploadDoneFile)
			continue
		}

		// 查看工作目录下是否存在该记录的质检文件
		qaFile := workDir + "/" + configWork.QaFileName
		if _, err := os.Stat(qaFile); os.IsNotExist(err) {
			// 不存在则跳过
			// fmt.Printf("质检文件不存在, 跳过: %v\n", qaFile)
			continue
		}
		qaDoneFile := workDir + "/" + configWork.WorkDoneFileName
		if _, err := os.Stat(qaDoneFile); os.IsNotExist(err) {
			// 不存在则跳过
			// fmt.Printf("质检完成标记文件不存在, 跳过: %v\n", qaDoneFile)
			continue
		}

		fmt.Println(workDir)

		// 读取质检文件
		qaData, err := os.ReadFile(qaFile)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		qaDataInit := qaDataInit{}
		err = json.Unmarshal(qaData, &qaDataInit)
		if err != nil {
			fmt.Printf("质检文件解析失败: %v, err:%v\n", qaData, err)
			return err
		}

		// var qaDataMap map[string]any
		if qaDataInit.Uid != uid {
			fmt.Printf("质检文件中的uid与记录ID不一致: %v, 跳过\n", qaDataInit.Uid)
			return fmt.Errorf("质检文件中的uid与记录ID不一致: %v", qaDataInit.Uid)
		}

		var scoreList [][]int
		var commentList []string
		var noPassCount int
		for index, value := range qaDataInit.Qa {
			_pageId := index + 1
			var _scoreList []int
			for _index, _v := range value.Result {
				rubricIndex := _index + 1
				_scoreList = append(_scoreList, cast.ToInt(_v.Pass))
				if !_v.Pass {
					noPassCount++
					if _v.Comment == `` {
						fmt.Printf("第%d个页面->第%d条rubrics->备注为空\n", _pageId, rubricIndex)
						return fmt.Errorf("第%d个页面->第%d条rubrics->备注为空", _pageId, rubricIndex)
					}
					commentList = append(commentList, fmt.Sprintf("第%d个页面->第%d条rubrics->%s", _pageId, rubricIndex, _v.Comment))
				}
			}
			scoreList = append(scoreList, _scoreList)
		}

		// 全部满分不用写备注
		if noPassCount == 0 {
			UploadRequiredFields = []string{
				"评分",
			}
		}

		scoreListJson, _err := json.Marshal(scoreList)
		if _err != nil {
			fmt.Printf("解析记录 %s 失败: %v\n", recordId, _err)
			continue
		}
		commentStr := ""
		for index, _comment := range commentList {
			commentStr += fmt.Sprintf("%d. %s\n", index+1, _comment)
		}
		qaDataMap := map[string]any{
			"评分": string(scoreListJson),
			"备注": commentStr,
		}

		uploadMap := make(map[string]any)
		var missingFields []string // 缺失字段
		for _, field := range UploadFields {
			if value, ok := qaDataMap[field]; !ok {
				// 如果字段不存在, 且是必填字段, 跳过
				if libraryUtils.InArray(field, UploadRequiredFields) {
					missingFields = append(missingFields, field)
					continue
				}
			} else {
				// 如果字段存在, 且是必填字段, 但值为空, 跳过
				if libraryUtils.InArray(field, UploadRequiredFields) && (value == nil || value == "") {
					missingFields = append(missingFields, field)
					continue
				} else {
					uploadMap[field] = value
				}
			}
		}

		if len(missingFields) > 0 {
			fmt.Printf("工作产出文件必填字段校验失败: %v, 跳过\n", missingFields)
			continue
		}

		_, err = uploadFeishuDoc(ctx, *item.RecordId, uploadMap)
		if err != nil {
			fmt.Printf("更新飞书多维表格记录字段失败: %v, err:%v\n", uploadMap, err)
			continue
		}

		// 创建上传完成标记文件
		if err := os.WriteFile(uploadDoneFile, []byte("1"), 0644); err != nil {
			fmt.Printf("创建上传完成标记文件失败: %v, err:%v\n", uploadDoneFile, err)
			continue
		}
		fmt.Printf("✓ 质检 %v 已更新飞书多维表格记录字段: %v\n\n", recordId, uploadMap)
	}
	// fmt.Println(feishuResp)
	// fmt.Printf("全部工作目录已处理: %s\n", workSpace)

	return nil

}

// 更新飞书多维表格记录字段
func uploadFeishuDoc(ctx context.Context, recordId string, Fields map[string]any) (resp *larkbitable.UpdateAppTableRecordResp, err error) {
	// 创建请求对象
	req := larkbitable.NewUpdateAppTableRecordReqBuilder().
		AppToken(FeishuDocAppToken).
		TableId(FeishuDocTableId).
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
