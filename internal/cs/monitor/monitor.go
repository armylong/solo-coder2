package monitor

// 排序方向
const (
	SortDirectionAsc  = "asc"  // 升序
	SortDirectionDesc = "desc" // 降序
)

// 基础请求
type BaseRequest struct{}

// 杀进程-请求
type KillProcessRequest struct {
	BaseRequest
	PID int32 `json:"pid" form:"pid"` // 进程ID
}

// 杀端口进程-请求
type KillPortRequest struct {
	BaseRequest
	Port int `json:"port" form:"port"` // 端口号
}

// 排序+分页-请求
type SortLimitRequest struct {
	BaseRequest
	SortBy        string `json:"sort_by" form:"sort_by"`                 // 排序字段
	SortDirection string `json:"sort_direction" form:"sort_direction"` // 排序方向: asc/desc
	Limit         int    `json:"limit" form:"limit"`                     // 返回数量
}

// 端口筛选-请求
type PortFilterRequest struct {
	BaseRequest
	Port          int    `json:"port" form:"port"`                       // 端口筛选
	SortBy        string `json:"sort_by" form:"sort_by"`                 // 排序字段
	SortDirection string `json:"sort_direction" form:"sort_direction"` // 排序方向
	Limit         int    `json:"limit" form:"limit"`                     // 返回数量
}

// CPU基本信息
type CPUInfo struct {
	ModelName     string   `json:"model_name"`     // 型号
	PhysicalCores int      `json:"physical_cores"` // 物理核心数
	LogicalCores  int      `json:"logical_cores"`  // 逻辑核心数
	Frequency     float64  `json:"frequency"`      // 主频(MHz)
	VendorID      string   `json:"vendor_id"`      // 厂商
	CacheSize     int32    `json:"cache_size"`     // 缓存(KB)
	Flags         []string `json:"flags"`          // 特性标志
}

// CPU使用率
type CPUUsage struct {
	User         float64   `json:"user"`           // 用户态(%)
	System       float64   `json:"system"`         // 系统态(%)
	Idle         float64   `json:"idle"`           // 空闲(%)
	Nice         float64   `json:"nice"`           // 低优先级(%)
	IOWait       float64   `json:"iowait"`         // IO等待(%)
	IRQ          float64   `json:"irq"`            // 硬件中断(%)
	SoftIRQ      float64   `json:"softirq"`        // 软件中断(%)
	Steal        float64   `json:"steal"`          // 虚拟机偷取(%)
	Guest        float64   `json:"guest"`          // 虚拟机(%)
	TotalUsage   float64   `json:"total_usage"`    // 总使用率(%)
	PerCoreUsage []float64 `json:"per_core_usage"` // 各核心使用率(%)
}

// CPU占用进程
type CPUProcessInfo struct {
	PID        int32   `json:"pid"`         // 进程ID
	Name       string  `json:"name"`        // 进程名
	CmdLine    string  `json:"cmd_line"`    // 完整命令行
	CPU        float64 `json:"cpu"`         // CPU使用率(%)
	Memory     float32 `json:"memory"`      // 内存使用率(%)
	MemoryMB   float32 `json:"memory_mb"`   // 内存(MB)
	User       string  `json:"user"`        // 运行用户
	StartTime  string  `json:"start_time"`  // 启动时间
	CreateTime int64   `json:"create_time"` // 创建时间戳(ms)
}

// CPU进程列表-响应
type CPUProcessListResponse struct {
	List  []CPUProcessInfo `json:"list"`  // 进程列表
	Total int              `json:"total"` // 总数
}

// 内存信息
type MemoryInfo struct {
	Total       uint64  `json:"total"`        // 总内存(字节)
	Used        uint64  `json:"used"`         // 已用(字节)
	Free        uint64  `json:"free"`         // 空闲(字节)
	Shared      uint64  `json:"shared"`       // 共享(字节)
	Buffers     uint64  `json:"buffers"`      // 缓冲区(字节)
	Cached      uint64  `json:"cached"`       // 缓存(字节)
	Available   uint64  `json:"available"`    // 可用(字节)
	UsedPercent float64 `json:"used_percent"` // 使用率(%)
	SwapTotal   uint64  `json:"swap_total"`   // 交换总量(字节)
	SwapUsed    uint64  `json:"swap_used"`    // 交换已用(字节)
	SwapFree    uint64  `json:"swap_free"`    // 交换空闲(字节)
}

// 磁盘信息
type DiskInfo struct {
	Device      string  `json:"device"`       // 设备名
	MountPoint  string  `json:"mount_point"`  // 挂载点
	FileSystem  string  `json:"file_system"`  // 文件系统
	Total       uint64  `json:"total"`        // 总容量(字节)
	Used        uint64  `json:"used"`         // 已用(字节)
	Free        uint64  `json:"free"`         // 空闲(字节)
	UsedPercent float64 `json:"used_percent"` // 使用率(%)
}

// 磁盘列表-响应
type DiskListResponse struct {
	List  []DiskInfo `json:"list"`  // 磁盘列表
	Total int        `json:"total"` // 总数
}

// 网络连接
type ConnectionInfo struct {
	Protocol   string `json:"protocol"`    // 协议: tcp/udp
	LocalAddr  string `json:"local_addr"`  // 本地地址
	LocalPort  int    `json:"local_port"`  // 本地端口
	RemoteAddr string `json:"remote_addr"` // 远程地址
	RemotePort int    `json:"remote_port"` // 远程端口
	State      string `json:"state"`       // 状态
	PID        int32  `json:"pid"`         // 进程ID
	Process    string `json:"process"`     // 进程名
}

// 端口信息
type PortInfo struct {
	Protocol  string `json:"protocol"`   // 协议
	Port      int    `json:"port"`       // 端口号
	LocalAddr string `json:"local_addr"` // 绑定地址
	State     string `json:"state"`      // 状态
	PID       int32  `json:"pid"`        // 进程ID
	Process   string `json:"process"`    // 进程名
}

// 网络接口
type InterfaceInfo struct {
	Name     string   `json:"name"`     // 接口名
	Addrs    []string `json:"addrs"`    // IP地址列表
	Flags    string   `json:"flags"`    // 标志
	MTU      int      `json:"mtu"`      // MTU值
	Hardware string   `json:"hardware"` // MAC地址
}

// 网络带宽
type NetworkBandwidthInfo struct {
	InterfaceName string  `json:"interface_name"` // 接口名
	BytesSent     uint64  `json:"bytes_sent"`     // 累计发送(字节)
	BytesRecv     uint64  `json:"bytes_recv"`     // 累计接收(字节)
	PacketsSent   uint64  `json:"packets_sent"`   // 累计发送包数
	PacketsRecv   uint64  `json:"packets_recv"`   // 累计接收包数
	SendRate      float64 `json:"send_rate"`      // 发送速率(字节/秒)
	RecvRate      float64 `json:"recv_rate"`      // 接收速率(字节/秒)
	TotalRate     float64 `json:"total_rate"`     // 总速率(字节/秒)
	UsagePercent  float64 `json:"usage_percent"`  // 使用率(%)
}

// 网络带宽列表-响应
type NetworkBandwidthListResponse struct {
	List  []NetworkBandwidthInfo `json:"list"`  // 带宽列表
	Total int                    `json:"total"` // 总数
}

// 网络占用进程
type NetworkProcessInfo struct {
	PID          int32   `json:"pid"`           // 进程ID
	Name         string  `json:"name"`          // 进程名
	LocalPort    int     `json:"local_port"`    // 本地端口
	RemotePort   int     `json:"remote_port"`   // 远程端口
	Protocol     string  `json:"protocol"`      // 协议
	State        string  `json:"state"`         // 状态
	Connections  int     `json:"connections"`   // 连接数
	UsagePercent float64 `json:"usage_percent"` // 占用百分比
}

// 网络进程列表-响应
type NetworkProcessListResponse struct {
	List  []NetworkProcessInfo `json:"list"`  // 进程列表
	Total int                   `json:"total"` // 总数
}

// 端口列表-响应
type PortListResponse struct {
	List  []PortInfo `json:"list"`  // 端口列表
	Total int        `json:"total"` // 总数
}

// GPU信息
type GPUInfo struct {
	Index         int     `json:"index"`          // GPU索引
	Vendor        string  `json:"vendor"`         // 厂商: NVIDIA/AMD
	Model         string  `json:"model"`          // 型号
	MemoryTotal   uint64  `json:"memory_total"`   // 显存总量(MB)
	MemoryUsed    uint64  `json:"memory_used"`    // 显存已用(MB)
	MemoryFree    uint64  `json:"memory_free"`    // 显存空闲(MB)
	UsagePercent  float64 `json:"usage_percent"`  // GPU使用率(%)
	MemoryPercent float64 `json:"memory_percent"` // 显存使用率(%)
	Temperature   float64 `json:"temperature"`    // 温度(°C)
}

// GPU列表-响应
type GPUListResponse struct {
	List  []GPUInfo `json:"list"`  // GPU列表
	Total int       `json:"total"` // 总数
}

// 进程信息
type ProcessInfo struct {
	PID        int32   `json:"pid"`         // 进程ID
	Name       string  `json:"name"`        // 进程名
	CmdLine    string  `json:"cmd_line"`    // 完整命令行
	CPU        float64 `json:"cpu"`         // CPU使用率(%)
	Memory     float32 `json:"memory"`      // 内存使用率(%)
	MemoryMB   float32 `json:"memory_mb"`   // 内存(MB)
	Status     string  `json:"status"`      // 状态
	PPID       int32   `json:"ppid"`        // 父进程ID
	NumThreads int32   `json:"num_threads"` // 线程数
	User       string  `json:"user"`        // 运行用户
	StartTime  string  `json:"start_time"`  // 启动时间
	CreateTime int64   `json:"create_time"` // 创建时间戳(ms)
}

// 进程列表-响应
type ProcessListResponse struct {
	List  []ProcessInfo `json:"list"`  // 进程列表
	Total int           `json:"total"` // 总数
}

// 系统信息
type SystemInfo struct {
	Hostname        string `json:"hostname"`         // 主机名
	OS              string `json:"os"`               // 操作系统
	Platform        string `json:"platform"`         // 平台
	PlatformVersion string `json:"platform_version"` // 平台版本
	KernelVersion   string `json:"kernel_version"`   // 内核版本
	Architecture    string `json:"architecture"`     // 架构
	CPUCount        int    `json:"cpu_count"`        // CPU核心数
	GoVersion       string `json:"go_version"`       // Go版本
}

// 运行时间
type UptimeInfo struct {
	Uptime    uint64 `json:"uptime"`     // 运行秒数
	BootTime  uint64 `json:"boot_time"`  // 启动时间戳
	UptimeStr string `json:"uptime_str"` // 格式化运行时间
}
