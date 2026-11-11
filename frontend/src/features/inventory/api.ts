import type {
  CreateInventoryMovementRequest,
  CreateInventoryMovementResponse,
  InventoryItem,
  InventoryItemType,
} from "./types";

const INVENTORY_URL = "/api/inventory";

// รองรับ FR-INV-01, CON-TECH-02: อ่านรายการ Inventory และค้นหาด้วยชื่อผ่าน API
export async function listInventoryItems(name = ""): Promise<InventoryItem[]> {
  const query = name ? `?name=${encodeURIComponent(name)}` : "";
  const response = await fetch(`${INVENTORY_URL}${query}`);
  if (!response.ok) {
    throw new Error("ไม่สามารถโหลดรายการสินค้าคงคลังได้");
  }
  return (await response.json()) as InventoryItem[];
}

// รองรับ FR-INV-01, CON-TECH-01, CON-TECH-02: สร้างรายการ Inventory ผ่าน API
export async function createInventoryItem(input: {
  name: string;
  item_type: InventoryItemType;
  reorder_point?: number;
}): Promise<InventoryItem> {
  const response = await fetch(INVENTORY_URL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error("ไม่สามารถเพิ่มรายการสินค้าคงคลังได้");
  }
  return (await response.json()) as InventoryItem;
}

// รองรับ FR-INV-01, DOM-INV-02, CON-TECH-01, CON-TECH-02: แก้ไขรายการ Inventory ผ่าน API
export async function updateInventoryItem(
  id: number,
  input: Pick<InventoryItem, "name" | "item_type" | "reorder_point">,
): Promise<InventoryItem> {
  const response = await fetch(`${INVENTORY_URL}/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error("ไม่สามารถแก้ไขรายการสินค้าคงคลังได้");
  }
  return (await response.json()) as InventoryItem;
}

// รองรับ FR-INV-01, CON-TECH-01, CON-TECH-02: ซ่อนรายการ Inventory ผ่าน API
export async function hideInventoryItem(id: number): Promise<void> {
  const response = await fetch(`${INVENTORY_URL}/${id}`, { method: "DELETE" });
  if (!response.ok) {
    throw new Error("ไม่สามารถซ่อนรายการสินค้าคงคลังได้");
  }
}

// รองรับ FR-INV-02, DOM-INV-01, NFR-DATA-01, CON-TECH-01, CON-TECH-02: บันทึก movement ผ่าน API และส่งผลลัพธ์/ข้อผิดพลาดกลับ
export async function createInventoryMovement(
  itemID: number,
  input: CreateInventoryMovementRequest,
): Promise<CreateInventoryMovementResponse> {
  const response = await fetch(`${INVENTORY_URL}/${itemID}/movements`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (response.status === 409) {
    throw new Error("จำนวนคงเหลือไม่เพียงพอสำหรับทำรายการ");
  }
  if (response.status === 400) {
    throw new Error("ข้อมูลรายการไม่ถูกต้อง กรุณาตรวจสอบข้อมูลที่กรอก");
  }
  if (!response.ok) {
    throw new Error("ไม่สามารถบันทึกรายการเคลื่อนไหวได้");
  }
  return (await response.json()) as CreateInventoryMovementResponse;
}

// รองรับ FR-INV-03, DOM-INV-02, CON-TECH-01, CON-TECH-02: อ่านรายการที่ backend แจ้งว่าอยู่ในระดับสต็อกขั้นต่ำ
export async function listLowStockAlerts(): Promise<InventoryItem[]> {
  const response = await fetch(`${INVENTORY_URL}/alerts`);
  if (!response.ok) {
    throw new Error("ไม่สามารถโหลดการแจ้งเตือนสต็อกขั้นต่ำได้");
  }
  return (await response.json()) as InventoryItem[];
}
