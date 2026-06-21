package monitor

import (
	"errors"

	monitorBusiness "github.com/armylong/armylong-go/internal/business/monitor"
	monitorCs "github.com/armylong/armylong-go/internal/cs/monitor"
)

// 系统监控控制器
type MonitorController struct{}

// 获取CPU基本信息
func (c *MonitorController) ActionGetCPUInfo(req *monitorCs.BaseRequest) (*monitorCs.CPUInfo, error) {
	return monitorBusiness.CPUBusiness.GetCPUInfo()
}

// 获取CPU使用率，采样1秒
func (c *MonitorController) ActionGetCPUUsage(req *monitorCs.BaseRequest) (*monitorCs.CPUUsage, error) {
	return monitorBusiness.CPUBusiness.GetCPUUsage()
}

// 获取CPU占用进程列表
func (c *MonitorController) ActionGetCPUProcesses(req *monitorCs.SortLimitRequest) (*monitorCs.CPUProcessListResponse, error) {
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "cpu"
	}

	sortDirection := req.SortDirection
	if sortDirection == "" {
		sortDirection = monitorCs.SortDirectionDesc
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	processes, err := monitorBusiness.CPUBusiness.GetCPUProcesses(sortBy, sortDirection, limit)
	if err != nil {
		return nil, err
	}

	return &monitorCs.CPUProcessListResponse{
		List:  processes,
		Total: len(processes),
	}, nil
}

// 获取内存使用情况
func (c *MonitorController) ActionGetMemoryInfo(req *monitorCs.BaseRequest) (*monitorCs.MemoryInfo, error) {
	return monitorBusiness.MemoryBusiness.GetMemoryUsage()
}

// 获取磁盘使用情况
func (c *MonitorController) ActionGetDiskInfo(req *monitorCs.BaseRequest) (*monitorCs.DiskListResponse, error) {
	disks, err := monitorBusiness.DiskBusiness.GetDiskUsage()
	if err != nil {
		return nil, err
	}

	return &monitorCs.DiskListResponse{
		List:  disks,
		Total: len(disks),
	}, nil
}

// 获取实时网络带宽，首次调用返回基础数据，第二次开始计算速率
func (c *MonitorController) ActionGetNetworkBandwidth(req *monitorCs.BaseRequest) (*monitorCs.NetworkBandwidthListResponse, error) {
	bandwidths, err := monitorBusiness.NetworkBusiness.GetNetworkBandwidth()
	if err != nil {
		return nil, err
	}

	return &monitorCs.NetworkBandwidthListResponse{
		List:  bandwidths,
		Total: len(bandwidths),
	}, nil
}

// 获取网络占用进程列表
func (c *MonitorController) ActionGetNetworkProcesses(req *monitorCs.SortLimitRequest) (*monitorCs.NetworkProcessListResponse, error) {
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "usage"
	}

	sortDirection := req.SortDirection
	if sortDirection == "" {
		sortDirection = monitorCs.SortDirectionDesc
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	processes, err := monitorBusiness.NetworkBusiness.GetNetworkProcesses(sortBy, sortDirection, limit)
	if err != nil {
		return nil, err
	}

	return &monitorCs.NetworkProcessListResponse{
		List:  processes,
		Total: len(processes),
	}, nil
}

// 获取端口信息，支持筛选和排序
func (c *MonitorController) ActionGetPortInfo(req *monitorCs.PortFilterRequest) (*monitorCs.PortListResponse, error) {
	portFilter := req.Port
	sortBy := req.SortBy

	sortDirection := req.SortDirection
	if sortDirection == "" {
		sortDirection = monitorCs.SortDirectionDesc
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	ports, err := monitorBusiness.NetworkBusiness.GetPortsWithFilter(portFilter, sortBy, sortDirection, limit)
	if err != nil {
		return nil, err
	}

	return &monitorCs.PortListResponse{
		List:  ports,
		Total: len(ports),
	}, nil
}

// 获取GPU信息，支持NVIDIA和AMD
func (c *MonitorController) ActionGetGPUInfo(req *monitorCs.BaseRequest) (*monitorCs.GPUListResponse, error) {
	gpus, err := monitorBusiness.GPUBusiness.GetGPUInfo()
	if err != nil {
		return &monitorCs.GPUListResponse{
			List:  []monitorCs.GPUInfo{},
			Total: 0,
		}, nil
	}

	return &monitorCs.GPUListResponse{
		List:  gpus,
		Total: len(gpus),
	}, nil
}

// 获取进程列表，支持排序和限制数量
func (c *MonitorController) ActionGetProcessList(req *monitorCs.SortLimitRequest) (*monitorCs.ProcessListResponse, error) {
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "cpu"
	}

	sortDirection := req.SortDirection
	if sortDirection == "" {
		sortDirection = monitorCs.SortDirectionDesc
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	processes, err := monitorBusiness.ProcessBusiness.GetProcessList(sortBy, sortDirection, limit)
	if err != nil {
		return nil, err
	}

	return &monitorCs.ProcessListResponse{
		List:  processes,
		Total: len(processes),
	}, nil
}

// 获取系统信息
func (c *MonitorController) ActionGetSystemInfo(req *monitorCs.BaseRequest) (*monitorCs.SystemInfo, error) {
	return monitorBusiness.SystemBusiness.GetSystemInfo()
}

// 获取系统运行时间
func (c *MonitorController) ActionGetUptime(req *monitorCs.BaseRequest) (*monitorCs.UptimeInfo, error) {
	return monitorBusiness.SystemBusiness.GetUptime()
}

// 按PID杀进程
func (c *MonitorController) ActionKillProcessByPID(req *monitorCs.KillProcessRequest) error {
	if req.PID <= 0 {
		return errors.New("无效的进程ID")
	}

	return monitorBusiness.ProcessBusiness.KillProcess(req.PID)
}

// 按端口号杀掉占用进程
func (c *MonitorController) ActionKillProcessByPort(req *monitorCs.KillPortRequest) error {
	if req.Port <= 0 || req.Port > 65535 {
		return errors.New("无效的端口号")
	}

	return monitorBusiness.NetworkBusiness.KillProcessByPort(req.Port)
}
