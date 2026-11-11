// ตรวจ AC-INV-06 ถึง AC-INV-12 (FR-INV-02, DOM-INV-01) ด้วย API จำลองผ่าน MSW
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { server } from "../../mocks/server";
import type {
  CreateInventoryMovementRequest,
  InventoryItem,
} from "./types";
import { StockMovementForm } from "./StockMovementForm";

const initialItems: InventoryItem[] = [
  {
    id: 1,
    name: "วิทยุสื่อสาร",
    item_type: "equipment",
    quantity_on_hand: 5,
    reorder_point: null,
  },
  {
    id: 2,
    name: "กระดาษ",
    item_type: "consumable",
    quantity_on_hand: 10,
    reorder_point: null,
  },
];

let items: InventoryItem[];
let movements: Array<CreateInventoryMovementRequest & { inventory_item_id: number }>;

beforeEach(() => {
  items = initialItems.map((item) => ({ ...item }));
  movements = [];

  server.use(
    http.get("/api/inventory", () => HttpResponse.json(items)),
    http.post("/api/inventory/:id/movements", async ({ params, request }) => {
      const inventoryItemID = Number(params.id);
      const movement = (await request.json()) as CreateInventoryMovementRequest;
      movements.push({ ...movement, inventory_item_id: inventoryItemID });

      const item = items.find((entry) => entry.id === inventoryItemID);
      if (!item) return HttpResponse.json({ error: "not_found" }, { status: 404 });

      const decreasesStock =
        movement.movement_type === "issue" || movement.movement_type === "borrow";
      const quantityOnHand = item.quantity_on_hand +
        (decreasesStock ? -movement.quantity : movement.quantity);
      const updatedItem = { ...item, quantity_on_hand: quantityOnHand };
      items = items.map((entry) => (entry.id === inventoryItemID ? updatedItem : entry));

      return HttpResponse.json(
        {
          movement: {
            id: movements.length,
            inventory_item_id: inventoryItemID,
            movement_type: movement.movement_type,
            quantity: movement.quantity,
            responsible_person: movement.responsible_person ?? null,
            purpose: movement.purpose ?? null,
            due_date: movement.due_date ?? null,
            created_at: "2026-11-11T00:00:00Z",
          },
          quantity_on_hand: quantityOnHand,
        },
        { status: 201 },
      );
    }),
  );
});

function chooseMovementType(type: "issue" | "borrow" | "return" | "receive") {
  fireEvent.change(screen.getByLabelText("ประเภท movement"), { target: { value: type } });
}

function fillRequiredMovementDetails() {
  fireEvent.change(screen.getByLabelText("ผู้รับผิดชอบ"), {
    target: { value: "สมชาย" },
  });
  fireEvent.change(screen.getByLabelText("วัตถุประสงค์"), {
    target: { value: "ใช้งานในหน่วยงาน" },
  });
}

describe("Stock Movement Form", () => {
  it("TestAC_INV_06 issues three consumables and records the movement", async () => {
    // Given: วัสดุสิ้นเปลืองคงเหลือ 10
    render(<StockMovementForm />);
    await screen.findByText("กระดาษ");
    fireEvent.change(screen.getByLabelText("รายการสินค้า"), { target: { value: "2" } });

    // When: เบิก 3 พร้อมผู้รับผิดชอบและวัตถุประสงค์
    fireEvent.change(screen.getByLabelText("จำนวน"), { target: { value: "3" } });
    fillRequiredMovementDetails();
    fireEvent.click(screen.getByRole("button", { name: "บันทึกเบิก" }));

    // Then: คงเหลือ 7 และบันทึกประวัติการเบิกหนึ่งรายการ
    expect(await screen.findByRole("status")).toHaveTextContent("คงเหลือ: 7");
    expect(movements).toHaveLength(1);
    expect(movements[0]).toMatchObject({
      inventory_item_id: 2,
      movement_type: "issue",
      quantity: 3,
      responsible_person: "สมชาย",
      purpose: "ใช้งานในหน่วยงาน",
    });
  });

  it("TestAC_INV_07 borrows one item with a due date and records the movement", async () => {
    // Given: อุปกรณ์ยืม-คืนคงเหลือ 5
    render(<StockMovementForm />);
    await screen.findByText("วิทยุสื่อสาร");

    // When: ยืม 1 พร้อมผู้รับผิดชอบ วัตถุประสงค์ และวันกำหนดคืน
    chooseMovementType("borrow");
    fireEvent.change(screen.getByLabelText("จำนวน"), { target: { value: "1" } });
    fillRequiredMovementDetails();
    fireEvent.change(screen.getByLabelText("วันกำหนดคืน"), {
      target: { value: "2026-11-20" },
    });
    fireEvent.click(screen.getByRole("button", { name: "บันทึกยืม" }));

    // Then: คงเหลือ 4 และมีประวัติการยืมหนึ่งรายการ
    expect(await screen.findByRole("status")).toHaveTextContent("คงเหลือ: 4");
    expect(movements).toHaveLength(1);
    expect(movements[0]).toMatchObject({
      inventory_item_id: 1,
      movement_type: "borrow",
      quantity: 1,
      responsible_person: "สมชาย",
      purpose: "ใช้งานในหน่วยงาน",
      due_date: "2026-11-20",
    });
  });

  it("TestAC_INV_08 returns one borrowed item and records the movement", async () => {
    // Given: อุปกรณ์ที่ถูกยืมไป 1 เหลือ 4
    items = items.map((item) =>
      item.id === 1 ? { ...item, quantity_on_hand: 4 } : item,
    );
    render(<StockMovementForm />);
    await screen.findByText("วิทยุสื่อสาร");

    // When: บันทึกการคืน
    chooseMovementType("return");
    fireEvent.change(screen.getByLabelText("จำนวน"), { target: { value: "1" } });
    fillRequiredMovementDetails();
    fireEvent.click(screen.getByRole("button", { name: "บันทึกคืน" }));

    // Then: คงเหลือ 5 และมีประวัติการคืนหนึ่งรายการ
    expect(await screen.findByRole("status")).toHaveTextContent("คงเหลือ: 5");
    expect(movements).toHaveLength(1);
    expect(movements[0]).toMatchObject({
      inventory_item_id: 1,
      movement_type: "return",
      quantity: 1,
    });
  });

  it("TestAC_INV_09 receives ten items and records the movement", async () => {
    // Given: รายการคงเหลือ 5
    items = items.map((item) =>
      item.id === 1 ? { ...item, quantity_on_hand: 5 } : item,
    );
    render(<StockMovementForm />);
    await screen.findByText("วิทยุสื่อสาร");

    // When: รับของเข้า 10
    chooseMovementType("receive");
    fireEvent.change(screen.getByLabelText("จำนวน"), { target: { value: "10" } });
    fireEvent.click(screen.getByRole("button", { name: "บันทึกรับเข้า" }));

    // Then: คงเหลือ 15 และมีประวัติรับเข้าหนึ่งรายการ
    expect(await screen.findByRole("status")).toHaveTextContent("คงเหลือ: 15");
    expect(movements).toHaveLength(1);
    expect(movements[0]).toMatchObject({
      inventory_item_id: 1,
      movement_type: "receive",
      quantity: 10,
    });
  });

  it("keeps the displayed stock unchanged when the API rejects an issue", async () => {
    // Given: รายการคงเหลือ 5 และ API ปฏิเสธคำขอเพราะยอดไม่เพียงพอ
    server.use(
      http.post("/api/inventory/:id/movements", () =>
        HttpResponse.json({ error: "insufficient_stock" }, { status: 409 }),
      ),
    );
    render(<StockMovementForm />);
    await screen.findByText("วิทยุสื่อสาร");

    // When: เจ้าหน้าที่ส่งคำขอเบิก และ API ตอบ 409
    fireEvent.change(screen.getByLabelText("จำนวน"), { target: { value: "6" } });
    fillRequiredMovementDetails();
    fireEvent.click(screen.getByRole("button", { name: "บันทึกเบิก" }));

    // Then: แสดงข้อผิดพลาดจากผล API และยอดไม่เปลี่ยน
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "จำนวนคงเหลือไม่เพียงพอสำหรับทำรายการ",
    );
    expect(screen.getByLabelText("จำนวนคงเหลือ")).toHaveTextContent("5");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });

  it.each(["issue", "borrow"] as const)(
    "TestAC_INV_10 rejects %s without a responsible person and leaves stock unchanged",
    async (type) => {
      // Given: รายการคงเหลือ 10
      render(<StockMovementForm />);
      await screen.findByText("กระดาษ");
      fireEvent.change(screen.getByLabelText("รายการสินค้า"), { target: { value: "2" } });

      // When: เบิกหรือยืมโดยไม่ระบุผู้รับผิดชอบ
      chooseMovementType(type);
      fireEvent.change(screen.getByLabelText("จำนวน"), { target: { value: "1" } });
      fireEvent.change(screen.getByLabelText("วัตถุประสงค์"), {
        target: { value: "ใช้ในหน่วยงาน" },
      });
      if (type === "borrow") {
        fireEvent.change(screen.getByLabelText("วันกำหนดคืน"), {
          target: { value: "2026-11-20" },
        });
      }
      fireEvent.click(
        screen.getByRole("button", {
          name: type === "issue" ? "บันทึกเบิก" : "บันทึกยืม",
        }),
      );

      // Then: ระบบปฏิเสธและคงเหลือ 10
      expect(screen.getByLabelText("ผู้รับผิดชอบ")).toBeRequired();
      expect(movements).toHaveLength(0);
      expect(screen.getByLabelText("จำนวนคงเหลือ")).toHaveTextContent("10");
    },
  );

  it.each(["issue", "borrow"] as const)(
    "TestAC_INV_11 rejects %s without a purpose and leaves stock unchanged",
    async (type) => {
      // Given: รายการคงเหลือ 10
      render(<StockMovementForm />);
      await screen.findByText("กระดาษ");
      fireEvent.change(screen.getByLabelText("รายการสินค้า"), { target: { value: "2" } });

      // When: เบิกหรือยืมโดยไม่ระบุวัตถุประสงค์
      chooseMovementType(type);
      fireEvent.change(screen.getByLabelText("จำนวน"), { target: { value: "1" } });
      fireEvent.change(screen.getByLabelText("ผู้รับผิดชอบ"), {
        target: { value: "สมชาย" },
      });
      if (type === "borrow") {
        fireEvent.change(screen.getByLabelText("วันกำหนดคืน"), {
          target: { value: "2026-11-20" },
        });
      }
      fireEvent.click(
        screen.getByRole("button", {
          name: type === "issue" ? "บันทึกเบิก" : "บันทึกยืม",
        }),
      );

      // Then: ระบบปฏิเสธและคงเหลือ 10
      expect(screen.getByLabelText("วัตถุประสงค์")).toBeRequired();
      expect(movements).toHaveLength(0);
      expect(screen.getByLabelText("จำนวนคงเหลือ")).toHaveTextContent("10");
    },
  );

  it("TestAC_INV_12 rejects borrowing without a due date and leaves stock unchanged", async () => {
    // Given: อุปกรณ์ยืม-คืนคงเหลือ 5
    render(<StockMovementForm />);
    await screen.findByText("วิทยุสื่อสาร");

    // When: ยืมโดยไม่ระบุวันกำหนดคืน
    chooseMovementType("borrow");
    fireEvent.change(screen.getByLabelText("จำนวน"), { target: { value: "1" } });
    fillRequiredMovementDetails();
    fireEvent.click(screen.getByRole("button", { name: "บันทึกยืม" }));

    // Then: ระบบปฏิเสธและคงเหลือ 5
    expect(screen.getByLabelText("วันกำหนดคืน")).toBeRequired();
    expect(movements).toHaveLength(0);
    await waitFor(() =>
      expect(screen.getByLabelText("จำนวนคงเหลือ")).toHaveTextContent("5"),
    );
  });
});
