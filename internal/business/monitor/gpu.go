package monitor

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	monitorCs "github.com/armylong/armylong-go/internal/cs/monitor"
)

type gpuBusiness struct{}

var GPUBusiness = &gpuBusiness{}

// 获取GPU信息，支持NVIDIA和AMD
func (b *gpuBusiness) GetGPUInfo() ([]monitorCs.GPUInfo, error) {
	var gpus []monitorCs.GPUInfo

	nvidiaGPUs, err := b.getNvidiaGPUInfo()
	if err == nil && len(nvidiaGPUs) > 0 {
		gpus = append(gpus, nvidiaGPUs...)
	}

	amdGPUs, err := b.getAMDGPUInfo()
	if err == nil && len(amdGPUs) > 0 {
		gpus = append(gpus, amdGPUs...)
	}

	if len(gpus) == 0 {
		return nil, fmt.Errorf("未检测到GPU设备")
	}

	return gpus, nil
}

// 通过nvidia-smi获取NVIDIA GPU信息
func (b *gpuBusiness) getNvidiaGPUInfo() ([]monitorCs.GPUInfo, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.total,memory.used,memory.free,utilization.gpu,utilization.memory,temperature.gpu", "--format=csv,noheader,nounits")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var gpus []monitorCs.GPUInfo

	for i, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ", ")
		if len(fields) < 8 {
			continue
		}

		gpu := monitorCs.GPUInfo{
			Vendor: "NVIDIA",
		}

		index, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err == nil {
			gpu.Index = index
		} else {
			gpu.Index = i
		}

		gpu.Model = strings.TrimSpace(fields[1])

		memTotal, _ := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64)
		gpu.MemoryTotal = memTotal

		memUsed, _ := strconv.ParseUint(strings.TrimSpace(fields[3]), 10, 64)
		gpu.MemoryUsed = memUsed

		memFree, _ := strconv.ParseUint(strings.TrimSpace(fields[4]), 10, 64)
		gpu.MemoryFree = memFree

		usage, _ := strconv.ParseFloat(strings.TrimSpace(fields[5]), 64)
		gpu.UsagePercent = usage

		memPercent, _ := strconv.ParseFloat(strings.TrimSpace(fields[6]), 64)
		gpu.MemoryPercent = memPercent

		temp, _ := strconv.ParseFloat(strings.TrimSpace(fields[7]), 64)
		gpu.Temperature = temp

		gpus = append(gpus, gpu)
	}

	return gpus, nil
}

// 获取AMD GPU信息，Linux用rocm-smi，Windows用WMIC
func (b *gpuBusiness) getAMDGPUInfo() ([]monitorCs.GPUInfo, error) {
	var gpus []monitorCs.GPUInfo

	if runtime.GOOS == "linux" {
		amdGPUs, err := b.getAMDGPUInfoViaRocm()
		if err == nil && len(amdGPUs) > 0 {
			gpus = append(gpus, amdGPUs...)
		}
	}

	if runtime.GOOS == "windows" {
		amdGPUs, err := b.getAMDGPUInfoViaWMIC()
		if err == nil && len(amdGPUs) > 0 {
			gpus = append(gpus, amdGPUs...)
		}
	}

	return gpus, nil
}

// 通过rocm-smi获取AMD GPU信息
func (b *gpuBusiness) getAMDGPUInfoViaRocm() ([]monitorCs.GPUInfo, error) {
	cmd := exec.Command("rocm-smi", "--showhw", "--csv")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var gpus []monitorCs.GPUInfo

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}

		gpu := monitorCs.GPUInfo{
			Vendor: "AMD",
			Index:  i - 1,
			Model:  strings.TrimSpace(fields[0]),
		}

		gpus = append(gpus, gpu)
	}

	return gpus, nil
}

// 通过WMIC获取Windows下AMD GPU信息
func (b *gpuBusiness) getAMDGPUInfoViaWMIC() ([]monitorCs.GPUInfo, error) {
	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "name,adapterram", "/format:csv")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var gpus []monitorCs.GPUInfo

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 3 {
			continue
		}

		model := strings.TrimSpace(fields[1])
		if strings.Contains(strings.ToLower(model), "amd") || strings.Contains(strings.ToLower(model), "radeon") {
			gpu := monitorCs.GPUInfo{
				Vendor: "AMD",
				Index:  len(gpus),
				Model:  model,
			}

			memBytes, _ := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64)
			gpu.MemoryTotal = memBytes / 1024 / 1024

			gpus = append(gpus, gpu)
		}
	}

	return gpus, nil
}

// FormatGPUInfoOutput 格式化GPU信息输出 (用于命令行)
func (b *gpuBusiness) FormatGPUInfoOutput(gpus []monitorCs.GPUInfo) string {
	if len(gpus) == 0 {
		return "未检测到GPU设备"
	}

	var sb strings.Builder
	for _, gpu := range gpus {
		sb.WriteString(fmt.Sprintf("===== GPU %d =====\n", gpu.Index))
		sb.WriteString(fmt.Sprintf("厂商:       %s\n", gpu.Vendor))
		sb.WriteString(fmt.Sprintf("型号:       %s\n", gpu.Model))
		if gpu.MemoryTotal > 0 {
			sb.WriteString(fmt.Sprintf("显存总量:   %.1f MB\n", float64(gpu.MemoryTotal)))
			sb.WriteString(fmt.Sprintf("显存已用:   %.1f MB\n", float64(gpu.MemoryUsed)))
			sb.WriteString(fmt.Sprintf("显存可用:   %.1f MB\n", float64(gpu.MemoryFree)))
		}
		sb.WriteString(fmt.Sprintf("GPU使用率:  %.1f%%\n", gpu.UsagePercent))
		if gpu.MemoryPercent > 0 {
			sb.WriteString(fmt.Sprintf("显存使用率: %.1f%%\n", gpu.MemoryPercent))
		}
		if gpu.Temperature > 0 {
			sb.WriteString(fmt.Sprintf("温度:       %.1f°C\n", gpu.Temperature))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
