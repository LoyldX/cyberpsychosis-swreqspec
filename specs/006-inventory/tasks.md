# Tasks: จัดการสินค้าคงคลังและอุปกรณ์
- Feature: จัดการสินค้าคงคลังและอุปกรณ์ (`006-inventory`)
- Spec ID: FR-INV-01, FR-INV-02, FR-INV-03, NFR-DATA-01, DOM-INV-01, DOM-INV-02, CON-DB-01 (AC ทั้งหมดเป็น Candidate)
- อ้างอิง plan.md: `specs/006-inventory/plan.md`
- วันที่: 4 ตุลาคม 2569
- สรุป: ทั้งหมด 10 task, รอ Q-xx 6 task (พร้อมทำ 4 task). ยังไม่มี task ใดถูกเริ่มทำ
- หมายเหตุ: ไฟล์ตามโครงค่าเริ่มต้น (React Vite / FastAPI) ทีมเลือกเอง ไม่ได้มาจาก spec

## รายการ task

### T-01 สร้างโมเดล item (reference, quantity, reorder point)
- รองรับ: FR-INV-01, DOM-INV-02, CON-DB-01
- ตรวจด้วย: AC-INV-06 (Candidate)
- ไฟล์ที่แตะ: `backend/app/models/inventory_item.py`, `backend/tests/test_inventory.py`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: `test_AC_INV_06_item_requires_minimum_alert_level` ผ่าน (ไม่บันทึกรายการที่ไม่มีระดับแจ้งเตือน)
- สถานะ: พร้อมทำ

### T-02 เพิ่มฟิลด์ประเภทพัสดุ (item type)
- รองรับ: FR-INV-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ (ค่าประเภทยังไม่กำหนด)
- ไฟล์ที่แตะ: `backend/app/models/inventory_item.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: มีฟิลด์ประเภทตามคำตอบ Q1
- สถานะ: รอ Q1

### T-03 CRUD และค้นหา item
- รองรับ: FR-INV-01
- ตรวจด้วย: AC-INV-01 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/inventory_items.py`, `backend/tests/test_inventory.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_INV_01_add_edit_delete_search_item` ผ่าน
- สถานะ: พร้อมทำ

### T-04 บันทึก stock movement พร้อมฟิลด์บังคับ
- รองรับ: FR-INV-02, DOM-INV-01
- ตรวจด้วย: AC-INV-02, AC-INV-05 (Candidate)
- ไฟล์ที่แตะ: `backend/app/models/stock_movement.py`, `backend/app/api/inventory_movements.py`, `backend/tests/test_inventory.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_INV_02_movement_history_recorded` และ `test_AC_INV_05_rejects_borrow_without_required_fields` ผ่าน
- สถานะ: พร้อมทำ

### T-05 ACID transaction สำหรับการตัดและเพิ่มยอด
- รองรับ: NFR-DATA-01
- ตรวจด้วย: AC-INV-04 (Candidate)
- ไฟล์ที่แตะ: `backend/app/services/stock_transaction.py`, `backend/tests/test_inventory_concurrency.py`
- ต้องทำหลัง: T-04
- เสร็จเมื่อ: `test_AC_INV_04_concurrent_updates_no_negative_stock` ผ่านบนฐานข้อมูลที่ทีมเลือก
- สถานะ: รอ Q4 (ACID ขึ้นกับชนิดฐานข้อมูล)

### T-06 แจ้งเตือนเมื่อถึง Reorder Point
- รองรับ: FR-INV-03, DOM-INV-02
- ตรวจด้วย: AC-INV-03 (Candidate)
- ไฟล์ที่แตะ: `backend/app/services/reorder_alert.py`, `backend/app/api/inventory_alerts.py`, `backend/tests/test_inventory.py`
- ต้องทำหลัง: T-01, T-04
- เสร็จเมื่อ: `test_AC_INV_03_alert_when_quantity_reaches_reorder_point` ผ่าน
- สถานะ: พร้อมทำ

### T-07 เชื่อม inventory กับฐานข้อมูลจริง
- รองรับ: CON-DB-01
- ตรวจด้วย: AC-INV-07 (Candidate)
- ไฟล์ที่แตะ: `backend/app/repositories/inventory_db.py`, `backend/tests/test_inventory_db.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_INV_07_inventory_stored_in_relational_db` ผ่าน ด้วยชนิดฐานข้อมูลที่ทีมเลือกตาม Q4
- สถานะ: รอ Q4

### T-08 สร้างหน้ารายการอุปกรณ์ (ช่องค้นหาและการแยกรายการถึงจุดสั่งซื้อ)
- รองรับ: FR-INV-01, FR-INV-03 (ข้อเสนอ UI-INV-01/03 เป็น Candidate)
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-INV-01/03 เป็น Candidate)
- ไฟล์ที่แตะ: `frontend/src/pages/InventoryList.jsx`, `frontend/src/api/mockInventory.js`
- ต้องทำหลัง: ไม่มี (ใช้ API จำลองตามสัญญาใน plan.md ข้อ 4)
- เสร็จเมื่อ: หน้าจอมีช่องค้นหาและแยกรายการถึง Reorder Point จาก API จำลอง
- สถานะ: รอ Q-xx (UI-INV-01/03 ยังไม่มีหมายเลข Q)

### T-09 สร้างฟอร์มเบิก/ยืม
- รองรับ: DOM-INV-01, FR-INV-02 (ข้อเสนอ UI-INV-02 เป็น Candidate)
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-INV-02 เป็น Candidate)
- ไฟล์ที่แตะ: `frontend/src/pages/StockMovementForm.jsx`
- ต้องทำหลัง: ไม่มี (ใช้ API จำลองตามสัญญาใน plan.md ข้อ 4)
- เสร็จเมื่อ: ฟอร์มบังคับกรอกผู้รับผิดชอบ วัตถุประสงค์ และวันคืนเมื่อเป็นการยืม
- สถานะ: รอ Q-xx (UI-INV-02 ยังไม่มีหมายเลข Q)

### T-10 ต่อหน้าจอกับ API จริงและทดสอบปลายทาง
- รองรับ: FR-INV-01, FR-INV-02, FR-INV-03, DOM-INV-01, CON-DB-01
- ตรวจด้วย: AC-INV-01, AC-INV-02, AC-INV-03, AC-INV-05, AC-INV-07 (Candidate)
- ไฟล์ที่แตะ: `frontend/src/api/inventory.js`, `frontend/src/pages/InventoryList.jsx`, `frontend/src/pages/StockMovementForm.jsx`
- ต้องทำหลัง: T-03, T-04, T-06, T-07, T-08, T-09
- เสร็จเมื่อ: ทดสอบปลายทางผ่านบนฐานข้อมูลที่ทีมเลือก
- สถานะ: รอ Q1 (ประเภทพัสดุ), Q4 (ชนิดฐานข้อมูล), Q-xx (UI-INV-01/02/03)

## ตารางตรวจความครบ

### AC ทั้งหมด (Candidate)
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-INV-01 (Candidate) | T-03, T-10 |
| AC-INV-02 (Candidate) | T-04, T-10 |
| AC-INV-03 (Candidate) | T-06, T-10 |
| AC-INV-04 (Candidate) | T-05 |
| AC-INV-05 (Candidate) | T-04, T-10 |
| AC-INV-06 (Candidate) | T-01 |
| AC-INV-07 (Candidate) | T-07, T-10 |

### Constraint ทั้งหมด
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-DB-01 | T-01 (model), T-07 (ฐานข้อมูลจริง), T-10 |
| DOM-INV-01 | T-04, T-09 |
| DOM-INV-02 | T-01, T-06, T-08 |

## สิ่งที่ยังไม่ทำ
- **Q1** Inventory ครอบคลุมพัสดุประเภทใดบ้าง: รอ T-02, T-10
- **Q4** เลือก PostgreSQL หรือ MySQL: รอ T-05, T-07, T-10
- ยังไม่มีหมายเลข Q: UI-INV-01/02/03 รอ T-08, T-09, T-10
- ยังไม่มีหมายเลข Q: ค่าระดับแจ้งเตือนขั้นต่ำของแต่ละรายการ (DOM-INV-02) ไม่มี task ที่กำหนดค่าตัวเลข
- หมายเหตุ NS-1 (ไม่ใช่ requirement): การจัดซื้ออุปกรณ์ใหม่ ไม่มี task
