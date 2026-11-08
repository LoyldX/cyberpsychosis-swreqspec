package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"gymapp/backend/internal/testutil"
)

// ทุก test ด้านล่างใช้ PostgreSQL จริงผ่าน testutil (CON-DB-01) งาน T-04 ไม่มี AC ตรง ๆ จึงตั้งชื่อตามพฤติกรรม.

// FR-AUTH-01, CON-DB-01: สมาชิกที่มีอยู่ถูกค้นเจอด้วยรหัสนักศึกษา และ hash ตรวจกับรหัสผ่านเดิมได้.
func TestMemberStore_FindByStudentID_ReturnsExistingMember(t *testing.T) {
	pool := testutil.ConnectPool(t)
	created := testutil.CreateMember(t, pool, "correct-horse-battery")
	store := NewMemberStore(pool)

	got, err := store.FindByStudentID(context.Background(), created.StudentID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.StudentID, got.StudentID)
	assert.NotEmpty(t, got.PasswordHash)
	assert.NotEqual(t, created.Password, got.PasswordHash, "ต้องเก็บ hash ไม่ใช่รหัสผ่านจริง")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(got.PasswordHash), []byte(created.Password)))
}

// FR-AUTH-01, CON-DB-01: รหัสนักศึกษาที่ไม่มีในระบบคืน ErrMemberNotFound.
func TestMemberStore_FindByStudentID_UnknownStudentReturnsNotFound(t *testing.T) {
	pool := testutil.ConnectPool(t)
	store := NewMemberStore(pool)

	got, err := store.FindByStudentID(context.Background(), "no-such-student-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMemberNotFound)
	assert.Equal(t, Member{}, got)
}

// ASM-05, CON-DB-01: error ของฐานข้อมูลอื่น (ที่นี่ context ถูกยกเลิก) ต้องไม่ถูกกลืนเป็น "ไม่พบ".
func TestMemberStore_FindByStudentID_CancelledContextIsNotReportedAsNotFound(t *testing.T) {
	pool := testutil.ConnectPool(t)
	created := testutil.CreateMember(t, pool, "irrelevant-password")
	store := NewMemberStore(pool)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.FindByStudentID(ctx, created.StudentID)

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrMemberNotFound)
	assert.ErrorIs(t, err, context.Canceled)
}

// CON-DB-01: input หน้าตาเป็น SQL injection ต้องถูกมองเป็นข้อความธรรมดา (query แบบมีพารามิเตอร์) จึงไม่พบสมาชิก.
func TestMemberStore_FindByStudentID_SQLInjectionInputReturnsNotFound(t *testing.T) {
	pool := testutil.ConnectPool(t)
	// มีสมาชิกอย่างน้อยหนึ่งคน เพื่อพิสูจน์ว่า ' OR '1'='1 ไม่ได้ทำให้ WHERE เป็นจริงทุกแถว
	testutil.CreateMember(t, pool, "some-password")
	store := NewMemberStore(pool)

	for _, input := range []string{
		`' OR '1'='1`,
		`x'; DROP TABLE members; --`,
	} {
		got, err := store.FindByStudentID(context.Background(), input)

		require.Error(t, err, "input: %q", input)
		assert.ErrorIs(t, err, ErrMemberNotFound, "input: %q", input)
		assert.Equal(t, Member{}, got, "input: %q", input)
	}
}
