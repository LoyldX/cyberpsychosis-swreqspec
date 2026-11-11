package inventory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"gymapp/backend/internal/testutil"
)

func TestInventoryModelOptionalFields(t *testing.T) {
	item := InventoryItem{}
	require.Nil(t, item.ReorderPoint)

	movement := InventoryMovement{}
	require.Nil(t, movement.ResponsiblePerson)
	require.Nil(t, movement.Purpose)
	require.Nil(t, movement.DueDate)
}

func TestInventoryMigrationUpAndDown(t *testing.T) {
	ctx := context.Background()
	dsn := newInventoryMigrationDatabase(t)

	m, err := testutil.NewMigrate(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })

	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close(context.Background()) })

	require.NoError(t, m.Steps(2))

	for _, table := range []string{"inventory_items", "inventory_movements"} {
		var relation *string
		require.NoError(t, conn.QueryRow(ctx, `SELECT to_regclass($1)::text`, "public."+table).Scan(&relation))
		require.NotNil(t, relation, "%s should be created by migration up", table)
	}

	var reorderPointNullable string
	require.NoError(t, conn.QueryRow(ctx, `
		SELECT is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = 'inventory_items'
		  AND column_name = 'reorder_point'`).Scan(&reorderPointNullable))
	require.Equal(t, "YES", reorderPointNullable)

	var quantityDefault string
	require.NoError(t, conn.QueryRow(ctx, `
		SELECT column_default
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = 'inventory_items'
		  AND column_name = 'quantity_on_hand'`).Scan(&quantityDefault))
	require.Equal(t, "0", quantityDefault)

	require.NoError(t, m.Steps(-1))

	for _, table := range []string{"inventory_movements", "inventory_items"} {
		var relation *string
		require.NoError(t, conn.QueryRow(ctx, `SELECT to_regclass($1)::text`, "public."+table).Scan(&relation))
		require.Nil(t, relation, "%s should be dropped by migration down", table)
	}
}

// newInventoryMigrationDatabase แยกฐานข้อมูลทดสอบเพื่อไม่ให้ migration down กระทบฐานข้อมูลร่วม (CON-DB-01).
func newInventoryMigrationDatabase(t *testing.T) string {
	t.Helper()

	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	name := "inventory_migrate_test_" + hex.EncodeToString(buf)

	admin, err := pgx.Connect(context.Background(), testutil.TestDSN())
	require.NoError(t, err, "เชื่อมต่อ PostgreSQL ไม่ได้ (เปิด docker compose up -d postgres หรือยัง?)")
	t.Cleanup(func() { _ = admin.Close(context.Background()) })

	_, err = admin.Exec(context.Background(), "CREATE DATABASE "+name)
	require.NoError(t, err, "ผู้ใช้ในฐานข้อมูลทดสอบต้องมีสิทธิ์ CREATEDB")
	t.Cleanup(func() {
		_, err := admin.Exec(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
		require.NoError(t, err)
	})

	u, err := url.Parse(testutil.TestDSN())
	require.NoError(t, err)
	u.Path = "/" + name
	return u.String()
}
