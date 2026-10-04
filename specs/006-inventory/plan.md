# แผน: จัดการสินค้าคงคลังและอุปกรณ์
Spec: `specs/006-inventory/spec.md` (Draft v2 — ยังไม่ผ่าน clarify) | สร้างเมื่อ 4 ตุลาคม 2569 | หน้านี้เป็นผลลัพธ์ที่ AI ร่าง คนตรวจต้องยืนยัน

## 1. สรุปแนวทาง
ระบบจัดการรายการอุปกรณ์และพัสดุ (เพิ่ม แก้ไข ลบ ค้นหา) ตาม FR-INV-01
บันทึกการเบิก ยืม-คืน และรับของเข้าสต็อกตาม FR-INV-02 พร้อมฟิลด์ผู้รับผิดชอบ วัตถุประสงค์ และกำหนดคืนตาม DOM-INV-01
ทุกการเพิ่มหรือตัดยอดอยู่ใน ACID transaction ตาม NFR-DATA-01 และแจ้งเตือนเมื่อถึง Reorder Point ตาม FR-INV-03 และ DOM-INV-02
ประเภทพัสดุรอ Q1 และชนิดฐานข้อมูลรอ Q4 จึงยังไม่กำหนดฟิลด์ประเภทและยังไม่เลือกเอนจินฐานข้อมูล

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| ฐานข้อมูลเชิงสัมพันธ์ (ACID) | CON-DB-01, NFR-DATA-01 | ยังไม่เลือก รอ Q4 (ค่าเดาคือ PostgreSQL) |
| หน้าบ้าน: React (Vite) | ทีมเลือกเอง ไม่ได้มาจาก spec | ใช้กับหน้าจอ UI-INV-01/02/03 (Candidate) |
| หลังบ้าน: Python FastAPI | ทีมเลือกเอง ไม่ได้มาจาก spec | ค่าเริ่มต้นของรายวิชา |

## 3. โมเดลข้อมูล
| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| Inventory item | item reference, quantity, reorder point (ระดับแจ้งเตือนขั้นต่ำ) | FR-INV-01, FR-INV-03, DOM-INV-02 |
| Stock movement | item reference, movement type (เบิก / ยืม / คืน / รับเข้า), quantity, responsible person, purpose, due date (เฉพาะการยืม) | FR-INV-02, DOM-INV-01, NFR-DATA-01 |

- ฟิลด์ประเภทพัสดุ (item type) ยังไม่สร้างจนกว่าจะตอบ Q1
- ค่า reorder point ของแต่ละรายการไม่ได้ระบุตัวเลข (ยังไม่มีหมายเลข Q)

## 4. API / หน้าจอ
| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| GET/POST/PUT/DELETE `/inventory/items` | รายการ item / ผลการจัดการรายการ | FR-INV-01 |
| POST `/inventory/movements` | ประเภท movement, จำนวน, ผู้รับผิดชอบ, วัตถุประสงค์, กำหนดคืน / ยอดใหม่ | FR-INV-02, DOM-INV-01, NFR-DATA-01 |
| GET `/inventory/alerts` | จุดสั่งซื้อและยอดคงเหลือ / รายการแจ้งเตือน | FR-INV-03, DOM-INV-02 |
| หน้าจอรายการอุปกรณ์ (UI-INV-01/03, Candidate) | รายการ + ช่องค้นหา / แยกรายการถึง Reorder Point | FR-INV-01, FR-INV-03 |
| ฟอร์มเบิก/ยืม (UI-INV-02, Candidate) | ข้อมูลการเบิก / บันทึกหรือแจ้งข้อผิดพลาด | DOM-INV-01, FR-INV-02 |

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| CON-DB-01 | Inventory item และ Stock movement (ข้อ 3) เก็บใน relational DB | ใช้แล้วเชิงออกแบบ แต่ชนิดฐานข้อมูลยังรอ Q4 |
| DOM-INV-01 | ฟิลด์ responsible person, purpose, due date ใน POST `/inventory/movements` (ข้อ 4) | ใช้แล้ว |
| DOM-INV-02 | ฟิลด์ reorder point และ GET `/inventory/alerts` (ข้อ 3, 4) | ใช้แล้ว (ค่าตัวเลขรายการยังเป็น TBD) |

## 6. แผนทดสอบจาก Acceptance Criteria
ทุก AC เป็น Candidate (ยังไม่ผ่านการยืนยันจากทีม)

| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-INV-01 (Candidate) | `test_AC_INV_01_add_edit_delete_search_item` | เพิ่ม แก้ไข ลบ ค้นหารายการ และตรวจจำนวนคงเหลือ |
| AC-INV-02 (Candidate) | `test_AC_INV_02_movement_history_recorded` | บันทึกเบิก ยืม-คืน และรับเข้า แล้วตรวจว่ามีประวัติ |
| AC-INV-03 (Candidate) | `test_AC_INV_03_alert_when_quantity_reaches_reorder_point` | ตั้งยอดให้เท่ากับ reorder point แล้วตรวจว่ามีการแจ้งเตือน |
| AC-INV-04 (Candidate) | `test_AC_INV_04_concurrent_updates_no_negative_stock` | ส่งการตัด/เพิ่มยอดพร้อมกัน ตรวจว่ายอดไม่ติดลบและไม่ผิด (จำนวนพร้อมกัน = ทีมเลือกเอง) |
| AC-INV-05 (Candidate) | `test_AC_INV_05_rejects_borrow_without_required_fields` | ส่งการเบิกหรือยืมที่ขาดผู้รับผิดชอบ วัตถุประสงค์ หรือวันคืนเมื่อยืม ตรวจว่าไม่บันทึก |
| AC-INV-06 (Candidate) | `test_AC_INV_06_item_requires_minimum_alert_level` | สร้างรายการโดยไม่มีระดับแจ้งเตือน ตรวจว่าบันทึกไม่ได้ (ค่าตัวเลข = TBD) |
| AC-INV-07 (Candidate) | `test_AC_INV_07_inventory_stored_in_relational_db` | บันทึกข้อมูลแล้วตรวจในฐานข้อมูลจริง (ชนิดรอ Q4) |

## 7. ลำดับงาน
1. ตอบ Q1 เรื่องประเภทพัสดุ และตอบ Q4 เรื่องชนิดฐานข้อมูล (ก่อนสร้างฟิลด์ประเภทและเชื่อม DB)
2. สร้าง Inventory item และ reorder point (FR-INV-01, DOM-INV-02, AC-INV-06, รอ Q1)
3. สร้าง CRUD และการค้นหา item (FR-INV-01, AC-INV-01, รอ Q1)
4. สร้าง Stock movement พร้อมฟิลด์บังคับ (FR-INV-02, DOM-INV-01, AC-INV-02, AC-INV-05)
5. ออกแบบ transaction แบบ ACID และการควบคุมพร้อมกัน (NFR-DATA-01, AC-INV-04, รอ Q4)
6. สร้างการแจ้งเตือน Reorder Point (FR-INV-03, AC-INV-03)
7. เชื่อมข้อมูลกับฐานข้อมูลจริง (CON-DB-01, AC-INV-07, รอ Q4)
8. สร้างหน้าจอ UI-INV-01/02/03 (Candidate, รอทีมยืนยัน)

## 8. สิ่งที่ยังไม่ทำ
- **Q1** Inventory ครอบคลุมพัสดุประเภทใดบ้าง
  ส่วนที่เกี่ยวข้องกับข้อนี้ (ฟิลด์ item type, ข้อ 7 ลำดับที่ 2 และ 3) จะยังไม่สร้างจนกว่าจะได้คำตอบ
- **Q4** เลือก PostgreSQL หรือ MySQL เป็นฐานข้อมูลหลัก
  ส่วนที่เกี่ยวข้องกับข้อนี้ (ACID transaction, การเชื่อม DB จริง, AC-INV-04 และ AC-INV-07) จะยังไม่สร้างจนกว่าจะได้คำตอบ
- ยังไม่มีหมายเลข Q: ค่าระดับแจ้งเตือนขั้นต่ำของแต่ละรายการ (DOM-INV-02)
- หมายเหตุ NS-1 (ไม่ใช่ requirement): การเพิ่มอุปกรณ์ประเภทใหม่เป็นเรื่องการจัดซื้อ ไม่อยู่ในแผนนี้
