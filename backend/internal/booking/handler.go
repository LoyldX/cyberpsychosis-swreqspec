package booking

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gymapp/backend/internal/auth"
)

// bookingService คือสิ่งที่ Handler ต้องใช้จาก Service (FR-BKG-01, FR-BKG-02).
type bookingService interface {
	Classes(ctx context.Context) ([]Class, error)
	Book(ctx context.Context, classID, memberID int64) (Booking, error)
	Booking(ctx context.Context, bookingID, memberID int64) (Booking, error)
}

// Handler คือ Gin handler ของการจองคลาส (CON-TECH-01, FR-BKG-01, FR-BKG-02).
type Handler struct {
	svc bookingService
}

// NewHandler สร้าง Handler จาก Service (CON-TECH-01).
func NewHandler(svc bookingService) *Handler {
	return &Handler{svc: svc}
}

// messages คือข้อความภาษาไทยของแต่ละรหัสเหตุผล ให้สมาชิกรู้ว่าทำไมจองไม่สำเร็จ (ASM-05, ASM-07, UI-BKG-02 เป็น Candidate).
var messages = map[string]string{
	CodeInvalidRequest:  "ข้อมูลการจองไม่ถูกต้อง",
	CodeClassNotFound:   "ไม่พบคลาสนี้",
	CodeClassStarted:    "คลาสนี้เริ่มแล้ว จองไม่ได้",
	CodeAlreadyBooked:   "คุณจองคลาสนี้ไว้แล้ว",
	CodeClassFull:       "คลาสนี้เต็มแล้ว",
	CodeTrainerOffShift: "จองไม่สำเร็จ เพราะคลาสอยู่นอกเวลาปฏิบัติงานของเทรนเนอร์",
	CodeTrainerBusy:     "จองไม่สำเร็จ เพราะเทรนเนอร์ไม่ว่าง มีคิวสอนคลาสอื่นในช่วงเวลานี้",
	CodeBookingNotFound: "ไม่พบการจองนี้",
	CodeUnavailable:     "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง",
}

// statusOf คือรหัส HTTP ของแต่ละรหัสเหตุผล: ไม่พบ = 404, ไม่ผ่านเงื่อนไขการจอง = 409 (plan ข้อ 4).
var statusOf = map[string]int{
	CodeClassNotFound:   http.StatusNotFound,
	CodeBookingNotFound: http.StatusNotFound,
	CodeClassStarted:    http.StatusConflict,
	CodeAlreadyBooked:   http.StatusConflict,
	CodeClassFull:       http.StatusConflict,
	CodeTrainerOffShift: http.StatusConflict,
	CodeTrainerBusy:     http.StatusConflict,
}

// createBookingRequest คือ body ของ POST /api/bookings ผู้จองมาจาก session ไม่ใช่จาก body (ASM-06).
type createBookingRequest struct {
	ClassID int64 `json:"class_id"`
}

// listClasses: GET /api/classes คืนคลาสที่ยังไม่ถึงเวลาเริ่ม (FR-BKG-01, plan D4).
func (h *Handler) listClasses(c *gin.Context) {
	classes, err := h.svc.Classes(c.Request.Context())
	if err != nil {
		respondUnavailable(c, "list classes", err)
		return
	}
	if classes == nil {
		classes = []Class{}
	}
	c.JSON(http.StatusOK, classes)
}

// createBooking: POST /api/bookings จองคลาสให้สมาชิกใน session ตอบ 201 หรือรหัสเหตุผล หรือ 503 โดยไม่บันทึก
// (FR-BKG-01, FR-BKG-02, DOM-BKG-01, ASM-05, ASM-06, ASM-07).
func (h *Handler) createBooking(c *gin.Context) {
	memberID, ok := memberIDFrom(c)
	if !ok {
		return
	}
	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ClassID <= 0 {
		respondReason(c, http.StatusBadRequest, CodeInvalidRequest)
		return
	}
	b, err := h.svc.Book(c.Request.Context(), req.ClassID, memberID)
	if err != nil {
		respondBookingError(c, "create booking", err)
		return
	}
	c.JSON(http.StatusCreated, b)
}

// getBooking: GET /api/bookings/:id คืนการจองของสมาชิกใน session ทีละรายการ ของคนอื่นตอบ 404
// (FR-BKG-01 AC-BKG-01, plan D5).
func (h *Handler) getBooking(c *gin.Context) {
	memberID, ok := memberIDFrom(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondReason(c, http.StatusBadRequest, CodeInvalidRequest)
		return
	}
	b, err := h.svc.Booking(c.Request.Context(), id, memberID)
	if err != nil {
		respondBookingError(c, "get booking", err)
		return
	}
	c.JSON(http.StatusOK, b)
}

// memberIDFrom อ่านสมาชิกที่ RequireSession ของ 001 ยืนยันแล้ว ถ้าไม่มีตอบ 401 (precondition "สมาชิกล็อกอินแล้ว", ASM-06).
func memberIDFrom(c *gin.Context) (int64, bool) {
	session, ok := auth.MemberFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, false
	}
	return session.MemberID, true
}

// respondBookingError ตอบรหัสเหตุผลของ error การจอง หรือ 503 ถ้าเป็นปัญหาของระบบ (ASM-05, ASM-07).
func respondBookingError(c *gin.Context, op string, err error) {
	if code, ok := ReasonCode(err); ok {
		respondReason(c, statusOf[code], code)
		return
	}
	respondUnavailable(c, op, err)
}

// respondReason ตอบ {"error": รหัส, "message": ข้อความไทย} (ASM-05).
func respondReason(c *gin.Context, status int, code string) {
	c.JSON(status, gin.H{"error": code, "message": messages[code]})
}

// respondUnavailable ตอบ 503 เมื่อฐานข้อมูลหรือแหล่งข้อมูลกะไม่ตอบ ไม่ retry (ASM-07).
func respondUnavailable(c *gin.Context, op string, err error) {
	if !errors.Is(err, context.Canceled) {
		log.Printf("booking: %s failed: %v", op, err)
	}
	respondReason(c, http.StatusServiceUnavailable, CodeUnavailable)
}
