package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"gymapp/backend/internal/testutil"
)

// failureRouterEnv ประกอบ router จริง + PostgreSQL จริง + Redis จริง สำหรับ test ของ T-07 (AC-AUTH-06, ASM-06).
type failureRouterEnv struct {
	handler http.Handler
	client  *redis.Client
	member  testutil.Member
}

func newFailureEnv(t *testing.T) failureRouterEnv {
	t.Helper()
	r, client, member := newLoginRouter(t)
	return failureRouterEnv{handler: r, client: client, member: member}
}

// unknownStudentID คืนรหัสนักศึกษาที่ไม่มีในระบบ (อิงจาก student_id สุ่มของสมาชิก test จึงไม่ชนกับข้อมูลจริง).
func (e failureRouterEnv) unknownStudentID() string { return e.member.StudentID + "-unknown" }

// ครอบคลุม AC-AUTH-06 (FR-AUTH-01, ASM-06): รหัสนักศึกษาไม่มีในระบบ กับรหัสผ่านผิด -> 401 body เหมือนกันทุกตัวอักษร
// ไม่มี refresh_token และไม่บอกว่าผิดข้อใด.
func TestAC_AUTH_06(t *testing.T) {
	env := newFailureEnv(t)

	unknownStudent := postLogin(env.handler, loginBody(t, env.unknownStudentID(), env.member.Password))
	wrongPassword := postLogin(env.handler, loginBody(t, env.member.StudentID, "definitely-wrong-password"))

	assert.Equal(t, http.StatusUnauthorized, unknownStudent.Code, unknownStudent.Body.String())
	assert.Equal(t, http.StatusUnauthorized, wrongPassword.Code, wrongPassword.Body.String())
	assert.Equal(t, unknownStudent.Body.String(), wrongPassword.Body.String(), "body ต้องเหมือนกันทุกตัวอักษร")
	assert.Equal(t, unknownStudent.Header().Get("Content-Type"), wrongPassword.Header().Get("Content-Type"))

	hints := []string{"password", "student", "not found", "not_found", "unknown", "no such", "exist", "รหัสผ่าน", "รหัสนักศึกษา", "ไม่พบ"}
	for name, w := range map[string]*httptest.ResponseRecorder{"unknown student": unknownStudent, "wrong password": wrongPassword} {
		body := strings.ToLower(w.Body.String())
		assert.NotContains(t, body, "refresh_token", name)
		for _, h := range hints {
			assert.NotContains(t, body, h, "%s: body ต้องไม่บอกว่าผิดข้อใด", name)
		}
	}
}

// ครอบคลุม ASM-06 (AC-AUTH-06): รหัสผ่านผิดต้องไม่สร้าง session ใน Redis ให้สมาชิกนั้น.
func TestAC_AUTH_06_WrongPasswordCreatesNoSession(t *testing.T) {
	env := newFailureEnv(t)

	for i := 0; i < 3; i++ {
		w := postLogin(env.handler, loginBody(t, env.member.StudentID, "wrong-password"))
		require.Equal(t, http.StatusUnauthorized, w.Code)
	}
	assert.Zero(t, sessionsOfStudent(t, env.client, env.member.StudentID), "รหัสผิดต้องไม่สร้าง session")

	// ตรวจควบคู่ว่า helper นับ session ได้จริง: ล็อกอินถูกต้องแล้วต้องเห็น 1
	loginSucceeds(t, env.handler, env.client, env.member)
	assert.Equal(t, 1, sessionsOfStudent(t, env.client, env.member.StudentID))
}

// ครอบคลุม ASM-06 (AC-AUTH-06) แบบแน่นอน ไม่พึ่งเวลา: ใช้ seam compareHash นับการเทียบ bcrypt
// กรณีไม่พบสมาชิกต้องเทียบกับ hash หลอก 1 ครั้ง (cost = dummyHashCost) ด้วยรหัสผ่านที่ผู้ใช้ส่งมา และผลต้องเป็น ErrInvalidCredentials.
func TestAC_AUTH_06_UnknownStudentRunsBcryptAgainstDummyHash(t *testing.T) {
	env := newFailureEnv(t)

	type call struct {
		hash     string
		password string
	}
	var calls []call
	orig := compareHash
	compareHash = func(hash, password []byte) error {
		calls = append(calls, call{string(hash), string(password)})
		return orig(hash, password)
	}
	t.Cleanup(func() { compareHash = orig })

	const sent = "any-password-from-user"
	w := postLogin(env.handler, loginBody(t, env.unknownStudentID(), sent))
	require.Equal(t, http.StatusUnauthorized, w.Code)

	require.Len(t, calls, 1, "ไม่พบสมาชิกต้องเทียบ bcrypt 1 ครั้ง")
	assert.Equal(t, string(dummyHash), calls[0].hash)
	assert.Equal(t, sent, calls[0].password)
	cost, err := bcrypt.Cost([]byte(calls[0].hash))
	require.NoError(t, err)
	assert.Equal(t, dummyHashCost, cost)

	// เทียบกับกรณีรหัสผ่านผิดของสมาชิกจริง: ก็เทียบ bcrypt 1 ครั้งเช่นกัน แต่กับ hash ของสมาชิก
	calls = nil
	w = postLogin(env.handler, loginBody(t, env.member.StudentID, "wrong-password"))
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Len(t, calls, 1)
	assert.NotEqual(t, string(dummyHash), calls[0].hash)
}

// ครอบคลุม ASM-06: hash หลอกต้องไม่มี input ใดผ่านได้ และเป็น bcrypt hash ที่ถูกต้อง (ไม่ใช่ค่าตายตัวจากโค้ด).
func TestAC_AUTH_06_DummyHashNeverMatches(t *testing.T) {
	_, err := bcrypt.Cost(dummyHash)
	require.NoError(t, err)

	for _, pw := range []string{"", "password", "correct-horse-battery", string(dummyHash), strings.Repeat("a", 72)} {
		err := bcrypt.CompareHashAndPassword(dummyHash, []byte(pw))
		assert.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword, "pw=%q", pw)
	}
}

// ครอบคลุม ASM-06 (AC-AUTH-06) ด้านเวลา: เวลาตอบเฉลี่ยของรหัสนักศึกษาที่ไม่มี ต้องไม่ต่ำกว่ากรณีรหัสผิดมากจนผิดสังเกต
// เกณฑ์กว้าง (>= 25% ของกรณีรหัสผิด) เพราะฝั่งไม่พบสมาชิกใช้ cost 10 ส่วนสมาชิก test ใช้ MinCost จึงช้ากว่าเสมอ ไม่ flaky.
func TestAC_AUTH_06_UnknownStudentResponseTimeNotFarBelowWrongPassword(t *testing.T) {
	env := newFailureEnv(t)
	const rounds = 5

	avg := func(studentID, password string) time.Duration {
		body := loginBody(t, studentID, password)
		postLogin(env.handler, body) // อุ่นเครื่อง (connection pool ฯลฯ) ไม่นับเวลา
		var total time.Duration
		for i := 0; i < rounds; i++ {
			start := time.Now()
			w := postLogin(env.handler, body)
			total += time.Since(start)
			require.Equal(t, http.StatusUnauthorized, w.Code)
		}
		return total / rounds
	}

	wrongPassword := avg(env.member.StudentID, "wrong-password")
	unknownStudent := avg(env.unknownStudentID(), env.member.Password)

	t.Logf("avg wrong password = %v, avg unknown student = %v", wrongPassword, unknownStudent)
	assert.GreaterOrEqual(t, unknownStudent, wrongPassword/4,
		"กรณีไม่พบสมาชิกเร็วกว่ากรณีรหัสผิดมาก แปลว่าอาจไม่ได้เทียบ bcrypt")
}
