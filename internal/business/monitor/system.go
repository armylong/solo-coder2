package monitor

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	monitorCs "github.com/armylong/armylong-go/internal/cs/monitor"
	"github.com/shirou/gopsutil/v4/host"
)

type systemBusiness struct{}

var SystemBusiness = &systemBusiness{}

// 获取系统信息
func (b *systemBusiness) GetSystemInfo() (*monitorCs.SystemInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	sysInfo := &monitorCs.SystemInfo{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		Architecture:    runtime.GOARCH,
		CPUCount:        runtime.NumCPU(),
		GoVersion:       runtime.Version(),
	}

	return sysInfo, nil
}

// 获取系统运行时间
func (b *systemBusiness) GetUptime() (*monitorCs.UptimeInfo, error) {
	uptime, err := host.Uptime()
	if err != nil {
		return nil, err
	}

	bootTime, err := host.BootTime()
	if err != nil {
		return nil, err
	}

	return &monitorCs.UptimeInfo{
		Uptime:    uptime,
		BootTime:  bootTime,
		UptimeStr: b.formatUptime(uptime),
	}, nil
}

// 格式化运行时间
func (b *systemBusiness) formatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d天", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d小时", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d分钟", minutes))
	}
	if len(parts) == 0 || secs > 0 {
		parts = append(parts, fmt.Sprintf("%d秒", secs))
	}

	return strings.Join(parts, " ")
}

// 格式化系统信息(命令行用)
func (b *systemBusiness) FormatSystemInfoOutput(info *monitorCs.SystemInfo) string {
	if info == nil {
		return "无法获取系统信息"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("主机名:       %s\n", info.Hostname))
	sb.WriteString(fmt.Sprintf("操作系统:     %s\n", info.OS))
	sb.WriteString(fmt.Sprintf("平台:         %s\n", info.Platform))
	sb.WriteString(fmt.Sprintf("平台版本:     %s\n", info.PlatformVersion))
	sb.WriteString(fmt.Sprintf("内核版本:     %s\n", info.KernelVersion))
	sb.WriteString(fmt.Sprintf("架构:         %s\n", info.Architecture))
	sb.WriteString(fmt.Sprintf("CPU核心数:    %d\n", info.CPUCount))
	sb.WriteString(fmt.Sprintf("Go版本:       %s\n", info.GoVersion))

	return sb.String()
}

// 格式化运行时间(命令行用)
func (b *systemBusiness) FormatUptimeOutput(info *monitorCs.UptimeInfo) string {
	if info == nil {
		return "无法获取运行时间"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("系统已运行: %s\n", info.UptimeStr))
	sb.WriteString(fmt.Sprintf("启动时间:   %s\n", time.Unix(int64(info.BootTime), 0).Format("2006-01-02 15:04:05")))

	return sb.String()
}
