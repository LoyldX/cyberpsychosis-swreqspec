// Package testutil เป็นตัวช่วยสำหรับ test เท่านั้น (ห้าม import จากโค้ด production)
// ใช้ PostgreSQL จริงตาม specs/000-shared/README.md §3 (CON-DB-01).
package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // driver ชื่อ scheme "pgx5" (CON-DB-01)
	_ "github.com/golang-migrate/migrate/v4/source/file"     // อ่านไฟล์ .sql จาก backend/migrations/
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// defaultTestDSN คือค่า dev ที่อยู่ใน docker-compose.yml อยู่แล้ว ใช้เป็นค่าเริ่มต้นของ test เท่านั้น
// (ไม่ใช่ค่าของ production; ตั้ง TEST_DATABASE_URL เพื่อ override).
const defaultTestDSN = "postgres://gym_user:gym_password_dev@localhost:5432/gym_management?sslmode=disable"

// TestDSN คืน DSN ของ PostgreSQL สำหรับ test จาก TEST_DATABASE_URL หรือค่า dev ข้างบน (CON-DB-01).
func TestDSN() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return defaultTestDSN
}

// MigrationsDir คืน path ของ backend/migrations โดยคำนวณจากตำแหน่งไฟล์นี้ ไม่ขึ้นกับ working directory (CON-DB-01).
func MigrationsDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("testutil: runtime.Caller failed (ห้ามคอมไพล์ test ด้วย -trimpath)")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}

// NewMigrate สร้าง golang-migrate ที่ชี้ไป DSN ที่กำหนดและโฟลเดอร์ backend/migrations (CON-DB-01).
func NewMigrate(dsn string) (*migrate.Migrate, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	// driver pgx v5 ของ golang-migrate ใช้ scheme "pgx5"
	u.Scheme = "pgx5"
	src := url.URL{Scheme: "file", Path: filepath.ToSlash(MigrationsDir())}
	return migrate.New(src.String(), u.String())
}

// MigrateUp รัน migration ขึ้นสุดท้ายบน DSN ที่กำหนด โดยถือว่า "ไม่มีอะไรต้องรัน" เป็นสำเร็จ
// golang-migrate ล็อกด้วย advisory lock จึงเรียกพร้อมกันจากหลาย package ได้ (CON-DB-01).
func MigrateUp(dsn string) error {
	m, err := NewMigrate(dsn)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

var (
	sharedOnce sync.Once
	sharedErr  error
)

// migrateSharedOnce รัน migrate up บนฐานข้อมูลร่วม ครั้งเดียวต่อ process (CON-DB-01).
// ไม่มีฟังก์ชัน down สำหรับฐานข้อมูลร่วม เพราะ package อื่นอาจใช้พร้อมกัน.
func migrateSharedOnce() error {
	sharedOnce.Do(func() { sharedErr = MigrateUp(TestDSN()) })
	return sharedErr
}

// ConnectPool ใช้ migration กับฐานข้อมูลร่วม (idempotent) แล้วคืน pgxpool ที่ปิดเองตอนจบ test (CON-DB-01).
func ConnectPool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	require.NoError(t, migrateSharedOnce(), "migrate up บนฐานข้อมูลทดสอบล้มเหลว (เปิด docker compose up -d postgres หรือยัง?)")
	pool, err := pgxpool.New(context.Background(), TestDSN())
	require.NoError(t, err)
	t.Cleanup(pool.Close) // ลงทะเบียนก่อน cleanup ของ CreateMember จึงถูกเรียกทีหลัง (LIFO)
	require.NoError(t, pool.Ping(context.Background()))
	return pool
}

// Member คือสมาชิกจำลองที่ CreateMember สร้าง (FR-AUTH-01).
type Member struct {
	ID        int64
	StudentID string
	Password  string // รหัสผ่านจริงสำหรับ test ใช้ล็อกอิน (ไม่ถูกเก็บในฐานข้อมูล)
}

// CreateMember สร้างสมาชิกจำลองสำหรับ TEST เท่านั้น ห้ามเรียกจากโค้ด production (Q-03: ยังไม่มีช่องทางสร้างสมาชิกจริง).
// ใช้ student_id สุ่มไม่ซ้ำ, bcrypt.MinCost เพื่อความเร็ว และ t.Cleanup ลบเฉพาะแถวนี้ (FR-AUTH-01, CON-DB-01).
func CreateMember(t testing.TB, pool *pgxpool.Pool, password string) Member {
	t.Helper()
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	studentID := "t" + hex.EncodeToString(buf)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)

	var id int64
	err = pool.QueryRow(context.Background(),
		`INSERT INTO members (student_id, password_hash) VALUES ($1, $2) RETURNING id`,
		studentID, string(hash)).Scan(&id)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `DELETE FROM members WHERE id = $1`, id)
		require.NoError(t, err)
	})
	return Member{ID: id, StudentID: studentID, Password: password}
}
