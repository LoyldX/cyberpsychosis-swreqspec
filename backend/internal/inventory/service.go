package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	movementTypeIssue   = "issue"
	movementTypeReceive = "receive"
	movementTypeBorrow  = "borrow"
	movementTypeReturn  = "return"
)

var (
	ErrInvalidMovement   = errors.New("inventory: invalid movement")
	ErrInsufficientStock = errors.New("inventory: insufficient stock")
)

// MovementCommand คือข้อมูลที่ใช้บันทึกการเบิก ยืม คืน หรือรับเข้าสินค้าคงคลัง (FR-INV-02, DOM-INV-01).
type MovementCommand struct {
	Type              string
	Quantity          int64
	ResponsiblePerson string
	Purpose           string
	DueDate           *time.Time
}

// MovementResult คืนประวัติ movement พร้อมยอดคงเหลือหลัง transaction (FR-INV-02, NFR-DATA-01).
type MovementResult struct {
	Movement       InventoryMovement
	QuantityOnHand int64
}

// Service จัดการการเปลี่ยนยอดสต็อกและบันทึกประวัติผ่าน Store (FR-INV-02, NFR-DATA-01, CON-DB-01).
type Service struct {
	store *Store
}

// NewService สร้าง service สำหรับจัดการ stock movement (FR-INV-02, NFR-DATA-01).
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// CreateMovement ตรวจข้อมูลการเคลื่อนไหวและสั่ง Store ทำทั้งสองอย่างใน transaction เดียว (FR-INV-02, DOM-INV-01, NFR-DATA-01).
func (s *Service) CreateMovement(ctx context.Context, itemID int64, command MovementCommand) (MovementResult, error) {
	if itemID < 1 || command.Quantity <= 0 {
		return MovementResult{}, ErrInvalidMovement
	}

	switch command.Type {
	case movementTypeIssue, movementTypeBorrow, movementTypeReturn:
		if strings.TrimSpace(command.ResponsiblePerson) == "" || strings.TrimSpace(command.Purpose) == "" {
			return MovementResult{}, ErrInvalidMovement
		}
		if command.Type == movementTypeBorrow && command.DueDate == nil {
			return MovementResult{}, ErrInvalidMovement
		}
		if command.Type != movementTypeBorrow && command.DueDate != nil {
			return MovementResult{}, ErrInvalidMovement
		}
	case movementTypeReceive:
		if command.DueDate != nil {
			return MovementResult{}, ErrInvalidMovement
		}
	default:
		return MovementResult{}, ErrInvalidMovement
	}

	movement, quantityOnHand, err := s.store.ApplyStockMovement(ctx, itemID, InventoryMovement{
		InventoryItemID:   itemID,
		MovementType:      command.Type,
		Quantity:          command.Quantity,
		ResponsiblePerson: optionalText(command.ResponsiblePerson),
		Purpose:           optionalText(command.Purpose),
		DueDate:           command.DueDate,
	})
	if err != nil {
		return MovementResult{}, fmt.Errorf("inventory: apply movement: %w", err)
	}
	return MovementResult{Movement: movement, QuantityOnHand: quantityOnHand}, nil
}

// ListLowStockItems returns inventory items that meet their configured reorder point (FR-INV-03, DOM-INV-02, CON-DB-01).
func (s *Service) ListLowStockItems(ctx context.Context) ([]InventoryItem, error) {
	items, err := s.store.ListLowStockItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("inventory: list low stock items: %w", err)
	}
	return items, nil
}

// optionalText แปลงข้อความว่างให้เป็น SQL NULL สำหรับฟิลด์ movement ที่ไม่บังคับ (FR-INV-02).
func optionalText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
