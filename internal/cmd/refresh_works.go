package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	libraryUtils "github.com/armylong/go-library/utils"
	"github.com/urfave/cli/v2"
)

// 轮询刷新工作目录，等待符合条件的任务出现
func RefreshWorksHandler(c *cli.Context) error {

	workSpace := c.String("work_space")
	if workSpace == "" {
		fmt.Println("错误: 请指定工作空间")
		return nil
	}
	hasFileNamesStr := c.String("has_file_names")
	noHasFileNamesStr := c.String("no_has_file_names")

	var hasFileNames []string
	if hasFileNamesStr != "" {
		hasFileNames = strings.Split(hasFileNamesStr, ",")
	}
	var noHasFileNames []string
	if noHasFileNamesStr != "" {
		noHasFileNames = strings.Split(noHasFileNamesStr, ",")
	}

	sleepTime := 5 * time.Second

	for {
		entries, err := os.ReadDir(workSpace)
		if err != nil {
			fmt.Printf("工作目录不存在 %s: %v\n", workSpace, err)
			return nil
		}

		hasEmptyWork := false
		for _, entry := range entries {
			subdirPath := filepath.Join(workSpace, entry.Name())
			subEntries, err := os.ReadDir(subdirPath)
			if err != nil {
				fmt.Printf("读取子目录 %s 失败: %v\n", subdirPath, err)
				return nil
			}
			var subEntryHasList []string
			var subEntryNoHasList []string
			for _, subEntry := range subEntries {
				if libraryUtils.InArray(subEntry.Name(), hasFileNames) {
					subEntryHasList = append(subEntryHasList, subEntry.Name())
					continue
				}
				if libraryUtils.InArray(subEntry.Name(), noHasFileNames) {
					subEntryNoHasList = append(subEntryNoHasList, subEntry.Name())
					continue
				}
			}
			if len(subEntryHasList) == len(hasFileNames) && len(subEntryNoHasList) != len(noHasFileNames) {
				hasEmptyWork = true
				break
			}
		}
		if hasEmptyWork {
			fmt.Printf("[%s] 任务刷新成功\n", time.Now().Format("2006-01-02 15:04:05"))
			break
		} else {
			time.Sleep(sleepTime)
		}
	}
	return nil
}
