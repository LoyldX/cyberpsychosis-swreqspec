package occupancy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/app"
	"gymapp/backend/internal/testutil"
)

// newTestRouter ประกอบ router ของ occupancy บน counter ที่กำหนด (CON-TECH-01).
func newTestRouter(store counter, clock Clock) (*gin.Engine, *Service) {
	gin.SetMode(gin.TestMode)
	svc := NewService(store, NewBroadcaster(), clock)
	return app.NewRouter(NewHandler(svc).Register), svc
}

// newDBRouter ประกอบ router บน Store จริงของ PostgreSQL และนาฬิกาจริง.
func newDBRouter(pool *pgxpool.Pool) *gin.Engine {
	router, _ := newTestRouter(NewStore(pool), nil)
	return router
}

// postSignal ยิง POST /api/gate/signals ด้วยทิศทาง d.
func postSignal(t *testing.T, router http.Handler, d string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(map[string]string{"direction": d})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/gate/signals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// postSignals ยิงสัญญาณทิศทาง d จำนวน n ครั้ง และต้องได้ 201 ทุกครั้ง.
func postSignals(t *testing.T, router http.Handler, d string, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		require.Equal(t, http.StatusCreated, postSignal(t, router, d).Code)
	}
}

// getCount ยิง GET /api/occupancy แล้วคืน count (ต้องได้ 200).
func getCount(t *testing.T, router http.Handler) int {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/occupancy", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	return decodeSnapshot(t, rec).Count
}

func decodeSnapshot(t *testing.T, rec *httptest.ResponseRecorder) Snapshot {
	t.Helper()
	var snap Snapshot
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &snap))
	return snap
}

// TestAC_BI_01_HTTP: Given วันปัจจุบันระบบรับสัญญาณผ่านเข้า 5 ครั้งและผ่านออก 2 ครั้ง
// When ผู้จัดการเปิด Dashboard (GET /api/occupancy) Then เห็นจำนวนปัจจุบันเป็น 3 (FR-BI-01).
func TestAC_BI_01_HTTP(t *testing.T) {
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)
	router := newDBRouter(pool)

	// Given
	postSignals(t, router, "in", 5)
	postSignals(t, router, "out", 2)
	// When / Then
	require.Equal(t, 3, getCount(t, router))
}

// TestAC_BI_02_HTTP: Given จำนวนปัจจุบัน 10 When รับสัญญาณผ่านเข้า 1 ครั้ง Then จำนวนปัจจุบันเป็น 11 (DOM-CAP-01).
func TestAC_BI_02_HTTP(t *testing.T) {
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)
	router := newDBRouter(pool)

	postSignals(t, router, "in", 10)
	rec := postSignal(t, router, "in")

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, 11, decodeSnapshot(t, rec).Count)
	require.Equal(t, 11, getCount(t, router))
}

// TestAC_BI_03_HTTP: Given จำนวนปัจจุบัน 10 When รับสัญญาณผ่านออก 1 ครั้ง Then จำนวนปัจจุบันเป็น 9 (DOM-CAP-01, ASM-01).
func TestAC_BI_03_HTTP(t *testing.T) {
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)
	router := newDBRouter(pool)

	postSignals(t, router, "in", 10)
	rec := postSignal(t, router, "out")

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, 9, decodeSnapshot(t, rec).Count)
	require.Equal(t, 9, getCount(t, router))
}

// TestAC_BI_05_HTTP: Given จำนวนปัจจุบัน 0 When รับสัญญาณผ่านออก 1 ครั้ง Then จำนวนยังเป็น 0 และสัญญาณนั้นถูกบันทึกไว้ (ASM-04).
func TestAC_BI_05_HTTP(t *testing.T) {
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)
	router := newDBRouter(pool)

	require.Equal(t, 0, getCount(t, router))
	rec := postSignal(t, router, "out")

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, 0, decodeSnapshot(t, rec).Count)
	var rows int
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT count(*) FROM gate_signals WHERE direction = 'out'`).Scan(&rows))
	require.Equal(t, 1, rows)
}

// TestAC_BI_07_HTTP: Given จำนวนปัจจุบัน 10 When backend รีสตาร์ตภายในวันเดียวกัน (pool/service/router ใหม่) Then จำนวนปัจจุบันยังเป็น 10 (ASM-02).
func TestAC_BI_07_HTTP(t *testing.T) {
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)
	postSignals(t, newDBRouter(pool), "in", 10)

	restarted, err := pgxpool.New(context.Background(), testutil.TestDSN())
	require.NoError(t, err)
	defer restarted.Close()

	require.Equal(t, 10, getCount(t, newDBRouter(restarted)))
}

// TestPostSignalRejectsInvalidDirection: direction อื่นนอกจาก in/out ได้ 400 และไม่ถูกบันทึก (IF-GATE-01) — ไม่ต้องต่อฐานข้อมูล.
func TestPostSignalRejectsInvalidDirection(t *testing.T) {
	router, _ := newTestRouter(&memoryCounter{}, newFakeClock(testNoon))
	for _, bad := range []string{"", "IN", "sideways"} {
		rec := postSignal(t, router, bad)
		require.Equal(t, http.StatusBadRequest, rec.Code, bad)
		require.JSONEq(t, `{"error":"invalid_direction"}`, rec.Body.String())
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/gate/signals", bytes.NewReader([]byte("not json"))))
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, 0, getCount(t, router))
}

// failingCounter จำลองฐานข้อมูลที่ใช้ไม่ได้ (plan D6).
type failingCounter struct{}

var errDatabaseDown = errors.New("database down")

func (failingCounter) RecordSignal(context.Context, Direction, time.Time) (int, error) {
	return 0, errDatabaseDown
}
func (failingCounter) CurrentCount(context.Context, time.Time) (int, error) {
	return 0, errDatabaseDown
}

// TestHandlerReturns503WhenDatabaseUnavailable: ฐานข้อมูลใช้ไม่ได้ → 503 ทั้งอ่านและบันทึก (plan D6, ยังไม่มี AC).
func TestHandlerReturns503WhenDatabaseUnavailable(t *testing.T) {
	router, _ := newTestRouter(failingCounter{}, newFakeClock(testNoon))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/occupancy", nil))
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.JSONEq(t, `{"error":"unavailable"}`, rec.Body.String())

	rec = postSignal(t, router, "in")
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.JSONEq(t, `{"error":"unavailable"}`, rec.Body.String())
}
