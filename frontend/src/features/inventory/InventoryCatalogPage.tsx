import { useEffect, useState, type FormEvent } from "react";
import { createInventoryItem, hideInventoryItem, listInventoryItems, updateInventoryItem } from "./api";
import type { InventoryItem, InventoryItemType } from "./types";

// รองรับ FR-INV-01, CON-TECH-02: จัดการค้นหา เพิ่ม แก้ไข ซ่อน และตรวจสอบยอดรายการ
export function InventoryCatalogPage() {
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [searchName, setSearchName] = useState("");
  const [createName, setCreateName] = useState("");
  const [createType, setCreateType] = useState<InventoryItemType>("equipment");
  const [createReorderPoint, setCreateReorderPoint] = useState("");
  const [editingItem, setEditingItem] = useState<InventoryItem | null>(null);
  const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    void listInventoryItems()
      .then(setItems)
      .catch(() => setError("ไม่สามารถโหลดรายการสินค้าคงคลังได้"));
  }, []);

  // รองรับ FR-INV-01, CON-TECH-02: ค้นหารายการด้วยชื่อผ่าน query ของ API
  async function handleSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      setItems(await listInventoryItems(searchName.trim()));
    } catch {
      setError("ไม่สามารถค้นหารายการสินค้าคงคลังได้");
    }
  }

  // รองรับ FR-INV-01, DOM-INV-02, CON-TECH-02: สร้างรายการใหม่และแสดงค่าที่ API คืนมา
  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    const reorderPoint = createReorderPoint.trim();
    const body = {
      name: createName.trim(),
      item_type: createType,
      ...(reorderPoint ? { reorder_point: Number(reorderPoint) } : {}),
    };

    try {
      const item = await createInventoryItem(body);
      setItems((currentItems) => [...currentItems, item]);
      setCreateName("");
      setCreateReorderPoint("");
    } catch {
      setError("ไม่สามารถเพิ่มรายการสินค้าคงคลังได้");
    }
  }

  // รองรับ FR-INV-01, DOM-INV-02, CON-TECH-02: บันทึกข้อมูลรายการที่แก้ไข
  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editingItem) return;
    setError("");

    try {
      const updatedItem = await updateInventoryItem(editingItem.id, {
        name: editingItem.name,
        item_type: editingItem.item_type,
        reorder_point: editingItem.reorder_point,
      });
      setItems((currentItems) =>
        currentItems.map((item) => (item.id === updatedItem.id ? updatedItem : item)),
      );
      setEditingItem(null);
    } catch {
      setError("ไม่สามารถแก้ไขรายการสินค้าคงคลังได้");
    }
  }

  // รองรับ FR-INV-01, CON-TECH-02: ซ่อนรายการจาก catalog ผ่าน API
  async function handleHide(item: InventoryItem) {
    setError("");
    try {
      await hideInventoryItem(item.id);
      setItems((currentItems) => currentItems.filter((current) => current.id !== item.id));
      if (editingItem?.id === item.id) setEditingItem(null);
      if (selectedItemId === item.id) setSelectedItemId(null);
    } catch {
      setError("ไม่สามารถซ่อนรายการสินค้าคงคลังได้");
    }
  }

  // รองรับ FR-INV-01, DOM-INV-02, CON-TECH-02: แสดงฟอร์มแก้ไขเฉพาะรายการที่เลือก
  function beginEditing(item: InventoryItem) {
    setEditingItem({ ...item });
  }

  return (
    <main>
      <h1>รายการสินค้าคงคลัง</h1>
      {error && <p role="alert">{error}</p>}

      <form aria-label="ค้นหารายการ" onSubmit={handleSearch}>
        <label>
          ค้นหารายการด้วยชื่อ
          <input
            value={searchName}
            onChange={(event) => setSearchName(event.target.value)}
          />
        </label>
        <button type="submit">ค้นหา</button>
      </form>

      <section aria-labelledby="create-item-heading">
        <h2 id="create-item-heading">เพิ่มรายการ</h2>
        <form onSubmit={handleCreate}>
          <label>
            ชื่อรายการใหม่
            <input
              required
              value={createName}
              onChange={(event) => setCreateName(event.target.value)}
            />
          </label>
          <label>
            ประเภท
            <select
              value={createType}
              onChange={(event) => setCreateType(event.target.value as InventoryItemType)}
            >
              <option value="equipment">อุปกรณ์ยืม-คืน</option>
              <option value="consumable">วัสดุสิ้นเปลือง</option>
            </select>
          </label>
          <label>
            Reorder Point (ไม่บังคับ)
            <input
              type="number"
              value={createReorderPoint}
              onChange={(event) => setCreateReorderPoint(event.target.value)}
            />
          </label>
          <button type="submit">เพิ่มรายการ</button>
        </form>
      </section>

      <section aria-labelledby="inventory-items-heading">
        <h2 id="inventory-items-heading">รายการทั้งหมด</h2>
        {items.length === 0 ? (
          <p>ไม่พบรายการสินค้าคงคลัง</p>
        ) : (
          <ul>
            {items.map((item) => (
              <li key={item.id}>
                {editingItem?.id === item.id ? (
                  <form aria-label={`แก้ไข ${item.name}`} onSubmit={handleUpdate}>
                    <label>
                      ชื่อรายการแก้ไข
                      <input
                        required
                        value={editingItem.name}
                        onChange={(event) =>
                          setEditingItem({ ...editingItem, name: event.target.value })
                        }
                      />
                    </label>
                    <label>
                      ประเภทที่แก้ไข
                      <select
                        value={editingItem.item_type}
                        onChange={(event) =>
                          setEditingItem({
                            ...editingItem,
                            item_type: event.target.value as InventoryItemType,
                          })
                        }
                      >
                        <option value="equipment">อุปกรณ์ยืม-คืน</option>
                        <option value="consumable">วัสดุสิ้นเปลือง</option>
                      </select>
                    </label>
                    <label>
                      Reorder Point ที่แก้ไข (ไม่บังคับ)
                      <input
                        type="number"
                        value={editingItem.reorder_point ?? ""}
                        onChange={(event) =>
                          setEditingItem({
                            ...editingItem,
                            reorder_point: event.target.value
                              ? Number(event.target.value)
                              : null,
                          })
                        }
                      />
                    </label>
                    <button type="submit">บันทึกการแก้ไข</button>
                    <button type="button" onClick={() => setEditingItem(null)}>
                      ยกเลิก
                    </button>
                  </form>
                ) : (
                  <>
                    <span>{item.name}</span>
                    <span> — คงเหลือ: {item.quantity_on_hand}</span>
                    <button
                      type="button"
                      aria-label={`ดูรายละเอียด ${item.name}`}
                      onClick={() => setSelectedItemId(item.id)}
                    >
                      ดูรายละเอียด
                    </button>
                    <button type="button" onClick={() => beginEditing(item)}>
                      แก้ไข {item.name}
                    </button>
                    <button type="button" onClick={() => void handleHide(item)}>
                      ซ่อน {item.name}
                    </button>
                    {selectedItemId === item.id && (
                      <section aria-label={`รายละเอียด ${item.name}`}>
                        <h3>{item.name}</h3>
                        <p>ประเภท: {item.item_type}</p>
                        <p>จำนวนคงเหลือ: {item.quantity_on_hand}</p>
                        <p>
                          Reorder Point: {item.reorder_point === null ? "ไม่ได้กำหนด" : item.reorder_point}
                        </p>
                      </section>
                    )}
                  </>
                )}
              </li>
            ))}
          </ul>
        )}
      </section>
    </main>
  );
}
