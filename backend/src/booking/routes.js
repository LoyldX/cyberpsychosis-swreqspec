/**
 * Booking Routes
 * SPEC: 005-class-booking (FR-BKG-01, FR-BKG-02, DOM-BKG-01, NFR-PERF-01)
 * Endpoints: create booking, cancel, list member bookings
 */

const express = require('express');
const { authenticateToken } = require('../shared/authMiddleware');
const {
  createBooking,
  cancelBooking,
  getMemberBookings,
} = require('./controller');

const router = express.Router();

/**
 * POST /api/booking
 * FR-BKG-01: Create a new booking
 * FR-BKG-02: Block booking if trainer unavailable (DOM-BKG-01)
 * NFR-PERF-01: Real-time validation (latency < 500ms target)
 *
 * Request: { class_id }
 * Response: { id, member_id, class_id, status, created_at }
 */
router.post('/', authenticateToken, async (req, res) => {
  await createBooking(req, res);
});

/**
 * DELETE /api/booking/:id
 * FR-BKG-01: Cancel a booking
 *
 * Response: { id, status }
 */
router.delete('/:id', authenticateToken, async (req, res) => {
  await cancelBooking(req, res);
});

/**
 * GET /api/booking/member/:memberId
 * FR-BKG-01: List member's bookings
 *
 * Response: [{ id, class_id, status, created_at, class: {...} }]
 */
router.get('/member/:memberId', authenticateToken, async (req, res) => {
  await getMemberBookings(req, res);
});

module.exports = router;
