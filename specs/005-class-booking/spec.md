# Feature: จองคลาสออกกำลังกาย
Spec ID: FR-BKG-01 | Source: `draft/refine-req.md` | Status: Draft

## Goal
สมาชิกเลือกและจองคลาสออกกำลังกายที่เปิดรับผ่านระบบออนไลน์ โดยไม่สามารถจองซ้อนกับเวลาที่เทรนเนอร์ไม่ว่าง

## Scope
### In scope
- FR-BKG-01 และ FR-BKG-02
- DOM-BKG-01
- NFR-PERF-01
### Out of scope
- การกำหนดค่า latency ที่เป็นตัวเลข เพราะ source ระบุว่ายังต้องกำหนด

## Requirements
- FR-BKG-01 ระบบต้องให้สมาชิกเลือกและจองคลาสที่เปิดรับทางออนไลน์
- FR-BKG-02 ระบบต้องตรวจสอบและบล็อกการจองซ้อนเมื่อเทรนเนอร์ติดสอนคลาสอื่นหรืออยู่นอกเวลาปฏิบัติงาน

## Quality Requirements
- NFR-PERF-01 การตรวจสอบคิวว่างและบล็อกคิวต้องประมวลผลทันที โดยทีมต้องกำหนด latency ที่ชัดเจน

## Domain Rules
- DOM-BKG-01 สมาชิกไม่สามารถจองคลาสในช่วงที่เทรนเนอร์มีคิวอื่นหรืออยู่นอกกะเวลาทำงาน

## Acceptance Criteria
ยังไม่มี Acceptance Criteria ใน `draft/refine-req.md` จึงยังไม่สร้าง AC ใหม่

## Assumptions & Open Questions
ไม่มี Open Question เฉพาะฟีเจอร์นี้ใน source แต่ค่า latency ของ NFR-PERF-01 ยังไม่ถูกกำหนด

## Traceability
| Source ID | รายการ |
|---|---|
| FR-BKG-01, FR-BKG-02 | Requirements |
| NFR-PERF-01 | Quality Requirements |
| DOM-BKG-01 | Domain Rules |

