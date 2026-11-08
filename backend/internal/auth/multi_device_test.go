package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/testutil"
)

// requireSessionOK เรียก GET /api/auth/session ด้วย token แล้วตรวจ 200 และ member.id/student_id ตรงกับสมาชิก (FR-AUTH-01, AC-AUTH-09, ASM-04).
func requireSessionOK(t *testing.T, r http.Handler, token string, m testutil.Member, device string) {
	t.Helper()
	w := getWithAuth(r, "/api/auth/session", "Bearer "+token)
	require.Equal(t, http.StatusOK, w.Code, "อุปกรณ์ %s ต้องใช้ขอใช้งานต่อได้: %s", device, w.Body.String())

	var resp struct {
		Member struct {
			ID        int64  `json:"id"`
			StudentID string `json:"student_id"`
		} `json:"member"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, m.ID, resp.Member.ID, "อุปกรณ์ %s", device)
	assert.Equal(t, m.StudentID, resp.Member.StudentID, "อุปกรณ์ %s", device)
}

// requireKeyExists ตรวจว่า key ของ token มีอยู่ใน Redis จริง โดยคำนวณ key เองใน test (CON-CACHE-01, AC-AUTH-09, ASM-04).
func requireKeyExists(t *testing.T, client *redis.Client, token, device string) {
	t.Helper()
	key := independentSessionKey(token)
	n, err := client.Exists(context.Background(), key).Result()
	require.NoError(t, err)
	require.EqualValues(t, 1, n, "ต้องมี key ของอุปกรณ์ %s ใน Redis: %s", device, key)
}

// ครอบคลุม AC-AUTH-09 (FR-AUTH-01, ASM-04): สมาชิกเดียวกันเข้าสู่ระบบสองอุปกรณ์ (A แล้ว B) ได้ token ต่างกัน
// ทั้งสอง token ใช้ GET /api/auth/session ได้ 200 มี key ใน Redis ทั้งคู่ และการเข้าสู่ระบบบน B ไม่รีเซ็ต/ลบ/ต่ออายุ Session ของ A
// ผ่าน router, PostgreSQL และ Redis จริง.
func TestAC_AUTH_09(t *testing.T) {
	r, client, member := newSessionRouter(t)
	ctx := context.Background()

	// อุปกรณ์ A เข้าสู่ระบบก่อน (คำขอแยกกัน = คนละอุปกรณ์); loginSucceeds ลงทะเบียนลบ key ใน t.Cleanup ให้แล้ว
	tokenA, _ := loginSucceeds(t, r, client, member)
	keyA := independentSessionKey(tokenA)
	requireSessionOK(t, r, tokenA, member, "A")

	valueBefore, err := client.Get(ctx, keyA).Bytes()
	require.NoError(t, err)
	pttlBefore, err := client.PTTL(ctx, keyA).Result()
	require.NoError(t, err)
	require.Positive(t, pttlBefore, "key ของ A ต้องมี TTL ก่อนที่ B จะเข้าสู่ระบบ")

	// รอให้ PTTL ลดลงอย่างเห็นได้ชัด เพื่อให้การ "ไม่ถูกรีเซ็ต/ต่ออายุ" วัดได้จริง (ASM-02)
	time.Sleep(1200 * time.Millisecond)

	// อุปกรณ์ B เข้าสู่ระบบด้วยสมาชิกเดียวกัน
	tokenB, _ := loginSucceeds(t, r, client, member)
	keyB := independentSessionKey(tokenB)

	// token ต่างกัน และ key ต่างกัน
	assert.NotEqual(t, tokenA, tokenB)
	assert.NotEqual(t, keyA, keyB)

	// ทั้งสอง key มีอยู่ใน Redis
	requireKeyExists(t, client, tokenA, "A")
	requireKeyExists(t, client, tokenB, "B")

	// Refresh Token ของ A ยังใช้ขอใช้งานต่อได้ และของ B ก็ใช้ได้ (member.id ถูกต้องทั้งคู่)
	requireSessionOK(t, r, tokenA, member, "A")
	requireSessionOK(t, r, tokenB, member, "B")

	// B เข้าสู่ระบบต้องไม่รีเซ็ต/ต่ออายุ/แก้ Session ของ A: ข้อมูลเดิม และ TTL ไม่เพิ่ม (ลดลงตามเวลา)
	valueAfter, err := client.Get(ctx, keyA).Bytes()
	require.NoError(t, err)
	assert.JSONEq(t, string(valueBefore), string(valueAfter), "ข้อมูล Session ของ A ต้องไม่เปลี่ยน")

	pttlAfter, err := client.PTTL(ctx, keyA).Result()
	require.NoError(t, err)
	assert.LessOrEqual(t, pttlAfter, pttlBefore, "TTL ของ A หลัง B เข้าสู่ระบบต้องไม่มากกว่าก่อน (ไม่ต่ออายุ)")
	assert.Less(t, pttlAfter, pttlBefore-time.Second, "TTL ของ A ต้องลดลงตามเวลาที่ผ่านไป ไม่ถูกรีเซ็ตเป็น 7 วัน")
	assert.Positive(t, pttlAfter)

	// อุปกรณ์ที่สามเข้าสู่ระบบ ทั้งสามต้องยังใช้ได้ (ASM-04 "ไม่จำกัดจำนวน" ตัวอย่าง 3 เครื่อง)
	tokenC, _ := loginSucceeds(t, r, client, member)
	assert.NotContains(t, []string{tokenA, tokenB}, tokenC)
	requireKeyExists(t, client, tokenC, "C")
	requireSessionOK(t, r, tokenA, member, "A")
	requireSessionOK(t, r, tokenB, member, "B")
	requireSessionOK(t, r, tokenC, member, "C")
	assert.Equal(t, 3, sessionsOfStudent(t, client, member.StudentID), "ต้องมี Session ของสมาชิกนี้ 3 รายการใน Redis")
}
