package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"gymapp/backend/internal/app"
	"gymapp/backend/internal/testutil"
)

const (
	invalidCredentialsBody = `{"error":"invalid_credentials"}`
	unavailableBody        = `{"error":"service_unavailable","message":"ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง"}`
)

// newLoginRouter ประกอบ router จริง + PostgreSQL จริง + Redis จริง ผ่าน app.NewRouter (CON-TECH-01, CON-DB-01, CON-CACHE-01).
func newLoginRouter(t *testing.T) (*gin.Engine, *redis.Client, testutil.Member) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	pool := testutil.ConnectPool(t)
	client := newTestRedis(t)
	member := testutil.CreateMember(t, pool, "correct-horse-battery")

	svc := NewService(NewMemberStore(pool), NewSessionStore(client, DefaultSessionTTL), DefaultSessionTTL)
	return app.NewRouter(NewHandler(svc).Register), client, member
}

// postLogin ส่ง POST /api/auth/login ด้วย body ดิบที่กำหนด.
func postLogin(r http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func loginBody(t *testing.T, studentID, password string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"student_id": studentID, "password": password})
	require.NoError(t, err)
	return string(b)
}

// independentSessionKey คำนวณ key ใน test เองโดยไม่เรียก SessionKey เพื่อพิสูจน์รูปแบบ key ตาม plan.md §3 (AC-AUTH-05).
func independentSessionKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "session:" + hex.EncodeToString(sum[:])
}

// loginSucceeds ล็อกอินด้วยรหัสถูกต้อง ตรวจ 200 แล้วลงทะเบียนลบ key ของ token นี้เองใน t.Cleanup.
func loginSucceeds(t *testing.T, r http.Handler, client *redis.Client, m testutil.Member) (token string, w *httptest.ResponseRecorder) {
	t.Helper()
	w = postLogin(r, loginBody(t, m.StudentID, m.Password))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	token, _ = resp["refresh_token"].(string)
	require.NotEmpty(t, token)

	key := independentSessionKey(token)
	t.Cleanup(func() { client.Del(context.Background(), key) })
	return token, w
}

// sessionsOfStudent นับ session ใน Redis ที่ student_id ตรงกัน (นับได้เฉพาะของ test นี้ เพราะ student_id สุ่มไม่ซ้ำ).
func sessionsOfStudent(t *testing.T, client *redis.Client, studentID string) int {
	t.Helper()
	ctx := context.Background()
	n := 0
	iter := client.Scan(ctx, 0, "session:*", 200).Iterator()
	for iter.Next(ctx) {
		raw, err := client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue // key อาจหมดอายุ/ถูก test อื่นลบระหว่างสแกน
		}
		var data SessionData
		if json.Unmarshal(raw, &data) == nil && data.StudentID == studentID {
			n++
		}
	}
	require.NoError(t, iter.Err())
	return n
}

// ครอบคลุม AC-AUTH-01 (FR-AUTH-01, ASM-01, ASM-02): รหัสถูกต้อง -> 200 พร้อม refresh_token และ expires_in = 604800.
func TestAC_AUTH_01(t *testing.T) {
	r, client, member := newLoginRouter(t)

	token, w := loginSucceeds(t, r, client, member)

	assert.NotEmpty(t, token)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.EqualValues(t, 604800, resp["expires_in"])
	assert.Contains(t, w.Body.String(), `"expires_in":604800`, "ต้องเป็นจำนวนเต็ม ไม่ใช่ทศนิยม/สตริง")
	assert.Len(t, resp, 2, "ต้องมีเฉพาะ refresh_token กับ expires_in")
}

// ครอบคลุม AC-AUTH-05 (CON-CACHE-01, ASM-02): หลังล็อกอินพบ session:<sha256 ของ token> ใน Redis จริง
// value มี member_id ตรงกัน และ TTL อยู่ในช่วง 7 วันลบเศษไม่เกิน 60 วินาที.
func TestAC_AUTH_05(t *testing.T) {
	r, client, member := newLoginRouter(t)
	ctx := context.Background()

	token, _ := loginSucceeds(t, r, client, member) // loginSucceeds ลงทะเบียนลบ key ใน t.Cleanup แล้ว

	key := independentSessionKey(token)
	exists, err := client.Exists(ctx, key).Result()
	require.NoError(t, err)
	require.EqualValues(t, 1, exists, "ต้องมี key %s ใน Redis", key)

	raw, err := client.Get(ctx, key).Bytes()
	require.NoError(t, err)
	var data struct {
		MemberID  int64  `json:"member_id"`
		StudentID string `json:"student_id"`
	}
	require.NoError(t, json.Unmarshal(raw, &data))
	assert.Equal(t, member.ID, data.MemberID)
	assert.Equal(t, member.StudentID, data.StudentID)
	assert.NotContains(t, string(raw), token, "value ต้องไม่มี token ดิบ")

	ttl, err := client.TTL(ctx, key).Result()
	require.NoError(t, err)
	assert.LessOrEqual(t, ttl, 7*24*time.Hour)
	assert.GreaterOrEqual(t, ttl, 7*24*time.Hour-60*time.Second)
}

// ครอบคลุม AC-AUTH-06 (ASM-06) ระดับ handler: รหัสผิด/ไม่พบรหัสนักศึกษา -> 401 body เดียวกันทุกตัวอักษร
// ไม่มี refresh_token และไม่เกิด session ใหม่ของสมาชิกนั้นใน Redis.
func TestLogin_WrongPasswordAndUnknownStudent_Return401WithoutToken(t *testing.T) {
	r, client, member := newLoginRouter(t)
	unknown := member.StudentID + "-does-not-exist"

	wrongPassword := postLogin(r, loginBody(t, member.StudentID, "wrong-password"))
	unknownStudent := postLogin(r, loginBody(t, unknown, member.Password))

	for name, w := range map[string]*httptest.ResponseRecorder{"wrong password": wrongPassword, "unknown student": unknownStudent} {
		assert.Equal(t, http.StatusUnauthorized, w.Code, name)
		assert.JSONEq(t, invalidCredentialsBody, w.Body.String(), name)
		assert.NotContains(t, w.Body.String(), "refresh_token", name)
	}
	assert.Equal(t, wrongPassword.Body.String(), unknownStudent.Body.String(), "ต้องเหมือนกันทุกตัวอักษร")

	assert.Zero(t, sessionsOfStudent(t, client, member.StudentID), "ล็อกอินล้มเหลวต้องไม่สร้าง session")
	assert.Zero(t, sessionsOfStudent(t, client, unknown))
}

// ASM-06: JSON เสีย / ฟิลด์หาย / ค่าว่าง / ชนิดผิด ตอบ 401 body เดียวกับรหัสผิด.
func TestLogin_MalformedOrIncompleteRequest_Returns401SameBody(t *testing.T) {
	r, client, member := newLoginRouter(t)

	cases := map[string]string{
		"malformed json":     `{"student_id": `,
		"empty body":         ``,
		"not an object":      `[]`,
		"missing password":   `{"student_id":"` + member.StudentID + `"}`,
		"missing student_id": `{"password":"` + member.Password + `"}`,
		"empty password":     `{"student_id":"` + member.StudentID + `","password":""}`,
		"wrong type":         `{"student_id":123,"password":"x"}`,
		"empty object":       `{}`,
		"password too long":  loginBody(t, member.StudentID, strings.Repeat("a", 100)),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			w := postLogin(r, body)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.JSONEq(t, invalidCredentialsBody, w.Body.String())
			assert.NotContains(t, w.Body.String(), "refresh_token")
		})
	}
	assert.Zero(t, sessionsOfStudent(t, client, member.StudentID))
}

// ASM-04: ล็อกอินสองครั้งได้ token/key คนละอัน และ key แรกไม่ถูกลบ.
func TestLogin_TwiceYieldsDistinctSessions(t *testing.T) {
	r, client, member := newLoginRouter(t)

	tokenA, _ := loginSucceeds(t, r, client, member)
	tokenB, _ := loginSucceeds(t, r, client, member)

	assert.NotEqual(t, tokenA, tokenB)
	assert.Equal(t, 2, sessionsOfStudent(t, client, member.StudentID))
}

// fakes สำหรับพิสูจน์ ASM-05 โดยไม่ต้องทำให้ฐานข้อมูลจริงล่ม (การพิสูจน์ด้วย blackhole จริงเป็นงาน T-08..T-11).
type fakeMembers struct {
	member Member
	err    error
}

func (f fakeMembers) FindByStudentID(context.Context, string) (Member, error) { return f.member, f.err }

type fakeSessions struct{ err error }

func (f fakeSessions) Create(context.Context, int64, string) (string, error) {
	return "fake-token", f.err
}

// ASM-05: error ของ dependency (ไม่ใช่ข้อมูลเข้าสู่ระบบผิด) ต้องเป็น 503 พร้อมข้อความให้ลองใหม่ และไม่กลายเป็น ErrInvalidCredentials.
func TestLogin_DependencyErrors_Return503NotInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	good := Member{ID: 1, StudentID: "s1", PasswordHash: mustHash(t, "pw")}
	boom := errors.New("boom")

	cases := map[string]*Service{
		"member lookup fails":   NewService(fakeMembers{err: boom}, fakeSessions{}, DefaultSessionTTL),
		"session create fails":  NewService(fakeMembers{member: good}, fakeSessions{err: boom}, DefaultSessionTTL),
		"corrupt password hash": NewService(fakeMembers{member: Member{ID: 1, StudentID: "s1", PasswordHash: "not-a-bcrypt-hash"}}, fakeSessions{}, DefaultSessionTTL),
	}
	for name, svc := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := svc.Login(context.Background(), "s1", "pw")
			require.Error(t, err)
			assert.NotErrorIs(t, err, ErrInvalidCredentials)

			w := postLogin(app.NewRouter(NewHandler(svc).Register), loginBody(t, "s1", "pw"))
			assert.Equal(t, http.StatusServiceUnavailable, w.Code)
			assert.JSONEq(t, unavailableBody, w.Body.String())
			assert.NotContains(t, w.Body.String(), "refresh_token")
		})
	}
}

// mustHash สร้าง bcrypt hash (MinCost เพื่อความเร็ว) สำหรับ fake member ใน test เท่านั้น.
func mustHash(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(h)
}
