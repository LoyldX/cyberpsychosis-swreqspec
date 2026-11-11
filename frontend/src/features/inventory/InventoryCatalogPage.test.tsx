// ตรวจ AC-INV-01 ถึง AC-INV-05 (FR-INV-01) ด้วย API จำลองผ่าน MSW
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { MemoryRouter } from "react-router";
import { App } from "../../App";
import { FAKE_REFRESH_TOKEN } from "../../mocks/handlers";
import { InventoryCatalogPage } from "./InventoryCatalogPage";
import type { InventoryItem } from "./types";
import { server } from "../../mocks/server";
import { REFRESH_TOKEN_KEY } from "../auth/api";

const inventoryItems: InventoryItem[] = [
  {
    id: 1,
    name: "กรรไกร",
    item_type: "equipment",
    quantity_on_hand: 10,
    reorder_point: null,
  },
  {
    id: 2,
    name: "กระดาษ",
    item_type: "consumable",
    quantity_on_hand: 4,
    reorder_point: 5,
  },
];

let items: InventoryItem[];

beforeEach(() => {
  items = inventoryItems.map((item) => ({ ...item }));
  server.use(
    http.get("/api/inventory", ({ request }) => {
      const name = new URL(request.url).searchParams.get("name")?.toLocaleLowerCase();
      return HttpResponse.json(
        name ? items.filter((item) => item.name.toLocaleLowerCase().includes(name)) : items,
      );
    }),
    http.post("/api/inventory", async ({ request }) => {
      const input = (await request.json()) as Omit<InventoryItem, "id" | "quantity_on_hand">;
      const created: InventoryItem = {
        ...input,
        id: 3,
        quantity_on_hand: 0,
        reorder_point: input.reorder_point ?? null,
      };
      items.push(created);
      return HttpResponse.json(created, { status: 201 });
    }),
    http.patch("/api/inventory/:id", async ({ params, request }) => {
      const input = (await request.json()) as Partial<InventoryItem>;
      const id = Number(params.id);
      items = items.map((item) => (item.id === id ? { ...item, ...input } : item));
      return HttpResponse.json(items.find((item) => item.id === id));
    }),
    http.delete("/api/inventory/:id", ({ params }) => {
      items = items.filter((item) => item.id !== Number(params.id));
      return new HttpResponse(null, { status: 204 });
    }),
  );
});

describe("Inventory Catalog", () => {
  it("TestAC_INV_01 creates a new catalog item", async () => {
    // Given: ยังไม่มีรายการชื่อกล่องปฐมพยาบาลใน inventory
    render(<InventoryCatalogPage />);
    await screen.findByText("กรรไกร");
    expect(screen.queryByText("กล่องปฐมพยาบาล")).not.toBeInTheDocument();

    // When: เจ้าหน้าที่เพิ่มรายการใหม่
    fireEvent.change(screen.getByLabelText("ชื่อรายการใหม่"), {
      target: { value: "กล่องปฐมพยาบาล" },
    });
    fireEvent.click(screen.getByRole("button", { name: "เพิ่มรายการ" }));

    // Then: รายการใหม่ปรากฏในรายการทั้งหมด
    expect(await screen.findByText("กล่องปฐมพยาบาล")).toBeInTheDocument();
    expect(screen.getByText("กล่องปฐมพยาบาล").closest("li")).toHaveTextContent(
      "คงเหลือ: 0",
    );
  });

  it("TestAC_INV_02 updates an existing catalog item", async () => {
    // Given: รายการกรรไกรมีอยู่แล้ว
    render(<InventoryCatalogPage />);
    await screen.findByText("กรรไกร");

    // When: เจ้าหน้าที่แก้ไขชื่อรายการ
    fireEvent.click(screen.getByRole("button", { name: "แก้ไข กรรไกร" }));
    fireEvent.change(screen.getByLabelText("ชื่อรายการแก้ไข"), {
      target: { value: "กรรไกรตัดผ้า" },
    });
    fireEvent.click(screen.getByRole("button", { name: "บันทึกการแก้ไข" }));

    // Then: เมื่อเรียกดูรายการซ้ำเห็นชื่อที่แก้ไข
    expect(await screen.findByText("กรรไกรตัดผ้า")).toBeInTheDocument();
    expect(screen.queryByText("กรรไกร", { exact: true })).not.toBeInTheDocument();
  });

  it("TestAC_INV_03 hides a catalog item", async () => {
    // Given: รายการกรรไกรมีอยู่ในรายการ
    render(<InventoryCatalogPage />);
    await screen.findByText("กรรไกร");

    // When: เจ้าหน้าที่ซ่อนรายการ
    fireEvent.click(screen.getByRole("button", { name: "ซ่อน กรรไกร" }));

    // Then: รายการไม่ปรากฏในรายการทั้งหมดอีก
    await waitFor(() => expect(screen.queryByText("กรรไกร", { exact: true })).not.toBeInTheDocument());
  });

  it("TestAC_INV_04 searches catalog items by name", async () => {
    // Given: มีรายการหลายรายการ
    render(<InventoryCatalogPage />);
    await screen.findByText("กรรไกร");
    expect(screen.getByText("กระดาษ")).toBeInTheDocument();

    // When: ค้นหาด้วยชื่อของรายการกรรไกร
    fireEvent.change(screen.getByLabelText("ค้นหารายการด้วยชื่อ"), {
      target: { value: "กรรไกร" },
    });
    fireEvent.click(screen.getByRole("button", { name: "ค้นหา" }));

    // Then: ผลลัพธ์มีกรรไกรและไม่มีรายการอื่น
    expect(await screen.findByText("กรรไกร")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("กระดาษ")).not.toBeInTheDocument());
  });

  it("TestAC_INV_05 shows the on-hand quantity in item details", async () => {
    // Given: รายการกรรไกรมีคงเหลือ 10
    render(<InventoryCatalogPage />);
    await screen.findByText("กรรไกร");

    // When: เจ้าหน้าที่เปิดดูรายละเอียดรายการ
    fireEvent.click(screen.getByRole("button", { name: "ดูรายละเอียด กรรไกร" }));

    // Then: เห็นจำนวนคงเหลือ 10
    expect(screen.getByRole("region", { name: "รายละเอียด กรรไกร" })).toHaveTextContent(
      "จำนวนคงเหลือ: 10",
    );
  });

  it("renders the catalog at the authenticated /inventory route", async () => {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, FAKE_REFRESH_TOKEN);
    render(
      <MemoryRouter initialEntries={["/inventory"]}>
        <App />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "รายการสินค้าคงคลัง" })).toBeInTheDocument();
    expect(await screen.findByText("กรรไกร")).toBeInTheDocument();
  });
});
