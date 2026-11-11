import { useEffect, useState } from "react";
import { listLowStockAlerts } from "./api";
import type { InventoryItem } from "./types";

// รองรับ FR-INV-03, DOM-INV-02, CON-TECH-02: แสดงรายการที่กำหนด Reorder Point และถึงระดับเตือน
export function LowStockAlerts() {
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    void listLowStockAlerts()
      .then((alerts) =>
        setItems(
          alerts.filter(
            (item) =>
              item.reorder_point !== null &&
              item.quantity_on_hand <= item.reorder_point,
          ),
        ),
      )
      .catch(() => setError("ไม่สามารถโหลดการแจ้งเตือนสต็อกขั้นต่ำได้"));
  }, []);

  return (
    <section aria-labelledby="low-stock-alerts-heading">
      <h1 id="low-stock-alerts-heading">แจ้งเตือนสต็อกขั้นต่ำ</h1>
      {error && <p role="alert">{error}</p>}
      {!error && items.length === 0 && <p>ไม่มีรายการที่ถึงจุดสั่งซื้อขั้นต่ำ</p>}
      {items.length > 0 && (
        <ul aria-label="รายการแจ้งเตือนสต็อกขั้นต่ำ">
          {items.map((item) => (
            <li key={item.id}>
              <strong>{item.name}</strong>
              <span> — คงเหลือ: {item.quantity_on_hand}</span>
              <span> — Reorder Point: {item.reorder_point}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
