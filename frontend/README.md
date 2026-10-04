# Frontend - Gym Management System

Monorepo containing React Native mobile app and React web portal for the Gym Management System.

## Project Structure

```
frontend/
├── shared/              # Shared types, hooks, API client
│   └── src/
│       ├── types/      # Spec-based types (User, Class, Booking, etc.)
│       ├── hooks/      # useAuth hook with FR-AUTH-01 session management
│       └── api/        # API client skeleton (todo: backend contract)
├── mobile/             # React Native member app
│   └── src/
│       ├── screens/    # Login, Booking, Notifications, Profile
│       ├── navigation/ # Navigation structure
│       └── components/ # Reusable components
└── web/                # React web staff portal + dashboard
    └── src/
        ├── pages/      # RosterPage, InventoryPage, OccupancyDashboard
        └── components/ # Reusable components
```

## Key Features Implemented

### Shared (`shared/`)
- **types/index.ts**: All spec-based types (User, Class, Booking, InventoryItem, OccupancyData)
- **hooks/useAuth.ts**: Session persistence with refresh token (FR-AUTH-01)
- **api/client.ts**: API client interface for backend integration

### Mobile App (React Native) - `mobile/`

#### LoginScreen
- **Spec**: 001-auth-session
- **Requirements**: FR-AUTH-01, UI-AUTH-01
- **Features**:
  - Student ID + Password login (ASSUMPTION: Q2 TBD on TTL)
  - Input validation
  - Loading state & error handling
  - 24-hour session TTL (TODO: Q2 - replace with backend config)

#### BookingCalendarScreen
- **Spec**: 005-class-booking
- **Requirements**: FR-BKG-01, FR-BKG-02, DOM-BKG-01, UI-BKG-01, UI-BKG-02
- **Features**:
  - Display available classes with trainer info
  - Block unavailable time slots (trainer busy/off-shift)
  - Real-time booking with error feedback
  - Mock trainer data (TODO: replace with API)
  - Accessibility labels for all interactive elements

### Web Portal (React) - `web/`

#### RosterPage
- **Spec**: 004-trainer-roster
- **Requirements**: FR-ROST-01, UI-ROST-01
- **Features**:
  - Display trainer work shifts and available slots
  - Differentiate available vs. off-shift times
  - Inline schedule editing (TODO: implement form)
  - Mock trainer data (TODO: replace with API)
  - Changes immediately affect booking availability

#### InventoryPage
- **Spec**: 006-inventory
- **Requirements**: FR-INV-01, FR-INV-02, FR-INV-03, DOM-INV-01, DOM-INV-02, UI-INV-01, UI-INV-02, UI-INV-03
- **Features**:
  - Search inventory by name/category (UI-INV-01)
  - Display current stock and reorder points
  - **Highlight low-stock items** in separate section (UI-INV-03)
  - Modal form for checkout/return (UI-INV-02)
  - Required fields: responsible_person, purpose, return_date (DOM-INV-01)
  - Mock inventory data (TODO: replace with API + ACID transactions)

#### OccupancyDashboard
- **Spec**: 007-occupancy-dashboard
- **Requirements**: FR-BI-01, NFR-PERF-02, DOM-CAP-01, UI-BI-01, UI-BI-02
- **Features**:
  - **Prominent user count display** (UI-BI-01)
  - Real-time capacity calculation (+1/-1 from door sensor)
  - **Auto-refresh without manual action** (UI-BI-02 with WebSocket placeholder)
  - Capacity status indicator (🟢 Comfortable / 🟡 Busy / 🔴 Crowded)
  - Activity log showing recent entries/exits
  - 2-minute occupancy trend chart
  - Mock data updates every 30s (TODO: NFR-PERF-02 - actual refresh rate TBD)
  - WebSocket placeholder ready for backend integration

## Tech Stack

**From Spec Constraints (Q4/CON-DB-01):**
- PostgreSQL (assumed; MySQL if Q4 chooses otherwise)
- Redis for session store (CON-CACHE-01)

**Team Choices (Not Constrained by Spec):**
- **Mobile**: React Native 0.73 + TypeScript
- **Web**: React 18 + TypeScript + Vite
- **Package Manager**: pnpm workspaces
- **Styling**: CSS modules + inline styles
- **HTTP**: Axios (ready to integrate)
- **Real-time**: Socket.io (placeholder for WebSocket)

## Setup Instructions

### Prerequisites
```bash
# Install Node.js 18+ and pnpm
npm install -g pnpm
```

### Install Dependencies
```bash
cd frontend
pnpm install
```

### Development

**Web Portal** (Vite dev server)
```bash
pnpm dev:web
# Opens at http://localhost:5173
```

**Mobile App** (React Native)
```bash
pnpm dev:mobile
# Follow React Native CLI instructions for iOS/Android
```

### Build

```bash
pnpm build:web    # React build for deployment
pnpm build:mobile # React Native bundle
```

## Open Questions & Assumptions

### Q2 - Session TTL & Dropout Rate
**Status**: TBD  
**Current Assumption**: 24 hours (hardcoded in shared/src/hooks/useAuth.ts)  
**Will Update**: When backend provides /config/session endpoint  
**Impact**: LoginScreen, useAuth hook, token refresh interval

### Q4 - Database Choice
**Status**: TBD (PostgreSQL assumed in IMPLEMENTATION_PLAN.md)  
**Impact**: Backend API schema; no direct impact on frontend (uses shared types)  
**Action**: Update API client once backend commits

### Member Auth Method
**Status**: TBD  
**Current Assumption**: student_id + password login  
**Impact**: LoginScreen form, User model  
**Action**: Adjust when backend authentication API is finalized

### NFR-PERF-02 - Dashboard Refresh Rate
**Status**: TBD  
**Current Simulation**: 30-second updates  
**Will Update**: When backend defines actual refresh rate  
**Impact**: OccupancyDashboard WebSocket subscription interval

## Integration Checklist (Waiting for Backend)

- [ ] **Auth API** (spec 001)
  - POST /auth/login → returns { access_token, refresh_token, user }
  - POST /auth/refresh → returns new tokens
  - Token structure & TTL from backend config

- [ ] **Booking API** (spec 005)
  - GET /classes?date=YYYY-MM-DD → returns Class[]
  - POST /bookings → creates booking, validates DOM-BKG-01
  - GET /trainer/:id/schedule → returns available slots

- [ ] **Roster API** (spec 004)
  - GET /trainers → returns Trainer[]
  - PATCH /trainer/:id/schedule → updates schedule (AC-004)

- [ ] **Inventory API** (spec 006)
  - GET /inventory → returns InventoryItem[]
  - POST /inventory/movements → records stock movement (NFR-DATA-01 ACID)
  - Response includes low_stock_alert flag

- [ ] **Occupancy API** (spec 007)
  - WebSocket: Subscribe to 'occupancy:update'
  - Webhook: POST /webhook/door-sensor → increments/decrements count

## Component IDs & Spec References

All files reference their spec requirements:

**Mobile Screens:**
- `LoginScreen.tsx` → FR-AUTH-01, UI-AUTH-01
- `BookingCalendarScreen.tsx` → FR-BKG-01, FR-BKG-02, DOM-BKG-01, UI-BKG-01, UI-BKG-02

**Web Pages:**
- `RosterPage.tsx` → FR-ROST-01, UI-ROST-01
- `InventoryPage.tsx` → FR-INV-01, FR-INV-02, FR-INV-03, DOM-INV-01, DOM-INV-02, UI-INV-01, UI-INV-02, UI-INV-03
- `OccupancyDashboard.tsx` → FR-BI-01, NFR-PERF-02, DOM-CAP-01, UI-BI-01, UI-BI-02

**Shared:**
- `useAuth.ts` → FR-AUTH-01, NFR-REL-01
- `types/index.ts` → All FR/NFR/DOM spec references

## Testing

Test files should be named with AC (Acceptance Criteria) IDs:
- `LoginScreen.test.tsx` should contain: `test_AC_AUTH_01_persistent_session`
- `BookingCalendarScreen.test.tsx` should contain: `test_AC_BKG_01_no_duplicate_bookings`

(Tests not yet implemented - waiting for Jest/React Native Testing Library setup)

## Next Steps

1. **Backend agent** starts API development in parallel
2. **Replace mock data** with real API calls as endpoints become available
3. **Implement WebSocket** for real-time occupancy updates
4. **Add E2E tests** with spec ID references
5. **Coordinate API contract** at checkpoints (Week 1-4 per IMPLEMENTATION_PLAN.md)

## Notes

- All code uses `// FR-*`, `// UI-*`, `// AC-*` comments linking to spec
- `// TODO: Q#` comments mark areas awaiting open question resolution
- `// ASSUMPTION:` comments document provisional decisions
- Mock data allows development to proceed while backend is ready
