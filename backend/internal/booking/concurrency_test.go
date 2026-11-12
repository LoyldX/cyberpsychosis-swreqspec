package booking

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// concurrencyRounds คือจำนวนรอบที่ยิงคำขอพร้อมกัน เพราะผลขึ้นกับจังหวะ จึงต้องผ่านทุกรอบ (tasks T-07).
const concurrencyRounds = 20

// bookingLatencyLimit คือเวลาที่ทั้งสองคำขอต้องได้ผล (NFR-PERF-01, ASM-02).
const bookingLatencyLimit = time.Second

// TestAC_BKG_05: Given คลาส X และ Y ของเทรนเนอร์เดียวกันมีเวลาทับซ้อนกันและยังไม่มีผู้จอง และสมาชิกสองคนที่ล็อกอินแล้ว
// When ส่งคำขอจอง X และ Y พร้อมกัน Then สำเร็จเพียงรายการเดียว อีกรายการถูกบล็อก และทั้งสองคำขอได้รับผลภายใน 1 วินาที
// (NFR-PERF-01, CON-DB-01; ASM-02, ASM-10, plan D2) — PostgreSQL จริง รันด้วย go test -race.
func TestAC_BKG_05(t *testing.T) {
	pool := newTestPool(t)
	first, second := newMember(t, pool), newMember(t, pool)
	shifts := fakeShifts{byTrainer: map[int64]Interval{}}
	for round := 1; round <= concurrencyRounds; round++ {
		shifts.byTrainer[int64(round)] = span("08:30", "13:00") // อยู่ในกะ ตรวจเฉพาะการแข่งกัน
	}
	r := newDBRouter(pool, shifts, first, second)

	for round := 1; round <= concurrencyRounds; round++ {
		trainer := int64(round) // เทรนเนอร์ใหม่ทุกรอบ จึงไม่มีคิวสอนจากรอบก่อน
		x := newClass(t, pool, trainer, span("10:00", "11:00"), 10)
		y := newClass(t, pool, trainer, span("10:30", "11:30"), 10)

		results := bookConcurrently(r, []bookingAttempt{{first, x}, {second, y}})

		var created, busy int
		for i, res := range results {
			assert.LessOrEqual(t, res.elapsed, bookingLatencyLimit, "รอบ %d คำขอ %d ช้าเกิน 1 วินาที (ASM-02)", round, i)
			switch res.rec.Code {
			case http.StatusCreated:
				created++
			case http.StatusConflict:
				busy++
				requireReason(t, res.rec, http.StatusConflict, CodeTrainerBusy)
			default:
				t.Fatalf("รอบ %d คำขอ %d ได้ %d: %s", round, i, res.rec.Code, res.rec.Body.String())
			}
		}
		require.Equal(t, 1, created, "รอบ %d ต้องสำเร็จเพียงรายการเดียว", round)
		require.Equal(t, 1, busy, "รอบ %d อีกรายการต้องถูกบล็อก", round)
		require.Equal(t, 1, countBookings(t, pool, x)+countBookings(t, pool, y), "รอบ %d ต้องมีการจองเพียง 1 แถว", round)
	}
}

// bookingAttempt คือคำขอจองหนึ่งรายการ: สมาชิกและคลาส.
type bookingAttempt struct {
	memberID int64
	classID  int64
}

// attemptResult คือผลของคำขอและเวลาที่ใช้.
type attemptResult struct {
	rec     *httptest.ResponseRecorder
	elapsed time.Duration
}

// bookConcurrently ปล่อยทุกคำขอพร้อมกันด้วย channel แล้วรอจนครบ คืนผลตามลำดับคำขอ.
func bookConcurrently(r http.Handler, attempts []bookingAttempt) []attemptResult {
	results := make([]attemptResult, len(attempts))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, a := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			began := time.Now()
			rec := postBooking(r, a.memberID, a.classID)
			results[i] = attemptResult{rec: rec, elapsed: time.Since(began)}
		}()
	}
	close(start)
	wg.Wait()
	return results
}
