package occupancy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/testutil"
)

// ac04Deadline คือเวลาสูงสุดที่ตัวเลขต้องเปลี่ยนหลังระบบรับสัญญาณ (ASM-06, NFR-PERF-02).
const ac04Deadline = time.Second

// openStream เปิด GET /api/occupancy/stream บนเซิร์ฟเวอร์จริง แล้วคืนช่องที่ได้ Snapshot ของทุก event "occupancy"
// และฟังก์ชันปิดการเชื่อมต่อ (เหมือนผู้จัดการปิดหน้า).
func openStream(t *testing.T, baseURL string) (<-chan Snapshot, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/occupancy/stream", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream"))

	events := make(chan Snapshot, 16)
	go func() {
		defer close(events)
		scanner := bufio.NewScanner(resp.Body)
		event := ""
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event:"):
				event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "data:") && event == sseEventName:
				var snap Snapshot
				if json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &snap) == nil {
					events <- snap
				}
			}
		}
	}()
	closeFn := func() { cancel(); _ = resp.Body.Close() }
	t.Cleanup(closeFn)
	return events, closeFn
}

// postSignalTo ยิง POST /api/gate/signals ไปยังเซิร์ฟเวอร์จริง และต้องได้ 201.
func postSignalTo(t *testing.T, baseURL, d string) {
	t.Helper()
	resp, err := http.Post(baseURL+"/api/gate/signals", "application/json", bytes.NewReader([]byte(`{"direction":"`+d+`"}`)))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

// assertAC04 คือขั้นตอน Given/When/Then ของ AC-BI-04 บนเซิร์ฟเวอร์ที่กำหนด.
func assertAC04(t *testing.T, server *httptest.Server, svc *Service) {
	// Given: ผู้จัดการเปิด Dashboard ค้างไว้ — event แรกคือจำนวนปัจจุบัน
	events, closeStream := openStream(t, server.URL)
	first := receiveWithin(t, events, ac04Deadline)
	require.Equal(t, 3, first.Count)

	// When: ระบบรับสัญญาณผ่านประตู
	postSignalTo(t, server.URL, "in")
	received := time.Now()

	// Then: ตัวเลขเปลี่ยนเองภายใน 1 วินาทีนับจากที่ระบบรับสัญญาณ โดยไม่ต้องเปิด stream ใหม่
	got := receiveWithin(t, events, ac04Deadline)
	require.Equal(t, 4, got.Count)
	require.Less(t, time.Since(received), ac04Deadline)

	// ปิดหน้าแล้วผู้ฟังต้องถูกถอดออก
	closeStream()
	require.Eventually(t, func() bool { return svc.hub.Subscribers() == 0 }, time.Second, 5*time.Millisecond)
}

// TestAC_BI_04: Given ผู้จัดการเปิด Dashboard ค้างไว้ When ระบบรับสัญญาณผ่านประตู
// Then ตัวเลขบน Dashboard เปลี่ยนเองภายใน 1 วินาทีนับจากที่ระบบรับสัญญาณ (NFR-PERF-02, ASM-06) — Store จริงบน PostgreSQL.
func TestAC_BI_04(t *testing.T) {
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)
	router, svc := newTestRouter(NewStore(pool), nil)
	server := httptest.NewServer(router)
	defer server.Close()
	postSignals(t, router, "in", 3)

	assertAC04(t, server, svc)
}

// TestAC_BI_04_InMemory ตรวจขั้นตอนเดียวกับ TestAC_BI_04 ด้วย counter ในหน่วยความจำ (ไม่ต้องต่อฐานข้อมูล) (NFR-PERF-02).
func TestAC_BI_04_InMemory(t *testing.T) {
	router, svc := newTestRouter(&memoryCounter{}, nil)
	server := httptest.NewServer(router)
	defer server.Close()
	postSignals(t, router, "in", 3)

	assertAC04(t, server, svc)
}

// TestStreamReturns503WhenDatabaseUnavailable: เปิด stream ตอนฐานข้อมูลใช้ไม่ได้ → 503 และไม่ค้างผู้ฟังไว้ (plan D6).
func TestStreamReturns503WhenDatabaseUnavailable(t *testing.T) {
	router, svc := newTestRouter(failingCounter{}, newFakeClock(testNoon))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/occupancy/stream", nil))
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.Equal(t, 0, svc.hub.Subscribers())
}
