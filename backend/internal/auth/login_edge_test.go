package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/app"
	"gymapp/backend/internal/testutil"
)

// password72 คือรหัสผ่านยาว 72 ไบต์พอดี (เทียบ fixture m3) ใช้ทดสอบขอบของ bcrypt (AC-AUTH-06, ASM-06).
var password72 = strings.Repeat("Ab1-", 18)

// newEdgeRouter ประกอบ router จริง + PostgreSQL จริง + Redis จริง กับสมาชิกที่ใช้รหัสผ่านที่กำหนด (CON-TECH-01, CON-DB-01, CON-CACHE-01).
func newEdgeRouter(t *testing.T, password string) (http.Handler, *redis.Client, testutil.Member) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	pool := testutil.ConnectPool(t)
	client := newTestRedis(t)
	member := testutil.CreateMember(t, pool, password)
	t.Cleanup(func() { deleteSessionsOfStudent(client, member.StudentID) })

	svc := NewService(NewMemberStore(pool), NewSessionStore(client, DefaultSessionTTL), DefaultSessionTTL)
	return app.NewRouter(NewHandler(svc).Register), client, member
}

// deleteSessionsOfStudent ลบเฉพาะ session ของ student_id สุ่มของ test นี้ (กันรั่วถ้ามี login ที่ไม่ควรสำเร็จหลุดผ่าน) ไม่แตะ key อื่น.
func deleteSessionsOfStudent(client *redis.Client, studentID string) {
	ctx := context.Background()
	iter := client.Scan(ctx, 0, "session:*", 200).Iterator()
	for iter.Next(ctx) {
		raw, err := client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		var data SessionData
		if json.Unmarshal(raw, &data) == nil && data.StudentID == studentID {
			client.Del(ctx, iter.Val())
		}
	}
}

// compareCall บันทึกการเรียก compareHash หนึ่งครั้ง.
type compareCall struct {
	hash     string
	password string
}

// spyCompareHash ครอบ seam compareHash เพื่อนับการเทียบ bcrypt แล้วคืนค่าเดิมใน t.Cleanup (test กลุ่มนี้ไม่รันขนานกัน).
func spyCompareHash(t *testing.T) *[]compareCall {
	t.Helper()
	var calls []compareCall
	orig := compareHash
	compareHash = func(hash, password []byte) error {
		calls = append(calls, compareCall{string(hash), string(password)})
		return orig(hash, password)
	}
	t.Cleanup(func() { compareHash = orig })
	return &calls
}

// captureLog เก็บ log มาตรฐานของ test นี้ไว้ตรวจว่าไม่มี ERROR/login failed แล้วคืนค่าเดิมใน t.Cleanup.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prevOut, prevFlags := log.Writer(), log.Flags()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(prevOut); log.SetFlags(prevFlags) })
	return &buf
}

// assertSame401 ตรวจว่า w เป็น 401 invalid_credentials ที่เหมือน reference ทุกไบต์ (status, body, header) และไม่มี refresh_token (AC-AUTH-06, ASM-06).
func assertSame401(t *testing.T, name string, w, reference *httptest.ResponseRecorder) {
	t.Helper()
	assert.Equal(t, http.StatusUnauthorized, w.Code, "%s: %s", name, w.Body.String())
	assert.Equal(t, invalidCredentialsBody, w.Body.String(), name)
	assert.Equal(t, reference.Body.Bytes(), w.Body.Bytes(), "%s: body ต้องเหมือนกันทุกไบต์", name)
	assert.Equal(t, reference.Header(), w.Header(), "%s: header ต้องเหมือนกัน", name)
	assert.NotContains(t, w.Body.String(), "refresh_token", name)
}

// ครอบคลุม AC-AUTH-06 (FR-AUTH-01, ASM-06): รหัสผ่านจริงยาว 72 ไบต์พอดี -> 200; ยาวกว่านี้ (73 ไบต์, 10 KB) ที่ 72 ไบต์แรกถูก
// ต้องเป็น 401 body เดียวกับกรณีไม่พบรหัสนักศึกษา ไม่มี refresh_token และไม่สร้าง session เพิ่ม (F-2).
func TestAC_AUTH_06_Password72BytesAcceptedLongerRejected(t *testing.T) {
	require.Len(t, password72, 72)
	r, client, member := newEdgeRouter(t, password72)

	unknown := postLogin(r, loginBody(t, member.StudentID+"-unknown", password72))
	require.Equal(t, http.StatusUnauthorized, unknown.Code)

	rejected := map[string]string{
		"72 bytes + 1 extra byte": password72 + "x",
		"72 bytes + 2 extra":      password72 + "xy",
		"73+ with space":          password72 + " ",
		"72 bytes + 10 KB":        password72 + strings.Repeat("z", 10*1024),
		"10 KB only":              strings.Repeat("Ab1-", 2560),
		"73 bytes unrelated":      strings.Repeat("q", 73),
	}
	for name, pw := range rejected {
		t.Run(name, func(t *testing.T) {
			assertSame401(t, name, postLogin(r, loginBody(t, member.StudentID, pw)), unknown)
		})
	}
	assert.Zero(t, sessionsOfStudent(t, client, member.StudentID), "รหัสที่ถูกปฏิเสธต้องไม่สร้าง session")

	t.Run("exactly 72 bytes still succeeds", func(t *testing.T) {
		token, w := loginSucceeds(t, r, client, member)
		assert.NotEmpty(t, token)
		assert.NotContains(t, w.Body.String(), "invalid_credentials")
		assert.Equal(t, 1, sessionsOfStudent(t, client, member.StudentID))
	})
}

// ครอบคลุม AC-AUTH-06 (ASM-06): ความยาวนับเป็น "ไบต์" ไม่ใช่ตัวอักษร: รหัส 24 ตัวอักษรไทย (72 ไบต์) ผ่าน แต่ 25 ตัว (75 ไบต์) ถูกปฏิเสธ
// และรหัส 40 ตัวอักษร 2 ไบต์ (80 ไบต์ แต่ 40 runes) ก็ถูกปฏิเสธ.
func TestAC_AUTH_06_PasswordLengthCountsBytesNotRunes(t *testing.T) {
	thai24 := strings.Repeat("ก", 24)
	require.Len(t, thai24, 72)
	require.Equal(t, 24, len([]rune(thai24)))
	r, client, member := newEdgeRouter(t, thai24)

	unknown := postLogin(r, loginBody(t, member.StudentID+"-unknown", thai24))
	require.Equal(t, http.StatusUnauthorized, unknown.Code)

	assertSame401(t, "25 runes = 75 bytes", postLogin(r, loginBody(t, member.StudentID, thai24+"ก")), unknown)
	assertSame401(t, "40 runes x 2 bytes = 80 bytes", postLogin(r, loginBody(t, member.StudentID, strings.Repeat("é", 40))), unknown)
	loginSucceeds(t, r, client, member) // 24 runes = 72 bytes ต้องผ่าน
}

// ครอบคลุม AC-AUTH-06 (ASM-06): รหัสผ่านยาวเกิน 72 ไบต์ถูกปฏิเสธใน Service (ไม่ใช่ handler) โดยไม่ query ฐานข้อมูล
// และทำงานเทียบ bcrypt กับ hash หลอก 1 ครั้งเหมือนกรณีไม่พบสมาชิก (ไม่ใช้ hash ของสมาชิก) เพื่อไม่ให้เวลาตอบเป็น oracle ใหม่.
func TestAC_AUTH_06_LongPasswordRejectedInServiceWithDummyCompare(t *testing.T) {
	var lookups atomic.Int32
	members := countingMembers{calls: &lookups, member: Member{ID: 1, StudentID: "s1", PasswordHash: mustHash(t, password72)}}
	svc := NewService(members, fakeSessions{}, DefaultSessionTTL)
	calls := spyCompareHash(t)

	for _, pw := range []string{password72 + "x", password72 + strings.Repeat("z", 10*1024)} {
		*calls = nil
		token, ttl, err := svc.Login(context.Background(), "s1", pw)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Empty(t, token)
		assert.Zero(t, ttl)

		require.Len(t, *calls, 1, "ต้องเทียบ bcrypt กับ hash หลอก 1 ครั้ง")
		assert.Equal(t, string(dummyHash), (*calls)[0].hash)
		assert.LessOrEqual(t, len((*calls)[0].password), maxPasswordBytes)
	}
	assert.Zero(t, lookups.Load(), "รหัสยาวเกิน 72 ไบต์ต้องไม่ query ฐานข้อมูล")

	// รหัสพอดี 72 ไบต์ต้องผ่านเส้นทางปกติ: query สมาชิกและเทียบกับ hash ของสมาชิก
	*calls = nil
	_, _, err := svc.Login(context.Background(), "s1", password72)
	require.NoError(t, err)
	assert.EqualValues(t, 1, lookups.Load())
	require.Len(t, *calls, 1)
	assert.NotEqual(t, string(dummyHash), (*calls)[0].hash)
}

// countingMembers เป็น memberFinder ปลอมที่นับจำนวนครั้งที่ถูกเรียก (แทน "การ query PostgreSQL") ใน test เท่านั้น.
type countingMembers struct {
	calls  *atomic.Int32
	member Member
	err    error
}

func (f countingMembers) FindByStudentID(context.Context, string) (Member, error) {
	f.calls.Add(1)
	return f.member, f.err
}

// nulStudentIDs คือรหัสนักศึกษาที่มี NUL ตำแหน่งต่าง ๆ (F-3).
var nulStudentIDs = map[string]string{
	"NUL in middle": "77\x000001",
	"NUL only":      "\x00",
	"NUL at end":    "6600000001\x00",
	"NUL at start":  "\x006600000001",
}

// ครอบคลุม AC-AUTH-06 (ASM-05, ASM-06): รหัสนักศึกษาที่มี NUL ถูกปฏิเสธใน Service เหมือนไม่พบสมาชิก:
// ErrInvalidCredentials, ไม่เรียก memberFinder เลย (ไม่มี query PostgreSQL) และเทียบ bcrypt กับ hash หลอก 1 ครั้งด้วยรหัสผ่านที่ส่งมา.
func TestAC_AUTH_06_NulStudentID_ServiceSkipsDatabaseAndRunsDummyCompare(t *testing.T) {
	for name, sid := range nulStudentIDs {
		t.Run(name, func(t *testing.T) {
			var lookups atomic.Int32
			// ถ้าถูกเรียกจะคืน error ทั่วไป (จะกลายเป็น 503) เพื่อให้เห็นชัดว่าไม่ควรถูกเรียก
			svc := NewService(countingMembers{calls: &lookups, err: errors.New("must not be called")}, fakeSessions{}, DefaultSessionTTL)
			calls := spyCompareHash(t)

			token, ttl, err := svc.Login(context.Background(), sid, "correct-horse-battery")
			assert.ErrorIs(t, err, ErrInvalidCredentials)
			assert.Empty(t, token)
			assert.Zero(t, ttl)
			assert.Zero(t, lookups.Load(), "ต้องไม่ query ฐานข้อมูล")
			require.Len(t, *calls, 1)
			assert.Equal(t, string(dummyHash), (*calls)[0].hash)
			assert.Equal(t, "correct-horse-battery", (*calls)[0].password)
		})
	}
}

// ครอบคลุม AC-AUTH-06 (ASM-05, ASM-06): ผ่าน router จริง + PostgreSQL จริง รหัสนักศึกษามี NUL (JSON \u0000) -> 401 ไม่ใช่ 503
// body/header เหมือนกรณีไม่พบรหัสนักศึกษาทุกไบต์ และไม่มีบรรทัด login failed / ERROR / SQLSTATE ใน log (F-3).
func TestAC_AUTH_06_NulStudentID_RealRouterReturns401NotServiceUnavailable(t *testing.T) {
	r, client, member := newEdgeRouter(t, "correct-horse-battery")
	logs := captureLog(t)

	unknown := postLogin(r, loginBody(t, member.StudentID+"-unknown", member.Password))
	require.Equal(t, http.StatusUnauthorized, unknown.Code)

	bodies := map[string]string{
		"json \\u0000 in middle": `{"student_id":"77\u00000001","password":"correct-horse-battery"}`,
		"json \\u0000 only":      `{"student_id":"\u0000","password":"correct-horse-battery"}`,
		"json \\u0000 at end":    `{"student_id":"` + member.StudentID + `\u0000","password":"` + member.Password + `"}`,
	}
	for name, sid := range nulStudentIDs {
		bodies["marshal "+name] = loginBody(t, sid, member.Password)
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			w := postLogin(r, body)
			assert.NotEqual(t, http.StatusServiceUnavailable, w.Code)
			assertSame401(t, name, w, unknown)
		})
	}

	assert.NotContains(t, logs.String(), "login failed")
	assert.NotContains(t, logs.String(), "SQLSTATE")
	assert.NotContains(t, logs.String(), "ERROR")
	assert.Zero(t, sessionsOfStudent(t, client, member.StudentID))
}

// ครอบคลุม ASM-05 (AC-AUTH-08): การแก้ F-2/F-3 ต้องไม่ทำให้ error จริงของ PostgreSQL/Redis กลายเป็น 401:
// error ทั่วไปของ memberFinder และของ session creator ยังต้องได้ 503 service_unavailable ผ่าน handler.
func TestAC_AUTH_08_DependencyErrorsStill503AfterInputGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	good := Member{ID: 1, StudentID: "s1", PasswordHash: mustHash(t, password72)}
	boom := errors.New("boom")

	cases := map[string]struct {
		svc      *Service
		student  string
		password string
	}{
		"member lookup generic error":             {NewService(fakeMembers{err: boom}, fakeSessions{}, DefaultSessionTTL), "s1", "pw"},
		"member lookup error, 72-byte password":   {NewService(fakeMembers{err: boom}, fakeSessions{}, DefaultSessionTTL), "s1", password72},
		"member lookup wrapped sqlstate-like":     {NewService(fakeMembers{err: errors.New("ERROR: connection refused (SQLSTATE 08006)")}, fakeSessions{}, DefaultSessionTTL), "s1", "pw"},
		"session create error, 72-byte password":  {NewService(fakeMembers{member: good}, fakeSessions{err: boom}, DefaultSessionTTL), "s1", password72},
		"session create error, ordinary password": {NewService(fakeMembers{member: Member{ID: 1, StudentID: "s1", PasswordHash: mustHash(t, "pw")}}, fakeSessions{err: boom}, DefaultSessionTTL), "s1", "pw"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := c.svc.Login(context.Background(), c.student, c.password)
			require.Error(t, err)
			assert.NotErrorIs(t, err, ErrInvalidCredentials)

			w := postLogin(app.NewRouter(NewHandler(c.svc).Register), loginBody(t, c.student, c.password))
			assert.Equal(t, http.StatusServiceUnavailable, w.Code)
			assert.JSONEq(t, unavailableBody, w.Body.String())
			assert.NotContains(t, w.Body.String(), "refresh_token")
		})
	}
}
