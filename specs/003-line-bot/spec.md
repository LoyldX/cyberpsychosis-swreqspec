# Feature: LINE Bot
Spec ID: FR-LINE-01, FR-LINE-02, IF-LINE-01 | Source: `draft/refine-req.md` (v5) | Status: Draft v2 — ยังไม่ผ่าน clarify | อัปเดต: 4 ตุลาคม 2569

## Goal
สมาชิกดูข้อมูลและสถานะการจองคลาสของตนเอง และถามคำถามทั่วไปเกี่ยวกับยิมผ่าน LINE Bot

## Scope
### In scope
- FR-LINE-01 สมาชิกเรียกดูข้อมูลและสถานะการจองคลาสของตนผ่าน LINE Bot
- FR-LINE-02 ระบบตอบคำถามทั่วไปที่พบบ่อยเกี่ยวกับยิมและการใช้งานอัตโนมัติ
- IF-LINE-01 เชื่อมต่อ LINE Messaging API
- UI-LINE-01, UI-LINE-02 อยู่ในหัวข้อ "ข้อเสนอ (Candidate)" ไม่ใช่ In scope จนกว่าทีมยืนยัน
### Out of scope
- วิธียืนยันตัวตนและผูกบัญชี LINE กับสมาชิก (Q5)
- ตัดออกทั้งระบบ (refine-req §2.1–2.2): POS และการขายหน้าร้าน; เครื่องสแกนเวลาและการซิงค์ HR; Data Privacy & Encryption (TLS 1.3, AES-256, PDPA); Nutrition (FR-NUT-01, IF-NUT-01); ระบบเช็กสถานะ/คิวเครื่องเล่น (Q10)

## Constraints
- IF-LINE-01 เชื่อมต่อ LINE Messaging API เพื่อให้บริการ LINE Bot และดึงข้อมูลการจอง

## Requirements (ยอมรับแล้ว)
| ID | ข้อกำหนด |
|---|---|
| FR-LINE-01 | สมาชิกต้องเรียกดูข้อมูลและสถานะการจองคลาสของตนผ่าน LINE Bot ได้ |
| FR-LINE-02 | ระบบต้องตอบคำถามทั่วไปที่พบบ่อยเกี่ยวกับยิมและการใช้งานอัตโนมัติผ่าน LINE Bot |

## Quality Requirements
ไม่มีใน spec นี้

## Domain Rules
ไม่มีใน spec นี้

## ข้อเสนอ (Candidate) — ยังไม่ใช่ requirement จนทีมยืนยัน
| ID | ข้อเสนอ | อ้างอิง | Open Question |
|---|---|---|---|
| UI-LINE-01 (Candidate) | คำตอบเรื่องการจองต้องแสดงคลาส วันเวลา และสถานะการจองของผู้ใช้คนนั้น (รูปแบบข้อความทีมเลือกเอง) | FR-LINE-01 | Q5 |
| UI-LINE-02 (Candidate) | ผู้ใช้เรียกดูคำถามที่พบบ่อยได้จากในห้องแชต LINE | FR-LINE-02 | ยังไม่มีหมายเลข Q (เสนอให้ทีมเพิ่มคำถามยืนยัน) |

## AC ข้อเสนอ (Candidate) — รอทีมยืนยัน
| AC ID | Given / When / Then | อ้างอิง | สถานะ |
|---|---|---|---|
| AC-LINE-01 (Candidate) | Given สมาชิกที่ผูกบัญชี LINE แล้ว / When พิมพ์ขอดูการจองของตน / Then ระบบตอบข้อมูลและสถานะการจองของสมาชิกคนนั้นเท่านั้น | FR-LINE-01 | การผูกบัญชีรอ Q5 |
| AC-LINE-02 (Candidate) | Given ผู้ใช้ถามคำถามที่พบบ่อยเกี่ยวกับยิม / When ส่งข้อความเข้า LINE Bot / Then ระบบตอบคำตอบอัตโนมัติ | FR-LINE-02 | ทดสอบได้ทันที (ชุด FAQ ทีมเลือกเอง) |
| AC-LINE-03 (Candidate) | Given LINE Bot รับข้อความจาก LINE / When ระบบส่งคำตอบผ่าน LINE Messaging API / Then ผู้ใช้ได้รับคำตอบ | IF-LINE-01 | กรณี LINE API ไม่ตอบสนอง = TBD (ไม่มี Q รองรับ) |

## Assumptions & Open Questions
- **Q5** ผู้ใช้ยืนยันตัวตนบน LINE Bot อย่างไรเพื่อผูกกับข้อมูลการจองของตน (ยังไม่มีคำตอบ ผู้ตอบที่น่าจะได้: ทีมพัฒนาระบบ)
- ยังไม่มีหมายเลข Q: พฤติกรรมเมื่อ LINE Messaging API ไม่ตอบสนอง (IF-LINE-01 ไม่ได้ระบุ)
- ไม่มี ASM ใหม่

## Traceability
| Source ID | ส่วนใน spec | ที่มา (refine-req v5) |
|---|---|---|
| FR-LINE-01, FR-LINE-02 | Requirements | v4 |
| IF-LINE-01 | Constraints | v4 |
| UI-LINE-01, UI-LINE-02 | ข้อเสนอ (Candidate) | v5 §6.4 |
| AC-LINE-01..03 | AC ข้อเสนอ | อนุมานจาก FR-LINE-01, FR-LINE-02, IF-LINE-01 |
