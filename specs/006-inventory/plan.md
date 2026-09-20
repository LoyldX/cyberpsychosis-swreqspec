# แผน: จัดการสินค้าคงคลังและอุปกรณ์

## 1. สรุปแนวทาง

ระบบจะจัดการรายการอุปกรณ์และพัสดุตาม FR-INV-01
ระบบจะบันทึกการเบิก ยืม-คืน และรับของเข้าตาม FR-INV-02
การเปลี่ยนยอดทุกครั้งต้องอยู่ใน ACID transaction ตาม NFR-DATA-01
ระบบจะแจ้งเตือนเมื่อยอดถึง Reorder Point ตาม FR-INV-03 และ DOM-INV-02
ประเภทสิ่งของและฐานข้อมูลยังรอคำตอบ Q1 และ Q4

## 2. เทคโนโลยีที่ใช้

| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| PostgreSQL หรือ MySQL | CON-DB-01 | ยังไม่เลือกตัวใดตัวหนึ่งตาม Q4 |
| หน้าบ้านและหลังบ้าน | ทีมเลือกเอง ไม่ได้มาจาก spec | ยังไม่ระบุเทคโนโลยี |

## 3. โมเดลข้อมูล

| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| Inventory item | item reference, item type, quantity, reorder point | FR-INV-01, FR-INV-03, DOM-INV-02 |
| Stock movement | item reference, movement type, quantity, responsible person, purpose, due date when borrowed | FR-INV-02, DOM-INV-01 |

ประเภท item ยังไม่กำหนดจนกว่าจะตอบ Q1

## 4. API / หน้าจอ

| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| GET/POST/PUT/DELETE `/inventory/items` | รายการ item / ผลการจัดการรายการ | FR-INV-01 |
| POST `/inventory/movements` | ประเภท movement, จำนวน, ผู้รับผิดชอบ, วัตถุประสงค์, กำหนดคืน / ยอดใหม่ | FR-INV-02, DOM-INV-01, NFR-DATA-01 |
| GET `/inventory/alerts` | จุดสั่งซื้อและยอดคงเหลือ / รายการแจ้งเตือน | FR-INV-03, DOM-INV-02 |

## 5. ตารางตรวจ Constraints

| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| CON-DB-01 | จัดเก็บ item และ stock movement ใน relational database | ใช้แล้ว แต่ชนิดฐานข้อมูลยังรอ Q4 |
| DOM-INV-01 | ฟิลด์ผู้รับผิดชอบ วัตถุประสงค์ และกำหนดคืน | ใช้แล้ว |
| DOM-INV-02 | ฟิลด์ reorder point และการสร้าง alert | ใช้แล้ว |

## 6. แผนทดสอบจาก Acceptance Criteria

ยังไม่มี Acceptance Criteria ใน spec จึงยังสร้าง test ที่อ้าง AC ไม่ได้

## 7. ลำดับงาน

1. ตอบ Q1 เรื่องประเภทสิ่งของใน inventory
2. ตอบ Q4 เรื่อง PostgreSQL หรือ MySQL
3. ออกแบบ item และ reorder point ตาม FR-INV-01, FR-INV-03, DOM-INV-02
4. ออกแบบ stock movement ตาม FR-INV-02 และ DOM-INV-01
5. ออกแบบ transaction และ concurrency control ตาม NFR-DATA-01
6. เพิ่ม Acceptance Criteria และทดสอบยอดติดลบ/การใช้งานพร้อมกัน

## 8. สิ่งที่ยังไม่ทำ

- Q1 inventory ครอบคลุมพัสดุประเภทใดบ้าง
  ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ
- Q4 จะเลือก PostgreSQL หรือ MySQL เป็นฐานข้อมูลหลัก
  ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ

