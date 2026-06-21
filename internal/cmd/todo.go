package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armylong/armylong-go/internal/common/webcache"
	"github.com/spf13/cast"
	"github.com/urfave/cli/v2"
)

// Todo任务管理
func TodoHandler(c *cli.Context) error {
	ctx := c.Context

	taskType := ""
	if c.NArg() > 0 {
		taskType = c.Args().Get(0)
	}
	taskId := c.Int64("task_id")
	title := c.String("title")
	desc := c.String("desc")
	sort := c.Int64("sort")
	expireAt := c.String("expire_at")

	if taskType == "" {
		fmt.Println("错误: task_type 不能为空")
		fmt.Println("可用命令: get, create, sort, complete, expire")
		return nil
	}

	switch taskType {
	case "get":
		taskData, err := getTodoTask(ctx, taskId)
		if err != nil {
			fmt.Printf("获取任务失败: %v\n", err)
			return nil
		}
		printTask(taskId, taskData)
		return nil
	case "create":
		res, err := createTodoTask(ctx, title, desc, sort, expireAt)
		if err != nil {
			fmt.Printf("创建任务失败: %v\n", err)
			return nil
		}
		if res {
			fmt.Println("✓ 任务创建成功")
		} else {
			fmt.Println("任务已存在，更新完成")
		}
		return nil
	case "sort":
		if taskId == 0 {
			fmt.Println("错误: sort 命令需要指定 task_id")
			return nil
		}
		err := updateTaskSort(ctx, taskId, sort)
		if err != nil {
			fmt.Printf("更新排序失败: %v\n", err)
			return nil
		}
		fmt.Printf("✓ 任务 %d 排序值已更新为 %d\n", taskId, sort)
		return nil
	case "complete":
		if taskId == 0 {
			fmt.Println("错误: complete 命令需要指定 task_id")
			return nil
		}
		err := completeTodoTask(ctx, taskId)
		if err != nil {
			fmt.Printf("完成任务失败: %v\n", err)
			return nil
		}
		fmt.Printf("✓ 任务 %d 已标记为完成\n", taskId)
		return nil
	case "expire":
		count, err := expireTodoTasks(ctx)
		if err != nil {
			fmt.Printf("检测过期任务失败: %v\n", err)
			return nil
		}
		if count > 0 {
			fmt.Printf("✓ 检测到 %d 个过期任务并已标记\n", count)
		} else {
			fmt.Println("✓ 未发现过期任务")
		}
		return nil
	default:
		fmt.Printf("未知命令: %s\n", taskType)
		fmt.Println("可用命令: get, create, sort, complete, expire")
	}
	return nil
}

// Todo任务请求
type TodoTaskRequest struct {
	TaskType     string    `json:"task_type"`
	TaskId       int64     `json:"task_id"`
	TaskData     *TaskData `json:"task_data"`
	TaskDataJson string    `json:"task_data_json"`
}

// 任务数据
type TaskData struct {
	Title       string `json:"title"`
	Desc        string `json:"desc"`
	Sort        int64  `json:"sort"`
	Status      int    `json:"status"`
	ExpireAt    string `json:"expire_at"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	DeletedAt   string `json:"deleted_at"`
	CompletedAt string `json:"completed_at"`
}

// 任务状态
const (
	statusDeleted   = 0
	statusNormal    = 1
	statusCompleted = 2
	statusExpired   = 3
)

// 获取任务存储key
func getTodoTaskKey(ctx context.Context) string {
	return "todo:task:map"
}

// 用时间戳生成任务ID
func generateTaskId(ctx context.Context) int64 {
	return cast.ToInt64(time.Now().Format("20060102150405"))
}

// 创建任务
func createTodoTask(ctx context.Context, title, desc string, sort int64, expireAt string) (bool, error) {
	if title == "" {
		return false, fmt.Errorf("title is empty")
	}
	if desc == "" {
		return false, fmt.Errorf("desc is empty")
	}

	if expireAt != "" {
		_, err := time.Parse("2006-01-02 15:04:05", expireAt)
		if err != nil {
			return false, fmt.Errorf("expire_at format error, expected: 2006-01-02 15:04:05")
		}
	}

	taskId := generateTaskId(ctx)
	fmt.Printf("生成任务id: %d\n", taskId)

	taskData := &TaskData{
		Title:     title,
		Desc:      desc,
		Sort:      sort,
		Status:    statusNormal,
		ExpireAt:  expireAt,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	taskDataJson, err := json.Marshal(taskData)
	if err != nil {
		return false, fmt.Errorf("json marshal error: %v", err)
	}

	res, err := webcache.RedisClient.HSet(ctx, getTodoTaskKey(ctx), taskId, string(taskDataJson)).Result()
	return res == 1, err
}

// 获取全部任务map
func getTodoTaskMap(ctx context.Context) (map[int64]TaskData, error) {
	taskKey := getTodoTaskKey(ctx)
	taskDataJson, err := webcache.RedisClient.HGetAll(ctx, taskKey).Result()
	if err != nil {
		return nil, err
	}
	taskMap := make(map[int64]TaskData)
	for taskIdStr, taskDataJson := range taskDataJson {
		taskId := cast.ToInt64(taskIdStr)
		taskData := TaskData{}
		if err := json.Unmarshal([]byte(taskDataJson), &taskData); err != nil {
			continue
		}
		taskMap[taskId] = taskData
	}
	return taskMap, nil
}

// 获取单个任务
func getTodoTask(ctx context.Context, taskId int64) (*TaskData, error) {
	if taskId == 0 {
		return nil, fmt.Errorf("task_id is empty")
	}
	taskMap, err := getTodoTaskMap(ctx)
	if err != nil {
		return nil, err
	}
	taskData, exists := taskMap[taskId]
	if !exists {
		return nil, fmt.Errorf("task_id %d not found", taskId)
	}
	return &taskData, nil
}

// 更新任务排序值
func updateTaskSort(ctx context.Context, taskId int64, sort int64) error {
	if taskId == 0 {
		return fmt.Errorf("task_id is empty")
	}

	taskData, err := getTodoTask(ctx, taskId)
	if err != nil {
		return err
	}

	taskData.Sort = sort
	taskData.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	taskDataJson, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}

	_, err = webcache.RedisClient.HSet(ctx, getTodoTaskKey(ctx), taskId, string(taskDataJson)).Result()
	return err
}

// 标记任务完成
func completeTodoTask(ctx context.Context, taskId int64) error {
	if taskId == 0 {
		return fmt.Errorf("task_id is empty")
	}

	taskData, err := getTodoTask(ctx, taskId)
	if err != nil {
		return err
	}

	if taskData.Status == statusDeleted {
		return fmt.Errorf("task has been deleted")
	}
	if taskData.Status == statusCompleted {
		return fmt.Errorf("task already completed")
	}

	taskData.Status = statusCompleted
	taskData.CompletedAt = time.Now().Format("2006-01-02 15:04:05")
	taskData.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	taskDataJson, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}

	_, err = webcache.RedisClient.HSet(ctx, getTodoTaskKey(ctx), taskId, string(taskDataJson)).Result()
	return err
}

// 批量检测并标记过期任务
func expireTodoTasks(ctx context.Context) (int, error) {
	taskMap, err := getTodoTaskMap(ctx)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	expireCount := 0

	for taskId, taskData := range taskMap {
		if taskData.Status != statusNormal {
			continue
		}

		if taskData.ExpireAt == "" {
			continue
		}

		expireTime, err := time.Parse("2006-01-02 15:04:05", taskData.ExpireAt)
		if err != nil {
			continue
		}

		if now.After(expireTime) {
			taskData.Status = statusExpired
			taskData.UpdatedAt = now.Format("2006-01-02 15:04:05")

			taskDataJson, err := json.Marshal(taskData)
			if err != nil {
				continue
			}

			_, err = webcache.RedisClient.HSet(ctx, getTodoTaskKey(ctx), taskId, string(taskDataJson)).Result()
			if err == nil {
				expireCount++
			}
		}
	}

	return expireCount, nil
}

// 状态文案映射
func getStatusText(status int) string {
	switch status {
	case statusDeleted:
		return "已删除"
	case statusNormal:
		return "正常"
	case statusCompleted:
		return "已完成"
	case statusExpired:
		return "已过期"
	default:
		return "未知"
	}
}

// 打印任务详情
func printTask(taskId int64, taskData *TaskData) {
	fmt.Println("========================================")
	fmt.Printf("任务ID:     %d\n", taskId)
	fmt.Printf("标题:       %s\n", taskData.Title)
	fmt.Printf("描述:       %s\n", taskData.Desc)
	fmt.Printf("状态:       %s\n", getStatusText(taskData.Status))
	fmt.Printf("排序值:     %d\n", taskData.Sort)
	if taskData.ExpireAt != "" {
		fmt.Printf("过期时间:   %s\n", taskData.ExpireAt)
	}
	fmt.Printf("创建时间:   %s\n", taskData.CreatedAt)
	fmt.Printf("更新时间:   %s\n", taskData.UpdatedAt)
	if taskData.CompletedAt != "" {
		fmt.Printf("完成时间:   %s\n", taskData.CompletedAt)
	}
	fmt.Println("========================================")
}
