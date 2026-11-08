// Package auth: handler.go คือ HTTP endpoint ของ Gin (CON-TECH-01, FR-AUTH-01).
package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler ผูก Service เข้ากับ route ของ Gin (CON-TECH-01, FR-AUTH-01).
type Handler struct {
	service  *Service
	sessions sessionReader // nil = ไม่ลง route GET /api/auth/session
	timeout  time.Duration // 0 = ไม่ตั้ง deadline (ASM-09)
}

// NewHandler สร้าง Handler จาก Service (FR-AUTH-01).
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// WithSessionReader ตั้งตัวอ่าน Session (Redis) เพื่อให้ Register ลง GET /api/auth/session ด้วย (FR-AUTH-01, CON-CACHE-01); คืน h เพื่อ chain.
func (h *Handler) WithSessionReader(r sessionReader) *Handler {
	h.sessions = r
	return h
}

// WithDependencyTimeout ตั้งเวลาสูงสุดที่รอ PostgreSQL/Redis ต่อ request; เกินแล้วตอบ 503 (AC-AUTH-07, AC-AUTH-08, ASM-05, ASM-09); d <= 0 = ไม่ตั้ง deadline; คืน h เพื่อ chain.
func (h *Handler) WithDependencyTimeout(d time.Duration) *Handler {
	h.timeout = d
	return h
}

// loginRequest คือ JSON ของ POST /api/auth/login ตาม plan.md §4.1 (FR-AUTH-01, ASM-01).
type loginRequest struct {
	StudentID string `json:"student_id"`
	Password  string `json:"password"`
}

// Register ลง route ของฟีเจอร์ auth บน router (FR-AUTH-01); ลง GET /api/auth/session เมื่อมี session reader โดยใช้ timeout เดียวกัน (AC-AUTH-07, ASM-09).
func (h *Handler) Register(r gin.IRouter) {
	r.POST("/api/auth/login", h.login)
	if h.sessions != nil {
		r.GET("/api/auth/session", RequireSession(h.sessions, h.timeout), h.session)
	}
}

// session รองรับ GET /api/auth/session: 200 {member:{id,student_id}} จาก Redis ไม่เรียก PostgreSQL (FR-AUTH-01, AC-AUTH-02, plan D2).
func (h *Handler) session(c *gin.Context) {
	data, ok := MemberFromContext(c)
	if !ok {
		respondUnauthorized(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"member": gin.H{"id": data.MemberID, "student_id": data.StudentID}})
}

// login รองรับ POST /api/auth/login: 200 {refresh_token, expires_in} | 401 invalid_credentials | 503 รวมกรณีเกิน timeout (FR-AUTH-01, AC-AUTH-01, AC-AUTH-08, ASM-05, ASM-06, ASM-09).
func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	// JSON เสีย ฟิลด์หาย หรือค่าว่าง ตอบเหมือนรหัสผิดทุกประการ (ASM-06)
	if err := c.ShouldBindJSON(&req); err != nil || req.StudentID == "" || req.Password == "" {
		respondInvalidCredentials(c)
		return
	}

	// deadline ต่อ request: PostgreSQL/Redis ไม่ตอบภายใน timeout -> ctx หมดเวลา -> error ไม่ใช่ ErrInvalidCredentials -> 503 (AC-AUTH-08, ASM-05, ASM-09)
	ctx := c.Request.Context()
	if h.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.timeout)
		defer cancel()
	}
	token, ttl, err := h.service.Login(ctx, req.StudentID, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			respondInvalidCredentials(c)
			return
		}
		// error อื่น (PostgreSQL/Redis) ปฏิเสธและให้ลองใหม่ ไม่ปล่อยผ่าน (ASM-05, CON-CACHE-01, CON-DB-01)
		// log เฉพาะ error ของ dependency ไม่ใส่รหัสผ่านหรือ token
		log.Printf("auth: login failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refresh_token": token,
		"expires_in":    int64(ttl.Seconds()),
	})
}

// respondInvalidCredentials ตอบ 401 ข้อความเดียวกันทุกกรณี ไม่บอกว่าผิดที่รหัสนักศึกษาหรือรหัสผ่าน (AC-AUTH-06, ASM-06).
func respondInvalidCredentials(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
}
