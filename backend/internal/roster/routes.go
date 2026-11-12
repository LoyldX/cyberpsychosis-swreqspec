package roster

import "github.com/gin-gonic/gin"

// Register ลงทะเบียน endpoint ตารางงานบน Gin โดยไม่ครอบ middleware ยืนยันตัวตน (FR-ROST-01, CON-TECH-01, ASM-06).
func (h *Handler) Register(r gin.IRouter) {
	r.GET("/api/trainers", h.listTrainers)
	r.GET("/api/trainers/:id/roster", h.listEntries)
	r.POST("/api/trainers/:id/roster", h.createEntry)
	r.PATCH("/api/roster/:entryId", h.updateEntry)
	r.POST("/api/roster/:entryId/cancel", h.cancelEntry)
	r.DELETE("/api/roster/:entryId", h.deleteEntry)
}
