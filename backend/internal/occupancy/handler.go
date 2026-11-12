package occupancy

import (
	"errors"
	"io"
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

// sseEventName คือชื่อ event ของ SSE ที่หน้า Dashboard ฟัง (plan §4.1).
const sseEventName = "occupancy"

// streamOccupancy: GET /api/occupancy/stream → text/event-stream
// ส่ง event "occupancy" {count, as_of} ทันทีที่เชื่อมต่อ ทุกครั้งที่บันทึกสัญญาณสำเร็จ และตอน 00:00 Asia/Bangkok
// เมื่อผู้ใช้ปิดหน้าหรือการเชื่อมต่อหลุด ผู้ฟังจะถูกถอดออกจาก Broadcaster (NFR-PERF-02, ASM-06, AC-BI-04, plan D1).
func (h *Handler) streamOccupancy(c *gin.Context) {
	ctx := c.Request.Context()
	// ลงทะเบียนก่อนอ่านจำนวน เพื่อไม่พลาดสัญญาณที่เข้ามาระหว่างนั้น
	updates, unsubscribe := h.svc.Subscribe()
	defer unsubscribe()

	snap, err := h.svc.Current(ctx)
	if err != nil {
		log.Printf("occupancy: stream current: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.SSEvent(sseEventName, snap)
	c.Writer.Flush()

	c.Stream(func(io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case next := <-updates:
			c.SSEvent(sseEventName, next)
			return true
		}
	})
}
