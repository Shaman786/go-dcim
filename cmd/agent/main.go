package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Shaman786/go-dcim/internal/metrics"
)

type TelemetryPayload struct {
	Hostname  string  `json:"hostname"`
	CPUUsage  float64 `json:"cpu_usage"`
	RAMTotal  uint64  `json:"ram_total"`
	RAMUsed   uint64  `json:"ram_used"`
	RAMFree   uint64  `json:"ram_free"`
	Timestamp string  `json:"timestamp"`
}

// Server URL (Change localhost to IP if running on different VMs)
func getServerURL() string {
	if url := os.Getenv("SERVER_URL"); url != "" {
		return url
	}
	return "http://localhost:8000/api/telemetry"
}

func main() {
	targetURL := getServerURL()
	log.Println("🔌 Starting Infrasight Agent (v1.0)...")
	log.Printf("📡 Target Server: %s", targetURL)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cpu, err := metrics.GetCPUUsage()
		if err != nil {
			log.Printf("Error reading CPU: %v", err)
			continue
		}

		totalMem, usedMem, freeMem, err := metrics.ReadMemoryStats()
		if err != nil {
			log.Printf("Error reading Memory: %v", err)
			continue
		}

		payload := TelemetryPayload{
			Hostname:  hostname,
			CPUUsage:  cpu,
			RAMTotal:  totalMem,
			RAMUsed:   usedMem,
			RAMFree:   freeMem,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		// Send HTTP POST Request
		sendTelemetry(payload, targetURL)
	}
}

func sendTelemetry(data TelemetryPayload, url string) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("JSON Error: %v", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("⚠️ Failed to send metrics: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Server returned status: %s", resp.Status)
	} else {
		fmt.Printf("✅ Sent metrics for %s (CPU: %.1f%%)\n", data.Hostname, data.CPUUsage)
	}
}
