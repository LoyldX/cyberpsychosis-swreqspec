package auth

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/app"
)

// ไฟล์นี้พิสูจน์ F-1 (AC-AUTH-07, AC-AUTH-08, ASM-05, ASM-09): dependency ไม่ตอบภายใน timeout -> 503 ไม่ค้าง.
//
// ข้อจำกัดที่ต้องรู้: test เหล่านี้ใช้ fake ที่ "บล็อกจนกว่า context จะถูกยกเลิก" ไม่ใช่ network blackhole จริง
// (ทีมข้ามงาน blackhole) จึงพิสูจน์ได้เฉพาะว่าชั้น HTTP ตั้ง deadline และปล่อย goroutine เมื่อหมดเวลา
// ส่วนพฤติกรรมบนเครือข่ายจริง (pgx/go-redis ยอมรับการยกเลิกของ ctx) coordinator ตรวจแยกต่างหากด้วย `docker pause`.

const (
	dtTimeout = 200 * time.Millisecond // timeout สั้นใน test (ค่าจริงเริ่มต้น 5s ตาม ASM-09 อยู่ใน config)
	dtMaxWait = 2 * time.Second        // เพดาน: ถ้าเกินนี้ถือว่า "ค้าง"
	dtToken   = "dtValidLookingToken_0123456789abcdefghijklmnopq"
)

// dtBlockingMembers เป็น memberFinder ที่บล็อกจนกว่า ctx จะถูกยกเลิก แล้วรายงาน ctx.Err() ที่เห็น (AC-AUTH-08, ASM-05).
type dtBlockingMembers struct{ observed chan error }

func newDTBlockingMembers() *dtBlockingMembers {
	return &dtBlockingMembers{observed: make(chan error, 1)}
}

func (f *dtBlockingMembers) FindByStudentID(ctx context.Context, _ string) (Member, error) {
	<-ctx.Done()
	f.observed <- ctx.Err()
	return Member{}, ctx.Err()
}

// dtBlockingCreator เป็น sessionCreator ที่บล็อกจนกว่า ctx จะถูกยกเลิก (Redis ไม่ตอบตอน login) (AC-AUTH-07, ASM-05).
type dtBlockingCreator struct{ observed chan error }

func (f *dtBlockingCreator) Create(ctx context.Context, _ int64, _ string) (string, error) {
	<-ctx.Done()
	f.observed <- ctx.Err()
	return "", ctx.Err()
}

// dtBlockingReader เป็น sessionReader ที่บล็อกจนกว่า ctx จะถูกยกเลิก (Redis ไม่ตอบตอนขอใช้งานต่อ) (AC-AUTH-07, ASM-05).
type dtBlockingReader struct{ observed chan error }

func newDTBlockingReader() *dtBlockingReader { return &dtBlockingReader{observed: make(chan error, 1)} }

func (f *dtBlockingReader) Get(ctx context.Context, _ string) (SessionData, error) {
	<-ctx.Done()
	f.observed <- ctx.Err()
	return SessionData{}, ctx.Err()
}

// dtSlowReader หน่วงแล้วคืนข้อมูล (ถ้า ctx ไม่ถูกยกเลิกก่อน) ใช้ทดสอบกรณีช้าแต่ยังทันเวลา (ASM-09).
type dtSlowReader struct {
	delay time.Duration
	data  SessionData
}

func (f dtSlowReader) Get(ctx context.Context, _ string) (SessionData, error) {
	select {
	case <-time.After(f.delay):
		return f.data, nil
	case <-ctx.Done():
		return SessionData{}, ctx.Err()
	}
}

// dtSlowMembers หน่วงแล้วคืนสมาชิก (ถ้า ctx ไม่ถูกยกเลิกก่อน) (ASM-09).
type dtSlowMembers struct {
	delay  time.Duration
	member Member
}

func (f dtSlowMembers) FindByStudentID(ctx context.Context, _ string) (Member, error) {
	select {
	case <-time.After(f.delay):
		return f.member, nil
	case <-ctx.Done():
		return Member{}, ctx.Err()
	}
}

// dtRequireObserved ตรวจว่า fake เห็น ctx ถูกยกเลิกเพราะหมดเวลา (goroutine ถูกปล่อย) (AC-AUTH-07, AC-AUTH-08, ASM-09).
func dtRequireObserved(t *testing.T, observed <-chan error) {
	t.Helper()
	select {
	case err := <-observed:
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(time.Second):
		t.Fatal("fake ไม่เห็นการยกเลิก ctx: goroutine ไม่ถูกปล่อย")
	}
}

// dtCaptureLog เปลี่ยนปลายทาง log มาที่ buffer ตลอด test นี้ เพื่อตรวจว่าไม่มีรหัสผ่าน/token หลุดลง log (ASM-05).
func dtCaptureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(orig) })
	return &buf
}

// dtGoodMember คืนสมาชิกปลอมที่รหัสผ่านคือ "pw" (MinCost เพื่อความเร็ว).
func dtGoodMember(t *testing.T) Member {
	t.Helper()
	return Member{ID: 1, StudentID: "s1", PasswordHash: mustHash(t, "pw")}
}

// dtLoginRouter ประกอบ router จริงของ Gin กับ handler ที่ตั้ง timeout ตามที่กำหนด (0 = ไม่ตั้ง) (FR-AUTH-01, ASM-09).
func dtLoginRouter(svc *Service, timeout time.Duration) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return app.NewRouter(NewHandler(svc).WithDependencyTimeout(timeout).Register)
}

// ครอบคลุม AC-AUTH-08 (CON-DB-01, ASM-05, ASM-09): PostgreSQL ไม่ตอบสนอง (fake บล็อกจน ctx ถูกยกเลิก) ตอนเข้าสู่ระบบ
// -> 503 service_unavailable ตาม body เดิมภายในเวลาใกล้ timeout ไม่ค้าง ไม่ใช่ 401 และไม่มี refresh_token.
// หมายเหตุ: ใช้ fake ที่บล็อก ไม่ใช่ network blackhole จริง; พฤติกรรมบนเครือข่ายจริงตรวจแยกด้วย docker pause.
func TestAC_AUTH_08(t *testing.T) {
	logs := dtCaptureLog(t)
	members := newDTBlockingMembers()
	r := dtLoginRouter(NewService(members, fakeSessions{}, DefaultSessionTTL), dtTimeout)

	start := time.Now()
	w := postLogin(r, loginBody(t, "s1", "super-secret-pw"))
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, unavailableBody, w.Body.String())
	assert.NotContains(t, w.Body.String(), "refresh_token")
	assert.GreaterOrEqual(t, elapsed, dtTimeout, "ต้องรอจนครบ timeout ก่อนปฏิเสธ")
	assert.Less(t, elapsed, dtMaxWait, "ต้องตอบไม่นานเกิน timeout มาก (ห้ามค้าง)")
	dtRequireObserved(t, members.observed)

	assert.Contains(t, logs.String(), "login failed")
	assert.NotContains(t, logs.String(), "super-secret-pw", "log ต้องไม่มีรหัสผ่าน")
}

// ครอบคลุม AC-AUTH-07 (CON-CACHE-01, ASM-05, ASM-09) ฝั่งขอใช้งานต่อ: Redis ไม่ตอบสนอง (fake reader บล็อก) ตอน GET /api/auth/session
// ด้วย bearer token ที่หน้าตาถูกต้อง -> 503 body เดิมภายในเวลาใกล้ timeout และ handler ที่ป้องกันด้วย RequireSession ไม่ถูกเรียก.
// หมายเหตุ: ใช้ fake ที่บล็อก ไม่ใช่ network blackhole จริง; พฤติกรรมบนเครือข่ายจริงตรวจแยกด้วย docker pause.
func TestAC_AUTH_07(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("route จริงจาก Handler.Register ใช้ timeout ของ handler", func(t *testing.T) {
		logs := dtCaptureLog(t)
		reader := newDTBlockingReader()
		svc := NewService(fakeMembers{}, fakeSessions{}, DefaultSessionTTL)
		r := app.NewRouter(NewHandler(svc).WithSessionReader(reader).WithDependencyTimeout(dtTimeout).Register)

		start := time.Now()
		w := getWithAuth(r, "/api/auth/session", "Bearer "+dtToken)
		elapsed := time.Since(start)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.JSONEq(t, unavailableBody, w.Body.String())
		assert.NotContains(t, w.Body.String(), `"member"`)
		assert.GreaterOrEqual(t, elapsed, dtTimeout)
		assert.Less(t, elapsed, dtMaxWait)
		dtRequireObserved(t, reader.observed)
		assert.NotContains(t, logs.String(), dtToken, "log ต้องไม่มี token")
	})

	t.Run("RequireSession พร้อม timeout หยุด chain ก่อนถึง handler", func(t *testing.T) {
		dtCaptureLog(t)
		reader := newDTBlockingReader()
		called := false
		r := app.NewRouter(func(g gin.IRouter) {
			g.GET("/p", RequireSession(reader, dtTimeout), func(c *gin.Context) {
				called = true
				c.Status(http.StatusOK)
			})
		})

		start := time.Now()
		w := getWithAuth(r, "/p", "Bearer "+dtToken)
		elapsed := time.Since(start)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.JSONEq(t, unavailableBody, w.Body.String())
		assert.False(t, called, "handler ต้องไม่ถูกเรียกเมื่อ Redis ไม่ตอบ")
		assert.GreaterOrEqual(t, elapsed, dtTimeout)
		assert.Less(t, elapsed, dtMaxWait)
		dtRequireObserved(t, reader.observed)
	})
}

// ครอบคลุม AC-AUTH-07 (CON-CACHE-01, ASM-05, ASM-09) ฝั่งเข้าสู่ระบบ: ค้นสมาชิกสำเร็จเร็ว แต่ Redis ไม่ตอบตอนสร้าง Session -> 503 ไม่ค้าง ไม่มี refresh_token.
func TestAC_AUTH_07_LoginSessionCreateBlocks_Returns503(t *testing.T) {
	dtCaptureLog(t)
	creator := &dtBlockingCreator{observed: make(chan error, 1)}
	r := dtLoginRouter(NewService(fakeMembers{member: dtGoodMember(t)}, creator, DefaultSessionTTL), dtTimeout)

	start := time.Now()
	w := postLogin(r, loginBody(t, "s1", "pw"))
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, unavailableBody, w.Body.String())
	assert.NotContains(t, w.Body.String(), "refresh_token")
	assert.GreaterOrEqual(t, elapsed, dtTimeout)
	assert.Less(t, elapsed, dtMaxWait)
	dtRequireObserved(t, creator.observed)
}

// ครอบคลุม ASM-09 (ถอยหลังเข้ากันได้): ไม่ตั้ง timeout (NewHandler เฉย ๆ หรือ 0) fake ที่เร็วยังใช้ได้ตามเดิม ทั้ง login และ session.
func TestDependencyTimeout_NotConfigured_FastFakeStillWorks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := NewService(fakeMembers{member: dtGoodMember(t)}, fakeSessions{}, DefaultSessionTTL)
	reader := fakeSessionReader{data: SessionData{MemberID: 1, StudentID: "s1"}}

	routers := map[string]*gin.Engine{
		"NewHandler ไม่เรียก WithDependencyTimeout": app.NewRouter(NewHandler(svc).WithSessionReader(reader).Register),
		"WithDependencyTimeout(0)":                  app.NewRouter(NewHandler(svc).WithSessionReader(reader).WithDependencyTimeout(0).Register),
	}
	for name, r := range routers {
		t.Run(name, func(t *testing.T) {
			w := postLogin(r, loginBody(t, "s1", "pw"))
			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), "refresh_token")

			w = getWithAuth(r, "/api/auth/session", "Bearer "+dtToken)
			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		})
	}

	// RequireSession แบบเดิมที่ไม่ส่ง timeout ก็ไม่มี deadline
	deadlineSeen := true
	r := app.NewRouter(func(g gin.IRouter) {
		g.GET("/p", RequireSession(deadlineProbe{&deadlineSeen}), func(c *gin.Context) { c.Status(http.StatusOK) })
	})
	assert.Equal(t, http.StatusOK, getWithAuth(r, "/p", "Bearer "+dtToken).Code)
	assert.False(t, deadlineSeen, "ไม่ตั้ง timeout ต้องไม่มี deadline บน ctx")
}

// deadlineProbe บันทึกว่า ctx ที่ได้รับมี deadline หรือไม่ (ASM-09).
type deadlineProbe struct{ hasDeadline *bool }

func (p deadlineProbe) Get(ctx context.Context, _ string) (SessionData, error) {
	_, *p.hasDeadline = ctx.Deadline()
	return SessionData{MemberID: 1, StudentID: "s1"}, nil
}

// ครอบคลุม ASM-09: dependency ช้าแต่ยังอยู่ใน timeout (50ms เทียบ 200ms) ต้องสำเร็จ 200 ไม่ถูกตัดเป็น 503.
func TestDependencyTimeout_SlowButWithinTimeout_Succeeds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const slow = 50 * time.Millisecond
	svc := NewService(dtSlowMembers{delay: slow, member: dtGoodMember(t)}, fakeSessions{}, DefaultSessionTTL)
	reader := dtSlowReader{delay: slow, data: SessionData{MemberID: 1, StudentID: "s1"}}
	r := app.NewRouter(NewHandler(svc).WithSessionReader(reader).WithDependencyTimeout(dtTimeout).Register)

	w := postLogin(r, loginBody(t, "s1", "pw"))
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "refresh_token")

	w = getWithAuth(r, "/api/auth/session", "Bearer "+dtToken)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
}

// ครอบคลุม ASM-05 (พฤติกรรมเดิม): error ทั่วไปของ dependency เมื่อมี timeout ตั้งไว้ยังเป็น 503 ทันที ไม่รอจนหมดเวลา.
func TestDependencyTimeout_GenericDependencyError_Still503(t *testing.T) {
	dtCaptureLog(t)
	boom := errors.New("boom")
	r := app.NewRouter(NewHandler(NewService(fakeMembers{err: boom}, fakeSessions{}, DefaultSessionTTL)).
		WithSessionReader(fakeSessionReader{err: boom}).WithDependencyTimeout(dtTimeout).Register)

	start := time.Now()
	w := postLogin(r, loginBody(t, "s1", "pw"))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, unavailableBody, w.Body.String())

	w = getWithAuth(r, "/api/auth/session", "Bearer "+dtToken)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, unavailableBody, w.Body.String())
	assert.Less(t, time.Since(start), dtTimeout, "error ทั่วไปต้องตอบทันที ไม่รอจนหมดเวลา")
}

// ครอบคลุม AC-AUTH-06, AC-AUTH-04 (ASM-06): 401 ของรหัสผิด/ไม่พบสมาชิก/token ไม่พบ ต้องไม่ถูกแปลงเป็น 503 เมื่อมี timeout ตั้งไว้.
func TestDependencyTimeout_InvalidCredentialsAndToken_Stay401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := NewService(fakeMembers{member: dtGoodMember(t)}, fakeSessions{}, DefaultSessionTTL)
	r := app.NewRouter(NewHandler(svc).WithSessionReader(fakeSessionReader{err: ErrSessionNotFound}).
		WithDependencyTimeout(dtTimeout).Register)

	w := postLogin(r, loginBody(t, "s1", "wrong-password"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, invalidCredentialsBody, w.Body.String())

	notFound := NewService(fakeMembers{err: ErrMemberNotFound}, fakeSessions{}, DefaultSessionTTL)
	w = postLogin(dtLoginRouter(notFound, dtTimeout), loginBody(t, "nobody", "pw"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, invalidCredentialsBody, w.Body.String())

	w = getWithAuth(r, "/api/auth/session", "Bearer "+dtToken)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, unauthorizedBody, w.Body.String())

	// ไม่มี header เลยก็ยัง 401 (ไม่เรียก reader)
	w = getWithAuth(r, "/api/auth/session", "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
