package occupancy

import "github.com/gin-gonic/gin"

// Register ลง endpoint ของ Occupancy Dashboard บน Gin (FR-BI-01, NFR-PERF-02, IF-GATE-01 ตัวจำลอง, CON-TECH-01).
func (h *Handler) Register(r gin.IRouter) {
	r.GET("/api/occupancy", h.getOccupancy)
	r.GET("/api/occupancy/stream", h.streamOccupancy)
	r.POST("/api/gate/signals", h.postSignal)
}
