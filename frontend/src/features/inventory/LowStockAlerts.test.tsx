// ตรวจ AC-INV-13 (FR-INV-03, DOM-INV-02) ด้วย API จำลองผ่าน MSW
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { server } from "../../mocks/server";
import type { InventoryItem } from "./types";
import { LowStockAlerts } from "./LowStockAlerts";

const alertResponse: InventoryItem[] = [
  {
    id: 1,
    name: "วิทยุสื่อสาร",
    item_type: "equipment",
    quantity_on_hand: 5,
    reorder_point: 5,
  },
  {
    id: 2,
    name: "กรรไกร",
    item_type: "equipment",
    quantity_on_hand: 2,
    reorder_point: null,
  },
  {
    id: 3,
    name: "กระดาษ",
    item_type: "consumable",
    quantity_on_hand: 8,
    reorder_point: 5,
  },
];

beforeEach(() => {
  server.use(
    http.get("/api/inventory/alerts", () => HttpResponse.json(alertResponse)),
  );
});

describe("Low Stock Alerts", () => {
  it("TestAC_INV_13 shows the item at its reorder point and excludes items without one", async () => {
    // Given: API จำลองมีรายการ A ที่ยอด 5/Reorder Point 5 และรายการ B ที่ไม่มี Reorder Point
    // When: เจ้าหน้าที่เปิดหน้าแจ้งเตือนสต็อกขั้นต่ำ
    render(<LowStockAlerts />);

    // Then: แสดง A ที่ถึงจุด และไม่แสดง B ที่ไม่มี Reorder Point
    const alertList = await screen.findByRole("list", {
      name: "รายการแจ้งเตือนสต็อกขั้นต่ำ",
    });
    expect(alertList).toHaveTextContent("วิทยุสื่อสาร");
    expect(alertList).toHaveTextContent("คงเหลือ: 5");
    expect(alertList).toHaveTextContent("Reorder Point: 5");
    expect(screen.queryByText("กรรไกร")).not.toBeInTheDocument();
    expect(screen.queryByText("กระดาษ")).not.toBeInTheDocument();
  });

  it("shows an error when the live alert API is unavailable", async () => {
    server.use(
      http.get("/api/inventory/alerts", () =>
        HttpResponse.json({ error: "internal_error" }, { status: 500 }),
      ),
    );
    render(<LowStockAlerts />);

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "ไม่สามารถโหลดการแจ้งเตือนสต็อกขั้นต่ำได้",
    );
  });
});
