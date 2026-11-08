package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/app"
	"gymapp/backend/internal/testutil"
)

const unauthorizedBody = `{"error":"unauthorized"}`

// newSessionRouter ประกอบ router จริง (PostgreSQL + Redis จริง) พร้อม GET /api/auth/session
// และ route ทดสอบ GET /test/protected ที่ห่อด้วย RequireSession (CON-TECH-01, CON-CACHE-01, FR-AUTH-01).
func newSessionRouter(t *testing.T) (*gin.Engine, *redis.Client, testutil.Member) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	pool := testutil.ConnectPool(t)
	client := newTestRedis(t)
	member := testutil.CreateMember(t, pool, "correct-horse-battery")

	store := NewSessionStore(client, DefaultSessionTTL)
	svc := NewService(NewMemberStore(pool), store, DefaultSessionTTL)
	protected := func(r gin.IRouter) {
		r.GET("/test/protected", RequireSession(store), func(c *gin.Context) {
			data, ok := MemberFromContext(c)
			if !ok {
				c.Status(http.StatusInternalServerError)
				return
			}
			c.JSON(http.StatusOK, gin.H{"member_id": data.MemberID})
		})
	}
	return app.NewRouter(NewHandler(svc).WithSessionReader(store).Register, protected), client, member
}

// getWithAuth ส่ง GET ไป path พร้อม header Authorization ตามที่กำหนด (ว่าง = ไม่ส่ง header).
func getWithAuth(r http.Handler, path, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ครอบคลุม AC-AUTH-02 (FR-AUTH-01, CON-CACHE-01, ASM-02): token จากการเข้าสู่ระบบใช้ขอใช้งานต่อได้ 200 ได้ member ถูกต้อง
// route ที่ห่อด้วย RequireSession เรียกได้โดยไม่ส่งรหัสผ่าน, TTL ของ key อยู่ในช่วง 7 วัน และการเรียก session ไม่ต่อ TTL.
func TestAC_AUTH_02(t *testing.T) {
	r, client, member := newSessionRouter(t)
	ctx := context.Background()

	token, _ := loginSucceeds(t, r, client, member)
	key := independentSessionKey(token)

	// TTL ทันทีหลังเข้าสู่ระบบ: 7 วัน ลบเศษไม่เกิน 60 วินาที (ASM-02)
	ttl, err := client.TTL(ctx, key).Result()
	require.NoError(t, err)
	assert.LessOrEqual(t, ttl, 7*24*time.Hour)
	assert.GreaterOrEqual(t, ttl, 7*24*time.Hour-60*time.Second)

	pttlBefore, err := client.PTTL(ctx, key).Result()
	require.NoError(t, err)

	// GET /api/auth/session ด้วย token อย่างเดียว -> 200 พร้อม member ที่ถูกต้อง
	w := getWithAuth(r, "/api/auth/session", "Bearer "+token)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	var resp struct {
		Member struct {
			ID        int64  `json:"id"`
			StudentID string `json:"student_id"`
		} `json:"member"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, member.ID, resp.Member.ID)
	assert.Equal(t, member.StudentID, resp.Member.StudentID)

	// route ทดสอบที่ห่อด้วย RequireSession เรียกได้ด้วย bearer token อย่างเดียว ไม่ส่งรหัสผ่าน
	w = getWithAuth(r, "/test/protected", "Bearer "+token)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.JSONEq(t, `{"member_id":`+strconv.FormatInt(member.ID, 10)+`}`, w.Body.String())

	// scheme ไม่สนตัวพิมพ์เล็กใหญ่
	w = getWithAuth(r, "/api/auth/session", "bearer "+token)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// การเรียกข้างบนต้องไม่ต่อ/รีเซ็ต TTL (ASM-02): PTTL ลดลงตามเวลา ไม่เพิ่ม
	time.Sleep(1200 * time.Millisecond)
	w = getWithAuth(r, "/api/auth/session", "Bearer "+token)
	require.Equal(t, http.StatusOK, w.Code)
	pttlAfter, err := client.PTTL(ctx, key).Result()
	require.NoError(t, err)
	assert.Less(t, pttlAfter, pttlBefore, "PTTL ต้องลดลง ไม่ถูกต่อหรือรีเซ็ต")
	assert.Greater(t, pttlAfter, pttlBefore-30*time.Second)
}

// ครอบคลุม AC-AUTH-04 (FR-AUTH-01): token สุ่มที่ไม่เคยออก และ token จริงที่แก้ 1 ตัวอักษร -> 401 body เดียวกัน.
func TestAC_AUTH_04(t *testing.T) {
	r, client, member := newSessionRouter(t)

	token, _ := loginSucceeds(t, r, client, member)

	// (1) token รูปแบบถูกต้องแต่ไม่เคยออก (สร้างด้วย store แล้วไม่เก็บ -> ใช้ token สุ่มจากเครื่องมือเดียวกันโดยไม่ Create)
	neverIssued := wellFormedUnissuedToken(t)
	// (2) token จริงแก้ตัวแรก 1 ตัวอักษรให้ต่างจากเดิม
	changed := []byte(token)
	if changed[0] == 'A' {
		changed[0] = 'B'
	} else {
		changed[0] = 'A'
	}
	require.NotEqual(t, token, string(changed))

	for name, tk := range map[string]string{"never_issued": neverIssued, "one_char_changed": string(changed)} {
		t.Run(name, func(t *testing.T) {
			w := getWithAuth(r, "/api/auth/session", "Bearer "+tk)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Equal(t, unauthorizedBody, w.Body.String())

			w = getWithAuth(r, "/test/protected", "Bearer "+tk)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Equal(t, unauthorizedBody, w.Body.String())
		})
	}
}

// wellFormedUnissuedToken สร้าง token สุ่มรูปแบบเดียวกับที่ระบบออก โดยไม่เก็บลง Redis (FR-AUTH-01; ใช้ใน test เท่านั้น).
func wellFormedUnissuedToken(t *testing.T) string {
	t.Helper()
	// ใช้ Create ของ store แล้วลบ key ทันที เพื่อได้ token ที่ถูกรูปแบบแต่ไม่มีอยู่ใน Redis
	client := newTestRedis(t)
	store := NewSessionStore(client, time.Minute)
	token, err := store.Create(context.Background(), 1, "never-issued")
	require.NoError(t, err)
	require.NoError(t, client.Del(context.Background(), SessionKey(token)).Err())
	return token
}

// ครอบคลุม FR-AUTH-01: header หาย / scheme ผิด / Bearer ไม่มี token / มีหลาย token -> 401 {"error":"unauthorized"}.
func TestSessionEndpoint_MalformedAuthorizationHeader(t *testing.T) {
	r, _, _ := newSessionRouter(t)

	cases := map[string]string{
		"missing_header":    "",
		"wrong_scheme":      "Basic dXNlcjpwYXNz",
		"bearer_no_token":   "Bearer",
		"bearer_empty":      "Bearer ",
		"token_only":        "abcdef",
		"two_tokens":        "Bearer aaa bbb",
		"bearer_with_space": "Bearer    ",
	}
	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			w := getWithAuth(r, "/api/auth/session", header)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Equal(t, unauthorizedBody, w.Body.String())
		})
	}
}

// fakeSessionReader คืนผล/error ตามที่ตั้ง ใช้พิสูจน์ทางแยกของ middleware โดยไม่ต้องทำให้ Redis จริงล่ม.
type fakeSessionReader struct {
	data SessionData
	err  error
}

func (f fakeSessionReader) Get(context.Context, string) (SessionData, error) { return f.data, f.err }

// ครอบคลุม ASM-05 (CON-CACHE-01): reader คืน error ทั่วไป -> 503 ข้อความให้ลองใหม่ และ middleware หยุด chain (handler ไม่ถูกเรียก)
// ส่วน ErrSessionNotFound -> 401 และ handler ก็ไม่ถูกเรียกเช่นกัน.
func TestRequireSession_ReaderErrorsAbortChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	run := func(err error) (*httptest.ResponseRecorder, bool) {
		called := false
		r := app.NewRouter(func(g gin.IRouter) {
			g.GET("/p", RequireSession(fakeSessionReader{err: err}), func(c *gin.Context) {
				called = true
				c.Status(http.StatusOK)
			})
		})
		return getWithAuth(r, "/p", "Bearer sometoken"), called
	}

	w, called := run(errors.New("redis: connection refused"))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, unavailableBody, w.Body.String())
	assert.False(t, called, "handler ต้องไม่ถูกเรียกเมื่อ Redis มีปัญหา")

	w, called = run(ErrSessionNotFound)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, unauthorizedBody, w.Body.String())
	assert.False(t, called)

	// ห่อ ErrSessionNotFound ด้วย %w ก็ยังนับเป็น not-found
	w, _ = run(errors.Join(errors.New("ctx"), ErrSessionNotFound))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ครอบคลุม FR-AUTH-01: ไม่ตั้ง session reader -> ไม่ลง GET /api/auth/session (login-only router ยังทำงาน).
func TestHandler_WithoutSessionReader_NoSessionRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := app.NewRouter(NewHandler(nil).Register)

	w := getWithAuth(r, "/api/auth/session", "Bearer x")
	assert.Equal(t, http.StatusNotFound, w.Code)
}
