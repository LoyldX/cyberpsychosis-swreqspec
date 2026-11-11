// รองรับ FR-INV-01, DOM-INV-02: ชนิดข้อมูลรายการและจุดสั่งซื้อที่ไม่บังคับ
export type InventoryItemType = "equipment" | "consumable";

// รองรับ FR-INV-01, DOM-INV-02: สัญญารายการ inventory จาก API
export interface InventoryItem {
  id: number;
  name: string;
  item_type: InventoryItemType;
  quantity_on_hand: number;
  reorder_point: number | null;
}

// รองรับ FR-INV-02: ประเภทการเคลื่อนไหวของรายการตาม API
export type InventoryMovementType = "issue" | "borrow" | "return" | "receive";

// รองรับ FR-INV-02, DOM-INV-01: ข้อมูล movement ที่ฟอร์มส่งให้ API
export interface CreateInventoryMovementRequest {
  movement_type: InventoryMovementType;
  quantity: number;
  responsible_person?: string;
  purpose?: string;
  due_date?: string;
}

// รองรับ FR-INV-02, DOM-INV-01: movement ที่ API บันทึกสำเร็จ
export interface InventoryMovement {
  id: number;
  inventory_item_id: number;
  movement_type: InventoryMovementType;
  quantity: number;
  responsible_person: string | null;
  purpose: string | null;
  due_date: string | null;
  created_at: string;
}

// รองรับ FR-INV-02, NFR-DATA-01: ผล movement และยอดหลังบันทึก
export interface CreateInventoryMovementResponse {
  movement: InventoryMovement;
  quantity_on_hand: number;
}
