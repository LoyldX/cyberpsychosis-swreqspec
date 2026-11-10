# แผน: จัดการสินค้าคงคลังและอุปกรณ์

## 1. สรุปแนวทาง
ฟีเจอร์นี้จัดการอุปกรณ์ยืม-คืนและวัสดุสิ้นเปลือง โดยไม่รวมสินค้าขายหรือทรัพย์สินอื่น (FR-INV-01, ASM-01).
เจ้าหน้าที่ใช้เว็บ React (Vite) และ Backend API ที่สร้างด้วย Go/Gin เชื่อม PostgreSQL (CON-TECH-01, CON-TECH-02, CON-DB-01).
เก็บรายการสินค้าพร้อมยอดคงเหลือและ Reorder Point ที่ไม่บังคับทุกรายการ (FR-INV-01, DOM-INV-02, ASM-02).
การเบิก ยืม คืน และรับเข้าเก็บประวัติและเปลี่ยนยอดภายใน transaction เดียว ป้องกันยอดติดลบเมื่อคำขอชนกัน (FR-INV-02, NFR-DATA-01, ASM-05, ASM-06).
แสดงการแจ้งเตือนสต็อกขั้นต่ำในเว็บเมื่อรายการที่ติดตามลดถึงจุดสั่งซื้อ; ไม่ทำระบบล็อกอินหรือกำหนดสิทธิ์ซึ่งอยู่นอกขอบเขต (FR-INV-03, ASM-03, Q-01).

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| Go และ Gin | CON-TECH-01 | Backend API |
| React (Vite) | CON-TECH-02 | เว็บแอปเท่านั้น ไม่มีแอปมือถือเนทีฟ |
| PostgreSQL | CON-DB-01 | ที่เก็บรายการ inventory และ movement |
| pgx/v5 (`pgxpool`) และ SQL โดยตรง | ทีมเลือกเอง ไม่ได้มาจาก spec | เป็นแนวทางร่วมใน `specs/000-shared/README.md`; ใช้ควบคุม transaction และ concurrency |
| golang-migrate และ migration แบบ SQL | ทีมเลือกเอง ไม่ได้มาจาก spec | แนวทางร่วมใน `specs/000-shared/README.md` |
| Go `testing`/`testify`, `httptest`, PostgreSQL จริงสำหรับ integration/concurrency test | ทีมเลือกเอง ไม่ได้มาจาก spec | ตามแนวทางทดสอบร่วมใน `specs/000-shared/README.md`; ไม่ใช้ mock แทน DB ใน AC-INV-14 |
| Vitest, Testing Library และ MSW | ทีมเลือกเอง ไม่ได้มาจาก spec | แนวทางร่วมสำหรับทดสอบหน้าจอด้วย API จำลอง |

## 3. โมเดลข้อมูล
| Entity | ฟิลด์หลักที่เสนอ | รองรับ |
|---|---|---|
| `inventory_items` | `id`, `name`, `item_type` (อุปกรณ์ยืม-คืน/วัสดุสิ้นเปลือง), `quantity_on_hand` (เริ่มต้น 0), `reorder_point` (nullable), `deleted_at` (nullable สำหรับซ่อนจาก catalog โดยยังเก็บประวัติ) | FR-INV-01, FR-INV-03, DOM-INV-02, AC-INV-01..05, AC-INV-13 |
| `inventory_movements` | `id`, `inventory_item_id`, `movement_type` (เบิก/ยืม/คืน/รับเข้า), `quantity`, `responsible_person`, `purpose`, `due_date` (เมื่อเป็นการยืม), `created_at` | FR-INV-02, DOM-INV-01, AC-INV-06..12 |
| การเปลี่ยนยอดและบันทึก movement | transaction เดียว: เปลี่ยน `quantity_on_hand` และเพิ่ม movement สำเร็จหรือล้มเหลวพร้อมกัน; ปฏิเสธยอดติดลบ | NFR-DATA-01, AC-INV-14, ASM-05, ASM-07 |

ฟิลด์ด้านบนเป็นข้อเสนอเชิงโครงสร้างเพื่อรองรับ ID ที่อ้างถึง ไม่เพิ่มขอบเขตหมวดสินค้าเกิน ASM-01 และไม่สร้างข้อมูลบัญชี/สิทธิ์เจ้าหน้าที่ (Q-01 อยู่นอก scope) รายละเอียดชนิดข้อมูลและการเชื่อม movement คืนกับรายการยืมเฉพาะรายการยังไม่กำหนดใน spec จึงไม่กำหนดกฎธุรกิจเพิ่มเติมในแผนนี้ (FR-INV-02).

## 4. API / หน้าจอ
| Method / path หรือหน้าจอ | Input / output หลัก | รองรับ |
|---|---|---|
| `POST /api/inventory` | ชื่อ, ประเภท, Reorder Point (เลือกได้) → รายการใหม่ที่มียอดเริ่มต้น 0 | FR-INV-01, DOM-INV-02 |
| `GET /api/inventory?name=` | คำค้นชื่อ → รายการที่ยังไม่ถูกซ่อนพร้อมยอดคงเหลือ | FR-INV-01 |
| `GET /api/inventory/{id}` | ID → รายละเอียดและยอดคงเหลือ | FR-INV-01 |
| `PATCH /api/inventory/{id}` | ฟิลด์รายการที่แก้ไข → รายการหลังแก้ | FR-INV-01 |
| `DELETE /api/inventory/{id}` | ID → ซ่อนจาก catalog โดยไม่ลบประวัติ | FR-INV-01, FR-INV-02, AC-INV-03 |
| `POST /api/inventory/{id}/movements` | ประเภท movement, จำนวน, ผู้รับผิดชอบ, วัตถุประสงค์, วันกำหนดคืนเมื่อยืม → ยอดใหม่และ movement; เมื่อ DB ไม่พร้อมให้ error และไม่มีการเปลี่ยนยอดบางส่วน | FR-INV-02, DOM-INV-01, NFR-DATA-01, ASM-04, ASM-05, ASM-07 |
| `GET /api/inventory/{id}/movements` | ID → ประวัติ movement ของรายการ | FR-INV-02 |
| `GET /api/inventory/alerts` | ไม่มี input → รายการที่ movement ทำให้ยอดถึง Reorder Point สำหรับแสดงบนเว็บ | FR-INV-03, DOM-INV-02, AC-INV-13 |
| Inventory Catalog / Item Detail | ค้นหา, ดู/แก้รายการและยอด, ซ่อนรายการจาก catalog; แสดงประวัติรายการ | FR-INV-01, FR-INV-02 |
| Stock Movement Form | เบิก/ยืม/คืน/รับเข้า; บังคับผู้รับผิดชอบและวัตถุประสงค์เมื่อเบิกหรือยืม และวันกำหนดคืนเมื่อยืม | FR-INV-02, DOM-INV-01 |
| Low Stock Alert ในเว็บ | แสดงรายการที่ถึง Reorder Point หลังการเคลื่อนไหวของสต็อก | FR-INV-03, ASM-02, ASM-03 |

เส้นทาง API เป็นข้อเสนอทางเทคนิคที่ยังไม่ใช่สัญญาเพิ่มจาก spec; ไม่รวม auth endpoints หรือการบังคับสิทธิ์ (Q-01 และ Out of scope).

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| CON-DB-01 | PostgreSQL สำหรับ `inventory_items`/`inventory_movements`; transaction และ integration test ใช้ PostgreSQL จริง; เมื่อ DB ไม่พร้อมให้ rollback, ปฏิเสธคำขอและแสดง error โดยไม่มีการเปลี่ยนยอดบางส่วน (ASM-07) | ใช้แล้ว |
| CON-TECH-01 | Go/Gin สำหรับ API และ handler | ใช้แล้ว |
| CON-TECH-02 | React (Vite) สำหรับ catalog, movement form และหน้าแจ้งเตือน; ไม่มี native app | ใช้แล้ว |
| DOM-INV-01 | Movement บันทึกผู้รับผิดชอบ/วัตถุประสงค์เมื่อเบิกหรือยืม และกำหนดวันคืนเมื่อยืม; ตรวจ input ก่อนเปลี่ยนยอด | ใช้แล้ว |
| DOM-INV-02 | `reorder_point` เป็น nullable; แจ้งเตือนรายการที่กำหนดค่าไว้เท่านั้น | ใช้แล้ว |
| IF | ไม่มี IF ใน spec นี้ | ไม่มี ID ให้ตรวจ |

## 6. แผนทดสอบจาก Acceptance Criteria
| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-INV-01 | `TestAC_INV_01` | สร้างรายการใหม่ผ่าน API แล้วตรวจว่าปรากฏใน catalog และมียอดเริ่มต้น 0 ตามคำตัดสินของทีม |
| AC-INV-02 | `TestAC_INV_02` | แก้ข้อมูลรายการ แล้วอ่านซ้ำและตรวจค่าที่แก้ |
| AC-INV-03 | `TestAC_INV_03` | ลบ/ซ่อนรายการ แล้วตรวจว่าไม่อยู่ใน catalog และ movement history เดิมยังอยู่ (การเก็บประวัติตามคำยืนยันทีม) |
| AC-INV-04 | `TestAC_INV_04` | ค้นชื่อในชุดรายการหลายรายการ แล้วตรวจว่าผลมีรายการที่ตรง |
| AC-INV-05 | `TestAC_INV_05` | เตรียมรายการยอด 10 แล้วอ่าน detail และตรวจยอด 10 |
| AC-INV-06 | `TestAC_INV_06` | เบิกวัสดุสิ้นเปลือง 3 จากยอด 10 พร้อมผู้รับผิดชอบ/วัตถุประสงค์ ตรวจยอด 7 และมี movement 1 รายการ |
| AC-INV-07 | `TestAC_INV_07` | ยืมอุปกรณ์ 1 จากยอด 5 พร้อมข้อมูลบังคับและวันคืน ตรวจยอด 4 และมีประวัติยืม |
| AC-INV-08 | `TestAC_INV_08` | คืนอุปกรณ์ที่ถูกยืม 1 จากยอด 4 ตรวจยอด 5 และมีประวัติคืน |
| AC-INV-09 | `TestAC_INV_09` | รับเข้า 10 จากยอด 5 ตรวจยอด 15 และมีประวัติรับเข้า |
| AC-INV-10 | `TestAC_INV_10` | ส่งคำขอเบิก/ยืมที่ไม่มีผู้รับผิดชอบ ตรวจว่าถูกปฏิเสธและยอดคงเดิม |
| AC-INV-11 | `TestAC_INV_11` | ส่งคำขอเบิก/ยืมที่ไม่มีวัตถุประสงค์ ตรวจว่าถูกปฏิเสธและยอดคงเดิม |
| AC-INV-12 | `TestAC_INV_12` | ส่งคำขอยืมที่ไม่มี due date ตรวจว่าถูกปฏิเสธและยอดคงเดิม |
| AC-INV-13 | `TestAC_INV_13` | ทดสอบ API/UI: รายการ A ลดจาก 6 ถึง 5 แล้วแสดงแจ้งเตือนบนเว็บ; รายการ B ที่ไม่มี Reorder Point ไม่แจ้งเตือน |
| AC-INV-14 | `TestAC_INV_14` | Integration test ด้วย PostgreSQL จริง: ใช้ goroutines/barrier ส่งเบิก 3 สองคำขอพร้อมกันจากยอด 5 ตรวจสำเร็จ 1, ปฏิเสธ 1, ยอดสุดท้าย 2 |

## 7. ลำดับงาน
1. เตรียม migration และโครงสร้าง `inventory_items`/`inventory_movements`; สร้างรายการด้วยยอด 0 และใช้ movement รับเข้าเพื่อเพิ่มยอด; ใช้ soft-delete เพื่อรักษาประวัติ (FR-INV-01, FR-INV-02, AC-INV-01, AC-INV-03, AC-INV-09).
2. สร้าง catalog API และทดสอบ create/read/update/search/delete ตาม AC (FR-INV-01, AC-INV-01..05).
3. สร้าง transaction/service สำหรับ issue, borrow, return, receive พร้อม validation ของ DOM-INV-01 และบันทึก movement (FR-INV-02, DOM-INV-01, AC-INV-06..12).
4. เพิ่มการป้องกันยอดติดลบด้วย atomic update/transaction และทดสอบ concurrency บน PostgreSQL จริง (NFR-DATA-01, AC-INV-14).
5. เพิ่มการตรวจ Reorder Point หลัง movement และ API สำหรับรายการที่เข้าเงื่อนไขแจ้งเตือน (FR-INV-03, DOM-INV-02, AC-INV-13).
6. สร้างหน้าเว็บ catalog/detail, movement form, และ alert ตาม flow ของ API (FR-INV-01..03, AC-INV-01..13).
7. เพิ่มชุดทดสอบ frontend/backend ครบทุก AC, รัน migration และ integration tests พร้อมตรวจ rollback เมื่อ DB error (AC-INV-01..14, ASM-07).

## 8. สิ่งที่ยังไม่ทำ
| รหัส | Open Question จาก spec | ขอบเขตที่ยังไม่สร้าง |
|---|---|---|
| Q-01 | วิธีระบุตัวตนเจ้าหน้าที่ก่อนใช้ระบบและความแตกต่างของสิทธิ์เจ้าหน้าที่ยังเป็นอย่างไร (Product Owner) | ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ: ไม่ทำ login/session และไม่กำหนด authorization ในแผน Inventory ตาม Out of scope |

### คำตัดสินของทีมระหว่างร่างแผน
- เมื่อลบรายการที่มี movement history ให้ซ่อนรายการจาก catalog และเก็บประวัติไว้; แผนจึงใช้ soft delete (FR-INV-02, AC-INV-03).
- รายการใหม่มียอดตั้งต้น 0 และเพิ่มยอดด้วย movement รับเข้า; API สร้างรายการจึงไม่รับจำนวนตั้งต้น (FR-INV-01, FR-INV-02, AC-INV-01, AC-INV-09).
