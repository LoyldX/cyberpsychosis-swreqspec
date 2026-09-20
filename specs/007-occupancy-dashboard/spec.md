# Feature: Occupancy Dashboard
Spec ID: FR-BI-01 | Source: `draft/refine-req.md` | Status: Draft

## Goal
แสดงจำนวนผู้ใช้บริการที่อยู่ภายในยิมจากข้อมูลการผ่านเข้า-ออกประตู

## Scope
### In scope
- FR-BI-01
- NFR-PERF-02, DOM-CAP-01, IF-GATE-01
### Out of scope
- การกำหนดรอบ refresh ที่เป็นตัวเลข เพราะ source ระบุว่ายังต้องกำหนด

## Constraints
- IF-GATE-01 เชื่อมต่อข้อมูลการผ่านเข้า-ออกจากประตูยิม

## Requirements
- FR-BI-01 ระบบต้องดึงข้อมูลการผ่านประตูมาประมวลผลและแสดงจำนวนผู้ใช้บริการภายในยิมแบบทันที

## Quality Requirements
- NFR-PERF-02 Dashboard ต้องอัปเดตข้อมูลอัตโนมัติเมื่อมีการผ่านประตู โดยรอบเวลาต้องให้ทีมกำหนด

## Domain Rules
- DOM-CAP-01 การคำนวณต้องเพิ่ม 1 เมื่อผ่านเข้า และลด 1 เมื่อผ่านออก

## Acceptance Criteria
ยังไม่มี Acceptance Criteria ใน `draft/refine-req.md` จึงยังไม่สร้าง AC ใหม่

## Assumptions & Open Questions
- Q3 ประตูทางเข้ายิมบันทึกทั้งขาเข้าและขาออกหรือไม่

## Traceability
| Source ID | รายการ |
|---|---|
| FR-BI-01 | Requirements |
| NFR-PERF-02 | Quality Requirements |
| DOM-CAP-01 | Domain Rules |
| IF-GATE-01 | Constraints |
