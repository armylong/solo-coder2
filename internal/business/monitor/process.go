package monitor

import (
	"fmt"
	"sort"
	"strings"
	"time"

	monitorCs "github.com/armylong/armylong-go/internal/cs/monitor"
	"github.com/shirou/gopsutil/v4/process"
)

type processBusiness struct{}

var ProcessBusiness = &processBusiness{}

// 获取进程列表，支持排序和限制数量
func (b *processBusiness) GetProcessList(sortBy string, sortDirection string, limit int) ([]monitorCs.ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	if sortDirection == "" {
		sortDirection = monitorCs.SortDirectionDesc
	}

	var result []monitorCs.ProcessInfo
	for _, p := range processes {
		info := b.convertProcess(p)
		result = append(result, info)
	}

	result = b.sortProcesses(result, sortBy, sortDirection)
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

// 转换gopsutil进程对象为ProcessInfo
func (b *processBusiness) convertProcess(p *process.Process) monitorCs.ProcessInfo {
	info := monitorCs.ProcessInfo{
		PID: p.Pid,
	}

	name, err := p.Name()
	if err == nil {
		info.Name = name
	}

	cmdline, err := p.Cmdline()
	if err == nil {
		info.CmdLine = cmdline
	}

	cpuPercent, err := p.CPUPercent()
	if err == nil {
		info.CPU = cpuPercent
	}

	memPercent, err := p.MemoryPercent()
	if err == nil {
		info.Memory = memPercent
	}

	memInfo, err := p.MemoryInfo()
	if err == nil && memInfo != nil {
		info.MemoryMB = float32(memInfo.RSS) / 1024 / 1024
	}

	status, err := p.Status()
	if err == nil && len(status) > 0 {
		info.Status = status[0]
	}

	ppid, err := p.Ppid()
	if err == nil {
		info.PPID = ppid
	}

	numThreads, err := p.NumThreads()
	if err == nil {
		info.NumThreads = numThreads
	}

	username, err := p.Username()
	if err == nil {
		info.User = username
	}

	createTime, err := p.CreateTime()
	if err == nil && createTime > 0 {
		info.CreateTime = createTime
		info.StartTime = formatTime(createTime)
	}

	return info
}

// 进程排序
func (b *processBusiness) sortProcesses(processes []monitorCs.ProcessInfo, sortBy string, sortDirection string) []monitorCs.ProcessInfo {
	isDesc := sortDirection == monitorCs.SortDirectionDesc

	switch sortBy {
	case "cpu":
		sort.Slice(processes, func(i, j int) bool {
			if isDesc {
				return processes[i].CPU > processes[j].CPU
			}
			return processes[i].CPU < processes[j].CPU
		})
	case "memory":
		sort.Slice(processes, func(i, j int) bool {
			if isDesc {
				return processes[i].Memory > processes[j].Memory
			}
			return processes[i].Memory < processes[j].Memory
		})
	case "pid":
		sort.Slice(processes, func(i, j int) bool {
			if isDesc {
				return processes[i].PID > processes[j].PID
			}
			return processes[i].PID < processes[j].PID
		})
	case "time":
		sort.Slice(processes, func(i, j int) bool {
			if isDesc {
				return processes[i].CreateTime > processes[j].CreateTime
			}
			return processes[i].CreateTime < processes[j].CreateTime
		})
	}
	return processes
}

// 获取CPU/内存占用最高的N个进程
func (b *processBusiness) GetTopProcesses(by string, limit int) ([]monitorCs.ProcessInfo, error) {
	return b.GetProcessList(by, monitorCs.SortDirectionDesc, limit)
}

// 杀掉指定进程
func (b *processBusiness) KillProcess(pid int32) error {
	if pid <= 0 {
		return fmt.Errorf("无效的进程ID: %d", pid)
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("找不到进程 %d: %v", pid, err)
	}

	return p.Kill()
}

// 按名称关键词查找进程
func (b *processBusiness) FindProcessByName(name string) ([]monitorCs.ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var matched []monitorCs.ProcessInfo
	nameLower := strings.ToLower(name)

	for _, p := range processes {
		pName, err := p.Name()
		if err != nil {
			continue
		}

		cmdline, _ := p.Cmdline()

		if strings.Contains(strings.ToLower(pName), nameLower) ||
			strings.Contains(strings.ToLower(cmdline), nameLower) {
			matched = append(matched, b.convertProcess(p))
		}
	}

	return matched, nil
}

// 获取单个进程信息
func (b *processBusiness) GetProcessInfo(pid int32) (*monitorCs.ProcessInfo, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("进程 %d 不存在", pid)
	}

	info := b.convertProcess(p)
	return &info, nil
}

// FormatProcessOutput 格式化进程输出 (用于命令行)
func (b *processBusiness) FormatProcessOutput(processes []monitorCs.ProcessInfo) string {
	if len(processes) == 0 {
		return "暂无进程信息"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-10s %-8s %-8s %-12s %-10s %s\n", "PID", "CPU%", "MEM%", "MEM(MB)", "PPID", "命令"))
	sb.WriteString(strings.Repeat("-", 80) + "\n")

	for _, p := range processes {
		sb.WriteString(fmt.Sprintf("%-10d %-8.1f %-8.1f %-12.1f %-10d %s\n",
			p.PID, p.CPU, p.Memory, p.MemoryMB, p.PPID, p.Name))
	}

	return sb.String()
}

// formatTime 格式化时间戳
func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	t := time.UnixMilli(timestamp)
	return t.Format("2006-01-02 15:04:05")
}
