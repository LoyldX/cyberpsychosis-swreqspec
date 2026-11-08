package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"sync"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// newThrowawayDB สร้างฐานข้อมูลชั่วคราว migrate_test_<random> บนเซิร์ฟเวอร์เดียวกับฐานข้อมูลทดสอบ
// คืน DSN ของฐานข้อมูลนั้น และลงทะเบียนลบทิ้งตอนจบ test (CON-DB-01).
// ใช้เพื่อไม่ให้ migrate down ไปกระทบฐานข้อมูลร่วมที่ package อื่นรันพร้อมกัน.
func newThrowawayDB(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	name := "migrate_test_" + hex.EncodeToString(buf)

	admin, err := pgx.Connect(ctx, TestDSN())
	require.NoError(t, err, "เชื่อมต่อ PostgreSQL ไม่ได้ (เปิด docker compose up -d postgres หรือยัง?)")
	t.Cleanup(func() { _ = admin.Close(context.Background()) })

	// ชื่อมาจาก hex ที่เราสร้างเอง จึงต่อสตริงเป็น identifier ได้อย่างปลอดภัย
	_, err = admin.Exec(ctx, "CREATE DATABASE "+name)
	require.NoError(t, err, "ผู้ใช้ในฐานข้อมูลทดสอบต้องมีสิทธิ์ CREATEDB")
	t.Cleanup(func() {
		_, err := admin.Exec(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
		assert.NoError(t, err)
	})

	u, err := url.Parse(TestDSN())
	require.NoError(t, err)
	u.Path = "/" + name
	return u.String()
}

type columnInfo struct {
	name, dataType, nullable string
	hasDefault, isIdentity   bool
}

func readMembersColumns(ctx context.Context, t *testing.T, conn *pgx.Conn) []columnInfo {
	t.Helper()
	rows, err := conn.Query(ctx, `
		SELECT column_name, data_type, is_nullable, column_default IS NOT NULL, is_identity = 'YES'
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'members'
		ORDER BY ordinal_position`)
	require.NoError(t, err)
	defer rows.Close()
	var cols []columnInfo
	for rows.Next() {
		var c columnInfo
		require.NoError(t, rows.Scan(&c.name, &c.dataType, &c.nullable, &c.hasDefault, &c.isIdentity))
		cols = append(cols, c)
	}
	require.NoError(t, rows.Err())
	return cols
}

// TestMembersMigration พิสูจน์ migration 0001 (CON-DB-01, FR-AUTH-01): up สร้าง members ตาม plan §3, down ลบตาราง
// รันบนฐานข้อมูลชั่วคราวของตัวเองเท่านั้น ไม่แตะฐานข้อมูลร่วม.
func TestMembersMigration(t *testing.T) {
	ctx := context.Background()
	dsn := newThrowawayDB(t)

	m, err := NewMigrate(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })

	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close(context.Background()) })

	// ก่อน up ยังไม่มีตาราง
	var reg *string
	require.NoError(t, conn.QueryRow(ctx, `SELECT to_regclass('public.members')::text`).Scan(&reg))
	require.Nil(t, reg, "ต้องไม่มีตาราง members ก่อน migrate up")

	require.NoError(t, m.Up())

	t.Run("columns", func(t *testing.T) {
		want := []columnInfo{
			{"id", "bigint", "NO", false, true},
			{"student_id", "text", "NO", false, false},
			{"password_hash", "text", "NO", false, false},
			{"created_at", "timestamp with time zone", "NO", true, false},
		}
		assert.Equal(t, want, readMembersColumns(ctx, t, conn), "คอลัมน์ต้องตรง plan §3 ทุกประการ ไม่มีคอลัมน์อื่น")

		var def string
		require.NoError(t, conn.QueryRow(ctx, `
			SELECT column_default FROM information_schema.columns
			WHERE table_name = 'members' AND column_name = 'created_at'`).Scan(&def))
		assert.Equal(t, "now()", def)
	})

	t.Run("constraints", func(t *testing.T) {
		rows, err := conn.Query(ctx, `
			SELECT c.contype::text, a.attname
			FROM pg_constraint c
			JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey)
			WHERE c.conrelid = 'public.members'::regclass AND c.contype IN ('p', 'u')
			ORDER BY c.contype, a.attname`)
		require.NoError(t, err)
		defer rows.Close()
		got := map[string]string{}
		for rows.Next() {
			var typ, col string
			require.NoError(t, rows.Scan(&typ, &col))
			got[typ] = col
		}
		require.NoError(t, rows.Err())
		assert.Equal(t, map[string]string{"p": "id", "u": "student_id"}, got, "PK ที่ id และ UNIQUE ที่ student_id")
	})

	t.Run("student_id unique", func(t *testing.T) {
		_, err := conn.Exec(ctx, `INSERT INTO members (student_id, password_hash) VALUES ('s-dup', 'h1')`)
		require.NoError(t, err)
		_, err = conn.Exec(ctx, `INSERT INTO members (student_id, password_hash) VALUES ('s-dup', 'h2')`)
		var pgErr *pgconn.PgError
		require.True(t, errors.As(err, &pgErr), "ต้องได้ PgError แต่ได้ %v", err)
		assert.Equal(t, "23505", pgErr.Code, "รหัสนักศึกษาซ้ำต้องผิด unique_violation")
	})

	t.Run("not null", func(t *testing.T) {
		_, err := conn.Exec(ctx, `INSERT INTO members (student_id) VALUES ('s-nohash')`)
		var pgErr *pgconn.PgError
		require.True(t, errors.As(err, &pgErr))
		assert.Equal(t, "23502", pgErr.Code, "password_hash ต้อง NOT NULL")
	})

	require.NoError(t, m.Down())

	require.NoError(t, conn.QueryRow(ctx, `SELECT to_regclass('public.members')::text`).Scan(&reg))
	assert.Nil(t, reg, "migrate down ต้องลบตาราง members")

	// up ซ้ำหลัง down ต้องใช้ได้ (migration ย้อนกลับได้จริง)
	require.NoError(t, m.Up())
	require.NoError(t, conn.QueryRow(ctx, `SELECT to_regclass('public.members')::text`).Scan(&reg))
	assert.NotNil(t, reg)
}

// TestMigrateUpIsIdempotent พิสูจน์ว่า MigrateUp เรียกซ้ำและเรียกพร้อมกันได้บนฐานข้อมูลร่วม (CON-DB-01).
func TestMigrateUpIsIdempotent(t *testing.T) {
	require.NoError(t, MigrateUp(TestDSN()))
	require.NoError(t, MigrateUp(TestDSN()), "รันซ้ำต้องผ่าน (ErrNoChange ถูกจัดการ)")

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- MigrateUp(TestDSN())
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		assert.NoError(t, err)
	}

	// ยืนยันว่า migrate ตรงๆ รอบถัดไปรายงาน ErrNoChange จริง
	m, err := NewMigrate(TestDSN())
	require.NoError(t, err)
	defer func() { _, _ = m.Close() }()
	assert.ErrorIs(t, m.Up(), migrate.ErrNoChange)
}

// TestCreateMember พิสูจน์ว่า CreateMember เก็บเฉพาะ bcrypt hash, student_id ไม่ซ้ำ และ cleanup ลบเฉพาะแถวของตน (FR-AUTH-01, CON-DB-01).
func TestCreateMember(t *testing.T) {
	ctx := context.Background()
	pool := ConnectPool(t)

	var a, b Member
	t.Run("create", func(t *testing.T) {
		a = CreateMember(t, pool, "secret-a")
		b = CreateMember(t, pool, "secret-b")
		assert.NotEqual(t, a.StudentID, b.StudentID)
		assert.NotEqual(t, a.ID, b.ID)

		var studentID, hash string
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT student_id, password_hash FROM members WHERE id = $1`, a.ID).Scan(&studentID, &hash))
		assert.Equal(t, a.StudentID, studentID)
		assert.NotEqual(t, "secret-a", hash, "ห้ามเก็บรหัสผ่านจริง")
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte("secret-a")))
		assert.Error(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrong")))
		cost, err := bcrypt.Cost([]byte(hash))
		require.NoError(t, err)
		assert.Equal(t, bcrypt.MinCost, cost)
	})

	// subtest จบแล้ว t.Cleanup ของมันทำงานแล้ว แถวต้องหายไป
	for _, m := range []Member{a, b} {
		var n int
		require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM members WHERE id = $1`, m.ID).Scan(&n))
		assert.Equal(t, 0, n, "cleanup ต้องลบแถวของ member %d", m.ID)
	}
}
