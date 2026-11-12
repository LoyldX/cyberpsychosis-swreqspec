package roster

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// rosterStore คือสิ่งที่ Handler ต้องใช้จาก Store (FR-ROST-01, CON-DB-01).
type rosterStore interface {
	ListTrainers(ctx context.Context) ([]Trainer, error)
	ListEntries(ctx context.Context, trainerID int64) ([]Entry, error)
	Create(ctx context.Context, trainerID int64, kind Kind, slot Slot) (Entry, error)
	Update(ctx context.Context, entryID int64, slot Slot, now time.Time) (Entry, error)
	Cancel(ctx context.Context, entryID int64, now time.Time) (Entry, error)
	Delete(ctx context.Context, entryID int64, now time.Time) error
}

// Clock ให้เวลาปัจจุบัน ใช้ตัดสินว่าคลาสเริ่มแล้วหรือยังตาม ASM-07/Q-11 (test กำหนดเวลาได้).
type Clock interface {
	Now() time.Time
}

// systemClock คือนาฬิกาของเครื่อง.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// Handler คือ Gin handler ของตารางงาน ไม่ต้องล็อกอิน (CON-TECH-01, FR-ROST-01, ASM-06).
type Handler struct {
	store rosterStore
	clock Clock
}

// NewHandler สร้าง Handler; clock เป็น nil ใช้นาฬิกาของเครื่อง (CON-TECH-01).
func NewHandler(store rosterStore, clock Clock) *Handler {
	if clock == nil {
		clock = systemClock{}
	}
	return &Handler{store: store, clock: clock}
}

// messages คือข้อความภาษาไทยของแต่ละรหัสเหตุผล (ASM-02, ASM-04, ASM-07, ASM-08).
var messages = map[string]string{
	CodeInvalidRequest:         "ข้อมูลตารางงานไม่ถูกต้อง (วันต้องเป็น 1–7 เวลาเป็น HH:MM และเวลาเริ่มต้องก่อนเวลาจบในวันเดียวกัน)",
	CodeTrainerNotFound:        "ไม่พบเทรนเนอร์นี้",
	CodeEntryNotFound:          "ไม่พบรายการนี้",
	CodeOverlap:                "บันทึกไม่ได้ เพราะเวลาทับซ้อนกับรายการอื่นของเทรนเนอร์คนนี้",
	CodeTeachOutsideWork:       "บันทึกไม่ได้ เพราะเวลาสอนต้องอยู่ภายในเวลาเข้างาน",
	CodeBookedClassOutsideWork: "บันทึกไม่ได้ เพราะมีคลาสที่สมาชิกจองแล้วจะอยู่นอกเวลาเข้างาน",
	CodeUnavailable:            "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง",
}

// statusOf คือรหัส HTTP ของแต่ละรหัสเหตุผล (plan ข้อ 4).
var statusOf = map[string]int{
	CodeInvalidRequest:         http.StatusBadRequest,
	CodeTrainerNotFound:        http.StatusNotFound,
	CodeEntryNotFound:          http.StatusNotFound,
	CodeOverlap:                http.StatusConflict,
	CodeTeachOutsideWork:       http.StatusConflict,
	CodeBookedClassOutsideWork: http.StatusConflict,
}

// slotRequest คือส่วนวันและเวลาของ body (ASM-03).
type slotRequest struct {
	Weekday   *int   `json:"weekday"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// createRequest คือ body ของ POST /api/trainers/:id/roster (ASM-04).
type createRequest struct {
	Kind string `json:"kind"`
	slotRequest
}

// toSlot แปลง body เป็น Slot ที่ผ่าน Validate แล้ว (ASM-03).
func (r slotRequest) toSlot() (Slot, error) {
	if r.Weekday == nil {
		return Slot{}, ErrInvalidEntry
	}
	start, err := ParseTimeOfDay(r.StartTime)
	if err != nil {
		return Slot{}, err
	}
	end, err := ParseTimeOfDay(r.EndTime)
	if err != nil {
		return Slot{}, err
	}
	s := Slot{Weekday: *r.Weekday, Start: start, End: end}
	return s, s.Validate()
}

// listTrainers: GET /api/trainers (FR-ROST-01, ASM-09).
func (h *Handler) listTrainers(c *gin.Context) {
	trainers, err := h.store.ListTrainers(c.Request.Context())
	if err != nil {
		respondError(c, "list trainers", err)
		return
	}
	if trainers == nil {
		trainers = []Trainer{}
	}
	c.JSON(http.StatusOK, trainers)
}

// listEntries: GET /api/trainers/:id/roster เรียกดูตารางของเทรนเนอร์ (AC-ROST-01..04, plan D5).
func (h *Handler) listEntries(c *gin.Context) {
	trainerID, ok := pathID(c, "id")
	if !ok {
		return
	}
	entries, err := h.store.ListEntries(c.Request.Context(), trainerID)
	if err != nil {
		respondError(c, "list entries", err)
		return
	}
	if entries == nil {
		entries = []Entry{}
	}
	c.JSON(http.StatusOK, entries)
}

// createEntry: POST /api/trainers/:id/roster เพิ่มเวลาเข้างานหรือเวลาสอน (AC-ROST-01, AC-ROST-02; ASM-02, ASM-04).
func (h *Handler) createEntry(c *gin.Context) {
	trainerID, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondReason(c, CodeInvalidRequest)
		return
	}
	kind, err := ParseKind(req.Kind)
	if err != nil {
		respondReason(c, CodeInvalidRequest)
		return
	}
	slot, err := req.toSlot()
	if err != nil {
		respondReason(c, CodeInvalidRequest)
		return
	}
	entry, err := h.store.Create(c.Request.Context(), trainerID, kind, slot)
	if err != nil {
		respondError(c, "create entry", err)
		return
	}
	c.JSON(http.StatusCreated, entry)
}

// updateEntry: PATCH /api/roster/:entryId แก้วันและเวลาทีละรายการ (AC-ROST-03; ASM-05, ASM-07).
func (h *Handler) updateEntry(c *gin.Context) {
	entryID, ok := pathID(c, "entryId")
	if !ok {
		return
	}
	var req slotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondReason(c, CodeInvalidRequest)
		return
	}
	slot, err := req.toSlot()
	if err != nil {
		respondReason(c, CodeInvalidRequest)
		return
	}
	entry, err := h.store.Update(c.Request.Context(), entryID, slot, h.clock.Now())
	if err != nil {
		respondError(c, "update entry", err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

// cancelEntry: POST /api/roster/:entryId/cancel ยกเลิกแต่ยังเก็บไว้ (ASM-01, ASM-07).
func (h *Handler) cancelEntry(c *gin.Context) {
	entryID, ok := pathID(c, "entryId")
	if !ok {
		return
	}
	entry, err := h.store.Cancel(c.Request.Context(), entryID, h.clock.Now())
	if err != nil {
		respondError(c, "cancel entry", err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

// deleteEntry: DELETE /api/roster/:entryId ลบออกจากระบบ (ASM-01, ASM-07).
func (h *Handler) deleteEntry(c *gin.Context) {
	entryID, ok := pathID(c, "entryId")
	if !ok {
		return
	}
	if err := h.store.Delete(c.Request.Context(), entryID, h.clock.Now()); err != nil {
		respondError(c, "delete entry", err)
		return
	}
	c.Status(http.StatusNoContent)
}

// pathID อ่านรหัสจาก path ถ้าไม่ใช่จำนวนเต็มบวกตอบ 400.
func pathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		respondReason(c, CodeInvalidRequest)
		return 0, false
	}
	return id, true
}

// respondError ตอบรหัสเหตุผลของ error หรือ 503 ถ้าเป็นปัญหาของระบบ ไม่บันทึกอะไร (ASM-02, ASM-08).
func respondError(c *gin.Context, op string, err error) {
	if code, ok := ReasonCode(err); ok {
		respondReason(c, code)
		return
	}
	if !errors.Is(err, context.Canceled) {
		log.Printf("roster: %s failed: %v", op, err)
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": CodeUnavailable, "message": messages[CodeUnavailable]})
}

// respondReason ตอบ {"error": รหัส, "message": ข้อความไทย}.
func respondReason(c *gin.Context, code string) {
	c.JSON(statusOf[code], gin.H{"error": code, "message": messages[code]})
}
