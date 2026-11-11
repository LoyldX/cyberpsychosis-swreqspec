// Package inventory เก็บโมเดลรายการและประวัติการเคลื่อนไหวของสินค้าคงคลัง (FR-INV-01, FR-INV-02, FR-INV-03, CON-DB-01).
package inventory

import "time"

// InventoryItem แทนรายการสินค้าคงคลังและจำนวนคงเหลือ (FR-INV-01, FR-INV-03, DOM-INV-02, CON-DB-01).
type InventoryItem struct {
	ID             int64
	Name           string
	ItemType       string
	QuantityOnHand int64
	ReorderPoint   *int64
	DeletedAt      *time.Time
}

// InventoryMovement แทนประวัติการเบิก ยืม คืน หรือรับเข้าของรายการ (FR-INV-02, DOM-INV-01, CON-DB-01).
type InventoryMovement struct {
	ID                int64
	InventoryItemID   int64
	MovementType      string
	Quantity          int64
	ResponsiblePerson *string
	Purpose           *string
	DueDate           *time.Time
	CreatedAt         time.Time
}
