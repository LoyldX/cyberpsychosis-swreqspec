# Feature: Push Notification
Spec ID: FR-NOTI-01 (หลัก) | Source: `draft/refine-req.md` (v5) | Status: Draft v2 — ยังไม่ผ่าน clarify | อัปเดต: 4 ตุลาคม 2569

## Goal
สมาชิกได้รับการแจ้งเตือนวันและเวลาเปิด-ปิดยิม รวมถึงข่าวสารประชาสัมพันธ์ผ่านแอปพลิเคชันสมาชิก (pain point ข้อ 5 ใน refine-req §1 ติดตามข่าวผ่าน Facebook Page เป็นหลัก)

## Scope
### In scope
- FR-NOTI-01 ส่ง push notification เรื่องวันและเวลาเปิด-ปิดยิม และข่าวสารประชาสัมพันธ์ไปยังแอปสมาชิก
- UI-NOTI-01 อยู่ในหัวข้อ "ข้อเสนอ (Candidate)" ไม่ใช่ In scope จนกว่าทีมยืนยัน
### Out of scope
- ความสัมพันธ์กับ Facebook Page (แทนหรือใช้คู่กัน ไม่ได้โพสต์ไป Facebook) รอ Q9
- ผู้รับกลุ่มใดและรอบเวลาส่งที่แน่นอน: refine-req ไม่ได้ระบุ (ยังไม่มีหมายเลข Q)
- ตัดออกทั้งระบบ (refine-req §2.1–2.2): POS และการขายหน้าร้าน; เครื่องสแกนเวลาและการซิงค์ HR; Data Privacy & Encryption (TLS 1.3, AES-256, PDPA); Nutrition (FR-NUT-01, IF-NUT-01); ระบบเช็กสถานะ/คิวเครื่องเล่น (Q10)

## Constraints
ไม่มี CON, DOM หรือ IF ที่ระบุช่องทาง push ใน refine-req (บริการ push ที่ใช้ยังไม่กำหนด)

## Requirements (ยอมรับแล้ว)
| ID | ข้อกำหนด |
|---|---|
| FR-NOTI-01 | ระบบต้องส่ง push notification เรื่องวันและเวลาเปิด-ปิดยิม และข่าวสารประชาสัมพันธ์ไปยังแอปสมาชิก |

## Quality Requirements
ไม่มีใน spec นี้

## Domain Rules
ไม่มีใน spec นี้

## ข้อเสนอ (Candidate) — ยังไม่ใช่ requirement จนทีมยืนยัน
| ID | ข้อเสนอ | อ้างอิง | Open Question |
|---|---|---|---|
| UI-NOTI-01 (Candidate) | ข้อความแจ้งเตือนเรื่องเปิด-ปิดยิมต้องแสดงวันและเวลาที่เปิด-ปิด | FR-NOTI-01 | Q9 |

## AC ข้อเสนอ (Candidate) — รอทีมยืนยัน
| AC ID | Given / When / Then | อ้างอิง | สถานะ |
|---|---|---|---|
| AC-NOTI-01 (Candidate) | Given มีข้อความเรื่องวันและเวลาเปิด-ปิดยิมหรือข่าวสารประชาสัมพันธ์ / When ระบบสร้างข้อความ push / Then ข้อความมีเนื้อหาเรื่องวันเวลาเปิด-ปิดหรือข่าวสารนั้น | FR-NOTI-01 | ทดสอบได้ทันที |
| AC-NOTI-02 (Candidate) | Given มีข้อความแจ้งเตือน / When ระบบส่ง / Then ข้อความถึงแอปของสมาชิกที่เป็นผู้รับ | FR-NOTI-01 | ผู้รับกลุ่มใดและความถี่ส่ง = TBD (ยังไม่มี Q โดยตรง; ช่องทางข่าวสารรอ Q9) |

## Assumptions & Open Questions
- **Q9** Push Notification จะมาแทน Facebook Page หรือใช้คู่กัน และระบบต้องโพสต์ไป Facebook Page ด้วยหรือไม่ (ยังไม่มีคำตอบ ผู้ตอบที่น่าจะได้: ผู้ดูแลยิม / ฝ่ายประชาสัมพันธ์)
- ยังไม่มีหมายเลข Q: ผู้รับกลุ่มใดและรอบเวลาส่ง (เสนอให้ทีมเพิ่มเป็นคำถามใหม่ ไม่กำหนดหมายเลขเอง)
- ไม่มี ASM ใหม่

## Traceability
| Source ID | ส่วนใน spec | ที่มา (refine-req v5) |
|---|---|---|
| FR-NOTI-01 | Requirements | v4 |
| UI-NOTI-01 | ข้อเสนอ (Candidate) | v5 §6.3 |
| AC-NOTI-01..02 | AC ข้อเสนอ | อนุมานจาก FR-NOTI-01 |
