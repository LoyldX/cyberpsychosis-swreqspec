# Gym Management System - Backend API

**Spec-Driven Development** | Phase 1: Foundation (MVP)

## Overview

Node.js + Express backend for gym management system with PostgreSQL + Redis. All endpoints are traced back to spec IDs (FR/NFR/DOM/CON/IF).

**Implemented Specs:**
- ✅ 001: Auth & Session (FR-AUTH-01, NFR-REL-01, CON-CACHE-01)
- ✅ 005: Class Booking (FR-BKG-01, FR-BKG-02, DOM-BKG-01, NFR-PERF-01)
- 🚧 004: Trainer Roster (FR-ROST-01) — skeleton
- 🚧 006: Inventory (FR-INV-01, NFR-DATA-01) — schema only
- 🚧 007: Occupancy (FR-BI-01, NFR-PERF-02) — schema only
- 🚧 002: Notifications (FR-NOTI-01) — skeleton
- 🚧 003: LINE Bot (FR-LINE-01, FR-LINE-02) — skeleton
- 🚧 008: Network Segmentation (NFR-SEC-01) — schema only

## Technology Stack

| Layer | Tech | From Spec? | Notes |
|-------|------|-----------|-------|
| Framework | Express.js 4.18 | Team choice | Fast, minimal, real-time ready |
| Database | PostgreSQL 16 | CON-DB-01 (Q4 assumed) | ACID support for inventory |
| ORM | Prisma 5.7 | Team choice | Type-safe, migrations |
| Session Store | Redis 7 | CON-CACHE-01 | Required for persistent login |
| Auth | JWT + Refresh Token | Q1 assumption | Stateless, mobile-friendly |
| Validation | Joi | Team choice | Runtime schema validation |

## Quick Start

### Prerequisites
- Node.js >= 18
- Docker & Docker Compose (for PostgreSQL + Redis)
- npm >= 9

### Setup

1. **Install dependencies:**
   ```bash
   cd backend
   npm install
   ```

2. **Create .env file (from .env.example):**
   ```bash
   cp .env.example .env
   # Edit .env with your secrets (dev defaults are safe for local)
   ```

3. **Start PostgreSQL + Redis:**
   ```bash
   docker-compose up -d
   # Wait ~10 seconds for containers to be ready
   ```

4. **Initialize database:**
   ```bash
   # Apply Prisma schema
   npm run db:push
   
   # (Optional) Seed with test data
   npm run db:seed
   ```

5. **Start development server:**
   ```bash
   npm run dev
   # Server runs on http://localhost:3000
   ```

### Verify Setup

```bash
# Health check
curl http://localhost:3000/health

# API endpoints
curl http://localhost:3000/api
```

## Development Workflow

### Database Changes

```bash
# Create a migration (after editing prisma/schema.prisma)
npm run db:migrate -- "add user table"

# Push schema changes to dev database
npm run db:push

# Open Prisma Studio (GUI for database)
npm run db:studio
```

### Testing

```bash
# Run unit tests
npm test

# Watch mode
npm test:watch

# Coverage report
npm test -- --coverage
```

### Code Quality

```bash
# Lint code
npm run lint

# Auto-fix issues
npm run lint:fix
```

## API Endpoints

### Auth (Spec 001)
- `POST /api/auth/login` — Login with student_id + password
- `POST /api/auth/refresh` — Refresh access token
- `POST /api/auth/logout` — Logout and revoke tokens
- `GET /api/auth/me` — Get current user profile

**Example Login:**
```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"student_id": "STU001", "password": "demo123"}'
```

Response:
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 900,
  "user": { "id": 1, "student_id": "STU001", "name": "John Doe" }
}
```

### Booking (Spec 005)
- `POST /api/booking` — Create booking (requires auth)
- `DELETE /api/booking/:id` — Cancel booking
- `GET /api/booking/member/:memberId` — List member's bookings

**Example Booking:**
```bash
curl -X POST http://localhost:3000/api/booking \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"class_id": 1}'
```

## Spec Assumptions & Open Questions

### Documented Assumptions (Marked in Code)

| Q# | Question | Assumption | Impact |
|----|----------|-----------|--------|
| Q1 | API contract (login format) | JWT payload = {user_id, student_id, iat, exp, type} | Auth controller |
| Q2 | Session TTL | 24 hours; max 0.1% auto-logout | Auth middleware, Redis config |
| Q3 | Trainer schedule model | Trainer shifts (recurring) + Class instances (one-off) | Booking validation |
| Q4 | Database choice | PostgreSQL (better ACID) | schema.prisma, docker-compose |

### Open Questions (Awaiting Team Input)

- **Latency target for NFR-PERF-01** — "Real-time" not defined; assuming < 500ms
- **Class capacity** — MVP assumes 1-to-1 trainer:student; extend later
- **Registration method** — Q6 undecided (QR vs online vs both)
- **Member data ownership** — Who creates member records? (Auth spec vs external system)

See `specs/*/spec.md` for full requirements and open questions.

## Project Structure

```
backend/
├── src/
│   ├── auth/                  # Spec 001: Auth & Session
│   │   ├── controller.js      # Login, refresh, logout, getCurrentUser
│   │   └── routes.js          # Endpoint definitions
│   ├── booking/               # Spec 005: Class Booking
│   │   ├── controller.js      # Create, cancel, list bookings + validation
│   │   └── routes.js          # Endpoint definitions
│   ├── roster/                # Spec 004: Trainer Roster (skeleton)
│   ├── inventory/             # Spec 006: Inventory (skeleton)
│   ├── occupancy/             # Spec 007: Occupancy Dashboard (skeleton)
│   ├── notification/          # Spec 002: Notifications (skeleton)
│   ├── linebot/               # Spec 003: LINE Bot (skeleton)
│   ├── network/               # Spec 008: Network Segmentation (skeleton)
│   ├── shared/
│   │   ├── redis.js           # Redis session store (CON-CACHE-01)
│   │   ├── jwt.js             # JWT token generation & verification
│   │   └── authMiddleware.js  # Express middleware for protected routes
│   └── index.js               # Main Express app, middleware setup
├── prisma/
│   └── schema.prisma          # Prisma ORM schema (source of truth for DB)
├── db/
│   ├── migrations/            # Prisma migrations
│   └── seeds/                 # Database seed scripts
├── tests/                     # Unit & integration tests
├── config/                    # Configuration files
├── .env.example               # Environment variables template
├── package.json               # Dependencies & scripts
└── README.md                  # This file
```

## Code Standards (Spec-Driven Development)

### Every Function Must Reference Spec IDs

```javascript
/**
 * Create a new booking
 * FR-BKG-01: Member books a class
 * FR-BKG-02: Block booking if trainer unavailable
 * DOM-BKG-01: Trainer availability constraint
 * NFR-PERF-01: Real-time validation (latency < 500ms)
 */
async function createBooking(req, res) { ... }
```

### All Assumptions Marked with Q#

```javascript
// ASSUMPTION Q2: Session TTL = 24 hours (from IMPLEMENTATION_PLAN.md)
const SESSION_TTL_SECONDS = SESSION_TTL_HOURS * 3600;

// ASSUMPTION Q3: Trainer shifts + Class instances (separate tables)
const available = await isTrainerAvailable(trainerId, scheduledAt);
```

### Test Names Use Acceptance Criteria (AC) IDs

```javascript
// test_AC_BKG_01_no_duplicate_bookings
it('should prevent duplicate bookings for same member+class', () => { ... });
```

## Database Schema

### Key Tables (from prisma/schema.prisma)

**Members** (Spec 001)
- student_id (unique, PK alternative)
- name, email, phone
- Created for auth & booking reference

**Classes & TrainerShifts** (Spec 005, 004)
- TrainerShift: recurring schedule (Mon-Fri 09:00-10:00)
- Class: scheduled instance (2024-10-05 09:00)
- Unique constraint: trainer + scheduled_at (prevent double-booking)

**Bookings** (Spec 005)
- member_id + class_id (unique to prevent duplicates)
- status: confirmed, cancelled, no_show
- Validates against trainer availability (DOM-BKG-01)

**InventoryItems & InventoryTransactions** (Spec 006)
- Items: equipment catalog with reorder_point
- Transactions: audit trail of all stock moves
- ACID-guaranteed via database transactions

**DoorAccessEvents & OccupancySnapshot** (Spec 007)
- Events: logs from IF-GATE-01 (door sensors)
- Snapshot: real-time occupancy count (updated on each event)

See `prisma/schema.prisma` for full schema with field descriptions.

## Troubleshooting

### PostgreSQL Connection Failed
```bash
# Check if container is running
docker ps | grep gym-db

# View logs
docker logs gym-db

# Restart
docker-compose restart postgres
```

### Redis Connection Failed
```bash
# Check if container is running
docker ps | grep gym-redis

# Test connection
redis-cli -h localhost -p 6379 ping
```

### Prisma Schema Sync Issues
```bash
# Regenerate Prisma client after schema changes
npm run db:push

# If that fails, nuke and recreate
docker-compose down -v  # Remove volumes
docker-compose up -d    # Restart fresh
npm run db:push
```

### Tests Failing
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Re-seed test database
npm run db:seed
npm test
```

## Next Steps (Roadmap)

### Phase 1 (Current)
- [x] Auth module with JWT + Redis sessions
- [x] Booking with trainer availability validation
- [ ] Implement all unit tests with AC IDs
- [ ] Integration tests for auth → booking flow

### Phase 2
- [ ] Roster CRUD (spec 004)
- [ ] Inventory transaction management (spec 006)
- [ ] Occupancy WebSocket (spec 007)

### Phase 3
- [ ] FCM notifications (spec 002)
- [ ] LINE Bot webhook handler (spec 003)

### Phase 4
- [ ] Network segmentation config API (spec 008)

## Contributing

1. **Read the spec** before coding (`specs/NNN-*/spec.md`)
2. **Reference spec IDs** in all function comments
3. **Document assumptions** with Q# numbers
4. **Write tests first** (TDD) — test names use AC IDs
5. **No feature creep** — only implement what's in spec, mark extensions as "team choice"

## References

- [Spec-Driven Development Guide](../AGENTS.md)
- [Implementation Plan](../IMPLEMENTATION_PLAN.md)
- [Full SRS](../docs/srs/srs.md)
- [Prisma Docs](https://www.prisma.io/docs)
- [Express Docs](https://expressjs.com/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)

---

**Status:** Phase 1 Foundation (MVP) | Last Updated: 2024-10-04
