package booking

import "github.com/gin-gonic/gin"

// Register ลงทะเบียน endpoint การจองบน Gin ทุกตัวอยู่หลัง requireSession (auth.RequireSession ของ 001)
// (FR-BKG-01, FR-BKG-02, CON-TECH-01; precondition "สมาชิกล็อกอินแล้ว").
func (h *Handler) Register(r gin.IRouter, requireSession gin.HandlerFunc) {
	api := r.Group("/api", requireSession)
	api.GET("/classes", h.listClasses)
	api.POST("/bookings", h.createBooking)
	api.GET("/bookings/:id", h.getBooking)
}
