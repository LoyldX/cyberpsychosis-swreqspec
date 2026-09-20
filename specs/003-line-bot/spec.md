# Feature: LINE Bot
Spec ID: IF-LINE-01 | Source: `draft/refine-req.md` | Status: Draft

## Goal
สมาชิกดูข้อมูลการจองคลาสและถามคำถามทั่วไปเกี่ยวกับยิมผ่าน LINE Bot

## Scope
### In scope
- FR-LINE-01 และ FR-LINE-02
- IF-LINE-01
### Out of scope
- วิธีการยืนยันตัวตนและการผูกบัญชี ซึ่งยังเป็น Q5

## Constraints
- IF-LINE-01 เชื่อมต่อ LINE Messaging API

## Requirements
- FR-LINE-01 สมาชิกต้องเรียกดูข้อมูลและสถานะการจองคลาสของตนผ่าน LINE Bot ได้
- FR-LINE-02 ระบบต้องตอบคำถามทั่วไปที่พบบ่อยเกี่ยวกับยิมและการใช้งานอัตโนมัติผ่าน LINE Bot

## Acceptance Criteria
ยังไม่มี Acceptance Criteria ใน `draft/refine-req.md` จึงยังไม่สร้าง AC ใหม่

## Assumptions & Open Questions
- Q5 ผู้ใช้งานยืนยันตัวตนบน LINE Bot อย่างไรเพื่อเชื่อมโยงกับข้อมูลการจองของตนเอง

## Traceability
| Source ID | รายการ |
|---|---|
| FR-LINE-01, FR-LINE-02 | Requirements |
| IF-LINE-01 | Constraints |

