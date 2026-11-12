package occupancy

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/testutil"
)

// fakeClock คือ Clock จำลองที่ test เลื่อนเวลาเองได้ (ASM-03, AC-BI-06).
type fakeClock struct {
	mu      sync.Mutex
	now     time.Time
	waiters []fakeWaiter
}

type fakeWaiter struct {
	deadline time.Time
	ch       chan time.Time
}

func newFakeClock(now time.Time) *fakeClock { return &fakeClock{now: now} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := make(chan time.Time, 1)
	c.waiters = append(c.waiters, fakeWaiter{deadline: c.now.Add(d), ch: ch})
	return ch
}

// Advance เลื่อนเวลาไปที่ t และปลุกตัวจับเวลาที่ถึงกำหนดแล้ว.
func (c *fakeClock) Advance(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
	kept := c.waiters[:0]
	for _, w := range c.waiters {
		if !w.deadline.After(t) {
			w.ch <- t
			continue
		}
		kept = append(kept, w)
	}
	c.waiters = kept
}

func (c *fakeClock) pendingWaiters() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.waiters)
}

// memoryCounter คือ counter ในหน่วยความจำสำหรับ test ของ Service ที่ไม่ต้องต่อฐานข้อมูล (ใช้กฎเดียวกับ Store).
type memoryCounter struct {
	mu      sync.Mutex
	signals []memorySignal
}

type memorySignal struct {
	at    time.Time
	after int
}

func (m *memoryCounter) RecordSignal(_ context.Context, d Direction, at time.Time) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	next := ApplySignal(m.countSince(StartOfDay(at)), d)
	m.signals = append(m.signals, memorySignal{at: at, after: next})
	return next, nil
}

func (m *memoryCounter) CurrentCount(_ context.Context, now time.Time) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.countSince(StartOfDay(now)), nil
}

func (m *memoryCounter) countSince(since time.Time) int {
	for i := len(m.signals) - 1; i >= 0; i-- {
		if !m.signals[i].at.Before(since) {
			return m.signals[i].after
		}
	}
	return 0
}

// receiveWithin รอ Snapshot จากช่องไม่เกิน d.
func receiveWithin(t *testing.T, ch <-chan Snapshot, d time.Duration) Snapshot {
	t.Helper()
	select {
	case s := <-ch:
		return s
	case <-time.After(d):
		t.Fatalf("ไม่ได้รับจำนวนใหม่ภายใน %v", d)
		return Snapshot{}
	}
}

// startMidnightReset รัน RunMidnightReset และรอจนตั้งตัวจับเวลาแล้ว (กันการเลื่อนเวลาก่อน goroutine พร้อม).
func startMidnightReset(t *testing.T, svc *Service, clock *fakeClock) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { svc.RunMidnightReset(ctx); close(done) }()
	t.Cleanup(func() { cancel(); <-done })
	require.Eventually(t, func() bool { return clock.pendingWaiters() == 1 }, time.Second, time.Millisecond)
}

var (
	testBeforeMidnight = time.Date(2026, 11, 12, 23, 59, 0, 0, bangkok)
	testMidnight       = time.Date(2026, 11, 13, 0, 0, 0, 0, bangkok)
)

// TestAC_BI_06: Given จำนวนปัจจุบัน 10 เมื่อเวลา 23:59 When ถึง 00:00 (Asia/Bangkok) ของวันถัดไป Then จำนวนปัจจุบันเป็น 0
// ใช้ Store จริงบน PostgreSQL + นาฬิกาจำลอง และตรวจว่าผู้ฟังได้รับ count 0 (DOM-CAP-01, ASM-03, plan D3).
func TestAC_BI_06(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)
	clock := newFakeClock(testBeforeMidnight)
	svc := NewService(NewStore(pool), NewBroadcaster(), clock)

	// Given
	for i := 0; i < 10; i++ {
		_, err := svc.Record(ctx, DirectionIn)
		require.NoError(t, err)
	}
	snap, err := svc.Current(ctx)
	require.NoError(t, err)
	require.Equal(t, 10, snap.Count)
	updates, unsubscribe := svc.Subscribe()
	defer unsubscribe()
	startMidnightReset(t, svc, clock)

	// When
	clock.Advance(testMidnight)

	// Then
	got := receiveWithin(t, updates, time.Second)
	require.Equal(t, 0, got.Count)
	snap, err = svc.Current(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, snap.Count)
}

// TestAC_BI_06_MidnightPublish ตรวจตรรกะเดียวกับ TestAC_BI_06 ด้วย counter ในหน่วยความจำ (ไม่ต้องต่อฐานข้อมูล)
// และตรวจว่าตัวจับเวลาตั้งรอบถัดไปหลังเที่ยงคืน (ASM-03, plan D3).
func TestAC_BI_06_MidnightPublish(t *testing.T) {
	ctx := context.Background()
	clock := newFakeClock(testBeforeMidnight)
	svc := NewService(&memoryCounter{}, NewBroadcaster(), clock)
	for i := 0; i < 10; i++ {
		_, err := svc.Record(ctx, DirectionIn)
		require.NoError(t, err)
	}
	updates, unsubscribe := svc.Subscribe()
	defer unsubscribe()
	startMidnightReset(t, svc, clock)

	clock.Advance(testMidnight)

	got := receiveWithin(t, updates, time.Second)
	require.Equal(t, 0, got.Count)
	require.Equal(t, testMidnight, got.AsOf)
	require.Eventually(t, func() bool { return clock.pendingWaiters() == 1 }, time.Second, time.Millisecond)
}

// TestServiceRecordPublishesToAllSubscribers ยืนยันว่าบันทึกสัญญาณแล้วผู้ฟังทุกรายได้รับจำนวนใหม่ (NFR-PERF-02).
func TestServiceRecordPublishesToAllSubscribers(t *testing.T) {
	ctx := context.Background()
	svc := NewService(&memoryCounter{}, NewBroadcaster(), newFakeClock(testNoon))
	a, unsubA := svc.Subscribe()
	defer unsubA()
	b, unsubB := svc.Subscribe()
	defer unsubB()

	snap, err := svc.Record(ctx, DirectionIn)
	require.NoError(t, err)
	require.Equal(t, 1, snap.Count)
	require.Equal(t, 1, receiveWithin(t, a, time.Second).Count)
	require.Equal(t, 1, receiveWithin(t, b, time.Second).Count)
}

// TestBroadcasterKeepsLatestForSlowSubscriber ยืนยันว่าผู้ฟังที่อ่านไม่ทันได้ค่าล่าสุด และ Publish ไม่ block (NFR-PERF-02).
func TestBroadcasterKeepsLatestForSlowSubscriber(t *testing.T) {
	hub := NewBroadcaster()
	ch, unsubscribe := hub.Subscribe()
	hub.Publish(Snapshot{Count: 1})
	hub.Publish(Snapshot{Count: 2})
	hub.Publish(Snapshot{Count: 3})
	require.Equal(t, 3, receiveWithin(t, ch, time.Second).Count)

	require.Equal(t, 1, hub.Subscribers())
	unsubscribe()
	unsubscribe() // เรียกซ้ำได้
	require.Equal(t, 0, hub.Subscribers())
}
