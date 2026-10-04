/**
 * Booking Controller
 * SPEC: 005-class-booking (FR-BKG-01, FR-BKG-02, DOM-BKG-01, NFR-PERF-01)
 * Handles class booking with real-time trainer availability validation
 */

const { PrismaClient } = require('@prisma/client');

const prisma = new PrismaClient();

/**
 * Check trainer availability for a specific time slot
 * DOM-BKG-01: Member cannot book if trainer has another class or is off-shift
 * NFR-PERF-01: Must be real-time (latency = TBD, assume < 500ms)
 *
 * ASSUMPTION Q3: Trainer availability = union of:
 *   1. Scheduled classes (no overlap allowed)
 *   2. Trainer shifts (class must be within shift hours)
 *
 * @param {number} trainerId - Trainer ID
 * @param {Date} scheduledAt - Requested class start time
 * @param {number} durationMins - Class duration in minutes (default 60)
 * @returns {Promise<boolean>} True if trainer is available
 */
async function isTrainerAvailable(trainerId, scheduledAt, durationMins = 60) {
  try {
    const classEndTime = new Date(scheduledAt.getTime() + durationMins * 60000);

    // Check 1: Trainer has no conflicting classes (FR-BKG-02)
    const conflictingClass = await prisma.class.findFirst({
      where: {
        trainerId,
        scheduledAt: {
          gte: scheduledAt,
          lt: classEndTime,
        },
      },
    });

    if (conflictingClass) {
      return false; // Trainer busy with another class
    }

    // Check 2: Trainer is on-shift (within trainer shift hours)
    const dayOfWeek = scheduledAt.getDay(); // 0=Sunday, 1=Monday, etc.
    const classHours = scheduledAt.getHours();
    const classMinutes = scheduledAt.getMinutes();
    const classTimeStr = `${String(classHours).padStart(2, '0')}:${String(classMinutes).padStart(2, '0')}`;

    const trainerShift = await prisma.trainerShift.findFirst({
      where: {
        trainerId,
        dayOfWeek,
        startTime: { lte: classTimeStr },
        endTime: { gte: classTimeStr },
      },
    });

    if (!trainerShift) {
      return false; // Trainer off-shift
    }

    return true; // Trainer available
  } catch (err) {
    console.error('Error checking trainer availability:', err);
    throw err;
  }
}

/**
 * Create a new booking
 * FR-BKG-01: Member books a class
 * FR-BKG-02: Block booking if trainer unavailable (DOM-BKG-01)
 * NFR-PERF-01: Real-time validation
 *
 * Request:
 *   POST /api/booking
 *   { class_id }
 *
 * Response:
 *   { id, member_id, class_id, status, created_at }
 *
 * @param {object} req - Express request (requires authenticateToken middleware)
 * @param {object} res - Express response
 */
async function createBooking(req, res) {
  try {
    if (!req.user) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Authentication required',
      });
    }

    const { class_id: classId } = req.body;
    const memberId = req.user.id;

    // Validation
    if (!classId) {
      return res.status(400).json({
        error: 'Bad Request',
        message: 'Missing class_id',
      });
    }

    // Fetch class details
    const classRecord = await prisma.class.findUnique({
      where: { id: classId },
      include: { trainer: true },
    });

    if (!classRecord) {
      return res.status(404).json({
        error: 'Not Found',
        message: 'Class not found',
      });
    }

    // Check if member already booked this class (prevent duplicates - FR-BKG-02)
    const existingBooking = await prisma.booking.findUnique({
      where: {
        memberId_classId: {
          memberId,
          classId,
        },
      },
    });

    if (existingBooking) {
      return res.status(409).json({
        error: 'Conflict',
        message: 'Member already booked this class',
      });
    }

    // Check trainer availability (DOM-BKG-01: FR-BKG-02)
    const available = await isTrainerAvailable(
      classRecord.trainerId,
      classRecord.scheduledAt,
      classRecord.durationMins,
    );

    if (!available) {
      return res.status(409).json({
        error: 'Conflict',
        message: 'Trainer is not available at this time',
        reason: 'trainer_unavailable', // For UI to display appropriate error
      });
    }

    // Check class capacity
    const bookingCount = await prisma.booking.count({
      where: {
        classId,
        status: 'confirmed',
      },
    });

    if (bookingCount >= classRecord.capacity) {
      return res.status(409).json({
        error: 'Conflict',
        message: 'Class is at capacity',
      });
    }

    // Create booking
    const booking = await prisma.booking.create({
      data: {
        memberId,
        classId,
        status: 'confirmed',
      },
    });

    res.status(201).json({
      id: booking.id,
      member_id: booking.memberId,
      class_id: booking.classId,
      status: booking.status,
      created_at: booking.createdAt,
    });
  } catch (err) {
    console.error('Create booking error:', err);
    res.status(500).json({
      error: 'Internal Server Error',
      message: 'Failed to create booking',
    });
  }
}

/**
 * Cancel a booking
 * FR-BKG-01: Member can cancel their booking
 *
 * Request:
 *   DELETE /api/booking/:id
 *
 * Response:
 *   { id, status }
 *
 * @param {object} req - Express request
 * @param {object} res - Express response
 */
async function cancelBooking(req, res) {
  try {
    if (!req.user) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Authentication required',
      });
    }

    const { id: bookingId } = req.params;
    const memberId = req.user.id;

    const booking = await prisma.booking.findUnique({
      where: { id: parseInt(bookingId, 10) },
    });

    if (!booking) {
      return res.status(404).json({
        error: 'Not Found',
        message: 'Booking not found',
      });
    }

    // Verify member owns this booking
    if (booking.memberId !== memberId) {
      return res.status(403).json({
        error: 'Forbidden',
        message: 'Cannot cancel another member\'s booking',
      });
    }

    // Update booking status
    const updated = await prisma.booking.update({
      where: { id: booking.id },
      data: {
        status: 'cancelled',
        cancelledAt: new Date(),
      },
    });

    res.json({
      id: updated.id,
      status: updated.status,
    });
  } catch (err) {
    console.error('Cancel booking error:', err);
    res.status(500).json({
      error: 'Internal Server Error',
      message: 'Failed to cancel booking',
    });
  }
}

/**
 * Get member's bookings
 * FR-BKG-01: Member views their bookings
 *
 * Request:
 *   GET /api/booking/member/:memberId
 *
 * Response:
 *   [{ id, class_id, status, created_at, class: { ... } }]
 *
 * @param {object} req - Express request
 * @param {object} res - Express response
 */
async function getMemberBookings(req, res) {
  try {
    if (!req.user) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Authentication required',
      });
    }

    const { memberId } = req.params;

    // Members can only see their own bookings
    if (parseInt(memberId, 10) !== req.user.id) {
      return res.status(403).json({
        error: 'Forbidden',
        message: 'Cannot view another member\'s bookings',
      });
    }

    const bookings = await prisma.booking.findMany({
      where: { memberId: parseInt(memberId, 10) },
      include: {
        class: {
          include: {
            trainer: {
              select: {
                id: true,
                name: true,
              },
            },
          },
        },
      },
      orderBy: { class: { scheduledAt: 'asc' } },
    });

    res.json(bookings);
  } catch (err) {
    console.error('Get member bookings error:', err);
    res.status(500).json({
      error: 'Internal Server Error',
      message: 'Failed to retrieve bookings',
    });
  }
}

module.exports = {
  isTrainerAvailable,
  createBooking,
  cancelBooking,
  getMemberBookings,
};
