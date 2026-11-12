package roster

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAC_ROST_03_HTTP: Given เทรนเนอร์ที่มีตารางอยู่แล้ว When เจ้าหน้าที่/ผู้จัดการปรับปรุงเวลาเข้างานหรือเวลาสอน
// Then เมื่อเรียกดูซ้ำเห็นค่าใหม่ (FR-ROST-01; แก้ทีละรายการ ASM-05).
func TestAC_ROST_03_HTTP(t *testing.T) {
	pool := newTestPool(t)
	trainer := newTrainer(t, pool, "เอ")
	r := newDBRouter(pool)
	var work, teach entryBody
	decode(t, postEntry(r, trainer, "work", 1, "09:00", "12:00"), &work)
	decode(t, postEntry(r, trainer, "teach", 1, "10:00", "11:00"), &teach)

	// ปรับเวลาเข้างานให้กว้างขึ้นก่อน แล้วย้ายเวลาสอนเข้าไปช่วงใหม่ (เวลาสอนต้องอยู่ในเวลาเข้างาน ASM-04)
	rec := doRequest(r, http.MethodPatch, fmt.Sprintf("/api/roster/%d", work.ID), `{"weekday":1,"start_time":"09:00","end_time":"17:00"}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	rec = doRequest(r, http.MethodPatch, fmt.Sprintf("/api/roster/%d", teach.ID), `{"weekday":1,"start_time":"15:00","end_time":"16:30"}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	assert.Equal(t, []entryBody{
		{ID: work.ID, TrainerID: trainer, Kind: "work", Weekday: 1, StartTime: "09:00", EndTime: "17:00"},
		{ID: teach.ID, TrainerID: trainer, Kind: "teach", Weekday: 1, StartTime: "15:00", EndTime: "16:30"},
	}, getRoster(t, r, trainer))
}

// TestAC_ROST_04_HTTP: Given เทรนเนอร์ A และ B มีตารางต่างกัน When ปรับปรุงตารางของ A Then ตารางของ B ไม่เปลี่ยน
// ทดสอบทั้งแก้ ยกเลิก และลบ (FR-ROST-01; ASM-01, ASM-05).
func TestAC_ROST_04_HTTP(t *testing.T) {
	pool := newTestPool(t)
	a := newTrainer(t, pool, "เอ")
	b := newTrainer(t, pool, "บี")
	r := newDBRouter(pool)

	var aMonday, aTuesday, aWednesday entryBody
	decode(t, postEntry(r, a, "work", 1, "09:00", "12:00"), &aMonday)
	decode(t, postEntry(r, a, "work", 2, "09:00", "12:00"), &aTuesday)
	decode(t, postEntry(r, a, "work", 3, "09:00", "12:00"), &aWednesday)
	require.Equal(t, http.StatusCreated, postEntry(r, b, "work", 1, "09:00", "12:00").Code) // เวลาเดียวกับ A
	require.Equal(t, http.StatusCreated, postEntry(r, b, "work", 2, "13:00", "17:00").Code)
	require.Equal(t, http.StatusCreated, postEntry(r, b, "teach", 2, "14:00", "15:00").Code)
	bBefore := getRoster(t, r, b)

	require.Equal(t, http.StatusOK, doRequest(r, http.MethodPatch, fmt.Sprintf("/api/roster/%d", aMonday.ID),
		`{"weekday":1,"start_time":"10:00","end_time":"18:00"}`).Code)
	require.Equal(t, http.StatusOK, doRequest(r, http.MethodPost, fmt.Sprintf("/api/roster/%d/cancel", aTuesday.ID), "").Code)
	require.Equal(t, http.StatusNoContent, doRequest(r, http.MethodDelete, fmt.Sprintf("/api/roster/%d", aWednesday.ID), "").Code)

	assert.Equal(t, bBefore, getRoster(t, r, b), "ตารางของ B ต้องเหมือนเดิมทุกฟิลด์")
	aAfter := getRoster(t, r, a)
	require.Len(t, aAfter, 2)
	assert.Equal(t, "10:00", aAfter[0].StartTime)
	assert.True(t, aAfter[1].Cancelled)
}
