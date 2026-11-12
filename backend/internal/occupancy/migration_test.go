package occupancy

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

// migrationVersionBeforeOccupancy คือ version ของ migration ก่อน 0003 (0002_create_inventory).
const migrationVersionBeforeOccupancy = 2

// TestMigration0003CreatesGateSignals ยืนยันว่า migration 0003 สร้างตาราง gate_signals ตาม plan §3 และย้อนกลับได้
// (DOM-CAP-01, FR-BI-01, IF-GATE-01; ASM-02, ASM-04) — ใช้ฐานข้อมูลชั่วคราวแยก เพื่อไม่ migrate down ฐานข้อมูลร่วม.
func TestMigration0003CreatesGateSignals(t *testing.T) {
	ctx := context.Background()
	admin := testutil.ConnectPool(t)
	dsn := createTempDatabase(t, admin)

	m, err := testutil.NewMigrate(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })
	require.NoError(t, m.Up())

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	// ข้อมูลถูกต้อง: บันทึกได้ (DOM-CAP-01)
	_, err = pool.Exec(ctx, `INSERT INTO gate_signals (direction, occupancy_after) VALUES ('in', 1), ('out', 0)`)
	require.NoError(t, err)

	// direction อื่นนอกจาก in/out ถูกปฏิเสธ (IF-GATE-01)
	_, err = pool.Exec(ctx, `INSERT INTO gate_signals (direction, occupancy_after) VALUES ('sideways', 0)`)
	requireCheckViolation(t, err)

	// occupancy_after ต่ำกว่า 0 ถูกปฏิเสธ (ASM-04)
	_, err = pool.Exec(ctx, `INSERT INTO gate_signals (direction, occupancy_after) VALUES ('out', -1)`)
	requireCheckViolation(t, err)
	pool.Close()

	// migrate down กลับไป version ก่อน 0003 แล้วตารางต้องหายไป
	require.NoError(t, m.Migrate(migrationVersionBeforeOccupancy))
	pool, err = pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()
	var exists bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT to_regclass('public.gate_signals') IS NOT NULL`).Scan(&exists))
	require.False(t, exists)
}

// requireCheckViolation ตรวจว่า error มาจาก CHECK constraint ของ PostgreSQL (SQLSTATE 23514).
func requireCheckViolation(t *testing.T, err error) {
	t.Helper()
	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr), "ต้องเป็น error จาก PostgreSQL: %v", err)
	require.Equal(t, "23514", pgErr.Code)
}

// createTempDatabase สร้างฐานข้อมูลชื่อสุ่มบนเซิร์ฟเวอร์ทดสอบ คืน DSN ของมัน และลบทิ้งตอนจบ test (CON-DB-01 ของ test เท่านั้น).
func createTempDatabase(t *testing.T, admin *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	name := "occupancy_migration_" + hex.EncodeToString(buf)

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
