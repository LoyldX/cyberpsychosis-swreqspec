# Feature: รักษา Session Login แบบ Persistent
Spec ID: FR-AUTH-01 (หลัก), NFR-REL-01, CON-CACHE-01 | Source: `draft/refine-req.md` (v5) | Status: Draft v2 — ยังไม่ผ่าน clarify | อัปเดต: 4 ตุลาคม 2569

## Goal
สมาชิกเข้าสู่ระบบและใช้งานต่อเนื่องได้ โดยระบบใช้ refresh token รักษา session เพื่อลดการถูก Auto-logout บ่อยครั้ง (pain point ข้อ 2 ใน refine-req §1)

## Scope
### In scope
- FR-AUTH-01 รักษา session login แบบ persistent ด้วย refresh token
- NFR-REL-01 การจัดการ session ผ่าน Redis ร่วมกับ refresh token (เกณฑ์ตัวเลขยังไม่มี ดู Q2)
- CON-CACHE-01 ใช้ Redis บริหาร session login
- UI-AUTH-01 อยู่ในหัวข้อ "ข้อเสนอ (Candidate)" ไม่ใช่ In scope จนกว่าทีมยืนยัน
### Out of scope
- อายุ session ที่แน่นอน และกลุ่มผู้ใช้ "User" ที่โดน Auto-logout (Q2)
- ตัดออกทั้งระบบ (refine-req §2.1–2.2): POS และการขายหน้าร้าน; เครื่องสแกนเวลาและการซิงค์ HR; Data Privacy & Encryption (TLS 1.3, AES-256, PDPA); Nutrition (FR-NUT-01, IF-NUT-01); ระบบเช็กสถานะ/คิวเครื่องเล่น (Q10)

## Constraints
- CON-CACHE-01 ใช้ Redis สำหรับบริหาร session login
- ไม่มี CON-DB-01 ใน spec นี้ (refine-req ไม่ได้ระบุว่า session อยู่ใน relational DB)

## Requirements (ยอมรับแล้ว)
| ID | ข้อกำหนด |
|---|---|
| FR-AUTH-01 | ระบบต้องรองรับ session login แบบ persistent ด้วย refresh token |

## Quality Requirements (ยอมรับแล้ว)
| ID | ข้อกำหนด |
|---|---|
| NFR-REL-01 | การจัดการ session ผ่าน Redis ร่วมกับ refresh token ต้องมีอัตรา session หลุดโดยไม่ตั้งใจไม่เกินเกณฑ์ที่กำหนด — อายุ session และอัตราหลุดสูงสุด = **TBD (Q2)** |

## Domain Rules
ไม่มีใน spec นี้

## ข้อเสนอ (Candidate) — ยังไม่ใช่ requirement จนทีมยืนยัน
| ID | ข้อเสนอ | อ้างอิง | Open Question |
|---|---|---|---|
| UI-AUTH-01 (Candidate) | เมื่อเปิดแอปภายในอายุ session ผู้ใช้เข้าใช้งานได้ทันทีโดยไม่ต้องกรอกข้อมูลล็อกอินซ้ำ | FR-AUTH-01, NFR-REL-01 | Q2 (ยืนยันอายุ session) |

## AC ข้อเสนอ (Candidate) — รอทีมยืนยัน
| AC ID | Given / When / Then | อ้างอิง | สถานะ |
|---|---|---|---|
| AC-AUTH-01 (Candidate) | Given สมาชิกเคยเข้าสู่ระบบและยังอยู่ในอายุ session / When เปิดแอปอีกครั้ง / Then เข้าใช้งานได้โดยใช้ refresh token โดยไม่ต้องล็อกอินซ้ำ | FR-AUTH-01 | อายุ session = TBD (Q2) |
| AC-AUTH-02 (Candidate) | Given session ที่กำลังใช้งาน / When วัดอัตราการหลุดโดยไม่ตั้งใจในช่วงทดสอบ / Then อัตราไม่เกินเกณฑ์ที่กำหนด | NFR-REL-01 | เกณฑ์ = TBD (Q2) |
| AC-AUTH-03 (Candidate) | Given มีการสร้าง session / When ตรวจที่เก็บ session / Then ข้อมูล session อยู่ใน Redis | CON-CACHE-01 | ทดสอบได้โดยไม่ต้องรอ Q |

## Assumptions & Open Questions
- **Q2** "User" ที่โดน Auto-logout คือผู้ใช้กลุ่มใด และ session ต้องคงอยู่นานเท่าใด (ยังไม่มีคำตอบ ผู้ตอบที่น่าจะได้: ทีม Mobile App / Product Owner)
- ไม่มี ASM ใหม่ (AI ไม่ได้ตัดสินใจแทนทีม)

## Traceability
| Source ID | ส่วนใน spec | ที่มา (refine-req v5) |
|---|---|---|
| FR-AUTH-01 | Requirements | v4 |
| NFR-REL-01 | Quality Requirements | v4 |
| CON-CACHE-01 | Constraints | v4 |
| UI-AUTH-01 | ข้อเสนอ (Candidate) | v5 §6.3 |
| AC-AUTH-01..03 | AC ข้อเสนอ | อนุมานจาก FR-AUTH-01, NFR-REL-01, CON-CACHE-01 |
