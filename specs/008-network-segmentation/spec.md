# Feature: Network Segmentation
Spec ID: NFR-SEC-01 | Source: `docs/srs/srs.md` (v5) | Status: Draft v1 | อัปเดต: 4 ตุลาคม 2569

## Goal
แยกเครือข่าย Wi-Fi ของพนักงานและของสมาชิกออกจากกันอย่างเด็ดขาด เพื่อป้องกันการโจรกรรมข้อมูล

## Scope
### In scope
- NFR-SEC-01 แยกเครือข่าย Wi-Fi ของพนักงานและสมาชิกด้วย VLAN
### Out of scope
- ตัดออกทั้งระบบ (refine-req §2.1–2.2): POS และการขายหน้าร้าน; เครื่องสแกนเวลาและการซิงค์ HR; Data Privacy & Encryption (TLS 1.3, AES-256, PDPA); Nutrition (FR-NUT-01, IF-NUT-01); ระบบเช็กสถานะเครื่องเล่นและคิวเครื่องเล่น (Q10)

## Constraints
ไม่มี CON, DOM หรือ IF ใน spec นี้

## Requirements
ไม่มี FR ใน spec นี้

## Quality Requirements (ยอมรับแล้ว)
| ID | ข้อกำหนด |
|---|---|
| NFR-SEC-01 | Network Segmentation (VLAN): แยกเครือข่าย Wi-Fi ของพนักงานและ Wi-Fi สมาชิกออกจากกันอย่างเด็ดขาด เพื่อป้องกันการโจรกรรมข้อมูล — เกณฑ์: Wi-Fi สมาชิกเข้าถึงเครือข่ายพนักงานไม่ได้ |

## Domain Rules
ไม่มีใน spec นี้

## Assumptions & Open Questions
- ยังไม่มีหมายเลข Q: ช่วง VLAN ของ Wi-Fi พนักงานและ Wi-Fi สมาชิก (เป็นการตัดสินใจของทีม ไม่ได้มาจาก spec)
- ยังไม่มีหมายเลข Q: อุปกรณ์เครือข่ายที่ใช้ รุ่นและผู้ผลิต รวมถึง topology (เป็นการตัดสินใจของทีม ไม่ได้มาจาก spec)
- ยังไม่มีหมายเลข Q: ตำแหน่งที่บันทึกการตั้งค่าเครือข่าย (spec นี้ไม่มี application code)
- หมายเหตุ: spec นี้ไม่มี application entity หรือ API จึงไม่มีโค้ดให้สร้าง

## Traceability
| Source | ส่วนใน spec | ที่มา (refine-req v5) |
|---|---|---|
| NFR-SEC-01 | Quality Requirements | §4 (v4) |
