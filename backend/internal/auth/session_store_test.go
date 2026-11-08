package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// defaultTestRedisAddr คือค่า dev ที่อยู่ใน docker-compose.yml (service redis) ใช้เป็นค่าเริ่มต้นของ test เท่านั้น
// ไม่ใช่ค่าของ production; ตั้ง TEST_REDIS_ADDR เพื่อ override (CON-CACHE-01).
const defaultTestRedisAddr = "localhost:6379"

// newTestRedis ต่อ Redis จริง (ไม่ mock) และ skip ไม่ได้: ถ้าต่อไม่ได้ test ต้องล้มเพื่อไม่ให้ผ่านเงียบ ๆ (CON-CACHE-01).
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = defaultTestRedisAddr
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = client.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, client.Ping(ctx).Err(), "ต้องเปิด Redis ก่อนรัน test: docker compose up -d redis (addr=%s)", addr)
	return client
}

// createTracked สร้าง Session แล้วลงทะเบียนลบ key ของ test นี้เองใน t.Cleanup (ไม่แตะ key อื่น ไม่ใช้ FLUSH*).
func createTracked(t *testing.T, store *SessionStore, client *redis.Client, memberID int64, studentID string) string {
	t.Helper()
	token, err := store.Create(context.Background(), memberID, studentID)
	require.NoError(t, err)
	key := SessionKey(token)
	t.Cleanup(func() { client.Del(context.Background(), key) })
	return token
}

// ครอบคลุม FR-AUTH-01: Create แล้ว Get ได้ MemberID/StudentID เดิม
func TestSessionStore_CreateThenGet_RoundTripsMemberData(t *testing.T) {
	client := newTestRedis(t)
	store := NewSessionStore(client, DefaultSessionTTL)

	token := createTracked(t, store, client, 42, "6512345678")

	got, err := store.Get(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, SessionData{MemberID: 42, StudentID: "6512345678"}, got)
}

// ครอบคลุม ASM-02: TTL ตอนสร้างด้วย DefaultSessionTTL อยู่ในช่วง 7 วัน ลบเศษไม่เกิน 60 วินาที
func TestSessionStore_Create_SetsTTLToSevenDays(t *testing.T) {
	client := newTestRedis(t)
	store := NewSessionStore(client, DefaultSessionTTL)

	token := createTracked(t, store, client, 1, "6500000001")

	ttl, err := client.TTL(context.Background(), SessionKey(token)).Result()
	require.NoError(t, err)
	assert.LessOrEqual(t, ttl, 7*24*time.Hour)
	assert.GreaterOrEqual(t, ttl, 7*24*time.Hour-60*time.Second)
}

// ครอบคลุม ASM-02 "ไม่ต่ออายุ": Get ต้องไม่รีเซ็ต/ต่อ TTL
func TestSessionStore_Get_DoesNotExtendTTL(t *testing.T) {
	client := newTestRedis(t)
	store := NewSessionStore(client, 30*time.Second)
	ctx := context.Background()

	token := createTracked(t, store, client, 2, "6500000002")
	key := SessionKey(token)

	before, err := client.PTTL(ctx, key).Result()
	require.NoError(t, err)
	require.Greater(t, before, time.Duration(0))

	time.Sleep(1100 * time.Millisecond)
	_, err = store.Get(ctx, token)
	require.NoError(t, err)

	after, err := client.PTTL(ctx, key).Result()
	require.NoError(t, err)
	assert.Greater(t, after, time.Duration(0), "key ต้องยังอยู่")
	assert.Less(t, after, before, "Get ต้องไม่ต่อ/รีเซ็ต TTL")
}

// ครอบคลุม ASM-04: Create สองครั้ง (สองอุปกรณ์) ได้ token/key คนละอัน และ key เดิมไม่ถูกลบ
func TestSessionStore_CreateTwice_YieldsDistinctTokensAndKeepsBoth(t *testing.T) {
	client := newTestRedis(t)
	store := NewSessionStore(client, DefaultSessionTTL)
	ctx := context.Background()

	tokenA := createTracked(t, store, client, 3, "6500000003")
	tokenB := createTracked(t, store, client, 3, "6500000003")

	assert.NotEqual(t, tokenA, tokenB)
	assert.NotEqual(t, SessionKey(tokenA), SessionKey(tokenB))

	n, err := client.Exists(ctx, SessionKey(tokenA), SessionKey(tokenB)).Result()
	require.NoError(t, err)
	assert.EqualValues(t, 2, n, "ทั้งสอง key ต้องมีอยู่")

	for _, tok := range []string{tokenA, tokenB} {
		got, err := store.Get(ctx, tok)
		require.NoError(t, err)
		assert.Equal(t, int64(3), got.MemberID)
	}
}

// ครอบคลุม CON-CACHE-01 / AC-AUTH-05 (ระดับ store): key ใช้ SHA-256 ของ token ไม่มี token ดิบใน key หรือ value
func TestSessionStore_StoresOnlyHashOfToken(t *testing.T) {
	client := newTestRedis(t)
	store := NewSessionStore(client, DefaultSessionTTL)
	ctx := context.Background()

	token := createTracked(t, store, client, 4, "6500000004")

	rawKeyExists, err := client.Exists(ctx, "session:"+token).Result()
	require.NoError(t, err)
	assert.EqualValues(t, 0, rawKeyExists, "ห้ามมี key ที่ใช้ token ดิบ")

	hashedKeyExists, err := client.Exists(ctx, SessionKey(token)).Result()
	require.NoError(t, err)
	assert.EqualValues(t, 1, hashedKeyExists)

	assert.Regexp(t, `^session:[0-9a-f]{64}$`, SessionKey(token))
	assert.NotContains(t, SessionKey(token), token)

	value, err := client.Get(ctx, SessionKey(token)).Result()
	require.NoError(t, err)
	assert.NotContains(t, value, token, "ค่าที่เก็บต้องไม่มี token ดิบ")
}

// ครอบคลุม AC-AUTH-04 (ระดับ store): token ที่ไม่รู้จัก/ถูกแก้ไข ได้ ErrSessionNotFound
func TestSessionStore_Get_UnknownOrGarbledToken_ReturnsNotFound(t *testing.T) {
	client := newTestRedis(t)
	store := NewSessionStore(client, DefaultSessionTTL)
	ctx := context.Background()

	token := createTracked(t, store, client, 5, "6500000005")

	for name, bad := range map[string]string{
		"unknown":  "this-token-was-never-issued-by-the-system",
		"empty":    "",
		"garbled":  token + "x",
		"truncate": token[:len(token)-1],
	} {
		t.Run(name, func(t *testing.T) {
			_, err := store.Get(ctx, bad)
			assert.ErrorIs(t, err, ErrSessionNotFound)
		})
	}
}

// ครอบคลุม AC-AUTH-03 (ระดับ store): หลังเกิน TTL ที่ฉีดไว้ (1 วินาที) Get ได้ ErrSessionNotFound
func TestSessionStore_Get_AfterTTLExpires_ReturnsNotFound(t *testing.T) {
	client := newTestRedis(t)
	store := NewSessionStore(client, 1*time.Second)
	ctx := context.Background()

	token := createTracked(t, store, client, 6, "6500000006")

	_, err := store.Get(ctx, token)
	require.NoError(t, err, "ก่อนหมดอายุต้องอ่านได้")

	time.Sleep(1500 * time.Millisecond)
	_, err = store.Get(ctx, token)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

// ครอบคลุม ASM-05: error ของ Redis (ที่นี่ context ถูกยกเลิก) ต้องไม่ถูกแปลงเป็น not-found เพื่อให้ชั้นบนตอบ 503 ได้
func TestSessionStore_Get_CancelledContext_ReturnsErrorOtherThanNotFound(t *testing.T) {
	client := newTestRedis(t)
	store := NewSessionStore(client, DefaultSessionTTL)

	token := createTracked(t, store, client, 7, "6500000007")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.Get(ctx, token)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrSessionNotFound)
	assert.ErrorIs(t, err, context.Canceled)
}
