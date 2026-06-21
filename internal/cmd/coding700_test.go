package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/urfave/cli/v2"
)

func TestCoding700NewHandler(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("repo_url", "https://github.com/gogs/gogs", "仓库URL")
	cliCtx := cli.NewContext(nil, flagSet, nil)
	err := Coding700NewHandler(cliCtx)
	fmt.Println(err)
}

func TestCoding700InitHandler(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	// flagSet.String("base", "base", "基础项目名称")
	flagSet.String("project", "0524_2", "项目名称")
	// flagSet.String("image", "repo", "镜像名称")
	// flagSet.Bool("git_init", false, "是否初始化git")
	cliCtx := cli.NewContext(nil, flagSet, nil)
	Coding700InitHandler(cliCtx)
}

func TestCoding700CollectTaskResultsHandler(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("project", "project_0528_4", "项目名称")
	flagSet.String("prompt_id", "1", "题目ID")
	flagSet.Int("rollout_id", 1, "子任务ID")
	cliCtx := cli.NewContext(nil, flagSet, nil)
	err := Coding700CollectTaskResultsHandler(cliCtx)
	fmt.Println(err)
}

func TestCoding700PromptInitHandler(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("project", "project_0527_2", "项目名称")
	flagSet.String("prompt_id", "1,2,3,4,5,6,7", "题目ID")
	// flagSet.Int("rollout_id", 8, "子任务ID")
	cliCtx := cli.NewContext(nil, flagSet, nil)
	Coding700ResetHandler(cliCtx)
}

func TestCoding700ResetHandler(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("project", "project_0527_2", "项目名称")
	flagSet.String("prompt_id", "1,2,3,4,5,6,7", "题目ID")
	// flagSet.Int("rollout_id", 8, "子任务ID")
	cliCtx := cli.NewContext(nil, flagSet, nil)
	Coding700ResetHandler(cliCtx)
}

func TestGetRepoRecord(t *testing.T) {
	cliCtx := &cli.Context{
		Context: context.Background(),
	}
	ret, err := Coding700GetRepoRecord(cliCtx.Context, "1", `https://github.com/armylong/20260516-01`)

	fmt.Println(larkcore.Prettify(ret))
	fmt.Println(err)
}

func TestGetPromptRecord(t *testing.T) {
	cliCtx := &cli.Context{
		Context: context.Background(),
	}
	ret, err := Coding700GetPromptRecord(cliCtx.Context, `recvjQynRX9hqm`, 2)
	fmt.Println(err)
	fmt.Println(larkcore.Prettify(ret))
}

func TestGetRolloutRecord(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cliCtx := &cli.Context{
		Context: ctx,
	}
	ret, err := Coding700GetRolloutRecord(cliCtx.Context, `recvjSJmvzjwn1`, 1)
	fmt.Println(err)
	fmt.Println(larkcore.Prettify(ret))
}

func TestFormatJsonFile(t *testing.T) {
	cliCtx := &cli.Context{
		Context: context.Background(),
	}
	ret, err := Coding700FormatJsonFile(cliCtx.Context, `project_0527_6`, true, false, []string{"repo_url"})
	fmt.Println(err)
	fmt.Println(larkcore.Prettify(ret))
}

func TestCoding700UploadTaskDataHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("project", "project_0609_1", "项目名称")
	// flagSet.Int("repo_uid", 13979, "仓库UID")
	flagSet.Bool("prompt_yn", true, "是否上传题目")
	flagSet.Bool("rollout_yn", false, "是否上传子任务")
	flagSet.String("repo_ignore_fields", "repo_url", "忽略的字段")
	// flagSet.String("repo_ignore_fields", "repo_url,language", "忽略的字段")

	cliCtx := cli.NewContext(nil, flagSet, nil)
	cliCtx.Context = ctx
	err := Coding700UploadTaskDataHandler(cliCtx)
	fmt.Println(err)
}

func TestTmp(t *testing.T) {
	s := "projectName"
	fmt.Println(s)
	sList := strings.Split(s, "/")
	fmt.Println(sList)
	s1 := sList[len(sList)-1]
	fmt.Println(s1)
}
