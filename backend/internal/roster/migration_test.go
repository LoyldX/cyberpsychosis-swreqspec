package roster

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/testutil"
)

// migrationVersionBeforeRoster คือ version ของ migration ก่อน 0005 (0004_create_class_booking).
const migrationVersionBeforeRoster = 4

// TestMigration0005CreatesTrainerRoster ยืนยันว่า migration 0005 สร้าง trainers และ roster_entries ตาม plan §3 และย้อนกลับได้
// (CON-DB-01, FR-ROST-01; ASM-01, ASM-03, ASM-04, ASM-09) — ใช้ฐานข้อมูลชั่วคราวแยก เพื่อไม่ migrate down ฐานข้อมูลร่วม.
func TestMigration0005CreatesTrainerRoster(t *testing.T) {
	ctx := context.Background()
	admin := testutil.ConnectPool(t)
	dsn := createTempDatabase(t, admin)

	m, err := testutil.NewMigrate(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })
	require.NoError(t, m.Up())

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	var trainerID int64
	require.NoError(t, pool.QueryRow(ctx, `INSERT INTO trainers (name) VALUES ('เทรนเนอร์ทดสอบ') RETURNING id`).Scan(&trainerID))

	insert := `INSERT INTO roster_entries (trainer_id, kind, weekday, start_time, end_time) VALUES ($1, $2, $3, $4, $5)`
	// รายการถูกต้องบันทึกได้ และ cancelled เริ่มต้นเป็น false (ASM-01)
	_, err = pool.Exec(ctx, insert, trainerID, "work", 1, "09:00", "12:00")
	require.NoError(t, err)
	var cancelled bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT cancelled FROM roster_entries WHERE trainer_id = $1`, trainerID).Scan(&cancelled))
	require.False(t, cancelled)

	// kind ต้องเป็น work/teach (ASM-04), weekday 1..7, เริ่มก่อนจบในวันเดียว (ASM-03)
	_, err = pool.Exec(ctx, insert, trainerID, "lunch", 1, "09:00", "12:00")
	requirePgCode(t, err, "23514")
	_, err = pool.Exec(ctx, insert, trainerID, "work", 0, "09:00", "12:00")
	requirePgCode(t, err, "23514")
	_, err = pool.Exec(ctx, insert, trainerID, "work", 8, "09:00", "12:00")
	requirePgCode(t, err, "23514")
	_, err = pool.Exec(ctx, insert, trainerID, "work", 1, "12:00", "12:00")
	requirePgCode(t, err, "23514")
	_, err = pool.Exec(ctx, insert, trainerID, "work", 5, "20:00", "02:00") // ข้ามเที่ยงคืน (Q-07)
	requirePgCode(t, err, "23514")

	// เทรนเนอร์ต้องมีอยู่ใน trainers (ASM-09)
	_, err = pool.Exec(ctx, insert, trainerID+1000, "work", 1, "09:00", "12:00")
	requirePgCode(t, err, "23503")
	pool.Close()

	// migrate down กลับไป version ก่อน 0005 แล้วทั้งสองตารางต้องหายไป
	require.NoError(t, m.Migrate(migrationVersionBeforeRoster))
	pool, err = pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()
	for _, table := range []string{"public.roster_entries", "public.trainers"} {
		var exists bool
		require.NoError(t, pool.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists))
		require.False(t, exists, table)
	}
}

// requirePgCode ตรวจว่า error มาจาก PostgreSQL ด้วย SQLSTATE ที่คาด (23514 CHECK, 23503 FK).
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
	name := "roster_migration_" + hex.EncodeToString(buf)

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
