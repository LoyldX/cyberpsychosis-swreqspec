// Package auth รวมโค้ดเข้าสู่ระบบและ Session ของสมาชิก (FR-AUTH-01, CON-DB-01, CON-CACHE-01).
package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Member คือสมาชิกที่อ่านจากตาราง members; PasswordHash เป็นผลจาก bcrypt ไม่ใช่รหัสผ่านจริง (FR-AUTH-01, CON-DB-01).
type Member struct {
	ID           int64
	StudentID    string
	PasswordHash string
}

// ErrMemberNotFound บอกว่าไม่มีสมาชิกที่ตรงกับรหัสนักศึกษา ใช้ errors.Is ตรวจ (FR-AUTH-01, AC-AUTH-06 ในงาน T-07).
var ErrMemberNotFound = errors.New("auth: member not found")

// MemberStore อ่านสมาชิกจาก PostgreSQL ด้วย pgxpool และ SQL ตรง (CON-DB-01).
type MemberStore struct {
	pool *pgxpool.Pool
}

// NewMemberStore สร้าง MemberStore จาก pool ที่เรียกใช้เป็นผู้ปิดเอง (CON-DB-01).
func NewMemberStore(pool *pgxpool.Pool) *MemberStore {
	return &MemberStore{pool: pool}
}

// FindByStudentID ค้นสมาชิกตามรหัสนักศึกษาจากตาราง members ด้วย query แบบมีพารามิเตอร์ (FR-AUTH-01, CON-DB-01).
// คืน ErrMemberNotFound เมื่อไม่มีแถว; error อื่นของฐานข้อมูลถูกห่อและส่งต่อ ไม่แปลงเป็น "ไม่พบ"
// เพื่อให้ชั้นบนแปลงเป็น 503 ได้ (ASM-05, AC-AUTH-08).
func (s *MemberStore) FindByStudentID(ctx context.Context, studentID string) (Member, error) {
	var m Member
	err := s.pool.QueryRow(ctx,
		`SELECT id, student_id, password_hash FROM members WHERE student_id = $1`,
		studentID,
	).Scan(&m.ID, &m.StudentID, &m.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Member{}, ErrMemberNotFound
		}
		return Member{}, fmt.Errorf("auth: find member by student_id: %w", err)
	}
	return m, nil
}
