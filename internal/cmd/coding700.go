package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	feishuCloudDocBusiness "github.com/armylong/armylong-go/internal/business/feishu/cloud_doc"
	feishuDriveBusiness "github.com/armylong/armylong-go/internal/business/feishu/cloud_file"
	libraryUtils "github.com/armylong/go-library/utils"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/spf13/cast"
	"github.com/urfave/cli/v2"
)

var workHome = ""

const (
	feishuDocAppToken = `ULuObnT5cajntxsxZiGc4YyOnDh`
	// feishuDocTableId  = `tblhLmLID4ZMBpDO` // 表1
	// feishuDocTableId = `tblTuI7jCdQopuL6` // 表2
	// feishuDocTableId = `tblKFV8xUJwiM2oJ` // 表3
	feishuDocTableId = `tblyZROCWHW0qaKI` // 表4
	// feishuDocTableId = `tblnwk23hoP1STWZ` // 表5-私库
	feishuDocViewId = `vewxWP7trZ`
)

var promptIds = []int{1, 2, 3, 4, 5, 6, 7}
var rolloutIds = []int{1, 2, 3, 4, 5}

// 第一份作业永远写死: Doubao-Seed-2.0-Code
// 第二份作业永远写死: GPT5.4
// 第三份作业永远写死: Gemini 3.1 pro
// 第四份作业永远写死: DeepSeek-v4
// 第五份作业不一样, 所有项目所有题目的第五份作业中按顺序轮流使用这三个模型: MinMax-M2.7 / GLM-5.1 / Qwen3.6-Plus
var rolloutModelName1_4 = []string{"Doubao-Seed-2.0-Code", "GPT5.4", "Gemini 3.1 pro", "DeepSeek-v4"}
var rolloutModelName5 = []string{"GLM-5.1", "Qwen3.6-Plus"}

var scoreRulesContent = "# 评分规则\n\n***每一轮的评分都要严格按照题目工作流的步骤执行: `%s/题目工作流.md` (经常会更新,每次评分都要重新查看)***"

var defaultRepoType = "私有仓库"

var repoTypes = []string{"公有仓库", "私有仓库"}

var registryAddr = "192.168.1.14:5000"

// 主力机器ssh别名
var masterMachineSshAddr = "mac24"

// 主力机器workHome路径
var masterWorkHome = "/Users/zhangzelong/works/coding700"

type container struct {
	Name      string
	ImageName string
	SshPort   int
	HttpPort  int
	PromptId  int
	Host      string
}

var Containers []container = []container{
	{
		Name:     "",
		SshPort:  2221,
		HttpPort: 81,
		PromptId: 1,
		Host:     `ssh://ubt`,
	},
	{
		Name:     "",
		SshPort:  2222,
		HttpPort: 82,
		PromptId: 2,
		Host:     `ssh://ubt`,
	},
	{
		Name:     "",
		SshPort:  2223,
		HttpPort: 83,
		PromptId: 3,
		Host:     `ssh://ubt`,
	},
	{
		Name:     "",
		SshPort:  2224,
		HttpPort: 84,
		PromptId: 4,
		Host:     `ssh://ubt2`,
	},
	{
		Name:     "",
		SshPort:  2225,
		HttpPort: 85,
		PromptId: 5,
		Host:     `ssh://ubt2`,
	},
	{
		Name:     "",
		SshPort:  2226,
		HttpPort: 86,
		PromptId: 6,
		Host:     `ssh://ubt2`,
	},
	{
		Name:     "",
		SshPort:  2227,
		HttpPort: 87,
		PromptId: 7,
		Host:     `ssh://ubt2`,
	},
}

func init() {
	workHome = filepath.Join(os.Getenv("HOME"), "works", "coding700")
	scoreRulesContent = fmt.Sprintf(scoreRulesContent, workHome)
}

func Coding700Handler(c *cli.Context) error {
	return nil
}

func Coding700NewHandler(c *cli.Context) error {
	repoUrl := c.String(`repo_url`)
	if repoUrl != "" {
		fmt.Printf("repoUrl: %s\n", repoUrl)
	}

	// 新建目录, 命名格式: project_0516_1(project_月日_序号, 年月都有前补零, 序号从1开始, 检测目录里已存在则序号+1)
	prefix := fmt.Sprintf("project_%s_", time.Now().Format("0102"))

	seq := 1
	entries, err := os.ReadDir(workHome)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	// 顺便计算上一份题目的最后一份作业的模型名称
	lastRollout5ModelName := ""
	for _, entry := range entries {
		name := entry.Name()
		// 不是目录或目录不是以 prefix 开头的,跳过
		if !entry.IsDir() || !strings.HasPrefix(name, prefix) {
			continue
		}
		parts := strings.SplitN(name, prefix, 2)
		if len(parts) == 2 {
			if n, err := strconv.Atoi(parts[1]); err == nil {
				if n+1 > seq {
					seq = n + 1
				}
			}
		}

		// 读取最后一个题目的最后一份作业里的rollout.json的model_name
		rolloutJsonPath := filepath.Join(workHome, name, fmt.Sprintf("prompt_%d", promptIds[len(promptIds)-1]), fmt.Sprintf("rollout_%d", rolloutIds[len(rolloutIds)-1]), "rollout.json")
		if _, err := os.Stat(rolloutJsonPath); os.IsNotExist(err) {
			return fmt.Errorf("上一个项目最后一题的rollout.json文件不存在: %s", rolloutJsonPath)
		}
		rolloutJsonData := Coding700RolloutData{}
		jsonData, err := os.ReadFile(rolloutJsonPath)
		if err != nil {
			return fmt.Errorf("读取rollout.json文件失败: %s %v", rolloutJsonPath, err)
		}
		if err := json.Unmarshal(jsonData, &rolloutJsonData); err != nil {
			return fmt.Errorf("解析rollout.json文件数据失败: %s, %v", rolloutJsonPath, err)
		}
		lastRollout5ModelName = rolloutJsonData.ModelName
	}
	newProjectName := fmt.Sprintf("%s%d", prefix, seq)

	// 新目录已存在就报错
	if _, err := os.Stat(filepath.Join(workHome, newProjectName)); err == nil {
		return fmt.Errorf("项目目录已存在: %s", filepath.Join(workHome, newProjectName))
	}

	if err := os.MkdirAll(filepath.Join(workHome, newProjectName), 0755); err != nil {
		return fmt.Errorf("新建目录失败: %v", err)
	}
	fmt.Printf("新建项目目录成功: %s\n", filepath.Join(workHome, newProjectName))

	// 拉取仓库
	if repoUrl != "" && repoUrl != "-" {
		// 如果不包含: https://github.com/, 则添加
		repoPrefix := "https://github.com/"
		if !strings.HasPrefix(repoUrl, repoPrefix) {
			return fmt.Errorf("repoURL必须以https://github.com/开头")
		}
		repoParts := strings.SplitN(repoUrl, repoPrefix, 2)
		if len(repoParts) != 2 {
			return fmt.Errorf("repoURL格式错误1")
		}
		cloneParts := strings.SplitN(repoParts[1], `/`, 2)
		if len(cloneParts) != 2 {
			return fmt.Errorf("repoURL格式错误2")
		}
		cloneUrl := fmt.Sprintf("git@github.com:%s/%s.git", cloneParts[0], cloneParts[1])
		cmd := exec.Command("sh", "-c", fmt.Sprintf(
			"git clone %s %s",
			cloneUrl, filepath.Join(workHome, newProjectName, "repo"),
		))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("拉取仓库失败: %v", err)
		}
		fmt.Printf("拉取仓库成功: %s\n", repoUrl)
	}

	// 新建repo.json文件 使用Coding700RepoData结构体 美化输出
	repoData := Coding700RepoData{
		RepoType:  defaultRepoType,
		Language:  "",
		RepoURL:   repoUrl,
		TaskCount: "7",
	}
	jsonData, err := json.MarshalIndent(repoData, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化repoData失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workHome, newProjectName, "repo.json"), jsonData, 0644); err != nil {
		return fmt.Errorf("写入repo.json文件失败: %v", err)
	}
	fmt.Printf("repo.json文件写入成功: %s\n", filepath.Join(workHome, newProjectName, "repo.json"))

	// 新建Dockerfile和Dockerfile.dev空文件
	if err := os.WriteFile(filepath.Join(workHome, newProjectName, "Dockerfile"), []byte{}, 0644); err != nil {
		return fmt.Errorf("写入Dockerfile文件失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workHome, newProjectName, "Dockerfile.dev"), []byte{}, 0644); err != nil {
		return fmt.Errorf("写入Dockerfile.dev文件失败: %v", err)
	}
	fmt.Printf("Dockerfile和Dockerfile.dev文件写入成功: %s\n", filepath.Join(workHome, newProjectName))

	// 读取~/.ssh/id_rsa.pub内容, 并写入到authorized_keys文件
	idRsaPub := []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC2myBaKw3O+ItKIPGwrdIg1ZteXHMo6pKISG/GzbJCI6AvHY2H18qZDxd2dIrrf63ffqCa0X33uK/AvMuHi9KutgjGWhLaCEW9yEk8bBYaCQsVc7oFK+RajuirqI1OKslxhecrVG+qRmuM0Ua1ylJMWmnw5WQSnbrqDdVSJRzzTghvnmbvb9sSYCp2C2vQv+lI0aXjB+nFDDDdIQamsf7jHn7R/H5Ic7TWIUxxJ7Id947k1VvhpTrshwIlqh7VMJEWmC8t7PF9745FauhvDhBnxwhLQ6sa49gcm/M4MzM6ShEOvj19IE/l3hxbQHY6PekLFoQyW6tTVnofcHigj4aqKwKsT40hMuTVqIwe+cZRRkb0JfSiUEvftgUgbtq386YpGwbq/ZPRS41cdpyEBpY26RY2Ric3FEdBDb6FzamvWEJPtI8EtHpoy7dZ/xU7vxElUjYtOLouAGlh/ywthkftIb2POGTSaDQf0aZxKfYMPxjAkmtJXu2DFKYWoIaZp9r1N0kJrfT+RwM7jyUCzwDeKI8VPV1MaFE08SyvfJn43xKsAdIT4sxBZDeMi6TEDrFYLTM2P6hzihFBMysfBh2jshv8nYfXY0q6+0pcgoJyD5zAorkJG+3DkZO2hfYxmuctrnirneiDJFKayUQ/7n4lEKxyGUTFpp5wHs7+wEtlyw== armylong@163.com`)
	if err := os.WriteFile(filepath.Join(workHome, newProjectName, "authorized_keys"), idRsaPub, 0644); err != nil {
		return fmt.Errorf("写入authorized_keys文件失败: %v", err)
	}
	fmt.Printf("authorized_keys文件写入成功: %s\n", filepath.Join(workHome, newProjectName, "authorized_keys"))

	// 构建作业目录
	for _, promptId := range promptIds {
		promptDir := filepath.Join(workHome, newProjectName, fmt.Sprintf("prompt_%d", promptId))
		if err := os.MkdirAll(promptDir, 0755); err != nil {
			return fmt.Errorf("新建目录失败: %v, promptDir: %s", err, promptDir)
		}
		fmt.Printf("新建目录成功: %s\n", promptDir)

		// 新建prompt.md空文件
		if err := os.WriteFile(filepath.Join(promptDir, "prompt.md"), []byte{}, 0644); err != nil {
			return fmt.Errorf("写入prompt.md文件失败: %v, promptDir: %s", err, promptDir)
		}
		fmt.Printf("prompt.md文件写入成功: %s\n", filepath.Join(promptDir, "prompt.md"))

		// 新建prompt.json文件, 内容是Coding700PromptData结构体
		promptData := Coding700PromptData{
			PromptId:   cast.ToString(promptId),
			Prompt:     "",
			Difficulty: "",
			Category:   "",
			TechStack:  "",
			ModuleTags: "",
		}
		jsonData, err := json.MarshalIndent(promptData, "", "    ")
		if err != nil {
			return fmt.Errorf("序列化promptData失败: %v", err)
		}
		if err := os.WriteFile(filepath.Join(promptDir, "prompt.json"), jsonData, 0644); err != nil {
			return fmt.Errorf("写入prompt.json文件失败: %v, promptDir: %s", err, promptDir)
		}
		fmt.Printf("prompt.json文件写入成功: %s\n", filepath.Join(promptDir, "prompt.json"))

		// 新建.trae/rules/目录, 并新建.trae/rules/project_rules.md文件, 内容写入"项目规则"
		if err := os.MkdirAll(filepath.Join(promptDir, ".trae/rules"), 0755); err != nil {
			return fmt.Errorf("新建目录失败: %v, promptDir: %s", err, promptDir)
		}
		if err := os.WriteFile(filepath.Join(promptDir, ".trae/rules/project_rules.md"), []byte(scoreRulesContent), 0644); err != nil {
			return fmt.Errorf("写入.trae/rules/project_rules.md文件失败: %v, promptDir: %s", err, promptDir)
		}
		fmt.Printf(".trae/rules/project_rules.md文件写入成功: %s\n", filepath.Join(promptDir, ".trae/rules/project_rules.md"))

		for _rollout_i, rolloutId := range rolloutIds {
			rolloutDir := filepath.Join(promptDir, fmt.Sprintf("rollout_%d", rolloutId))
			if err := os.MkdirAll(rolloutDir, 0755); err != nil {
				return fmt.Errorf("新建目录失败: %v, promptDir: %s", err, promptDir)
			}
			fmt.Printf("新建轮次作业目录成功: %s\n", rolloutDir)

			_modelName := ""
			if (_rollout_i + 1) < 5 {
				_modelName = rolloutModelName1_4[_rollout_i]
			} else {
				if lastRollout5ModelName == "" {
					lastRollout5ModelName = rolloutModelName5[0]
				} else {
					for _i, modelName := range rolloutModelName5 {
						if modelName == lastRollout5ModelName {
							if _i == len(rolloutModelName5)-1 {
								// 如果匹配到是最后一个轮询模型, 则使用第一个轮询模型
								lastRollout5ModelName = rolloutModelName5[0]
							} else {
								// 如果匹配到不是最后一个轮询模型, 则使用下一个轮询模型
								lastRollout5ModelName = rolloutModelName5[_i+1]
							}
							break
						}
					}
				}
				_modelName = lastRollout5ModelName
			}
			// 写入rollout.json文件, 内容是Coding700RolloutData结构体
			rolloutData := Coding700RolloutData{
				RolloutID:   cast.ToString(rolloutId),
				SessionID:   "",
				ModelName:   _modelName,
				Score:       "",
				ScoreReason: "",
			}
			rolloutJsonData, err := json.MarshalIndent(rolloutData, "", "    ")
			if err != nil {
				return fmt.Errorf("序列化rolloutData失败: %v", err)
			}
			if err := os.WriteFile(filepath.Join(rolloutDir, "rollout.json"), rolloutJsonData, 0644); err != nil {
				return fmt.Errorf("写入rollout.json文件失败: %v, rolloutDir: %s", err, rolloutDir)
			}
			fmt.Printf("rollout.json文件写入成功: %s\n", filepath.Join(rolloutDir, "rollout.json"))
		}
	}

	return nil
}

// 初始化项目
func Coding700InitHandler(c *cli.Context) error {
	projectName := c.String(`project`)
	gitInitYn := c.Bool(`git_init`)

	if projectName == "" {
		return fmt.Errorf("项目名称不能为空")
	}

	imageName := projectName

	// 默认都在这个目录进行操作 workHome/[projectName]/repo
	projectDir := filepath.Join(workHome, projectName)
	repoDir := filepath.Join(projectDir, "repo")
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("仓库目录不存在: %s", repoDir)
	}

	remoteRepoUrl := ""
	if gitInitYn {
		// 删除远程仓库 gh repo delete armylong/longt --yes
		cmd := exec.Command("sh", "-c", fmt.Sprintf(
			"gh api repos/armylong/%s",
			projectName,
		))
		output, err := cmd.CombinedOutput()
		if err == nil {
			cmd = exec.Command("sh", "-c", fmt.Sprintf(
				"gh repo delete armylong/%s --yes",
				projectName,
			))
			cmd.Dir = repoDir
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("gh repo delete 失败: %v, output: %s", err, string(output))
			}
			fmt.Printf("gh repo delete 成功: armylong/%s\n", projectName)
		} else {
			if strings.Contains(string(output), "404") {
				fmt.Printf("远程仓库不存在: armylong/%s\n", projectName)
			} else {
				return fmt.Errorf("gh api 查询仓库失败: %v, output: %s", err, string(output))
			}
		}

		// 删除.git目录
		gitDir := filepath.Join(repoDir, ".git")
		os.RemoveAll(gitDir)
		fmt.Printf("git目录删除成功: %s\n", gitDir)

		// 初始化git仓库: git init
		cmd = exec.Command("sh", "-c", "git init")
		cmd.Dir = repoDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git init 失败: %v, output: %s", err, string(output))
		}
		fmt.Printf("git init 成功: %s\n", repoDir)

		// 添加所有文件到git仓库: git add .
		cmd = exec.Command("sh", "-c", "git add .")
		cmd.Dir = repoDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git add 失败: %v, output: %s", err, string(output))
		}
		fmt.Printf("git add 成功: %s\n", repoDir)

		// 提交所有文件到git仓库: git commit -m "first commit"
		cmd = exec.Command("sh", "-c", `git commit -m "first commit"`)
		cmd.Dir = repoDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git commit 失败: %v, output: %s", err, string(output))
		}
		fmt.Printf("git commit 成功: %s\n", repoDir)

		// 推送本地仓库到github: gh repo create [projectName] --public --source=. --push
		cmd = exec.Command("sh", "-c", fmt.Sprintf(
			"gh repo create %s --public --source=. --push",
			projectName,
		))
		cmd.Dir = repoDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("gh repo create 失败: %v, output: %s", err, string(output))
		}
		fmt.Printf("gh repo create 成功: %s\n", repoDir)

		// 获取远程仓库url gh repo view armylong/05161 --json url --jq .url
		cmd = exec.Command("sh", "-c", fmt.Sprintf(
			"gh repo view armylong/%s --json url --jq .url",
			projectName,
		))
		if output, err := cmd.CombinedOutput(); err == nil {
			remoteRepoUrl = strings.TrimSpace(string(output))
			fmt.Printf("远程仓库url: %s\n", remoteRepoUrl)
		} else {
			return fmt.Errorf("gh repo view 获取远程仓库地址失败: %v, output: %s", err, string(output))
		}
	}

	// 将repo目录压缩zip（解压后直接是代码文件，不含repo目录层）
	zipFile := filepath.Join(projectDir, "repo.zip")
	cmd := exec.Command("sh", "-c", fmt.Sprintf(
		"zip -r %s .",
		zipFile,
	))
	cmd.Dir = repoDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("压缩repo目录失败: %v, output: %s", err, string(output))
	}
	fmt.Printf("repo.zip压缩成功: %s\n", zipFile)

	// 删除所有没有tag且最近未使用的镜像 docker image prune -f
	exec.Command("sh", "-c", "docker image prune -f").Run()
	fmt.Printf("所有无用镜像清理成功\n")

	// 删除旧镜像
	exec.Command("sh", "-c", fmt.Sprintf(
		"docker rmi -f %s",
		imageName,
	)).Run()
	fmt.Printf("旧镜像删除成功: %s\n", imageName)

	// 构建新镜像
	if err := exec.Command("sh", "-c", fmt.Sprintf(
		"docker image inspect %s",
		imageName,
	)).Run(); err != nil {
		cmd := exec.Command("sh", "-c", fmt.Sprintf(
			"docker build -t %s -f Dockerfile.dev .",
			imageName,
		))
		cmd.Dir = projectDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("docker build 失败: %v", err)
		}
		fmt.Printf("镜像构建成功: %s\n", imageName)
	} else {
		fmt.Printf("镜像已存在, 跳过构建: %s\n", imageName)
	}

	// 推送至docker私有仓库
	registryImage := fmt.Sprintf("%s/%s", registryAddr, imageName)
	if output, err := exec.Command("sh", "-c", fmt.Sprintf(
		"docker tag %s %s", imageName, registryImage,
	)).CombinedOutput(); err != nil {
		return fmt.Errorf("推送至docker私有仓库 docker tag 失败: %v, output: %s", err, string(output))
	}
	fmt.Printf("推送至docker私有仓库 docker tag 成功: %s -> %s\n", imageName, registryImage)
	if output, err := exec.Command("sh", "-c", fmt.Sprintf(
		"docker push %s", registryImage,
	)).CombinedOutput(); err != nil {
		return fmt.Errorf("推送至docker私有仓库 docker push 失败: %v, output: %s", err, string(output))
	}
	fmt.Printf("推送至docker私有仓库 docker push 成功: %s\n", registryImage)

	// 负载均衡机器拉取镜像
	sshAddrList := getContainersSshAddr()
	if len(sshAddrList) == 0 {
		return fmt.Errorf("没有容器可拉取镜像")
	}
	for _, sshAddr := range sshAddrList {
		sshHost := strings.TrimPrefix(sshAddr, "ssh://")
		if output, err := exec.Command("sh", "-c", fmt.Sprintf(
			"ssh %s 'docker pull %s'", sshHost, registryImage,
		)).CombinedOutput(); err != nil {
			return fmt.Errorf("负载均衡机器拉取镜像 ssh %s docker pull 失败: %v, output: %s", sshHost, err, string(output))
		}
		fmt.Printf("负载均衡机器拉取镜像 ssh %s docker pull 成功: %s\n", sshHost, registryImage)
	}

	return nil
}

// 获取容器的所有ssh地址
func getContainersSshAddr() (sshAddrList []string) {
	for _, container := range Containers {
		if container.Host != "" && !libraryUtils.InArray(container.Host, sshAddrList) {
			sshAddrList = append(sshAddrList, container.Host)
		}
	}
	return sshAddrList
}

// 重置题目ID对应的容器
func Coding700ResetHandler(c *cli.Context) error {
	projectName := c.String(`project`)
	promptIdStr := c.String(`prompt_id`)
	if projectName == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	if promptIdStr == "" {
		return fmt.Errorf("prompt_id is required")
	}
	promptIdStrList := strings.Split(promptIdStr, ",")

	// 重置题目环境前是否等待指定的rollout.json文件处理完成
	if !c.IsSet(`rollout_id`) {
		return fmt.Errorf("命令执行失败, 缺少作业id参数: rollout_id")
	}
	rolloutId := c.Int(`rollout_id`)
	if rolloutId > 0 {
		if len(promptIdStrList) > 1 {
			return fmt.Errorf("子任务检查只能用于单个题目环境")
		}
		promptId := cast.ToInt(promptIdStrList[0])
		for {
			time.Sleep(5 * time.Second)

			rolloutData, err := getRolloutJsonData(projectName, promptId, rolloutId)
			if err != nil {
				fmt.Printf("获取rollout.json文件失败: %v, 已经进入人工处理环节, 为了防止冲突, 请不要做任何操作, 耐心等待即可...\n", err)
				continue
			}

			if err := checkRolloutFields(rolloutData); err != nil {
				fmt.Printf("rollout.json文件字段校验失败: %v, 已经进入人工处理环节, 为了防止冲突, 请不要做任何操作, 耐心等待即可...\n", err)
				continue
			}

			break
		}
		// 删除rollout目录下的app目录
		rolloutAppDir := filepath.Join(workHome, projectName, fmt.Sprintf("prompt_%d", promptId), fmt.Sprintf("rollout_%d", rolloutId), "app")
		if err := os.RemoveAll(rolloutAppDir); err != nil {
			return fmt.Errorf("删除rollout目录下的app目录失败: %v", err)
		}
		fmt.Printf("rollout目录下的app目录删除成功: %s\n", rolloutAppDir)
	}

	for _, promptIdStr := range promptIdStrList {
		promptId := cast.ToInt(promptIdStr)
		fmt.Printf("重置题目 %d 的环境\n", promptId)
		// 重启置容器
		container, err := getContainerByPromptId(projectName, promptId)
		if err != nil {
			return err
		}
		if container == nil {
			return fmt.Errorf("prompt_id %d 未找到对应的容器", promptId)
		}

		if rolloutId == rolloutIds[len(rolloutIds)-1] {
			if err := deleteContainer(container); err != nil {
				return err
			}
			fmt.Printf("最后一轮: 容器 %s 删除成功\n", container.Name)
		} else {
			if err := resetContainer(container); err != nil {
				return err
			}
			fmt.Printf("容器 %s 重置成功启动 (http=%d, ssh=%d)\n", container.Name, container.HttpPort, container.SshPort)
		}
	}

	return nil
}

// 重启容器
func resetContainer(container *container) error {
	if err := deleteContainer(container); err != nil {
		return err
	} else {
		fmt.Printf("容器 %s 删除成功\n", container.Name)
	}
	if err := createContainer(container); err != nil {
		return err
	} else {
		fmt.Printf("容器 %s 启动成功\n", container.Name)
	}
	return nil
}

func deleteContainer(container *container) error {
	// 删除旧容器
	cmd := exec.Command("sh", "-c", fmt.Sprintf(
		"docker rm -f %s",
		container.Name,
	))
	if container.Host != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", container.Host))
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("删除容器 %s 失败: %v", container.Name, err)
	}
	return nil
}

// 创建容器
func createContainer(container *container) error {
	// if output, err := exec.Command("sh", "-c", fmt.Sprintf(
	// 	"docker run -d -p %d:%d -p %d:22 -v %s/.trae-cn-server:/root/.trae-cn-server --name %s %s",
	// 	httpPort, repoPort, sshPort, workHome, containerName, imageName,
	// )).CombinedOutput(); err != nil {
	// 	return fmt.Errorf("docker run %s 失败: %v, output: %s", containerName, err, string(output))
	// }
	runCmd := exec.Command("sh", "-c", fmt.Sprintf(
		"docker run -d -P -p %d:22 --name %s %s",
		container.SshPort, container.Name, container.ImageName,
	))
	if container.Host != "" {
		runCmd.Env = append(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", container.Host))
	}
	if err := runCmd.Run(); err != nil {
		return fmt.Errorf("docker run %s 失败: %v", container.Name, err)
	}

	time.Sleep(1 * time.Second)

	checkCmd := exec.Command("sh", "-c", fmt.Sprintf(
		"docker inspect -f '{{.State.Running}}' %s", container.Name,
	))
	if container.Host != "" {
		checkCmd.Env = append(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", container.Host))
	}
	output, _ := checkCmd.Output()
	if strings.TrimSpace(string(output)) != "true" {
		logsCmd := exec.Command("sh", "-c", fmt.Sprintf("docker logs %s", container.Name))
		if container.Host != "" {
			logsCmd.Env = append(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", container.Host))
		}
		logs, _ := logsCmd.CombinedOutput()
		return fmt.Errorf("容器 %s 启动后立即退出, 日志: %s", container.Name, string(logs))
	}

	envCmd := exec.Command("sh", "-c", fmt.Sprintf(
		"docker exec %s sh -c 'export -p > /etc/profile.d/docker-env.sh'", container.Name,
	))
	if container.Host != "" {
		envCmd.Env = append(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", container.Host))
	}
	// 将容器内的环境变量写入/etc/profile.d/docker-env.sh, 使SSH登录后也能读取
	if err := envCmd.Run(); err != nil {
		return fmt.Errorf("导出环境变量到 %s 失败: %v", container.Name, err)
	}
	return nil
}

// 收集任务成果 collectTaskResults
func Coding700CollectTaskResultsHandler(c *cli.Context) error {

	projectName := c.String(`project`)
	if projectName == "" {
		return fmt.Errorf("projectName is required")
	}

	promptId := c.Int(`prompt_id`)
	if promptId == 0 {
		return fmt.Errorf("请输入题目ID")
	}

	rolloutId := c.Int(`rollout_id`)
	if rolloutId == 0 {
		return fmt.Errorf("请输入子任务ID")
	}

	// 题目对应的容器
	container, err := getContainerByPromptId(projectName, promptId)
	if err != nil {
		return err
	}
	if container == nil {
		return fmt.Errorf("prompt_id %d 未找到对应的容器", promptId)
	}

	projectDir := filepath.Join(workHome, projectName)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("项目目录不存在: %s", projectDir)
	}
	// 不存在就创建提示词目录
	promptDir := filepath.Join(projectDir, fmt.Sprintf("prompt_%d", promptId))
	if _, err := os.Stat(promptDir); os.IsNotExist(err) {
		if err := os.MkdirAll(promptDir, 0755); err != nil {
			return fmt.Errorf("创建提示词目录 %s 失败: %v", promptDir, err)
		}
	}

	// 不存在就创建子任务目录
	rolloutDir := filepath.Join(promptDir, fmt.Sprintf("rollout_%d", rolloutId))
	if _, err := os.Stat(rolloutDir); os.IsNotExist(err) {
		if err := os.MkdirAll(rolloutDir, 0755); err != nil {
			return fmt.Errorf("创建子任务目录 %s 失败: %v", rolloutDir, err)
		}
	}

	// 拷贝diff文件 (prompt_id - 1) * 5 + rollout_id.patch
	diffFileName := getDiffFileName(promptId, rolloutId)
	patchFilePath := filepath.Join(rolloutDir, diffFileName)
	diffCmd := exec.Command("sh", "-c", fmt.Sprintf(
		"docker exec %s sh -c 'cd /app && git add -A && git diff --cached'",
		container.Name,
	))
	if container.Host != "" {
		diffCmd.Env = append(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", container.Host))
	}
	output, err := diffCmd.Output()
	if err != nil {
		return fmt.Errorf("diff %s 生成 失败: %v", diffFileName, err)
	}
	if err := os.WriteFile(patchFilePath, output, 0644); err != nil {
		return fmt.Errorf("写入%s文件失败: %v", diffFileName, err)
	}
	fmt.Printf("diff %s 生成 成功: %s\n", diffFileName, patchFilePath)

	// 拷贝/app目录
	cpCmd := exec.Command("sh", "-c", fmt.Sprintf(
		"docker cp %s:/app %s",
		container.Name, rolloutDir,
	))
	if container.Host != "" {
		cpCmd.Env = append(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", container.Host))
	}
	if err := cpCmd.Run(); err != nil {
		return fmt.Errorf("拷贝容器 %s 的/app目录到 %s 失败: %v", container.Name, rolloutDir, err)
	}
	fmt.Printf("app目录拷贝 成功: %s\n", rolloutDir)

	return nil
}

func getRolloutJsonData(projectName string, promptId, rolloutId int) (*Coding700RolloutData, error) {
	// 解析rollout.json文件为: Coding700RolloutData这个结构体
	rolloutJsonPath := filepath.Join(workHome, projectName, fmt.Sprintf("prompt_%d", promptId), fmt.Sprintf("rollout_%d", rolloutId), "rollout.json")
	if _, err := os.Stat(rolloutJsonPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("子任务目录里没有rollout.json文件: %s", rolloutJsonPath)
	}
	rolloutJsonData, err := os.ReadFile(rolloutJsonPath)
	if err != nil {
		return nil, fmt.Errorf("读取工作文件 %s 失败: %v", rolloutJsonPath, err)
	}
	var rolloutData Coding700RolloutData
	if err := json.Unmarshal(rolloutJsonData, &rolloutData); err != nil {
		return nil, fmt.Errorf("解析工作文件 %s 失败: %v", rolloutJsonPath, err)
	}
	return &rolloutData, nil
}

func getDiffFileName(promptId, rolloutId int) string {
	return fmt.Sprintf("%d.patch", (promptId-1)*5+rolloutId)
}

func getContainerByPromptId(projectName string, promptId int) (*container, error) {
	var container *container
	for _, c := range Containers {
		if c.PromptId == promptId {
			container = &c
			break
		}
	}
	// 根据容器名称获取容器的镜像名称
	if container != nil {
		container.Name = fmt.Sprintf("%s___%d", projectName, promptId)
		if container.Host != "" {
			container.ImageName = fmt.Sprintf("%s/%s", registryAddr, projectName)
		} else {
			container.ImageName = projectName
		}
	}

	return container, nil
}

type Coding700FeishuData struct {
	RecordID             string `json:"record_id,omitempty"`
	Coding700RepoData    *Coding700RepoData
	Coding700PromptData  *Coding700PromptData
	Coding700RolloutData *Coding700RolloutData
}

type Coding700RepoData struct {
	RepoType         string                   `json:"repo_type"`
	Language         string                   `json:"language"`
	RepoURL          string                   `json:"repo_url,omitempty"`
	TaskCount        string                   `json:"task_count"`
	EnvironmentNotes string                   `json:"environment_notes"`
	Repo             []Coding700FeishuFileObj `json:"repo,omitempty"`
	Dockerfile       []Coding700FeishuFileObj `json:"dockerfile,omitempty"`
	Uid              string                   `json:"uid,omitempty"`
}

type Coding700PromptData struct {
	PromptId     string   `json:"prompt_index"`
	Prompt       string   `json:"prompt,omitempty"`
	Difficulty   string   `json:"difficulty"`  // 难度
	Category     string   `json:"category"`    // 分类
	TechStack    string   `json:"tech_stack"`  // 技术栈
	ModuleTags   string   `json:"module_tags"` // 涉及模块
	ParentRecord []string `json:"父记录,omitempty"`
}

type Coding700RolloutData struct {
	RolloutID    string                   `json:"rollout_id"`
	SessionID    string                   `json:"session_id"`
	ModelName    string                   `json:"model_name"` // 模型名称
	Score        string                   `json:"score"`
	ScoreReason  string                   `json:"score_reason"`
	GitDiff      []Coding700FeishuFileObj `json:"git_diff,omitempty"`
	Notes        string                   `json:"notes,omitempty"`
	ParentRecord []string                 `json:"父记录,omitempty"`

	Prompt string `json:"prompt,omitempty"`
}

type Coding700FeishuFileObj struct {
	FileToken string `json:"file_token"`
}

// 上传任务数据
func Coding700UploadTaskDataHandler(c *cli.Context) error {

	ctx := c.Context
	projectName := c.String("project")
	promptYn := c.Bool("prompt_yn")
	rolloutYn := c.Bool("rollout_yn")
	repoIgnoreFieldsStr := c.String("repo_ignore_fields")
	cleanInstallModules := c.Bool("clean_install_modules")

	repoIgnoreFields := []string{}
	if repoIgnoreFieldsStr != "" {
		repoIgnoreFields = strings.Split(repoIgnoreFieldsStr, ",")
	}

	if projectName == "" {
		return fmt.Errorf("项目名称不能为空")
	}

	// 循环WorkHome下所有目录, 判断是否为未上传题目目录, 判断目录中是否有: configWork.WorkDoneFileName
	dirs, err := os.ReadDir(workHome)
	if err != nil {
		return fmt.Errorf("读取目录 %s 失败: %v", workHome, err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		if dir.Name() != projectName {
			continue
		}

		fmt.Printf("项目: %s\n", projectName)

		projectDir := filepath.Join(workHome, projectName)
		if cleanInstallModules {
			cmd := exec.Command("sh", "-c", fmt.Sprintf(
				"find %s -type d -name 'node_modules' -prune -exec rm -rf {} +",
				projectDir,
			))
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("删除node_modules目录失败: %v, output: %s", err, string(output))
			}
		}

		// 检查repo.json(1个), prompt.json(7个), rollout.json(35个) 是否存在, 以及格式是否正确
		feishuDataList, err := Coding700FormatJsonFile(ctx, projectName, promptYn, rolloutYn, repoIgnoreFields)
		if err != nil {
			return fmt.Errorf("格式化json文件失败: %v", err)
		}

		// 循环feishuDataList, 上传到飞ishu
		repoRecordId := ""
		beforePromptRecordId := ""
		for _, feishuData := range feishuDataList {
			logStr := ""
			if feishuData.Coding700RepoData != nil {
				// 上传任务到飞书(这个放到最前面, 是为了拿第一个记录id)
				_, err = uploadFeishuData(ctx, &feishuData, repoIgnoreFields)
				repoRecordId = feishuData.RecordID

				logStr = fmt.Sprintf("recordID: %s, 记录UID", feishuData.RecordID)
			} else if feishuData.Coding700PromptData != nil {
				// 上传任务到飞书
				feishuData.Coding700PromptData.ParentRecord = []string{repoRecordId}
				_, err = uploadFeishuData(ctx, &feishuData, repoIgnoreFields)
				beforePromptRecordId = feishuData.RecordID

				logStr = fmt.Sprintf("recordID: %s, - promptId: %s", feishuData.RecordID, feishuData.Coding700PromptData.PromptId)
			} else if feishuData.Coding700RolloutData != nil {
				// 上传任务到飞书
				feishuData.Coding700RolloutData.ParentRecord = []string{beforePromptRecordId}
				_, err = uploadFeishuData(ctx, &feishuData, repoIgnoreFields)

				logStr = fmt.Sprintf("recordID: %s, -- rolloutID: %s", feishuData.RecordID, feishuData.Coding700RolloutData.RolloutID)
			} else {
				return fmt.Errorf("飞书记录结构体异常, 实际 %s", feishuData.RecordID)
			}

			if feishuData.RecordID == "" {
				return fmt.Errorf("上传任务到飞书失败, %s, err: %v", logStr, err)
			}
			if err != nil {
				return fmt.Errorf("上传任务到飞书失败, %s, err: %v", logStr, err)
			} else {
				fmt.Printf("上传任务到飞书成功, %s\n", logStr)
			}
		}
		// // 上传完成后, 写入上传完成标记文件
		// if err := os.WriteFile(uploadDoneFile, []byte("1"), 0644); err != nil {
		// 	return fmt.Errorf("写入上传完成标记文件 %s 失败: %v", uploadDoneFile, err)
		// }

		// 将题目回传至主力机器mac
		if promptYn && !rolloutYn {
			// 判断当前机器是不是masterMachineSshAddr, 如果不是的话, rsync把当前项目传到主力机器上
			if output, err := exec.Command("sh", "-c", fmt.Sprintf(
				"ssh -o BatchMode=yes -o ConnectTimeout=1 %s 'echo ok'", masterMachineSshAddr,
			)).CombinedOutput(); err == nil && strings.TrimSpace(string(output)) == "ok" {
				masterProjectDir := filepath.Join(masterWorkHome, projectName)
				if output, err := exec.Command("sh", "-c", fmt.Sprintf(
					"rsync -avz %s/ %s:%s/", projectDir, masterMachineSshAddr, masterProjectDir,
				)).CombinedOutput(); err != nil {
					return fmt.Errorf("rsync到主力机器失败: %v, output: %s", err, string(output))
				}
				fmt.Printf("rsync到主力机器成功: %s -> %s:%s\n", projectDir, masterMachineSshAddr, masterProjectDir)
			}
		}
	}

	return nil
}

// 检查repo.json(1个), prompt.json(7个), rollout.json(35个) 是否存在, 以及格式是否正确
func Coding700FormatJsonFile(ctx context.Context, projectName string, checkPromptYn, checkRolloutYn bool, repoIgnoreFields []string) (ret []Coding700FeishuData, err error) {
	defer func() {
		// 预期条数
		expectedCount := 43
		if !checkPromptYn {
			expectedCount -= 7 + 35
		} else if !checkRolloutYn {
			expectedCount -= 35
		}
		if len(ret) != expectedCount {
			err = fmt.Errorf("json文件数量不对, 期望 %d 个记录, 实际 %d, err: %v", expectedCount, len(ret), err)
		}
	}()
	workDir := filepath.Join(workHome, projectName)
	// 检查repo.json是否存在, 以及格式是否正确
	repoJsonDir := filepath.Join(workDir, "repo.json")
	if _, err = os.Stat(repoJsonDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("repo.json 不存在, 项目名称: %s", projectName)
	}
	// 解析repo.json文件
	// 解析repo.json
	repoJsonData, err := os.ReadFile(repoJsonDir)
	if err != nil {
		return nil, fmt.Errorf("读取工作文件 %s 失败: %v", repoJsonDir, err)
	}
	var repoData Coding700RepoData
	if err = json.Unmarshal(repoJsonData, &repoData); err != nil {
		return nil, fmt.Errorf("解析工作文件 %s 失败: %v", repoJsonDir, err)
	}

	// repo.json字段检查
	if !libraryUtils.InArray(`repo_type`, repoIgnoreFields) && !libraryUtils.InArray(repoData.RepoType, repoTypes) {
		return nil, fmt.Errorf("repo_type 只能是 公有仓库 或 私有仓库")
	}
	if !libraryUtils.InArray(`language`, repoIgnoreFields) && repoData.Language == "" {
		return nil, fmt.Errorf("language 不能为空")
	}
	if !libraryUtils.InArray(`repo_url`, repoIgnoreFields) && repoData.RepoURL == "" {
		return nil, fmt.Errorf("repo_url 不能为空")
	}
	taskCount := cast.ToInt(repoData.TaskCount)
	if taskCount <= 0 {
		return nil, fmt.Errorf("task_count 必须大于0")
	}

	repoRecordID, err := Coding700GetRepoRecord(ctx, repoData.Uid, repoData.RepoURL)
	if err != nil {
		return nil, fmt.Errorf("获取仓库记录, 飞书报错: %v", err)
	}

	if !libraryUtils.InArray(`repo.zip`, repoIgnoreFields) {
		repoZipFileToken, err := feishuDriveBusiness.DriveBusiness.Upload(ctx, filepath.Join(workDir, "repo.zip"), feishuDocAppToken)
		if err != nil {
			return nil, fmt.Errorf("上传仓库文件失败: %v", err)
		}
		if repoZipFileToken == "" {
			return nil, fmt.Errorf("repo.zip 不能为空")
		}
		repoData.Repo = []Coding700FeishuFileObj{
			{
				FileToken: repoZipFileToken,
			},
		}
	}

	if !libraryUtils.InArray(`Dockerfile`, repoIgnoreFields) {
		dockerfileToken, err := feishuDriveBusiness.DriveBusiness.Upload(ctx, filepath.Join(workDir, "Dockerfile"), feishuDocAppToken)
		if err != nil {
			return nil, fmt.Errorf("上传Dockerfile文件失败: %v", err)
		}
		if dockerfileToken == "" {
			return nil, fmt.Errorf("Dockerfile 不能为空")
		}
		repoData.Dockerfile = []Coding700FeishuFileObj{
			{
				FileToken: dockerfileToken,
			},
		}
	}

	ret = append(ret, Coding700FeishuData{
		RecordID:          repoRecordID,
		Coding700RepoData: &repoData,
	})

	// 循环检查prompt_[1-7].json是否存在, 以及格式是否正确
	if checkPromptYn {
		for promptId := 1; promptId <= 7; promptId++ {
			promptDir := filepath.Join(workDir, fmt.Sprintf("prompt_%d", promptId))
			promptJsonPath := filepath.Join(promptDir, "prompt.json")
			if _, err = os.Stat(promptJsonPath); os.IsNotExist(err) {
				return nil, fmt.Errorf("prompt_id: %d, prompt_%d.json 不存在", promptId, promptId)
			}
			// 解析prompt.json文件
			promptJsonData, err := os.ReadFile(promptJsonPath)
			if err != nil {
				return nil, fmt.Errorf("prompt_id: %d, 读取工作文件 %s 失败: %v", promptId, promptJsonPath, err)
			}
			var promptData Coding700PromptData
			if err = json.Unmarshal(promptJsonData, &promptData); err != nil {
				return nil, fmt.Errorf("prompt_id: %d, 解析工作工作文件 %s 失败: %v", promptId, promptJsonPath, err)
			}

			// prompt.json字段检查
			promptId := cast.ToInt(promptData.PromptId)
			if !libraryUtils.InArray(promptId, []int{1, 2, 3, 4, 5, 6, 7}) {
				return nil, fmt.Errorf("prompt_id: %d, prompt_id 只能是 1-7", promptId)
			}
			if !libraryUtils.InArray(promptData.Difficulty, []string{"简单", "中等", "困难"}) {
				return nil, fmt.Errorf("prompt_id: %d, difficulty 只能是 简单, 中等, 难难", promptId)
			}
			if !libraryUtils.InArray(promptData.Category, []string{"代码生成", "Bug修复", "Bug修复/调试", "Bug 修复 / 调试", "代码重构", "功能迭代", "测试", "代码理解与分析", "DevOps/工程化"}) {
				return nil, fmt.Errorf("prompt_id: %d, category 只能是 代码生成, Bug修复, Bug修复/调试, Bug 修复 / 调试, 代码重构, 功能迭代, 测试, 代码理解与分析, DevOps/工程化", promptId)
			}
			if promptData.Category == "Bug修复/调试" || promptData.Category == "Bug修复" {
				promptData.Category = "Bug 修复 / 调试"
			}
			if promptData.TechStack == "" {
				return nil, fmt.Errorf("prompt_id: %d, tech_stack 不能为空", promptId)
			}
			if promptData.ModuleTags == "" {
				return nil, fmt.Errorf("prompt_id: %d, module_tags 不能为空", promptId)
			}

			// 读取prompt.md文件内容, 写入到promptData.Prompt
			promptMdPath := filepath.Join(promptDir, `prompt.md`)
			if _, err = os.Stat(promptMdPath); os.IsNotExist(err) {
				return nil, fmt.Errorf("prompt_id: %d, prompt.md 不存在", promptId)
			}
			promptMdData, err := os.ReadFile(promptMdPath)
			if err != nil {
				return nil, fmt.Errorf("prompt_id: %d, 读取工作文件 %s 失败: %v", promptId, promptMdPath, err)
			}
			promptData.Prompt = string(promptMdData)
			if promptData.Prompt == "" {
				return nil, fmt.Errorf("prompt_id: %d, prompt 不能为空", promptId)
			}

			promptRecordID, err := Coding700GetPromptRecord(ctx, repoRecordID, promptId)
			if err != nil {
				return nil, fmt.Errorf("prompt_id: %d, 获取prompt记录, 飞书报错: %v", promptId, err)
			}
			promptData.ParentRecord = []string{repoRecordID}
			ret = append(ret, Coding700FeishuData{
				RecordID:            promptRecordID,
				Coding700PromptData: &promptData,
			})

			// 循环检查rollout[1-5].json是否存在, 以及格式是否正确
			if checkRolloutYn {
				for rolloutId := 1; rolloutId <= 5; rolloutId++ {
					rolloutRecordID, err := Coding700GetRolloutRecord(ctx, promptRecordID, rolloutId)
					if err != nil {
						return nil, fmt.Errorf("rollout_id: %d, 获取rollout记录, 飞书报错: %v", rolloutId, err)
					}

					rolloutDir := filepath.Join(promptDir, fmt.Sprintf("rollout_%d", rolloutId))

					rolloutData, err := getRolloutJsonData(projectName, promptId, rolloutId)
					if err != nil {
						return nil, fmt.Errorf("rollout_id: %d, 获取rollout.json文件数据失败: %v", rolloutId, err)
					}

					rolloutData.Prompt = promptData.Prompt

					diffFileName := getDiffFileName(promptId, rolloutId)
					rolloutDiffFilePath := filepath.Join(rolloutDir, diffFileName)
					if _, err = os.Stat(rolloutDiffFilePath); os.IsNotExist(err) {
						return nil, fmt.Errorf("rollout_id: %d, %s 不存在", rolloutId, diffFileName)
					}

					if err = checkRolloutFields(rolloutData); err != nil {
						return nil, err
					}

					gitDiffFileToken, err := feishuDriveBusiness.DriveBusiness.Upload(ctx, fmt.Sprintf("%s", rolloutDiffFilePath), feishuDocAppToken)
					if err != nil {
						return nil, fmt.Errorf("rollout_id: %d, 上传%s文件失败: %v", rolloutId, diffFileName, err)
					}
					if gitDiffFileToken == "" {
						return nil, fmt.Errorf("rollout_id: %d, %s git_diff 不能为空", rolloutId, diffFileName)
					}
					rolloutData.GitDiff = []Coding700FeishuFileObj{
						{
							FileToken: gitDiffFileToken,
						},
					}
					rolloutData.ParentRecord = []string{promptRecordID}
					ret = append(ret, Coding700FeishuData{
						RecordID:             rolloutRecordID,
						Coding700RolloutData: rolloutData,
					})
				}
			}
		}
	}
	return ret, nil
}

func checkRolloutFields(rolloutData *Coding700RolloutData) error {
	// rollout.json字段检查
	if !libraryUtils.InArray(rolloutData.RolloutID, []string{`1`, `2`, `3`, `4`, `5`}) {
		return fmt.Errorf("rollout_id 只能是 1-5")
	}
	if rolloutData.SessionID == "" {
		return fmt.Errorf("session_id 不能为空")
	}
	if !libraryUtils.InArray(rolloutData.ModelName, []string{"Doubao-Seed-2.0-Code", "GPT5.4", "Gemini 3.1 pro", "DeepSeek-v4", "MinMax-M2.7", "GLM-5.1", "Qwen3.6-Plus"}) {
		return fmt.Errorf("model_name 只能是 Doubao-Seed-2.0-Code, GPT5.4, Gemini 3.1 pro, DeepSeek-v4, MinMax-M2.7, GLM-5.1, Qwen3.6-Plus")
	}
	if !libraryUtils.InArray(rolloutData.Score, []string{"0", "1", "2"}) {
		return fmt.Errorf("score 只能是 0, 1, 2")
	}
	if rolloutData.RolloutID == "1" && rolloutData.Score != "0" {
		return fmt.Errorf("第一轮作业必须是0分(其他轮不限制)")
	}
	if rolloutData.ScoreReason == "" {
		return fmt.Errorf("score_reason 不能为空")
	}
	return nil
}

func Coding700GetRepoRecord(ctx context.Context, repoUidStr, repoUrl string) (ret string, err error) {
	defer func() {
		fmt.Printf("检查repo记录 %s %s\n", repoUrl, ret)
	}()

	repoUID := cast.ToInt(repoUidStr)
	if repoUID <= 0 && repoUrl == "" {
		return "", fmt.Errorf("repo_uid 和 repo_url 不能同时为空")
	}

	var conditions []*larkbitable.Condition
	if repoUID > 0 {
		conditions = []*larkbitable.Condition{
			{
				FieldName: larkcore.StringPtr(`题目 id`),
				Operator:  larkcore.StringPtr(`is`),
				Value:     []string{fmt.Sprintf("%d", repoUID)},
			},
		}
	} else if repoUrl != "" && repoUrl != "-" {
		conditions = []*larkbitable.Condition{
			{
				FieldName: larkcore.StringPtr(`repo_url`),
				Operator:  larkcore.StringPtr(`is`),
				Value:     []string{repoUrl},
			},
		}
	}

	time.Sleep(1 * time.Second)
	req := larkbitable.NewSearchAppTableRecordReqBuilder().
		AppToken(feishuDocAppToken).
		TableId(feishuDocTableId).
		PageSize(1).
		Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
			ViewId(feishuDocViewId).
			Filter(&larkbitable.FilterInfo{
				Conjunction: larkcore.StringPtr(`and`),
				Conditions:  conditions,
			}).
			// AutomaticFields().
			Build()).
		Build()

	feishuResp, err := feishuCloudDocBusiness.BaseTablesBusiness.SearchBaseTables(ctx, req)
	if err != nil {
		return "", err
	}
	if feishuResp == nil || feishuResp.Data == nil || len(feishuResp.Data.Items) == 0 {
		return "", nil
	}

	return *feishuResp.Data.Items[0].RecordId, nil
}

func Coding700GetPromptRecord(ctx context.Context, repoRecordId string, promptId int) (ret string, err error) {
	defer func() {
		fmt.Printf("- 检查prompt记录 repoRecordId:%s, prompt_index:%d %s\n", repoRecordId, promptId, ret)
	}()

	time.Sleep(1 * time.Second)
	req := larkbitable.NewSearchAppTableRecordReqBuilder().
		AppToken(feishuDocAppToken).
		TableId(feishuDocTableId).
		PageSize(1).
		Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
			ViewId(feishuDocViewId).
			Filter(&larkbitable.FilterInfo{
				Conjunction: larkcore.StringPtr(`and`),
				Conditions: []*larkbitable.Condition{
					{
						FieldName: larkcore.StringPtr(`prompt_index`),
						Operator:  larkcore.StringPtr(`is`),
						Value:     []string{fmt.Sprintf("%d", promptId)},
					},
					{
						FieldName: larkcore.StringPtr(`父记录`),
						Operator:  larkcore.StringPtr(`is`),
						Value:     []string{repoRecordId},
					},
				},
			}).
			// AutomaticFields().
			Build()).
		Build()

	feishuResp, err := feishuCloudDocBusiness.BaseTablesBusiness.SearchBaseTables(ctx, req)
	if err != nil {
		return "", err
	}
	if feishuResp == nil || feishuResp.Data == nil || len(feishuResp.Data.Items) == 0 {
		return "", nil
	}

	return *feishuResp.Data.Items[0].RecordId, nil
}

func Coding700GetRolloutRecord(ctx context.Context, promptRecordId string, rolloutId int) (ret string, err error) {
	defer func() {
		fmt.Printf("-- 检查rollout记录 promptRecordId:%s, rollout_id:%d %s\n", promptRecordId, rolloutId, ret)
	}()

	time.Sleep(1 * time.Second)
	req := larkbitable.NewSearchAppTableRecordReqBuilder().
		AppToken(feishuDocAppToken).
		TableId(feishuDocTableId).
		PageSize(1).
		Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
			ViewId(feishuDocViewId).
			Filter(&larkbitable.FilterInfo{
				Conjunction: larkcore.StringPtr(`and`),
				Conditions: []*larkbitable.Condition{
					{
						FieldName: larkcore.StringPtr(`rollout_id`),
						Operator:  larkcore.StringPtr(`is`),
						Value:     []string{fmt.Sprintf("%d", rolloutId)},
					},
					{
						FieldName: larkcore.StringPtr(`父记录`),
						Operator:  larkcore.StringPtr(`is`),
						Value:     []string{promptRecordId},
					},
				},
			}).
			// AutomaticFields().
			Build()).
		Build()

	feishuResp, err := feishuCloudDocBusiness.BaseTablesBusiness.SearchBaseTables(ctx, req)
	if err != nil {
		return "", err
	}
	if feishuResp == nil || feishuResp.Data == nil || len(feishuResp.Data.Items) == 0 {
		return "", nil
	}

	return *feishuResp.Data.Items[0].RecordId, nil
}

func uploadFeishuData(ctx context.Context, feishuData *Coding700FeishuData, repoIgnoreFields []string) (ret string, err error) {
	time.Sleep(1 * time.Second)
	Fields := make(map[string]any)
	switch {
	case feishuData.Coding700RepoData != nil:
		b, _ := json.Marshal(feishuData.Coding700RepoData)
		json.Unmarshal(b, &Fields)
		for _, ignore := range repoIgnoreFields {
			delete(Fields, ignore)
		}
		delete(Fields, `uid`)
	case feishuData.Coding700PromptData != nil:
		b, _ := json.Marshal(feishuData.Coding700PromptData)
		json.Unmarshal(b, &Fields)
	case feishuData.Coding700RolloutData != nil:
		b, _ := json.Marshal(feishuData.Coding700RolloutData)
		json.Unmarshal(b, &Fields)
	default:
		return "", fmt.Errorf("不支持的记录类型")
	}

	if feishuData.RecordID == "" {
		feishuResp, err := addFeishuRecord(ctx, Fields)
		if err != nil {
			fmt.Println(larkcore.Prettify(feishuResp))
			return "", fmt.Errorf("添加飞书多维表格记录失败: %w", err)
		}
		feishuData.RecordID = *feishuResp.Data.Record.RecordId
		return feishuData.RecordID, nil
	} else {
		feishuResp, err := updateFeishuRecord(ctx, feishuData.RecordID, Fields)
		if err != nil {
			fmt.Println(larkcore.Prettify(feishuResp))
			return "", fmt.Errorf("更新飞书多维表格记录失败: %w", err)
		}
		return feishuData.RecordID, nil
	}
}

func addFeishuRecord(ctx context.Context, Fields map[string]any) (resp *larkbitable.CreateAppTableRecordResp, err error) {
	time.Sleep(1 * time.Second)
	// 创建请求对象
	req := larkbitable.NewCreateAppTableRecordReqBuilder().
		AppToken(feishuDocAppToken).
		TableId(feishuDocTableId).
		AppTableRecord(larkbitable.NewAppTableRecordBuilder().
			Fields(Fields).
			Build()).
		Build()
	resp, err = feishuCloudDocBusiness.BaseTablesBusiness.CreateBaseTables(ctx, req)

	// fmt.Println(larkcore.Prettify(resp))
	return
}

func updateFeishuRecord(ctx context.Context, recordId string, Fields map[string]any) (resp *larkbitable.UpdateAppTableRecordResp, err error) {
	time.Sleep(1 * time.Second)
	// 创建请求对象
	req := larkbitable.NewUpdateAppTableRecordReqBuilder().
		AppToken(feishuDocAppToken).
		TableId(feishuDocTableId).
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

func Coding700UploadInitRepoHandler(c *cli.Context) error {
	ctx := c.Context

	repoType := c.String(`repo_type`)
	repoURL := c.String(`repo_url`)
	if !libraryUtils.InArray(repoType, repoTypes) {
		return fmt.Errorf("repo_type 只能是 公有仓库 或 私有仓库")
	}

	if repoURL != "" {
		fmt.Println(repoURL)
	}

	feishuData := &Coding700FeishuData{
		Coding700RepoData: &Coding700RepoData{
			RepoType:  repoType,
			TaskCount: "7",
		},
	}
	if repoURL != "" {
		feishuData.Coding700RepoData.RepoURL = repoURL
	}
	ret, err := uploadFeishuData(ctx, feishuData, nil)
	if err != nil {
		return fmt.Errorf("上传飞书多维表格记录返回err: %w", err)
	}
	if ret == "" {
		return fmt.Errorf("上传飞书多维表格记录失败(recordId为空)")
	}
	fmt.Printf("上传飞书多维表格记录成功, recordId: %s\n", ret)

	return nil
}
