package monitor

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	monitorCs "github.com/armylong/armylong-go/internal/cs/monitor"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

var (
	prevNetStats  map[string]net.IOCountersStat
	prevNetTime   time.Time
	netStatsMutex sync.Mutex
)

type networkBusiness struct{}

var NetworkBusiness = &networkBusiness{}

// 获取所有网络连接
func (b *networkBusiness) GetConnections() ([]monitorCs.ConnectionInfo, error) {
	conns, err := net.Connections("all")
	if err != nil {
		return nil, err
	}

	var connections []monitorCs.ConnectionInfo
	for _, conn := range conns {
		protocol := b.getProtocolName(conn.Type)
		state := b.getConnectionState(conn.Status)

		connections = append(connections, monitorCs.ConnectionInfo{
			Protocol:   protocol,
			LocalAddr:  conn.Laddr.IP,
			LocalPort:  int(conn.Laddr.Port),
			RemoteAddr: conn.Raddr.IP,
			RemotePort: int(conn.Raddr.Port),
			State:      state,
			PID:        conn.Pid,
		})
	}

	return connections, nil
}

// 连接类型转协议名
func (b *networkBusiness) getProtocolName(connType uint32) string {
	switch connType {
	case 1:
		return "tcp"
	case 2:
		return "udp"
	default:
		return "unknown"
	}
}

// 格式化连接状态
func (b *networkBusiness) getConnectionState(status string) string {
	if status == "" {
		return "-"
	}
	return status
}

// 获取端口占用情况
func (b *networkBusiness) GetPortUsage() ([]monitorCs.PortInfo, error) {
	conns, err := net.Connections("all")
	if err != nil {
		return nil, err
	}

	portMap := make(map[string]monitorCs.PortInfo)
	for _, conn := range conns {
		if conn.Laddr.Port == 0 {
			continue
		}

		key := fmt.Sprintf("%s:%d", conn.Laddr.IP, conn.Laddr.Port)
		if _, exists := portMap[key]; !exists {
			protocol := b.getProtocolName(conn.Type)
			portMap[key] = monitorCs.PortInfo{
				Protocol:  protocol,
				Port:      int(conn.Laddr.Port),
				LocalAddr: conn.Laddr.IP,
				State:     b.getConnectionState(conn.Status),
				PID:       conn.Pid,
			}

			if conn.Pid > 0 {
				p, err := process.NewProcess(conn.Pid)
				if err == nil {
					name, _ := p.Name()
					portInfo := portMap[key]
					portInfo.Process = name
					portMap[key] = portInfo
				}
			}
		}
	}

	var ports []monitorCs.PortInfo
	for _, port := range portMap {
		ports = append(ports, port)
	}

	return ports, nil
}

// 获取端口列表，支持筛选和排序
func (b *networkBusiness) GetPortsWithFilter(portFilter int, sortBy string, sortDirection string, limit int) ([]monitorCs.PortInfo, error) {
	ports, err := b.GetPortUsage()
	if err != nil {
		return nil, err
	}

	if sortDirection == "" {
		sortDirection = monitorCs.SortDirectionDesc
	}
	isDesc := sortDirection == monitorCs.SortDirectionDesc

	var filtered []monitorCs.PortInfo
	for _, p := range ports {
		if portFilter > 0 && p.Port != portFilter {
			continue
		}
		filtered = append(filtered, p)
	}

	switch sortBy {
	case "port":
		sort.Slice(filtered, func(i, j int) bool {
			if isDesc {
				return filtered[i].Port > filtered[j].Port
			}
			return filtered[i].Port < filtered[j].Port
		})
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

// 根据端口查找占用进程
func (b *networkBusiness) FindProcessByPort(port int) (*monitorCs.PortInfo, error) {
	ports, err := b.GetPortUsage()
	if err != nil {
		return nil, err
	}

	for _, p := range ports {
		if p.Port == port {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("端口 %d 未被占用", port)
}

// 杀掉占用指定端口的进程
func (b *networkBusiness) KillProcessByPort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("无效的端口号: %d", port)
	}

	portInfo, err := b.FindProcessByPort(port)
	if err != nil {
		return err
	}

	if portInfo.PID == 0 {
		return fmt.Errorf("无法获取占用端口 %d 的进程ID", port)
	}

	return ProcessBusiness.KillProcess(portInfo.PID)
}

// 按PID杀掉网络占用进程
func (b *networkBusiness) KillProcessByPID(pid int32) error {
	return ProcessBusiness.KillProcess(pid)
}

// GetInterfaces 获取网络接口信息
// 返回所有网络接口的详细信息
func (b *networkBusiness) GetInterfaces() ([]monitorCs.InterfaceInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var interfaces []monitorCs.InterfaceInfo
	for _, iface := range ifaces {
		info := monitorCs.InterfaceInfo{
			Name:  iface.Name,
			MTU:   iface.MTU,
			Flags: strings.Join(iface.Flags, ","),
		}

		if iface.HardwareAddr != "" {
			info.Hardware = iface.HardwareAddr
		}

		for _, addr := range iface.Addrs {
			info.Addrs = append(info.Addrs, addr.Addr)
		}

		interfaces = append(interfaces, info)
	}

	return interfaces, nil
}

// GetNetworkBandwidth 获取实时网络带宽
// 通过计算两次采样的差值来获取实时速率
// 首次调用返回基础数据，第二次调用开始计算速率
func (b *networkBusiness) GetNetworkBandwidth() ([]monitorCs.NetworkBandwidthInfo, error) {
	netStatsMutex.Lock()
	defer netStatsMutex.Unlock()

	currentStats, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()
	currentStatsMap := make(map[string]net.IOCountersStat)
	for _, stat := range currentStats {
		currentStatsMap[stat.Name] = stat
	}

	var result []monitorCs.NetworkBandwidthInfo

	for name, stat := range currentStatsMap {
		info := monitorCs.NetworkBandwidthInfo{
			InterfaceName: name,
			BytesSent:     stat.BytesSent,
			BytesRecv:     stat.BytesRecv,
			PacketsSent:   stat.PacketsSent,
			PacketsRecv:   stat.PacketsRecv,
		}

		if prevNetStats != nil && !prevNetTime.IsZero() {
			if prevStat, exists := prevNetStats[name]; exists {
				duration := currentTime.Sub(prevNetTime).Seconds()
				if duration > 0 {
					info.SendRate = float64(stat.BytesSent-prevStat.BytesSent) / duration
					info.RecvRate = float64(stat.BytesRecv-prevStat.BytesRecv) / duration
					info.TotalRate = info.SendRate + info.RecvRate

					maxRate := 100.0 * 1024 * 1024
					if info.TotalRate > 0 {
						info.UsagePercent = info.TotalRate / maxRate * 100
						if info.UsagePercent > 100 {
							info.UsagePercent = 100
						}
					}
				}
			}
		}

		result = append(result, info)
	}

	prevNetStats = currentStatsMap
	prevNetTime = currentTime

	return result, nil
}

// GetNetworkProcesses 获取网络占用进程列表
// sortBy: 排序字段 (usage, connections, port)
// sortDirection: 排序方向 (asc/desc, 默认desc)
// limit: 返回数量限制
func (b *networkBusiness) GetNetworkProcesses(sortBy string, sortDirection string, limit int) ([]monitorCs.NetworkProcessInfo, error) {
	conns, err := net.Connections("all")
	if err != nil {
		return nil, err
	}

	if sortDirection == "" {
		sortDirection = monitorCs.SortDirectionDesc
	}
	isDesc := sortDirection == monitorCs.SortDirectionDesc

	processMap := make(map[int32]*monitorCs.NetworkProcessInfo)
	totalConns := len(conns)

	for _, conn := range conns {
		if conn.Pid <= 0 {
			continue
		}

		if info, exists := processMap[conn.Pid]; exists {
			info.Connections++
		} else {
			protocol := b.getProtocolName(conn.Type)
			state := b.getConnectionState(conn.Status)

			processMap[conn.Pid] = &monitorCs.NetworkProcessInfo{
				PID:         conn.Pid,
				LocalPort:   int(conn.Laddr.Port),
				RemotePort:  int(conn.Raddr.Port),
				Protocol:    protocol,
				State:       state,
				Connections: 1,
			}

			if conn.Pid > 0 {
				p, err := process.NewProcess(conn.Pid)
				if err == nil {
					name, _ := p.Name()
					processMap[conn.Pid].Name = name
				}
			}
		}
	}

	var result []monitorCs.NetworkProcessInfo
	for _, info := range processMap {
		if totalConns > 0 {
			info.UsagePercent = float64(info.Connections) / float64(totalConns) * 100
		}
		result = append(result, *info)
	}

	switch sortBy {
	case "usage":
		sort.Slice(result, func(i, j int) bool {
			if isDesc {
				return result[i].UsagePercent > result[j].UsagePercent
			}
			return result[i].UsagePercent < result[j].UsagePercent
		})
	case "connections":
		sort.Slice(result, func(i, j int) bool {
			if isDesc {
				return result[i].Connections > result[j].Connections
			}
			return result[i].Connections < result[j].Connections
		})
	case "port":
		sort.Slice(result, func(i, j int) bool {
			if isDesc {
				return result[i].LocalPort > result[j].LocalPort
			}
			return result[i].LocalPort < result[j].LocalPort
		})
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

// FormatConnectionsOutput 格式化连接输出 (用于命令行)
func (b *networkBusiness) FormatConnectionsOutput(connections []monitorCs.ConnectionInfo) string {
	if len(connections) == 0 {
		return "暂无网络连接"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-8s %-22s %-22s %-12s %s\n",
		"协议", "本地地址", "远程地址", "状态", "PID"))
	sb.WriteString(strings.Repeat("-", 80) + "\n")

	count := 0
	for _, conn := range connections {
		if count >= 20 {
			sb.WriteString(fmt.Sprintf("... 还有 %d 个连接 ...\n", len(connections)-20))
			break
		}

		local := fmt.Sprintf("%s:%d", conn.LocalAddr, conn.LocalPort)
		remote := fmt.Sprintf("%s:%d", conn.RemoteAddr, conn.RemotePort)
		if conn.RemotePort == 0 {
			remote = conn.RemoteAddr
		}

		pidStr := "-"
		if conn.PID > 0 {
			pidStr = strconv.Itoa(int(conn.PID))
		}

		sb.WriteString(fmt.Sprintf("%-8s %-22s %-22s %-12s %s\n",
			conn.Protocol, local, remote, conn.State, pidStr))
		count++
	}

	return sb.String()
}

// FormatPortsOutput 格式化端口输出 (用于命令行)
func (b *networkBusiness) FormatPortsOutput(ports []monitorCs.PortInfo) string {
	if len(ports) == 0 {
		return "暂无端口占用"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-8s %-8s %-22s %-12s %-10s %s\n",
		"协议", "端口", "地址", "状态", "PID", "进程"))
	sb.WriteString(strings.Repeat("-", 80) + "\n")

	for _, p := range ports {
		process := p.Process
		if process == "" {
			process = "-"
		}
		pidStr := "-"
		if p.PID > 0 {
			pidStr = strconv.Itoa(int(p.PID))
		}

		sb.WriteString(fmt.Sprintf("%-8s %-8d %-22s %-12s %-10s %s\n",
			p.Protocol, p.Port, p.LocalAddr, p.State, pidStr, process))
	}

	return sb.String()
}

// FormatInterfacesOutput 格式化网络接口输出 (用于命令行)
func (b *networkBusiness) FormatInterfacesOutput(interfaces []monitorCs.InterfaceInfo) string {
	if len(interfaces) == 0 {
		return "暂无网络接口信息"
	}

	var sb strings.Builder
	for _, iface := range interfaces {
		sb.WriteString(fmt.Sprintf("接口: %s\n", iface.Name))
		sb.WriteString(fmt.Sprintf("  MAC:  %s\n", iface.Hardware))
		sb.WriteString(fmt.Sprintf("  MTU:  %d\n", iface.MTU))
		sb.WriteString(fmt.Sprintf("  标志: %s\n", iface.Flags))
		sb.WriteString("  地址:\n")
		for _, addr := range iface.Addrs {
			sb.WriteString(fmt.Sprintf("    %s\n", addr))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
