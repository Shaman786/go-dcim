// Package metrics provides low-level functions to parse Linux kernel data (/proc)
// for CPU and Memory usage statistics.
package metrics

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ReadMemoryStats returns Total, Used, and Free memory in bytes
// We return 'free' now so it is no longer an "unused variable"
func ReadMemoryStats() (uint64, uint64, uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()

	var total, free, available uint64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSuffix(parts[0], ":")
		value, _ := strconv.ParseUint(parts[1], 10, 64)

		switch key {
		case "MemTotal":
			total = value * 1024
		case "MemFree":
			free = value * 1024 // Now we are actually storing this!
		case "MemAvailable":
			available = value * 1024
		}
	}

	// Used = Total - Available (Standard Linux calculation)
	used := total - available

	// We return 'free' as the 3rd value
	return total, used, free, nil
}

// getCPUSample reads /proc/stat
func getCPUSample() (uint64, uint64, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 5 {
			return 0, 0, fmt.Errorf("invalid /proc/stat format")
		}

		parse := func(s string) uint64 {
			val, _ := strconv.ParseUint(s, 10, 64)
			return val
		}

		user := parse(fields[1])
		nice := parse(fields[2])
		system := parse(fields[3])
		idleTicks := parse(fields[4])
		iowait := parse(fields[5])
		irq := parse(fields[6])
		softirq := parse(fields[7])
		steal := parse(fields[8])

		idleVal := idleTicks + iowait
		nonIdleVal := user + nice + system + irq + softirq + steal
		totalVal := idleVal + nonIdleVal

		return idleVal, totalVal, nil
	}
	return 0, 0, fmt.Errorf("could not read cpu line")
}

// GetCPUUsage calculates usage percentage
func GetCPUUsage() (float64, error) {
	idle1, total1, err := getCPUSample()
	if err != nil {
		return 0, err
	}

	time.Sleep(200 * time.Millisecond)

	idle2, total2, err := getCPUSample()
	if err != nil {
		return 0, err
	}

	idleTicks := float64(idle2 - idle1)
	totalTicks := float64(total2 - total1)

	if totalTicks == 0 {
		return 0, nil
	}

	return (totalTicks - idleTicks) / totalTicks * 100, nil
}
