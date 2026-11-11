package inventory

import "github.com/gin-gonic/gin"

// Register ลงทะเบียน endpoint จัดการ catalog บน Gin (FR-INV-01, CON-TECH-01).
func (h *Handler) Register(r gin.IRouter) {
	r.POST("/api/inventory", h.createItem)
	r.GET("/api/inventory", h.listItems)
	r.GET("/api/inventory/alerts", h.listLowStockItems)
	r.GET("/api/inventory/:id", h.getItem)
	r.PATCH("/api/inventory/:id", h.updateItem)
	r.DELETE("/api/inventory/:id", h.deleteItem)
	r.POST("/api/inventory/:id/movements", h.createMovement)
}
