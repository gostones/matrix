package util

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"runtime"
)

type MyInfo struct {
	Os   string `json:"os"`
	Arch string `json:"arch"`

	Cpu    CpuStat    `json:"cpu"`
	Memory MemoryStat `json:"memory"`
	Disk   DiskStat   `json:"disk"`
	Host   HostStat   `json:"host"`
}

type HostStat struct {
	Hostname string `json:"hostname"`
	Uptime   uint64 `json:"uptime"`
}

type CpuStat struct {
	ModelName   string  `json:"modelName"`
	Cores       int32   `json:"cores"`
	Mhz         float64 `json:"mhz"`
	UsedPercent float64 `json:"usedPercent"`
}

type MemoryStat struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
	Free        uint64  `json:"free"`
}

type DiskStat struct {
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

func GetMyInfo() *MyInfo {
	my := &MyInfo{}

	my.Os = runtime.GOOS
	my.Arch = runtime.GOARCH

	// cpu
	if v, err := cpu.Info(); err == nil {
		my.Cpu.Cores = v[0].Cores
		my.Cpu.ModelName = v[0].ModelName
		my.Cpu.Mhz = v[0].Mhz
	}
	if v, err := cpu.Percent(0, true); err == nil {
		var percent float64
		for _, p := range v {
			percent += p
		}
		percent = percent / float64(len(v))
		my.Cpu.UsedPercent = percent
	}

	// memory
	if v, err := mem.VirtualMemory(); err == nil {
		my.Memory.Total = v.Total
		my.Memory.Available = v.Available
		my.Memory.Used = v.Used
		my.Memory.UsedPercent = v.UsedPercent
		my.Memory.Free = v.Free
	}

	// disk
	if v, err := disk.Usage("/"); err == nil {
		my.Disk.Fstype = v.Fstype
		my.Disk.Total = v.Total
		my.Disk.Free = v.Free
		my.Disk.Used = v.Used
		my.Disk.UsedPercent = v.UsedPercent
	}

	// host or machine kernel, uptime, platform Info
	if v, err := host.Info(); err == nil {
		my.Host.Hostname = v.Hostname
		my.Host.Uptime = v.Uptime
	}

	// wifi network
	return my
}
