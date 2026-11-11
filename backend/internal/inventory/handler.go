package inventory

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler จัดการคำขอ catalog ของ inventory ผ่าน Gin (FR-INV-01, CON-TECH-01, CON-DB-01).
type Handler struct {
	store   *Store
	service *Service
}

// NewHandler สร้าง handler สำหรับ catalog inventory (FR-INV-01, CON-TECH-01).
func NewHandler(store *Store) *Handler {
	return &Handler{store: store, service: NewService(store)}
}

type createItemRequest struct {
	Name         string `json:"name" binding:"required"`
	ItemType     string `json:"item_type" binding:"required"`
	ReorderPoint *int64 `json:"reorder_point"`
}

type inventoryItemResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	ItemType       string `json:"item_type"`
	QuantityOnHand int64  `json:"quantity_on_hand"`
	ReorderPoint   *int64 `json:"reorder_point"`
}

type createMovementRequest struct {
	Type              string  `json:"movement_type" binding:"required"`
	Quantity          int64   `json:"quantity" binding:"required,gt=0"`
	ResponsiblePerson string  `json:"responsible_person"`
	Purpose           string  `json:"purpose"`
	DueDate           *string `json:"due_date"`
}

type inventoryMovementResponse struct {
	ID                int64     `json:"id"`
	InventoryItemID   int64     `json:"inventory_item_id"`
	MovementType      string    `json:"movement_type"`
	Quantity          int64     `json:"quantity"`
	ResponsiblePerson *string   `json:"responsible_person"`
	Purpose           *string   `json:"purpose"`
	DueDate           *string   `json:"due_date"`
	CreatedAt         time.Time `json:"created_at"`
}

type createMovementResponse struct {
	Movement       inventoryMovementResponse `json:"movement"`
	QuantityOnHand int64                     `json:"quantity_on_hand"`
}

// createItem สร้างรายการ inventory โดย store กำหนดยอดเริ่มต้นเป็นศูนย์ (FR-INV-01, CON-DB-01).
func (h *Handler) createItem(c *gin.Context) {
	var req createItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	item, err := h.store.CreateItem(c.Request.Context(), InventoryItem{
		Name:         req.Name,
		ItemType:     req.ItemType,
		ReorderPoint: req.ReorderPoint,
	})
	if err != nil {
		log.Printf("inventory: create item failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusCreated, toInventoryItemResponse(item))
}

// listItems คืนรายการ catalog ที่มองเห็นได้และค้นหาด้วยชื่อบางส่วนได้ (FR-INV-01, CON-DB-01).
func (h *Handler) listItems(c *gin.Context) {
	items, err := h.store.ListItems(c.Request.Context(), c.Query("name"))
	if err != nil {
		log.Printf("inventory: list items failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	response := make([]inventoryItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toInventoryItemResponse(item))
	}
	c.JSON(http.StatusOK, response)
}

// listLowStockItems returns visible items whose quantity is at or below their reorder point (FR-INV-03, DOM-INV-02, CON-TECH-01, CON-DB-01).
func (h *Handler) listLowStockItems(c *gin.Context) {
	items, err := h.service.ListLowStockItems(c.Request.Context())
	if err != nil {
		log.Printf("inventory: list low stock items failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	response := make([]inventoryItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toInventoryItemResponse(item))
	}
	c.JSON(http.StatusOK, response)
}

// getItem อ่านรายการ catalog ที่มองเห็นได้ตาม ID (FR-INV-01, CON-DB-01).
func (h *Handler) getItem(c *gin.Context) {
	id, ok := parseItemID(c)
	if !ok {
		return
	}

	item, err := h.store.GetItem(c.Request.Context(), id)
	if err != nil {
		h.respondItemError(c, err, "get item")
		return
	}
	c.JSON(http.StatusOK, toInventoryItemResponse(item))
}

// updateItem แก้เฉพาะฟิลด์ที่ส่งมาใน PATCH และ null ใช้ล้าง Reorder Point ที่ไม่บังคับ (FR-INV-01, DOM-INV-02, CON-DB-01).
func (h *Handler) updateItem(c *gin.Context) {
	id, ok := parseItemID(c)
	if !ok {
		return
	}

	var fields map[string]json.RawMessage
	if err := c.ShouldBindJSON(&fields); err != nil || len(fields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	current, err := h.store.GetItem(c.Request.Context(), id)
	if err != nil {
		h.respondItemError(c, err, "get item for update")
		return
	}

	for field, raw := range fields {
		switch field {
		case "name":
			if string(raw) == "null" || json.Unmarshal(raw, &current.Name) != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
				return
			}
		case "item_type":
			if string(raw) == "null" || json.Unmarshal(raw, &current.ItemType) != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
				return
			}
		case "reorder_point":
			if string(raw) == "null" {
				current.ReorderPoint = nil
				continue
			}
			var reorderPoint int64
			if err := json.Unmarshal(raw, &reorderPoint); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
				return
			}
			current.ReorderPoint = &reorderPoint
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}
	}

	item, err := h.store.UpdateItem(c.Request.Context(), current)
	if err != nil {
		h.respondItemError(c, err, "update item")
		return
	}
	c.JSON(http.StatusOK, toInventoryItemResponse(item))
}

// deleteItem ซ่อนรายการจาก catalog โดยไม่ลบประวัติ movement (FR-INV-01, FR-INV-02, CON-DB-01).
func (h *Handler) deleteItem(c *gin.Context) {
	id, ok := parseItemID(c)
	if !ok {
		return
	}

	if err := h.store.SoftDeleteItem(c.Request.Context(), id); err != nil {
		h.respondItemError(c, err, "hide item")
		return
	}
	c.Status(http.StatusNoContent)
}

// createMovement เบิก ยืม คืน หรือรับเข้าสต็อกพร้อมบันทึกประวัติใน transaction เดียว (FR-INV-02, DOM-INV-01, NFR-DATA-01, CON-DB-01).
func (h *Handler) createMovement(c *gin.Context) {
	id, ok := parseItemID(c)
	if !ok {
		return
	}

	var req createMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}
		dueDate = &parsed
	}

	result, err := h.service.CreateMovement(c.Request.Context(), id, MovementCommand{
		Type:              req.Type,
		Quantity:          req.Quantity,
		ResponsiblePerson: req.ResponsiblePerson,
		Purpose:           req.Purpose,
		DueDate:           dueDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidMovement):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		case errors.Is(err, ErrItemNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		case errors.Is(err, ErrInsufficientStock):
			c.JSON(http.StatusConflict, gin.H{"error": "insufficient_stock"})
		default:
			log.Printf("inventory: create movement failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}
	c.JSON(http.StatusCreated, createMovementResponse{
		Movement: inventoryMovementResponse{
			ID:                result.Movement.ID,
			InventoryItemID:   result.Movement.InventoryItemID,
			MovementType:      result.Movement.MovementType,
			Quantity:          result.Movement.Quantity,
			ResponsiblePerson: result.Movement.ResponsiblePerson,
			Purpose:           result.Movement.Purpose,
			DueDate:           formatDueDate(result.Movement.DueDate),
			CreatedAt:         result.Movement.CreatedAt,
		},
		QuantityOnHand: result.QuantityOnHand,
	})
}

// formatDueDate คืนวันกำหนดคืนในรูปแบบวันที่ที่ API รับเข้ามา (FR-INV-02, DOM-INV-01).
func formatDueDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}

// parseItemID ตรวจสอบ ID ใน path ของ catalog (FR-INV-01, CON-TECH-01).
func parseItemID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return 0, false
	}
	return id, true
}

// respondItemError แปลงกรณีไม่พบรายการและข้อผิดพลาดฐานข้อมูลเป็น HTTP response ที่ชัดเจน (FR-INV-01, CON-DB-01).
func (h *Handler) respondItemError(c *gin.Context, err error, operation string) {
	if errors.Is(err, ErrItemNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	log.Printf("inventory: %s failed: %v", operation, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
}

// toInventoryItemResponse คืนเฉพาะฟิลด์ catalog โดยไม่เปิดเผยข้อมูล soft-delete ภายใน (FR-INV-01).
func toInventoryItemResponse(item InventoryItem) inventoryItemResponse {
	return inventoryItemResponse{
		ID:             item.ID,
		Name:           item.Name,
		ItemType:       item.ItemType,
		QuantityOnHand: item.QuantityOnHand,
		ReorderPoint:   item.ReorderPoint,
	}
}
