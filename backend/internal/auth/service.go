// Package auth: service.go คือ logic เข้าสู่ระบบ (FR-AUTH-01, ASM-01).
package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials แทนทั้ง "ไม่พบรหัสนักศึกษา" และ "รหัสผ่านผิด" โดยไม่แยกให้ผู้เรียกรู้ (FR-AUTH-01, ASM-06).
var ErrInvalidCredentials = errors.New("auth: invalid credentials")

// dummyHashCost คือ cost ของ hash หลอก: ใช้ bcrypt.DefaultCost (10) ซึ่งเป็นค่าเริ่มต้นของ GenerateFromPassword
// จึงใกล้เคียง hash จริงที่สร้างด้วยค่าเริ่มต้น (ทีมเลือกเอง ไม่ได้มาจาก spec; spec ไม่กำหนด cost) (FR-AUTH-01, ASM-06).
const dummyHashCost = bcrypt.DefaultCost

// maxPasswordBytes คือความยาวสูงสุด (หน่วยไบต์) ที่ bcrypt ใช้จริง; bcrypt ใน x/crypto ตัดไบต์ที่เกินทิ้งเงียบ ๆ ตอนเทียบ
// จึงต้องปฏิเสธรหัสผ่านที่ยาวกว่านี้เองก่อนเทียบ (ทีมเลือกเอง ไม่ได้มาจาก spec; spec ไม่กำหนดความยาวรหัสผ่าน) (AC-AUTH-06, ASM-06).
const maxPasswordBytes = 72

// compareHash คือ seam ของ bcrypt.CompareHashAndPassword ให้ test นับการเรียกได้ ใน production ชี้ไปฟังก์ชันจริงเสมอ (ASM-06).
var compareHash = bcrypt.CompareHashAndPassword

// dummyHash คือ bcrypt hash หลอกที่สร้างครั้งเดียวตอนโหลดแพ็กเกจ จากรหัสสุ่ม 32 ไบต์ที่ทิ้งทันที
// จึงไม่มี input ใดเทียบตรงได้ และผลเทียบกับมันไม่ถูกใช้ตัดสินสิทธิ์ (FR-AUTH-01, ASM-06, AC-AUTH-06).
var dummyHash = newDummyHash()

// newDummyHash สร้าง hash หลอกจากรหัสสุ่มที่ไม่ถูกเก็บไว้; panic ได้เฉพาะตอนเริ่มโปรแกรมเมื่อ crypto/rand หรือ bcrypt ใช้ไม่ได้ (ASM-06).
func newDummyHash() []byte {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		panic(fmt.Sprintf("auth: dummy hash: crypto/rand: %v", err))
	}
	h, err := bcrypt.GenerateFromPassword(secret, dummyHashCost)
	if err != nil {
		panic(fmt.Sprintf("auth: dummy hash: bcrypt: %v", err))
	}
	return h
}

// memberFinder คือส่วนของ MemberStore ที่ Service ใช้ แยกเป็น interface เพื่อให้ task ถัดไปใส่ตัวปลอมได้ (FR-AUTH-01, CON-DB-01).
type memberFinder interface {
	FindByStudentID(ctx context.Context, studentID string) (Member, error)
}

// sessionCreator คือส่วนของ SessionStore ที่ Service ใช้ แยกเป็น interface เพื่อให้ task ถัดไปใส่ตัวปลอมได้ (FR-AUTH-01, CON-CACHE-01).
type sessionCreator interface {
	Create(ctx context.Context, memberID int64, studentID string) (string, error)
}

// Service รวม logic เข้าสู่ระบบด้วยรหัสนักศึกษา+รหัสผ่าน แล้วออก Refresh Token (FR-AUTH-01, ASM-01, CON-CACHE-01).
type Service struct {
	members  memberFinder
	sessions sessionCreator
	ttl      time.Duration
}

// NewService สร้าง Service; ttl คืนเป็น expires_in ให้ผู้เรียก production ส่ง DefaultSessionTTL (ASM-02).
func NewService(members memberFinder, sessions sessionCreator, ttl time.Duration) *Service {
	return &Service{members: members, sessions: sessions, ttl: ttl}
}

// rejectWithDummyCompare เทียบ bcrypt กับ hash หลอกแล้วทิ้งผล คืน ErrInvalidCredentials เสมอ
// ใช้กับ input ที่ไม่ควรไปถึงฐานข้อมูล/bcrypt ของสมาชิก ให้เสียเวลาใกล้เคียงกรณีไม่พบรหัสนักศึกษา (AC-AUTH-06, ASM-06).
func rejectWithDummyCompare(password string) (string, time.Duration, error) {
	// ตัดที่ 72 ไบต์ก่อนเทียบ: ผลถูกทิ้ง และ bcrypt เสียเวลาเท่าเดิมไม่ว่าไลบรารีจะตัดเองหรือปฏิเสธรหัสยาวทันที
	if len(password) > maxPasswordBytes {
		password = password[:maxPasswordBytes]
	}
	_ = compareHash(dummyHash, []byte(password))
	return "", 0, ErrInvalidCredentials
}

// Login ตรวจรหัสผ่านด้วย bcrypt แล้วสร้าง Session ใน Redis คืน token กับอายุ (FR-AUTH-01, AC-AUTH-01, AC-AUTH-05, AC-AUTH-06, ASM-04).
// ไม่พบสมาชิกหรือรหัสผ่านไม่ตรงคืน ErrInvalidCredentials (ASM-06); error ของ PostgreSQL/Redis ถูกห่อและส่งต่อ
// ไม่แปลงเป็น ErrInvalidCredentials เพื่อให้ชั้นบนตอบ 503 (ASM-05). ไม่บันทึกหรือใส่รหัสผ่าน/token ลงใน error.
// รหัสผ่านยาวเกิน 72 ไบต์ และรหัสนักศึกษาที่มี NUL ถูกปฏิเสธที่นี่ก่อนถึงฐานข้อมูล/bcrypt ของสมาชิก
// เพราะ bcrypt ตัดรหัสที่ยาวเกิน 72 ไบต์ทิ้งเงียบ ๆ และ PostgreSQL ไม่รับ 0x00 ใน text (AC-AUTH-06, ASM-05, ASM-06).
func (s *Service) Login(ctx context.Context, studentID, password string) (string, time.Duration, error) {
	// รหัสผ่านยาวเกิน 72 ไบต์ (นับไบต์ ไม่ใช่ตัวอักษร) หรือรหัสนักศึกษามี NUL ไม่มีทางตรงกับสมาชิกใด:
	// ปฏิเสธเหมือนไม่พบสมาชิก ไม่ query ฐานข้อมูล (AC-AUTH-06, ASM-06)
	if len(password) > maxPasswordBytes || strings.ContainsRune(studentID, 0) {
		return rejectWithDummyCompare(password)
	}

	member, err := s.members.FindByStudentID(ctx, studentID)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			// ไม่พบสมาชิก: ยังเทียบ bcrypt กับ hash หลอกให้เสียเวลาใกล้เคียงกรณีรหัสผิด แล้วทิ้งผล
			// คืน ErrInvalidCredentials เสมอ ไม่ว่าผลเทียบจะเป็นอย่างไร (AC-AUTH-06, ASM-06)
			return rejectWithDummyCompare(password)
		}
		return "", 0, fmt.Errorf("auth: login lookup: %w", err)
	}

	if err := compareHash([]byte(member.PasswordHash), []byte(password)); err != nil {
		// รหัสผ่านไม่ตรง = ข้อมูลเข้าสู่ระบบไม่ถูกต้อง (รหัสยาวเกิน 72 ไบต์ถูกกันไว้ข้างบนแล้ว ถึงตรงนี้ไม่เกิน 72)
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) || errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", 0, ErrInvalidCredentials
		}
		// hash ในฐานข้อมูลเสีย ฯลฯ ไม่ใช่ความผิดของสมาชิก จึงไม่ซ่อนเป็น ErrInvalidCredentials
		return "", 0, fmt.Errorf("auth: verify password: %w", err)
	}

	token, err := s.sessions.Create(ctx, member.ID, member.StudentID)
	if err != nil {
		return "", 0, fmt.Errorf("auth: login create session: %w", err)
	}
	return token, s.ttl, nil
}
