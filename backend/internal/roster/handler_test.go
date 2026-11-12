package roster

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixedClock คือ Clock จำลองที่คืนเวลาเดียวเสมอ.
type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

// newRouter ประกอบ Gin router ของตารางงานกับ store ที่กำหนด (ไม่มี middleware ยืนยันตัวตน ASM-06).
func newRouter(store rosterStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(store, fixedClock{changeNow}).Register(r)
	return r
}

// newDBRouter คือ router ที่ใช้ PostgreSQL จริง.
func newDBRouter(pool *pgxpool.Pool) *gin.Engine { return newRouter(NewStore(pool)) }

// doRequest ส่งคำขอโดยไม่มี Authorization (ASM-06) แล้วคืนผล.
func doRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// entryBody คือรูปแบบรายการใน JSON ตามสัญญา plan ข้อ 4.
type entryBody struct {
	ID        int64  `json:"id"`
	TrainerID int64  `json:"trainer_id"`
	Kind      string `json:"kind"`
	Weekday   int    `json:"weekday"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Cancelled bool   `json:"cancelled"`
}

// postEntry ส่ง POST /api/trainers/:id/roster.
func postEntry(r http.Handler, trainerID int64, kind string, weekday int, start, end string) *httptest.ResponseRecorder {
	return doRequest(r, http.MethodPost, fmt.Sprintf("/api/trainers/%d/roster", trainerID),
		fmt.Sprintf(`{"kind":%q,"weekday":%d,"start_time":%q,"end_time":%q}`, kind, weekday, start, end))
}

// getRoster เรียก GET /api/trainers/:id/roster แล้วคืนรายการ.
func getRoster(t *testing.T, r http.Handler, trainerID int64) []entryBody {
	t.Helper()
	rec := doRequest(r, http.MethodGet, fmt.Sprintf("/api/trainers/%d/roster", trainerID), "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var entries []entryBody
	decode(t, rec, &entries)
	return entries
}

// decode แปลง JSON ใน recorder เป็น v.
func decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), v), rec.Body.String())
}

// requireReason ตรวจรหัส HTTP รหัสเหตุผล และมีข้อความไทย.
func requireReason(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	require.Equal(t, status, rec.Code, rec.Body.String())
	var body struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	decode(t, rec, &body)
	assert.Equal(t, code, body.Error)
	assert.NotEmpty(t, body.Message)
}

// TestAC_ROST_01_HTTP: Given เทรนเนอร์ที่ยังไม่มีเวลาเข้างาน When เจ้าหน้าที่/ผู้จัดการกำหนดเวลาเข้างานให้เทรนเนอร์นั้น
// Then เวลาเข้างานถูกบันทึกและเรียกดูได้ (FR-ROST-01; ไม่ต้องล็อกอิน ASM-06).
func TestAC_ROST_01_HTTP(t *testing.T) {
	pool := newTestPool(t)
	trainer := newTrainer(t, pool, "เอ")
	r := newDBRouter(pool)
	require.Empty(t, getRoster(t, r, trainer), "Given: ยังไม่มีเวลาเข้างาน")

	rec := postEntry(r, trainer, "work", 1, "09:00", "12:00")
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var created entryBody
	decode(t, rec, &created)
	want := entryBody{ID: created.ID, TrainerID: trainer, Kind: "work", Weekday: 1, StartTime: "09:00", EndTime: "12:00"}
	assert.Equal(t, want, created)

	assert.Equal(t, []entryBody{want}, getRoster(t, r, trainer))
}

// TestAC_ROST_02_HTTP: Given เทรนเนอร์ที่ยังไม่มีเวลาสอน (มีเวลาเข้างานแล้ว เพราะเวลาสอนต้องอยู่ในเวลาเข้างาน ASM-04)
// When เจ้าหน้าที่/ผู้จัดการกำหนดเวลาสอนให้เทรนเนอร์นั้น Then เวลาสอนถูกบันทึกและเรียกดูได้ (FR-ROST-01).
func TestAC_ROST_02_HTTP(t *testing.T) {
	pool := newTestPool(t)
	trainer := newTrainer(t, pool, "เอ")
	r := newDBRouter(pool)
	require.Equal(t, http.StatusCreated, postEntry(r, trainer, "work", 1, "09:00", "12:00").Code)

	rec := postEntry(r, trainer, "teach", 1, "10:00", "11:00")
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var created entryBody
	decode(t, rec, &created)
	assert.Equal(t, "teach", created.Kind)

	entries := getRoster(t, r, trainer)
	require.Len(t, entries, 2)
	assert.Equal(t, entryBody{ID: created.ID, TrainerID: trainer, Kind: "teach", Weekday: 1, StartTime: "10:00", EndTime: "11:00"}, entries[1])
}

// TestListTrainersHTTP: GET /api/trainers คืนเทรนเนอร์ทั้งหมดเรียงตามรหัส และ [] เมื่อไม่มี (ASM-09).
func TestListTrainersHTTP(t *testing.T) {
	pool := newTestPool(t)
	r := newDBRouter(pool)
	rec := doRequest(r, http.MethodGet, "/api/trainers", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `[]`, rec.Body.String())

	a := newTrainer(t, pool, "เอ")
	b := newTrainer(t, pool, "บี")
	rec = doRequest(r, http.MethodGet, "/api/trainers", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, fmt.Sprintf(`[{"id":%d,"name":"เอ"},{"id":%d,"name":"บี"}]`, a, b), rec.Body.String())
}

// TestCreateEntryErrorsHTTP: รูปแบบผิด = 400, ไม่พบเทรนเนอร์ = 404, ทับซ้อนหรือเวลาสอนนอกเวลาเข้างาน = 409 และไม่บันทึก
// (ASM-02, ASM-03, ASM-04, ASM-10).
func TestCreateEntryErrorsHTTP(t *testing.T) {
	pool := newTestPool(t)
	trainer := newTrainer(t, pool, "เอ")
	r := newDBRouter(pool)
	path := fmt.Sprintf("/api/trainers/%d/roster", trainer)

	for _, body := range []string{
		`not json`,
		`{"kind":"work","start_time":"09:00","end_time":"12:00"}`,
		`{"kind":"lunch","weekday":1,"start_time":"09:00","end_time":"12:00"}`,
		`{"kind":"work","weekday":8,"start_time":"09:00","end_time":"12:00"}`,
		`{"kind":"work","weekday":1,"start_time":"9am","end_time":"12:00"}`,
		`{"kind":"work","weekday":1,"start_time":"12:00","end_time":"09:00"}`,
	} {
		requireReason(t, doRequest(r, http.MethodPost, path, body), http.StatusBadRequest, CodeInvalidRequest)
	}
	requireReason(t, doRequest(r, http.MethodPost, "/api/trainers/abc/roster", `{}`), http.StatusBadRequest, CodeInvalidRequest)
	requireReason(t, postEntry(r, trainer+1000, "work", 1, "09:00", "12:00"), http.StatusNotFound, CodeTrainerNotFound)
	requireReason(t, doRequest(r, http.MethodGet, fmt.Sprintf("/api/trainers/%d/roster", trainer+1000), ""), http.StatusNotFound, CodeTrainerNotFound)

	requireReason(t, postEntry(r, trainer, "teach", 1, "10:00", "11:00"), http.StatusConflict, CodeTeachOutsideWork)
	require.Equal(t, http.StatusCreated, postEntry(r, trainer, "work", 1, "09:00", "12:00").Code)
	requireReason(t, postEntry(r, trainer, "work", 1, "11:00", "13:00"), http.StatusConflict, CodeOverlap)
	require.Equal(t, http.StatusCreated, postEntry(r, trainer, "work", 1, "12:00", "15:00").Code, "ต่อกันพอดีได้ (ASM-10)")
	assert.Len(t, getRoster(t, r, trainer), 2)
}

// TestChangeEntryHTTP: แก้/ยกเลิก/ลบรายการที่ไม่มีได้ 404, ยกเลิกได้ 200 พร้อม cancelled, ลบได้ 204 (ASM-01, ASM-05).
func TestChangeEntryHTTP(t *testing.T) {
	pool := newTestPool(t)
	trainer := newTrainer(t, pool, "เอ")
	r := newDBRouter(pool)

	requireReason(t, doRequest(r, http.MethodPatch, "/api/roster/999999", `{"weekday":1,"start_time":"09:00","end_time":"10:00"}`), http.StatusNotFound, CodeEntryNotFound)
	requireReason(t, doRequest(r, http.MethodPost, "/api/roster/999999/cancel", ""), http.StatusNotFound, CodeEntryNotFound)
	requireReason(t, doRequest(r, http.MethodDelete, "/api/roster/999999", ""), http.StatusNotFound, CodeEntryNotFound)

	var monday, tuesday entryBody
	decode(t, postEntry(r, trainer, "work", 1, "09:00", "12:00"), &monday)
	decode(t, postEntry(r, trainer, "work", 2, "09:00", "12:00"), &tuesday)
	requireReason(t, doRequest(r, http.MethodPatch, fmt.Sprintf("/api/roster/%d", monday.ID), `{"weekday":1,"start_time":"12:00","end_time":"09:00"}`), http.StatusBadRequest, CodeInvalidRequest)

	rec := doRequest(r, http.MethodPost, fmt.Sprintf("/api/roster/%d/cancel", monday.ID), "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var cancelled entryBody
	decode(t, rec, &cancelled)
	assert.True(t, cancelled.Cancelled)

	rec = doRequest(r, http.MethodDelete, fmt.Sprintf("/api/roster/%d", tuesday.ID), "")
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	assert.Empty(t, rec.Body.String())

	entries := getRoster(t, r, trainer)
	require.Len(t, entries, 1)
	assert.True(t, entries[0].Cancelled, "รายการที่ยกเลิกยังเห็นในตาราง (ASM-01, plan D5)")
}

// failingStore คือ store ที่ฐานข้อมูลไม่ตอบทุกครั้ง (ASM-08).
type failingStore struct{}

var errDatabaseDown = errors.New("database down")

func (failingStore) ListTrainers(context.Context) ([]Trainer, error)     { return nil, errDatabaseDown }
func (failingStore) ListEntries(context.Context, int64) ([]Entry, error) { return nil, errDatabaseDown }
func (failingStore) Create(context.Context, int64, Kind, Slot) (Entry, error) {
	return Entry{}, errDatabaseDown
}
func (failingStore) Update(context.Context, int64, Slot, time.Time) (Entry, error) {
	return Entry{}, errDatabaseDown
}
func (failingStore) Cancel(context.Context, int64, time.Time) (Entry, error) {
	return Entry{}, errDatabaseDown
}
func (failingStore) Delete(context.Context, int64, time.Time) error { return errDatabaseDown }

// TestRosterUnavailableHTTP: ฐานข้อมูลไม่ตอบ ทุก endpoint ตอบ 503 unavailable พร้อมข้อความให้ลองใหม่ (ASM-08).
func TestRosterUnavailableHTTP(t *testing.T) {
	r := newRouter(failingStore{})
	for _, req := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/trainers", ""},
		{http.MethodGet, "/api/trainers/1/roster", ""},
		{http.MethodPost, "/api/trainers/1/roster", `{"kind":"work","weekday":1,"start_time":"09:00","end_time":"12:00"}`},
		{http.MethodPatch, "/api/roster/1", `{"weekday":1,"start_time":"09:00","end_time":"12:00"}`},
		{http.MethodPost, "/api/roster/1/cancel", ""},
		{http.MethodDelete, "/api/roster/1", ""},
	} {
		requireReason(t, doRequest(r, req.method, req.path, req.body), http.StatusServiceUnavailable, CodeUnavailable)
	}
}
