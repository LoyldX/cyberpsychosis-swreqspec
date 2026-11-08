package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/app"
	"gymapp/backend/internal/testutil"
)

// expiryTestTTL คืออายุ Session สั้นที่ใช้เฉพาะ test นี้ (ทีมเลือกเอง ไม่ได้มาจาก spec); production ใช้ DefaultSessionTTL (ASM-02).
const expiryTestTTL = 1 * time.Second

// newExpiryRouter ประกอบ router จริง (PostgreSQL + Redis จริง) โดยส่ง TTL สั้นให้ทั้ง NewSessionStore และ NewService (AC-AUTH-03, ASM-02).
func newExpiryRouter(t *testing.T, ttl time.Duration) (*gin.Engine, *redis.Client, testutil.Member) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	pool := testutil.ConnectPool(t)
	client := newTestRedis(t)
	member := testutil.CreateMember(t, pool, "correct-horse-battery")

	store := NewSessionStore(client, ttl)
	svc := NewService(NewMemberStore(pool), store, ttl)
	return app.NewRouter(NewHandler(svc).WithSessionReader(store).Register), client, member
}

// waitForKeyGone โพลว่า key หายจาก Redis จริงจนถึง deadline (ไม่ใช้ sleep ตายตัว) คืน true เมื่อ key หายแล้ว (AC-AUTH-03, ASM-02).
func waitForKeyGone(ctx context.Context, client *redis.Client, key string, deadline time.Duration) (bool, error) {
	end := time.Now().Add(deadline)
	for {
		n, err := client.Exists(ctx, key).Result()
		if err != nil {
			return false, err
		}
		if n == 0 {
			return true, nil
		}
		if time.Now().After(end) {
			return false, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// ครอบคลุม AC-AUTH-03 (FR-AUTH-01, ASM-02): Refresh Token ที่เกินอายุแล้ว ขอใช้งานต่อด้วย GET /api/auth/session -> 401 {"error":"unauthorized"}.
//
// หมายเหตุตามความจริง: รอ 7 วันจริงใน test ไม่ได้ จึงสร้าง SessionStore และ Service ด้วย TTL สั้น (1 วินาที) ผ่าน router จริง + PostgreSQL/Redis จริง
// test นี้พิสูจน์ "กลไกหมดอายุ" (key หายตาม TTL แล้ว token ถูกปฏิเสธ) ส่วนค่า "7 วัน" เองพิสูจน์แยกใน TestAC_AUTH_02 (ตรวจ TTL ของ key = 604800 วินาที)
// สองหลักฐานนี้ใช้คู่กันจึงจะครบ AC-AUTH-03 ทั้งข้อ.
func TestAC_AUTH_03(t *testing.T) {
	r, client, member := newExpiryRouter(t, expiryTestTTL)
	ctx := context.Background()

	// เข้าสู่ระบบจริงผ่าน POST /api/auth/login -> 200 และ expires_in ตรงกับ TTL สั้น (หน่วยวินาทีเต็ม)
	token, loginResp := loginSucceeds(t, r, client, member)
	var resp struct {
		ExpiresIn int64 `json:"expires_in"`
	}
	require.NoError(t, json.Unmarshal(loginResp.Body.Bytes(), &resp))
	assert.Equal(t, int64(expiryTestTTL/time.Second), resp.ExpiresIn)
	key := independentSessionKey(token)

	// ตัวควบคุม: ทันทีหลังเข้าสู่ระบบ (ยังไม่หมดอายุ) token ต้องใช้ได้ -> 200 ไม่ใช่ปฏิเสธเพราะเหตุอื่น
	w := getWithAuth(r, "/api/auth/session", "Bearer "+token)
	require.Equal(t, http.StatusOK, w.Code, "ตัวควบคุม: token ที่ยังไม่หมดอายุต้องใช้ได้ "+w.Body.String())
	exists, err := client.Exists(ctx, key).Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), exists, "ตัวควบคุม: key ต้องยังอยู่ใน Redis ก่อนหมดอายุ")

	// รอจน Redis ลบ key ตาม TTL (โพลด้วย deadline เผื่อมาก ไม่ใช้ sleep ตายตัวเพื่อไม่ให้ test ไม่เสถียร)
	gone, err := waitForKeyGone(ctx, client, key, 10*expiryTestTTL)
	require.NoError(t, err)
	require.True(t, gone, "key ต้องหายจาก Redis หลังเกิน TTL")

	// Then: ขอใช้งานต่อด้วย token ที่หมดอายุแล้ว -> ระบบปฏิเสธ 401 body เดียวกับกรณี token ไม่ถูกต้อง
	w = getWithAuth(r, "/api/auth/session", "Bearer "+token)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, unauthorizedBody, w.Body.String())

	// key ต้องไม่ถูกสร้างคืนจากการเรียกที่ถูกปฏิเสธ (ไม่ต่ออายุ/ไม่ฟื้น Session) (ASM-02)
	exists, err = client.Exists(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists, "key ต้องไม่อยู่ใน Redis หลังหมดอายุ")
}
