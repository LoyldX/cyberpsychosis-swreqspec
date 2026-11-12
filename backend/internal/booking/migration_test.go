package booking

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/testutil"
)

// migrationVersionBeforeBooking คือ version ของ migration ก่อน 0004 (0003_create_gate_signals).
const migrationVersionBeforeBooking = 3

// TestMigration0004CreatesClassBooking ยืนยันว่า migration 0004 สร้าง classes และ bookings ตาม plan §3 และย้อนกลับได้
// (CON-DB-01, FR-BKG-01; ASM-03, ASM-05, ASM-06) — ใช้ฐานข้อมูลชั่วคราวแยก เพื่อไม่ migrate down ฐานข้อมูลร่วม.
func TestMigration0004CreatesClassBooking(t *testing.T) {
	ctx := context.Background()
	admin := testutil.ConnectPool(t)
	dsn := createTempDatabase(t, admin)

	m, err := testutil.NewMigrate(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })
	require.NoError(t, m.Up())

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	start := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	var memberID, classID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO members (student_id, password_hash) VALUES ('m1', 'x') RETURNING id`).Scan(&memberID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO classes (trainer_id, starts_at, ends_at, capacity) VALUES (1, $1, $2, 10) RETURNING id`,
		start, end).Scan(&classID))

	// เวลาจบต้องหลังเวลาเริ่ม และจำนวนที่นั่งต้องมากกว่า 0 (ASM-03)
	_, err = pool.Exec(ctx, `INSERT INTO classes (trainer_id, starts_at, ends_at, capacity) VALUES (1, $1, $1, 10)`, start)
	requirePgCode(t, err, "23514")
	_, err = pool.Exec(ctx, `INSERT INTO classes (trainer_id, starts_at, ends_at, capacity) VALUES (1, $1, $2, 0)`, start, end)
	requirePgCode(t, err, "23514")

	// การจองถูกต้องบันทึกได้ (FR-BKG-01) แต่จองคลาสเดิมซ้ำไม่ได้ (ASM-05)
	_, err = pool.Exec(ctx, `INSERT INTO bookings (class_id, member_id) VALUES ($1, $2)`, classID, memberID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO bookings (class_id, member_id) VALUES ($1, $2)`, classID, memberID)
	requirePgCode(t, err, "23505")

	// สมาชิกต้องมีอยู่ใน members ของ 001 (ASM-06) และคลาสต้องมีอยู่จริง
	_, err = pool.Exec(ctx, `INSERT INTO bookings (class_id, member_id) VALUES ($1, $2)`, classID, memberID+1000)
	requirePgCode(t, err, "23503")
	_, err = pool.Exec(ctx, `INSERT INTO bookings (class_id, member_id) VALUES ($1, $2)`, classID+1000, memberID)
	requirePgCode(t, err, "23503")
	pool.Close()

	// migrate down กลับไป version ก่อน 0004 แล้วทั้งสองตารางต้องหายไป
	require.NoError(t, m.Migrate(migrationVersionBeforeBooking))
	pool, err = pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()
	for _, table := range []string{"public.bookings", "public.classes"} {
		var exists bool
		require.NoError(t, pool.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists))
		require.False(t, exists, table)
	}
}

// requirePgCode ตรวจว่า error มาจาก PostgreSQL ด้วย SQLSTATE ที่คาด (23514 CHECK, 23505 UNIQUE, 23503 FK).
func requirePgCode(t *testing.T, err error, code string) {
	t.Helper()
	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr), "ต้องเป็น error จาก PostgreSQL: %v", err)
	require.Equal(t, code, pgErr.Code)
}

// createTempDatabase สร้างฐานข้อมูลชื่อสุ่มบนเซิร์ฟเวอร์ทดสอบ คืน DSN ของมัน และลบทิ้งตอนจบ test (CON-DB-01 ของ test เท่านั้น).
func createTempDatabase(t *testing.T, admin *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	name := "booking_migration_" + hex.EncodeToString(buf)

	_, err = admin.Exec(ctx, `CREATE DATABASE `+name)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(), `DROP DATABASE IF EXISTS `+name+` WITH (FORCE)`)
	})

	u, err := url.Parse(testutil.TestDSN())
	require.NoError(t, err)
	u.Path = "/" + name
	return u.String()
}
