# Feature: Network Segmentation
Spec ID: NFR-SEC-01 | Source: `draft/refine-req.md` (v5) | Status: Draft v2 — ยังไม่ผ่าน clarify | อัปเดต: 4 ตุลาคม 2569

## Goal
แยกเครือข่าย Wi-Fi ของพนักงานและของสมาชิกออกจากกันอย่างเด็ดขาด เพื่อป้องกันการโจรกรรมข้อมูล

## Scope
### In scope
- NFR-SEC-01 แยกเครือข่าย Wi-Fi ของพนักงานและสมาชิกด้วย VLAN
### Out of scope
- รุ่นและผู้ผลิตอุปกรณ์เครือข่าย และ topology ที่ refine-req ไม่ได้กำหนด (ทีมเลือกเอง ไม่ได้มาจาก spec)
- ระบบ application ใด ๆ (spec นี้ไม่มี entity หรือ API ของแอป)
- ตัดออกทั้งระบบ (refine-req §2.1–2.2): POS และการขายหน้าร้าน; เครื่องสแกนเวลาและการซิงค์ HR; Data Privacy & Encryption (TLS 1.3, AES-256, PDPA); Nutrition (FR-NUT-01, IF-NUT-01); ระบบเช็กสถานะ/คิวเครื่องเล่น (Q10)
- หมายเหตุ (ไม่ใช่ requirement) NS-2: แอร์เย็นและอุณหภูมิเหมาะสมต่อการออกกำลังกาย เป็นคำชมสภาพปัจจุบัน ไม่ใช่ฟังก์ชันที่ต้องพัฒนา

## Constraints
ไม่มี CON, DOM หรือ IF ใน spec นี้

## Requirements
ไม่มี FR ใน spec นี้

## Quality Requirements (ยอมรับแล้ว)
| ID | ข้อกำหนด |
|---|---|
| NFR-SEC-01 | Wi-Fi ของพนักงานและ Wi-Fi ของสมาชิกต้องแยกออกจากกันด้วย VLAN โดย Wi-Fi สมาชิกเข้าถึงเครือข่ายพนักงานไม่ได้ |

## Domain Rules
ไม่มีใน spec นี้

## ข้อเสนอ (Candidate) — ยังไม่ใช่ requirement จนทีมยืนยัน
ไม่มี UI requirement ของฟีเจอร์นี้ใน refine-req v5

## AC ข้อเสนอ (Candidate) — รอทีมยืนยัน
| AC ID | Given / When / Then | อ้างอิง | สถานะ |
|---|---|---|---|
| AC-NET-01 (Candidate) | Given อุปกรณ์เชื่อมต่อ Wi-Fi ของสมาชิก / When พยายามเข้าถึงเครือข่ายพนักงาน / Then เข้าถึงไม่ได้ | NFR-SEC-01 | ทดสอบได้ทันทีเมื่อมีอุปกรณ์เครือข่าย |
| AC-NET-02 (Candidate) | Given มี Wi-Fi พนักงานและ Wi-Fi สมาชิก / When ตรวจการตั้งค่า / Then ทั้งสองเครือข่ายแยกด้วย VLAN | NFR-SEC-01 | ช่วง VLAN และอุปกรณ์ = ทีมเลือกเอง ไม่ได้มาจาก spec |

## Assumptions & Open Questions
- ไม่มี Open Question เฉพาะฟีเจอร์นี้ใน refine-req หัวข้อ 8.2
- ไม่มี ASM ใหม่ (ไม่มี ID หรือการตัดสินใจใหม่ในสเปคนี้)

## Traceability
| Source ID | ส่วนใน spec | ที่มา (refine-req v5) |
|---|---|---|
| NFR-SEC-01 | Quality Requirements | v4 |
| AC-NET-01..02 | AC ข้อเสนอ | อนุมานจาก NFR-SEC-01 |
