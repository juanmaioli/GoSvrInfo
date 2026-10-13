package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type DiskSpace struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Free string `json:"free"`
}

type NetworkInterface struct {
	Name      string `json:"name"`
	MacAddr   string `json:"mac_address"`
	IPAddress string `json:"ip_address"`
}

type OperatingSystem struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	KernelVersion string `json:"kernel_version"`
}

type GPUInfo struct {
	Name          string `json:"name"`
	DriverVersion string `json:"driver_version"`
	Memory        string `json:"memory"`
}

type HostInfo struct {
	Time             int64              `json:"time"`
	SystemUptime     int64              `json:"system_uptime"`
	Hostname         string             `json:"hostname"`
	IPPublicAddress  string             `json:"ip_public_address"`
	City             string             `json:"city"`
	Country          string             `json:"country"`
	Latitude         float64            `json:"latitude"`
	Longitude        float64            `json:"longitude"`
	Timezone         string             `json:"timezone"`
	ISP              string             `json:"isp"`
	CPU              string             `json:"CPU"`
	Cores            int                `json:"cores"`
	CPUSpeed         string             `json:"cpu_speed"`
	CPUUsage         string             `json:"cpu_usage"`
	RAM              string             `json:"RAM"`
	RAMUsage         string             `json:"ram_usage"`
	DiskSpace        []DiskSpace        `json:"disk_space"`
	NetworkInterfaces []NetworkInterface `json:"network_interfaces"`
	OperatingSystem  OperatingSystem    `json:"operating_system"`
	GPU              GPUInfo            `json:"GPU"`
}

type IPApiResponse struct {
	Status   string  `json:"status"`
	Country  string  `json:"country"`
	City     string  `json:"city"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Timezone string  `json:"timezone"`
	ISP      string  `json:"isp"`
	Query    string  `json:"query"`
}

func getUptime() int64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) > 0 {
		val, err := strconv.ParseFloat(fields[0], 64)
		if err == nil {
			return int64(val)
		}
	}
	return 0
}

func getCPUInfo() (string, int, string) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "Desconocido", 0, "N/A"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	modelName := ""
	cpuMHz := 0.0
	cores := 0

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		if key == "model name" && modelName == "" {
			modelName = val
		}
		if key == "cpu MHz" && cpuMHz == 0.0 {
			if mhz, err := strconv.ParseFloat(val, 64); err == nil {
				cpuMHz = mhz
			}
		}
		if key == "processor" {
			cores++
		}
	}

	speed := "N/A"
	if cpuMHz > 0 {
		speed = fmt.Sprintf("%.2f GHz", cpuMHz/1000.0)
	}

	return modelName, cores, speed
}

func readCPUSample() (idle, total uint64, err error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[0] == "cpu" {
			var vals []uint64
			for _, f := range fields[1:] {
				v, err := strconv.ParseUint(f, 10, 64)
				if err == nil {
					vals = append(vals, v)
				}
			}
			if len(vals) >= 4 {
				// user, nice, system, idle, iowait...
				var sum uint64
				for _, v := range vals {
					sum += v
				}
				idleVal := vals[3]
				if len(vals) >= 5 {
					idleVal += vals[4] // include iowait in idle
				}
				return idleVal, sum, nil
			}
		}
	}
	return 0, 0, fmt.Errorf("no cpu found")
}

func getCPUUsage() string {
	idle1, total1, err := readCPUSample()
	if err != nil {
		return "0%"
	}
	time.Sleep(200 * time.Millisecond)
	idle2, total2, err := readCPUSample()
	if err != nil || total2 <= total1 {
		return "0%"
	}

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)
	usage := 100.0 * (1.0 - (idleDelta / totalDelta))
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	return fmt.Sprintf("%.0f%%", usage)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.0f %s", float64(bytes)/float64(div), units[exp])
}

func getRAMInfo() (string, string) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return "0 GB", "0%"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var totalKB, availKB uint64

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 {
			key := strings.TrimSuffix(fields[0], ":")
			val, _ := strconv.ParseUint(fields[1], 10, 64)
			if key == "MemTotal" {
				totalKB = val
			} else if key == "MemAvailable" {
				availKB = val
			}
		}
	}

	if totalKB == 0 {
		return "0 GB", "0%"
	}

	usedKB := totalKB - availKB
	usagePct := (float64(usedKB) / float64(totalKB)) * 100.0
	totalBytes := totalKB * 1024

	return formatBytes(totalBytes), fmt.Sprintf("%.0f%%", usagePct)
}

func getDisks() []DiskSpace {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil
	}
	defer file.Close()

	var disks []DiskSpace
	seenDevs := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		dev := fields[0]
		mountPoint := fields[1]

		if !strings.HasPrefix(dev, "/dev/") || seenDevs[dev] {
			continue
		}
		// Ignorar loop devices y ramdisks
		if strings.HasPrefix(dev, "/dev/loop") || strings.HasPrefix(dev, "/dev/ram") {
			continue
		}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountPoint, &stat); err != nil {
			continue
		}

		totalBytes := stat.Blocks * uint64(stat.Bsize)
		freeBytes := stat.Bavail * uint64(stat.Bsize)

		if totalBytes == 0 {
			continue
		}

		seenDevs[dev] = true
		disks = append(disks, DiskSpace{
			Name: filepath.Base(dev),
			Size: formatBytes(totalBytes),
			Free: formatBytes(freeBytes),
		})
	}

	return disks
}

func getNetworkInterfaces() []NetworkInterface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var res []NetworkInterface
	for _, iface := range ifaces {
		// Ignorar loopback, caídas e interfaces virtuales de contenedores (veth/docker/br-)
		if (iface.Flags&net.FlagLoopback != 0) || (iface.Flags&net.FlagUp == 0) {
			continue
		}
		if strings.HasPrefix(iface.Name, "veth") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		ipStr := ""
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				ipStr = ip.String()
				break
			}
		}

		if ipStr != "" || iface.HardwareAddr.String() != "" {
			res = append(res, NetworkInterface{
				Name:      iface.Name,
				MacAddr:   iface.HardwareAddr.String(),
				IPAddress: ipStr,
			})
		}
	}
	return res
}

func getOSInfo() OperatingSystem {
	osInfo := OperatingSystem{
		Name:          "Linux",
		Version:       "Unknown",
		KernelVersion: "Unknown",
	}

	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err == nil {
		charsToByte := func(chars [65]int8) string {
			buf := make([]byte, 0, len(chars))
			for _, b := range chars {
				if b == 0 {
					break
				}
				buf = append(buf, byte(b))
			}
			return string(buf)
		}
		osInfo.KernelVersion = charsToByte(uname.Release)
	}

	file, err := os.Open("/etc/os-release")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				val := strings.Trim(parts[1], `"`)
				if key == "NAME" && osInfo.Name == "Linux" {
					osInfo.Name = val
				} else if key == "VERSION" {
					osInfo.Version = val
				}
			}
		}
	}

	return osInfo
}

func getGPUInfo() GPUInfo {
	gpu := GPUInfo{
		Name:          "N/A",
		DriverVersion: "N/A",
		Memory:        "N/A",
	}

	// 1. Probar nvidia-smi
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,driver_version,memory.total", "--format=csv,noheader,nounits")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		fields := strings.Split(strings.TrimSpace(out.String()), ",")
		if len(fields) >= 3 {
			gpu.Name = strings.TrimSpace(fields[0])
			gpu.DriverVersion = strings.TrimSpace(fields[1])
			memMB, err := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64)
			if err == nil {
				gpu.Memory = formatBytes(memMB * 1024 * 1024)
			} else {
				gpu.Memory = strings.TrimSpace(fields[2]) + " MB"
			}
			return gpu
		}
	}

	// 2. Si no hay nvidia-smi, chequear drm / lspci
	pciCmd := exec.Command("sh", "-c", "lspci | grep -i 'vga\\|3d\\|display'")
	var pciOut bytes.Buffer
	pciCmd.Stdout = &pciOut
	if err := pciCmd.Run(); err == nil && len(pciOut.String()) > 0 {
		lines := strings.Split(strings.TrimSpace(pciOut.String()), "\n")
		if len(lines) > 0 {
			line := lines[0]
			idx := strings.Index(line, ": ")
			if idx != -1 {
				gpu.Name = strings.TrimSpace(line[idx+2:])
			} else {
				gpu.Name = line
			}
		}
	}

	return gpu
}

func getGeoAndPublicIP() IPApiResponse {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get("http://ip-api.com/json/")
	if err != nil {
		return IPApiResponse{}
	}
	defer resp.Body.Close()

	var data IPApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return IPApiResponse{}
	}
	return data
}

func main() {
	now := time.Now().Unix()
	uptime := getUptime()

	hostname, _ := os.Hostname()
	geo := getGeoAndPublicIP()

	cpuModel, cores, cpuSpeed := getCPUInfo()
	cpuUsage := getCPUUsage()
	ramTotal, ramUsage := getRAMInfo()
	disks := getDisks()
	netIfaces := getNetworkInterfaces()
	osInfo := getOSInfo()
	gpuInfo := getGPUInfo()

	host := HostInfo{
		Time:              now,
		SystemUptime:      uptime,
		Hostname:          hostname,
		IPPublicAddress:   geo.Query,
		City:              geo.City,
		Country:           geo.Country,
		Latitude:          geo.Lat,
		Longitude:         geo.Lon,
		Timezone:          geo.Timezone,
		ISP:               geo.ISP,
		CPU:               cpuModel,
		Cores:             cores,
		CPUSpeed:          cpuSpeed,
		CPUUsage:          cpuUsage,
		RAM:               ramTotal,
		RAMUsage:          ramUsage,
		DiskSpace:         disks,
		NetworkInterfaces: netIfaces,
		OperatingSystem:   osInfo,
		GPU:               gpuInfo,
	}

	jsonBytes, err := json.MarshalIndent(host, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializando JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))
}
