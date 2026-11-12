package occupancy

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler ให้บริการ API ของ Occupancy Dashboard ผ่าน Gin (FR-BI-01, NFR-PERF-02, IF-GATE-01, CON-TECH-01).
// ทุก endpoint ไม่ต้องยืนยันตัวตน (ASM-07, plan D4/D5).
type Handler struct {
	svc *Service
}

// NewHandler สร้าง Handler บน Service (CON-TECH-01).
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type signalRequest struct {
	Direction string `json:"direction"`
}

// getOccupancy: GET /api/occupancy → 200 {count, as_of}; ฐานข้อมูลใช้ไม่ได้ → 503 (FR-BI-01, AC-BI-01, plan D6).
func (h *Handler) getOccupancy(c *gin.Context) {
	snap, err := h.svc.Current(c.Request.Context())
	if err != nil {
		log.Printf("occupancy: get current: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, snap)
}

// postSignal: POST /api/gate/signals {direction: "in"|"out"} → 201 {count, as_of}
// เป็นทางเข้าของ "ตัวจำลองสัญญาณ" เท่านั้น ไม่ใช่โปรโตคอลจริงของประตู (IF-GATE-01, Q-01, plan D4)
// direction อื่น → 400 invalid_direction; ฐานข้อมูลใช้ไม่ได้ → 503 และไม่บันทึกสัญญาณ (DOM-CAP-01, plan D6).
func (h *Handler) postSignal(c *gin.Context) {
	var req signalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_direction"})
		return
	}
	d, err := ParseDirection(req.Direction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_direction"})
		return
	}
	snap, err := h.svc.Record(c.Request.Context(), d)
	if err != nil {
		if errors.Is(err, ErrInvalidDirection) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_direction"})
			return
		}
		log.Printf("occupancy: record signal: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	c.JSON(http.StatusCreated, snap)
}
