package monitor

import (
	"fmt"
	"sort"
	"strings"
	"time"

	monitorCs "github.com/armylong/armylong-go/internal/cs/monitor"
	"github.com/shirou/gopsutil/v4/cpu"
)

type cpuBusiness struct{}

var CPUBusiness = &cpuBusiness{}

// 获取CPU基本信息
func (b *cpuBusiness) GetCPUInfo() (*monitorCs.CPUInfo, error) {
	infos, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	if len(infos) == 0 {
		return &monitorCs.CPUInfo{}, nil
	}

	info := infos[0]

	physicalCores, _ := cpu.Counts(false)
	logicalCores, _ := cpu.Counts(true)

	return &monitorCs.CPUInfo{
		ModelName:     info.ModelName,
		PhysicalCores: physicalCores,
		LogicalCores:  logicalCores,
		Frequency:     info.Mhz,
		VendorID:      info.VendorID,
		CacheSize:     info.CacheSize,
		Flags:         info.Flags,
	}, nil
}

// 获取CPU使用率，采样1秒
func (b *cpuBusiness) GetCPUUsage() (*monitorCs.CPUUsage, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	perCorePercentages, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, err
	}

	timesStat, err := cpu.Times(false)
	if err != nil || len(timesStat) == 0 {
		usage := &monitorCs.CPUUsage{
			TotalUsage:   percentages[0],
			PerCoreUsage: perCorePercentages,
		}
		return usage, nil
	}

	stat := timesStat[0]
	total := stat.User + stat.System + stat.Idle + stat.Nice + stat.Iowait + stat.Irq + stat.Softirq + stat.Steal + stat.Guest

	usage := &monitorCs.CPUUsage{
		User:         stat.User,
		System:       stat.System,
		Idle:         stat.Idle,
		Nice:         stat.Nice,
		IOWait:       stat.Iowait,
		IRQ:          stat.Irq,
		SoftIRQ:      stat.Softirq,
		Steal:        stat.Steal,
		Guest:        stat.Guest,
		TotalUsage:   percentages[0],
		PerCoreUsage: perCorePercentages,
	}

	if total > 0 {
		usage.User = stat.User / total * 100
		usage.System = stat.System / total * 100
		usage.Idle = stat.Idle / total * 100
		usage.Nice = stat.Nice / total * 100
		usage.IOWait = stat.Iowait / total * 100
		usage.IRQ = stat.Irq / total * 100
		usage.SoftIRQ = stat.Softirq / total * 100
		usage.Steal = stat.Steal / total * 100
		usage.Guest = stat.Guest / total * 100
	}

	return usage, nil
}

// 获取按CPU使用率排序的进程列表
func (b *cpuBusiness) GetCPUProcesses(sortBy string, sortDirection string, limit int) ([]monitorCs.CPUProcessInfo, error) {
	if sortDirection == "" {
		sortDirection = monitorCs.SortDirectionDesc
	}

	processes, err := ProcessBusiness.GetProcessList(sortBy, sortDirection, limit)
	if err != nil {
		return nil, err
	}

	var result []monitorCs.CPUProcessInfo
	for _, p := range processes {
		result = append(result, monitorCs.CPUProcessInfo{
			PID:        p.PID,
			Name:       p.Name,
			CmdLine:    p.CmdLine,
			CPU:        p.CPU,
			Memory:     p.Memory,
			MemoryMB:   p.MemoryMB,
			User:       p.User,
			StartTime:  p.StartTime,
			CreateTime: p.CreateTime,
		})
	}

	if sortBy == "cpu" {
		sort.Slice(result, func(i, j int) bool {
			if sortDirection == monitorCs.SortDirectionAsc {
				return result[i].CPU < result[j].CPU
			}
			return result[i].CPU > result[j].CPU
		})
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

// 按PID杀进程
func (b *cpuBusiness) KillProcessByPID(pid int32) error {
	return ProcessBusiness.KillProcess(pid)
}

// 获取CPU占用最高的N个进程
func (b *cpuBusiness) GetTopCPUProcesses(limit int) ([]monitorCs.CPUProcessInfo, error) {
	return b.GetCPUProcesses("cpu", monitorCs.SortDirectionDesc, limit)
}

// 格式化CPU信息(命令行用)
func (b *cpuBusiness) FormatCPUInfoOutput(info *monitorCs.CPUInfo) string {
	if info == nil {
		return "无法获取CPU信息"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("型号:       %s\n", info.ModelName))
	sb.WriteString(fmt.Sprintf("物理核心:   %d\n", info.PhysicalCores))
	sb.WriteString(fmt.Sprintf("逻辑核心:   %d\n", info.LogicalCores))
	if info.Frequency > 0 {
		sb.WriteString(fmt.Sprintf("频率:       %.2f MHz\n", info.Frequency))
	}
	if info.VendorID != "" {
		sb.WriteString(fmt.Sprintf("厂商:       %s\n", info.VendorID))
	}
	if info.CacheSize != 0 {
		sb.WriteString(fmt.Sprintf("缓存:       %d KB\n", info.CacheSize))
	}

	return sb.String()
}

// 格式化CPU使用率(命令行用)
func (b *cpuBusiness) FormatCPUUsageOutput(usage *monitorCs.CPUUsage) string {
	if usage == nil {
		return "无法获取CPU使用率"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("总使用率:   %.1f%%\n", usage.TotalUsage))
	sb.WriteString(fmt.Sprintf("用户态:     %.1f%%\n", usage.User))
	sb.WriteString(fmt.Sprintf("系统态:     %.1f%%\n", usage.System))
	sb.WriteString(fmt.Sprintf("空闲:       %.1f%%\n", usage.Idle))

	if usage.Nice > 0 {
		sb.WriteString(fmt.Sprintf("Nice:       %.1f%%\n", usage.Nice))
	}
	if usage.IOWait > 0 {
		sb.WriteString(fmt.Sprintf("IO等待:     %.1f%%\n", usage.IOWait))
	}
	if usage.IRQ > 0 {
		sb.WriteString(fmt.Sprintf("硬件中断:   %.1f%%\n", usage.IRQ))
	}
	if usage.SoftIRQ > 0 {
		sb.WriteString(fmt.Sprintf("软件中断:   %.1f%%\n", usage.SoftIRQ))
	}

	return sb.String()
}
