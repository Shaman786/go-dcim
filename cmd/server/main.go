package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// TelemetryPayload matches the struct sent by the Agent
type TelemetryPayload struct {
	Hostname  string  `json:"hostname"`
	CPUUsage  float64 `json:"cpu_usage"`
	RAMTotal  uint64  `json:"ram_total"`
	RAMUsed   uint64  `json:"ram_used"`
	RAMFree   uint64  `json:"ram_free"`
	Timestamp string  `json:"timestamp"`
}

// In-Memory Store (Simulating a database like Redis/Postgres)
// We use a Mutex to prevent "Race Conditions" (Interview Keyword!)
var (
	serverStore = make(map[string]TelemetryPayload)
	storeMutex  sync.RWMutex
)

func main() {
	// 1. Initialize Gin Router
	r := gin.Default()

	// 2. Define the Ingestion Endpoint (Agent sends data here)
	r.POST("/api/telemetry", func(c *gin.Context) {
		var payload TelemetryPayload

		// Bind JSON body to struct
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Store data safely using Mutex
		storeMutex.Lock()
		serverStore[payload.Hostname] = payload
		storeMutex.Unlock()

		log.Printf("📥 Received Data from %s: CPU %.2f%% | RAM Used: %d MB",
			payload.Hostname, payload.CPUUsage, payload.RAMUsed/1024/1024)

		c.JSON(http.StatusOK, gin.H{"status": "received"})
	})

	// 3. Define the Dashboard Endpoint (Frontend reads data here)
	r.GET("/api/nodes", func(c *gin.Context) {
		storeMutex.RLock()
		defer storeMutex.RUnlock()

		// Convert map to slice for JSON array
		nodes := make([]TelemetryPayload, 0, len(serverStore))
		for _, data := range serverStore {
			nodes = append(nodes, data)
		}

		c.JSON(http.StatusOK, nodes)
	})

	// 4. Start Server on Port 8080
	log.Println("🚀 Control Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
