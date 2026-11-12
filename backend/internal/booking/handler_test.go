package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/auth"
)

// fakeSessions คือแหล่ง Session จำลองของ auth.RequireSession: token → สมาชิก (001 ใช้ Redis จริง ที่นี่ไม่ต้องใช้).
type fakeSessions map[string]int64

// Get คืน SessionData ของ token หรือ auth.ErrSessionNotFound.
func (f fakeSessions) Get(_ context.Context, token string) (auth.SessionData, error) {
	id, ok := f[token]
	if !ok {
		return auth.SessionData{}, auth.ErrSessionNotFound
	}
	return auth.SessionData{MemberID: id, StudentID: fmt.Sprintf("member-%d", id)}, nil
}

// tokenOf คือ token จำลองของสมาชิก id.
func tokenOf(memberID int64) string { return fmt.Sprintf("token-%d", memberID) }

// newRouter ประกอบ Gin router ของการจองกับ service และ session จำลองของสมาชิกที่ระบุ.
func newRouter(svc bookingService, memberIDs ...int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	sessions := fakeSessions{}
	for _, id := range memberIDs {
		sessions[tokenOf(id)] = id
	}
	r := gin.New()
	NewHandler(svc).Register(r, auth.RequireSession(sessions))
	return r
}

// newDBRouter คือ router ที่ใช้ PostgreSQL จริง กะจำลอง และเวลา testNow.
func newDBRouter(pool *pgxpool.Pool, shifts ShiftChecker, memberIDs ...int64) *gin.Engine {
	return newRouter(NewService(NewStore(pool), shifts, fixedClock{testNow}), memberIDs...)
}

// doRequest ส่งคำขอพร้อม token (ว่าง = ไม่ส่ง Authorization) แล้วคืนผล.
func doRequest(r http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// postBooking ส่ง POST /api/bookings ของสมาชิก memberID สำหรับคลาส classID.
func postBooking(r http.Handler, memberID, classID int64) *httptest.ResponseRecorder {
	return doRequest(r, http.MethodPost, "/api/bookings", tokenOf(memberID), fmt.Sprintf(`{"class_id":%d}`, classID))
}

// errorBody คือรูปแบบ error ของ API การจอง.
type errorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// decode แปลง JSON ใน recorder เป็น v.
func decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), v), rec.Body.String())
}

// requireReason ตรวจรหัส HTTP รหัสเหตุผล และมีข้อความไทย (ASM-05).
func requireReason(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	require.Equal(t, status, rec.Code, rec.Body.String())
	var body errorBody
	decode(t, rec, &body)
	assert.Equal(t, code, body.Error)
	assert.NotEmpty(t, body.Message)
}

// TestAC_BKG_01_HTTP: Given สมาชิกที่ล็อกอินแล้ว และคลาสที่ยังไม่เริ่ม มีที่นั่ง เทรนเนอร์ว่างและอยู่ในกะ
// When สมาชิกจองคลาสนั้น Then การจองสำเร็จและเรียกดูการจองนั้นได้ (FR-BKG-01).
func TestAC_BKG_01_HTTP(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	classID := newClass(t, pool, 1, span("10:00", "11:00"), 10) // อยู่ในกะ 09:00–12:00 ของเทรนเนอร์ 1
	r := newDBRouter(pool, morningShift, member)

	rec := postBooking(r, member, classID)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var created Booking
	decode(t, rec, &created)
	assert.Positive(t, created.ID)
	assert.Equal(t, classID, created.ClassID)
	assert.Equal(t, member, created.MemberID)

	rec = doRequest(r, http.MethodGet, fmt.Sprintf("/api/bookings/%d", created.ID), tokenOf(member), "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var fetched Booking
	decode(t, rec, &fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, classID, fetched.ClassID)
	assert.Equal(t, member, fetched.MemberID)
	assert.True(t, created.CreatedAt.Equal(fetched.CreatedAt))
	assert.Equal(t, 1, countBookings(t, pool, classID))
}

// TestListClassesHTTP: GET /api/classes คืน JSON array ของคลาสที่ยังไม่เริ่ม และ [] เมื่อไม่มี (FR-BKG-01, plan D4).
func TestListClassesHTTP(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	r := newDBRouter(pool, morningShift, member)

	rec := doRequest(r, http.MethodGet, "/api/classes", tokenOf(member), "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `[]`, rec.Body.String())

	newClass(t, pool, 1, span("07:00", "08:00"), 10)
	upcoming := newClass(t, pool, 1, span("10:00", "11:00"), 12)
	rec = doRequest(r, http.MethodGet, "/api/classes", tokenOf(member), "")
	require.Equal(t, http.StatusOK, rec.Code)
	var classes []Class
	decode(t, rec, &classes)
	require.Len(t, classes, 1)
	assert.Equal(t, upcoming, classes[0].ID)
	assert.Equal(t, int64(1), classes[0].TrainerID)
	assert.Equal(t, 12, classes[0].Capacity)
	assert.True(t, classes[0].StartsAt.Equal(at("10:00")))
	assert.True(t, classes[0].EndsAt.Equal(at("11:00")))
}

// TestCreateBookingErrorsHTTP: body ผิด = 400, ไม่พบคลาส = 404, จองซ้ำ = 409 พร้อมรหัสเหตุผลและข้อความไทย (ASM-05).
func TestCreateBookingErrorsHTTP(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	r := newDBRouter(pool, morningShift, member)

	for _, body := range []string{`not json`, `{}`, `{"class_id":0}`, `{"class_id":"1"}`} {
		requireReason(t, doRequest(r, http.MethodPost, "/api/bookings", tokenOf(member), body),
			http.StatusBadRequest, CodeInvalidRequest)
	}
	requireReason(t, postBooking(r, member, 999_999), http.StatusNotFound, CodeClassNotFound)

	classID := newClass(t, pool, 1, span("10:00", "11:00"), 10)
	require.Equal(t, http.StatusCreated, postBooking(r, member, classID).Code)
	requireReason(t, postBooking(r, member, classID), http.StatusConflict, CodeAlreadyBooked)
	assert.Equal(t, 1, countBookings(t, pool, classID))
}

// TestGetBookingOfOtherMemberHTTP: ดูการจองของคนอื่นได้ 404 เหมือนไม่มีอยู่ และ id ผิดรูปได้ 400 (plan D5).
func TestGetBookingOfOtherMemberHTTP(t *testing.T) {
	pool := newTestPool(t)
	owner, other := newMember(t, pool), newMember(t, pool)
	classID := newClass(t, pool, 1, span("10:00", "11:00"), 10)
	r := newDBRouter(pool, morningShift, owner, other)

	rec := postBooking(r, owner, classID)
	require.Equal(t, http.StatusCreated, rec.Code)
	var created Booking
	decode(t, rec, &created)

	path := fmt.Sprintf("/api/bookings/%d", created.ID)
	requireReason(t, doRequest(r, http.MethodGet, path, tokenOf(other), ""), http.StatusNotFound, CodeBookingNotFound)
	requireReason(t, doRequest(r, http.MethodGet, "/api/bookings/abc", tokenOf(owner), ""), http.StatusBadRequest, CodeInvalidRequest)
}

// failingService คือ service ที่ฐานข้อมูลไม่ตอบทุกครั้ง (ASM-07).
type failingService struct{}

var errDatabaseDown = errors.New("database down")

func (failingService) Classes(context.Context) ([]Class, error) { return nil, errDatabaseDown }
func (failingService) Book(context.Context, int64, int64) (Booking, error) {
	return Booking{}, errDatabaseDown
}
func (failingService) Booking(context.Context, int64, int64) (Booking, error) {
	return Booking{}, errDatabaseDown
}

// TestBookingUnavailableHTTP: ฐานข้อมูลไม่ตอบ ทุก endpoint ตอบ 503 unavailable พร้อมข้อความให้ลองใหม่ (ASM-07).
func TestBookingUnavailableHTTP(t *testing.T) {
	r := newRouter(failingService{}, 1)
	requireReason(t, doRequest(r, http.MethodGet, "/api/classes", tokenOf(1), ""), http.StatusServiceUnavailable, CodeUnavailable)
	requireReason(t, postBooking(r, 1, 5), http.StatusServiceUnavailable, CodeUnavailable)
	requireReason(t, doRequest(r, http.MethodGet, "/api/bookings/5", tokenOf(1), ""), http.StatusServiceUnavailable, CodeUnavailable)
}

// TestBookingRequiresSessionHTTP: ไม่มี token หรือ token ไม่ถูกต้อง ทุก endpoint ได้ 401 (precondition "สมาชิกล็อกอินแล้ว", 001).
func TestBookingRequiresSessionHTTP(t *testing.T) {
	r := newRouter(failingService{}, 1)
	for _, req := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/classes", ""},
		{http.MethodPost, "/api/bookings", `{"class_id":1}`},
		{http.MethodGet, "/api/bookings/1", ""},
	} {
		for _, token := range []string{"", "wrong-token"} {
			rec := doRequest(r, req.method, req.path, token, req.body)
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "%s %s token=%q", req.method, req.path, token)
		}
	}
}
