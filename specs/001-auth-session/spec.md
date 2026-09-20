# Feature: รักษา Session Login แบบ Persistent
Spec ID: FR-AUTH-01 | Source: `draft/refine-req.md` | Status: Draft

## Goal
สมาชิกเข้าสู่ระบบและใช้งานต่อเนื่องได้โดยระบบใช้ refresh token เพื่อรักษา session

## Scope
### In scope
- การรักษา session login แบบ persistent ตาม FR-AUTH-01
### Out of scope
- การกำหนดอายุ session ที่แน่นอนและกลุ่มผู้ใช้ของ “User 2” (Q2)

## Constraints
- CON-CACHE-01 ใช้ Redis สำหรับบริหาร session login

## Requirements
- FR-AUTH-01 ระบบต้องรองรับ session login แบบ persistent ด้วย refresh token

## Quality Requirements
- NFR-REL-01 การจัดการ session ต้องใช้ Redis ร่วมกับ refresh token และต้องมีเกณฑ์อัตราการหลุดที่ทีมกำหนด

## Acceptance Criteria
ยังไม่มี Acceptance Criteria ใน `draft/refine-req.md` จึงยังไม่สร้าง AC ใหม่

## Assumptions & Open Questions
- Q2 “User 2” หมายถึงผู้ใช้กลุ่มใด และ session ต้องคงอยู่นานเท่าใด

## Traceability
| Source ID | รายการ |
|---|---|
| FR-AUTH-01 | Requirements |
| NFR-REL-01 | Quality Requirements |
| CON-CACHE-01 | Constraints |

