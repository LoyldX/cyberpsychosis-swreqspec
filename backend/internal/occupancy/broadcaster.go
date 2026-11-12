package occupancy

import (
	"sync"
	"time"
)

// Snapshot คือจำนวนปัจจุบัน ณ เวลาหนึ่ง ที่ส่งให้ Dashboard (FR-BI-01, NFR-PERF-02).
type Snapshot struct {
	Count int       `json:"count"`
	AsOf  time.Time `json:"as_of"`
}

// Broadcaster กระจาย Snapshot ใหม่ไปยังผู้ฟังทุกราย (เช่น การเชื่อมต่อ SSE ของ Dashboard) (NFR-PERF-02, ASM-06).
// ผู้ฟังที่อ่านไม่ทันจะได้ค่าล่าสุดแทนค่าเก่า เพื่อไม่ให้ผู้ฟังรายเดียวทำให้รายอื่นช้า.
type Broadcaster struct {
	mu   sync.Mutex
	subs map[chan Snapshot]struct{}
}

// NewBroadcaster สร้าง Broadcaster ที่ยังไม่มีผู้ฟัง (NFR-PERF-02).
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subs: make(map[chan Snapshot]struct{})}
}

// Subscribe ลงทะเบียนผู้ฟังใหม่ คืนช่องรับ Snapshot และฟังก์ชันยกเลิก (เรียกซ้ำได้) (NFR-PERF-02).
func (b *Broadcaster) Subscribe() (<-chan Snapshot, func()) {
	ch := make(chan Snapshot, 1)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subs, ch)
			b.mu.Unlock()
		})
	}
}

// Publish ส่ง Snapshot ให้ผู้ฟังทุกรายโดยไม่ block; ถ้าช่องของผู้ฟังเต็ม จะแทนค่าเก่าด้วยค่าใหม่ (NFR-PERF-02, ASM-06).
func (b *Broadcaster) Publish(s Snapshot) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- s:
			continue
		default:
		}
		select { // ทิ้งค่าเก่าที่ยังไม่ถูกอ่าน
		case <-ch:
		default:
		}
		select {
		case ch <- s:
		default:
		}
	}
}

// Subscribers คืนจำนวนผู้ฟังที่ลงทะเบียนอยู่ (ใช้ตรวจว่าการเชื่อมต่อที่ปิดแล้วถูกถอดออก).
func (b *Broadcaster) Subscribers() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subs)
}
