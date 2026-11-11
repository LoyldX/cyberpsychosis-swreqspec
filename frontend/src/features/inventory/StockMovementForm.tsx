import { useEffect, useState, type FormEvent } from "react";
import { createInventoryMovement, listInventoryItems } from "./api";
import type {
  CreateInventoryMovementResponse,
  InventoryItem,
  InventoryMovementType,
} from "./types";

const movementLabels: Record<InventoryMovementType, string> = {
  issue: "เบิก",
  borrow: "ยืม",
  return: "คืน",
  receive: "รับเข้า",
};

// รองรับ FR-INV-02, DOM-INV-01, CON-TECH-02: บันทึก movement ที่เลือกผ่าน API
export function StockMovementForm() {
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [selectedItemId, setSelectedItemId] = useState("");
  const [movementType, setMovementType] = useState<InventoryMovementType>("issue");
  const [quantity, setQuantity] = useState("");
  const [responsiblePerson, setResponsiblePerson] = useState("");
  const [purpose, setPurpose] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [error, setError] = useState("");
  const [result, setResult] = useState<CreateInventoryMovementResponse | null>(null);

  useEffect(() => {
    void listInventoryItems()
      .then((loadedItems) => {
        setItems(loadedItems);
        if (loadedItems.length > 0) setSelectedItemId(String(loadedItems[0].id));
      })
      .catch(() => setError("ไม่สามารถโหลดรายการสินค้าคงคลังได้"));
  }, []);

  const selectedItem = items.find((item) => String(item.id) === selectedItemId);
  const requiresResponsibleDetails = movementType !== "receive";

  // รองรับ FR-INV-02, DOM-INV-01: ส่งข้อมูล movement ตามรายการและชนิดที่เลือก
  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setResult(null);

    if (!selectedItem || !quantity || Number(quantity) < 1) {
      setError("กรุณาเลือกรายการและระบุจำนวนที่ถูกต้อง");
      return;
    }

    const requestBody = {
      movement_type: movementType,
      quantity: Number(quantity),
      ...(requiresResponsibleDetails
        ? {
            responsible_person: responsiblePerson.trim(),
            purpose: purpose.trim(),
          }
        : {}),
      ...(movementType === "borrow" ? { due_date: dueDate } : {}),
    };

    try {
      const savedResult = await createInventoryMovement(selectedItem.id, requestBody);
      setItems((currentItems) =>
        currentItems.map((item) =>
          item.id === selectedItem.id
            ? { ...item, quantity_on_hand: savedResult.quantity_on_hand }
            : item,
        ),
      );
      setResult(savedResult);
      setQuantity("");
      setResponsiblePerson("");
      setPurpose("");
      setDueDate("");
    } catch (requestError) {
      setError(
        requestError instanceof Error
          ? requestError.message
          : "ไม่สามารถบันทึกรายการเคลื่อนไหวได้",
      );
    }
  }

  return (
    <section aria-labelledby="stock-movement-heading">
      <h1 id="stock-movement-heading">บันทึกการเคลื่อนไหวสต็อก</h1>
      {error && <p role="alert">{error}</p>}
      {items.length === 0 && !error && <p>ไม่มีรายการสินค้าคงคลังให้เลือก</p>}

      <form onSubmit={handleSubmit}>
        <label>
          รายการสินค้า
          <select
            required
            value={selectedItemId}
            onChange={(event) => {
              setSelectedItemId(event.target.value);
              setResult(null);
            }}
          >
            {items.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </label>

        {selectedItem && (
          <p aria-label="จำนวนคงเหลือ">
            คงเหลือปัจจุบัน: {selectedItem.quantity_on_hand}
          </p>
        )}

        <label>
          ประเภท movement
          <select
            value={movementType}
            onChange={(event) => {
              setMovementType(event.target.value as InventoryMovementType);
              setResult(null);
            }}
          >
            <option value="issue">เบิก</option>
            <option value="borrow">ยืม</option>
            <option value="return">คืน</option>
            <option value="receive">รับเข้า</option>
          </select>
        </label>

        <label>
          จำนวน
          <input
            required
            type="number"
            min="1"
            step="1"
            value={quantity}
            onChange={(event) => setQuantity(event.target.value)}
          />
        </label>

        {requiresResponsibleDetails && (
          <>
            <label>
              ผู้รับผิดชอบ
              <input
                required
                value={responsiblePerson}
                onChange={(event) => setResponsiblePerson(event.target.value)}
              />
            </label>
            <label>
              วัตถุประสงค์
              <input
                required
                value={purpose}
                onChange={(event) => setPurpose(event.target.value)}
              />
            </label>
          </>
        )}

        {movementType === "borrow" && (
          <label>
            วันกำหนดคืน
            <input
              required
              type="date"
              value={dueDate}
              onChange={(event) => setDueDate(event.target.value)}
            />
          </label>
        )}

        <button type="submit" disabled={!selectedItem}>
          บันทึก{movementLabels[movementType]}
        </button>
      </form>

      {result && (
        <p role="status">
          บันทึก{movementLabels[result.movement.movement_type]}สำเร็จ — คงเหลือ:{" "}
          {result.quantity_on_hand}
        </p>
      )}
    </section>
  );
}
