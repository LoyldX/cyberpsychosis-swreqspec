// Package auth: middleware.go คือ middleware ตรวจ Session จาก Redis ที่ฟีเจอร์อื่นใช้ซ้ำได้ (CON-TECH-01, CON-CACHE-01, FR-AUTH-01).
package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// sessionContextKey คือ key ที่เก็บ SessionData ใน gin.Context (FR-AUTH-01, AC-AUTH-02).
const sessionContextKey = "auth.session"

// sessionReader คือสิ่งที่ RequireSession ต้องใช้อ่าน Session; *SessionStore ตรงตามนี้ (CON-CACHE-01).
type sessionReader interface {
	Get(ctx context.Context, token string) (SessionData, error)
}

// RequireSession ตรวจ header "Authorization: Bearer <token>" กับ Session ใน Redis แล้วใส่ SessionData ใน context
// 401 {"error":"unauthorized"} เมื่อ header ผิดรูปหรือไม่พบ Session, 503 เมื่อ Redis มีปัญหา ไม่ปล่อยผ่าน
// ไม่เรียก PostgreSQL (plan D2) และไม่ต่ออายุ TTL; timeout (ตัวเลือก, > 0) คือ deadline ของการอ่าน Session เกินแล้วตอบ 503
// (FR-AUTH-01, CON-CACHE-01, AC-AUTH-02, AC-AUTH-04, AC-AUTH-07, ASM-02, ASM-05, ASM-09).
func RequireSession(reader sessionReader, timeout ...time.Duration) gin.HandlerFunc {
	var limit time.Duration
	if len(timeout) > 0 {
		limit = timeout[0]
	}
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			respondUnauthorized(c)
			return
		}

		ctx := c.Request.Context()
		if limit > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, limit)
			defer cancel()
		}
		data, err := reader.Get(ctx, token)
		if err != nil {
			if errors.Is(err, ErrSessionNotFound) {
				respondUnauthorized(c)
				return
			}
			// error อื่น (Redis ไม่พร้อม) ปฏิเสธและให้ลองใหม่ (ASM-05, CON-CACHE-01); ไม่ log token
			log.Printf("auth: read session failed: %v", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง",
			})
			return
		}

		c.Set(sessionContextKey, data)
		c.Next()
	}
}

// MemberFromContext คืนสมาชิกที่ RequireSession ยืนยันแล้ว; ok เป็น false ถ้าไม่ได้ผ่าน middleware (FR-AUTH-01, AC-AUTH-02).
func MemberFromContext(c *gin.Context) (SessionData, bool) {
	v, exists := c.Get(sessionContextKey)
	if !exists {
		return SessionData{}, false
	}
	data, ok := v.(SessionData)
	return data, ok
}

// bearerToken แยก token จาก "Bearer <token>" (scheme ไม่สนตัวพิมพ์เล็กใหญ่, ต้องมี token เดียว) (FR-AUTH-01, AC-AUTH-04).
func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return parts[1], true
}

// respondUnauthorized ตอบ 401 และหยุด chain ด้วยข้อความเดียวกันทุกกรณี (FR-AUTH-01, AC-AUTH-04).
func respondUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
