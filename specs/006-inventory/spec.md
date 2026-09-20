# Feature: จัดการสินค้าคงคลังและอุปกรณ์
Spec ID: FR-INV-01 | Source: `draft/refine-req.md` | Status: Draft

## Goal
เจ้าหน้าที่จัดการรายการอุปกรณ์และพัสดุ ติดตามการเคลื่อนไหวของสต็อก และรับการแจ้งเตือนเมื่อถึงระดับขั้นต่ำ

## Scope
### In scope
- FR-INV-01, FR-INV-02, FR-INV-03
- NFR-DATA-01, DOM-INV-01, DOM-INV-02
- CON-DB-01
### Out of scope
- POS และการขายสินค้าหน้าร้าน
- ประเภทพัสดุที่แน่นอน ซึ่งยังเป็น Q1

## Constraints
- CON-DB-01 ใช้ PostgreSQL หรือ MySQL จัดเก็บข้อมูลสินค้าคงคลัง โดยยังไม่เลือกตัวใดตัวหนึ่ง

## Requirements
- FR-INV-01 เจ้าหน้าที่ต้องเพิ่ม แก้ไข ลบ ค้นหา และตรวจสอบจำนวนคงเหลือของอุปกรณ์และพัสดุได้
- FR-INV-02 ระบบต้องบันทึกการเบิกวัสดุ การยืม-คืนอุปกรณ์ และการรับของใหม่เข้าสต็อก
- FR-INV-03 ระบบต้องแจ้งเตือนเจ้าหน้าที่เมื่อจำนวนคงเหลือถึง Reorder Point

## Quality Requirements
- NFR-DATA-01 การเพิ่มหรือตัดยอดต้องทำงานแบบ ACID transaction เพื่อป้องกันยอดผิดพลาดหรือติดลบจากการใช้งานพร้อมกัน

## Domain Rules
- DOM-INV-01 การเบิกหรือยืมต้องบันทึกผู้รับผิดชอบ วัตถุประสงค์ และกำหนดคืนเมื่อเป็นการยืม
- DOM-INV-02 อุปกรณ์และของใช้จำเป็นต้องมีระดับแจ้งเตือนขั้นต่ำ

## Acceptance Criteria
ยังไม่มี Acceptance Criteria ใน `draft/refine-req.md` จึงยังไม่สร้าง AC ใหม่

## Assumptions & Open Questions
- Q1 inventory ครอบคลุมพัสดุประเภทใดบ้าง
- Q4 จะเลือก PostgreSQL หรือ MySQL เป็นฐานข้อมูลหลัก

## Traceability
| Source ID | รายการ |
|---|---|
| FR-INV-01, FR-INV-02, FR-INV-03 | Requirements |
| NFR-DATA-01 | Quality Requirements |
| DOM-INV-01, DOM-INV-02 | Domain Rules |
| CON-DB-01 | Constraints |

