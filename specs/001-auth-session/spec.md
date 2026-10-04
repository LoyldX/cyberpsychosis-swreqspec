# Feature: รักษา Session Login แบบ Persistent
Spec ID: FR-AUTH-01, NFR-REL-01, CON-CACHE-01 | Source: `req/refine-req.md` (v5) | Status: Draft v1 | อัปเดต: 4 ตุลาคม 2569

## Goal
สมาชิกเข้าสู่ระบบและใช้งานต่อเนื่องได้ โดยระบบรักษา session ด้วย Refresh Token เพื่อลดการถูก Auto-logout บ่อยครั้ง (pain point ข้อ 2 ใน refine-req §1)

## Scope
### In scope
- FR-AUTH-01 รองรับการเข้าสู่ระบบของสมาชิกและรักษา session ด้วย refresh token
- NFR-REL-01 การจัดการ session ผ่าน Redis ร่วมกับ refresh token
- CON-CACHE-01 ใช้ Redis บริหาร session login
### Out of scope
- ตัดออกทั้งระบบ (refine-req §2.1–2.2): POS และการขายหน้าร้าน; เครื่องสแกนเวลาและการซิงค์ HR; Data Privacy & Encryption (TLS 1.3, AES-256, PDPA); Nutrition (FR-NUT-01, IF-NUT-01); ระบบเช็กสถานะเครื่องเล่นและคิวเครื่องเล่น (Q10)

## Constraints
- CON-CACHE-01 ใช้ Redis สำหรับบริหาร session login

## Requirements (ยอมรับแล้ว)
| ID | ข้อกำหนด |
|---|---|
| FR-AUTH-01 | **รักษา Session Login แบบ Persistent**: รองรับการเข้าสู่ระบบของสมาชิกและรักษา Session ด้วย Refresh Token เพื่อป้องกันการถูก Auto-logout บ่อยครั้ง |

## Quality Requirements (ยอมรับแล้ว)
| ID | ข้อกำหนด |
|---|---|
| NFR-REL-01 | Session Persistence & Reliability: การจัดการ Session ผ่าน Redis ร่วมกับ Refresh Token ต้องมีอัตรา Session หลุดโดยไม่ตั้งใจไม่เกินเกณฑ์ที่กำหนด — อายุ Session และอัตราหลุดสูงสุด = **TBD (Q2)** |

## Domain Rules
ไม่มีใน spec นี้

## Assumptions & Open Questions
- **Q2** "User" ที่โดน Auto-logout คือผู้ใช้กลุ่มใด และ Session ต้องคงอยู่นานเท่าใด (ยังไม่มีคำตอบ ค่าเดา: สมาชิกทุกคน และอายุ session ตามค่าที่ทีมพัฒนาตั้งเอง ผู้ตอบที่น่าจะได้: ทีม Mobile App / Product Owner) เกี่ยวกับ FR-AUTH-01, NFR-REL-01 — อายุ session ย้ายมาจาก Out of scope เพราะยังไม่ได้ตัดสิน
- **Q7** ผู้ใช้บริการมีเฉพาะนักศึกษาหรือไม่ และผู้ใช้ที่ไม่มีรหัสนักศึกษา (เช่น บุคลากร บุคคลภายนอก) ลงทะเบียนอย่างไร (ยังไม่มีคำตอบ ค่าเดา: มีเฉพาะนักศึกษา ผู้ตอบที่น่าจะได้: ผู้ดูแลยิม) เกี่ยวกับ FR-AUTH-01
- ผลต่อ FR-AUTH-01: วิธีเข้าสู่ระบบของผู้ไม่มีรหัสนักศึกษา
- ยังไม่มีหมายเลข Q: สมาชิกเข้าสู่ระบบด้วยวิธีใด (refine-req ระบุเพียง "รองรับการเข้าสู่ระบบของสมาชิก" ไม่ระบุวิธี)
- ยังไม่มีหมายเลข Q: ใครเป็นเจ้าของข้อมูลสมาชิก (member record) เพราะ 001, 003 และ 005 อ้างถึงสมาชิก
- คำถามข้อเสนอ UI (ไม่มี ID): เมื่อเปิดแอปภายในอายุ session ควรให้เข้าใช้งานได้ทันทีโดยไม่ต้องกรอกข้อมูลล็อกอินซ้ำหรือไม่? (ยังไม่มีหมายเลข Q เกี่ยวข้องกับ Q2)
- หมายเหตุ: ค่าหมดอายุของ session (TTL) ยังไม่ถูกตั้งค่า จนกว่าจะตอบ Q2

## Traceability
| Source | ส่วนใน spec | ที่มา (refine-req v5) |
|---|---|---|
| FR-AUTH-01 | Requirements | §3 บรรทัดที่ 69 (v4) |
| NFR-REL-01 | Quality Requirements | §4 (v4) |
| CON-CACHE-01 | Constraints | §7 (v4) |
| ข้อเสนอ UI เข้าใช้งานต่อเนื่อง | Assumptions & Open Questions | §6.3 (ไม่มี ID ใน spec) |
