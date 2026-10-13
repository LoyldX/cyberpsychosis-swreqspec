/**
 * Gym Management System - Backend API
 * Spec-Driven Development: All routes/controllers trace to spec IDs (FR/NFR/DOM/CON/IF)
 */

require('express-async-errors');
require('dotenv').config();

const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const morgan = require('morgan');

// Route imports
const authRoutes = require('./auth/routes');
const bookingRoutes = require('./booking/routes');
// const rosterRoutes = require('./roster/routes');        // Phase 2
// const inventoryRoutes = require('./inventory/routes');  // Phase 2
// const occupancyRoutes = require('./occupancy/routes');  // Phase 2
// const notificationRoutes = require('./notification/routes'); // Phase 2
// const linebotRoutes = require('./linebot/routes');      // Phase 2
// const networkRoutes = require('./network/routes');      // Phase 2

const app = express();

// ============================================================================
// MIDDLEWARE
// ============================================================================

// Security
app.use(helmet());

// CORS (enable for frontend agents to connect)
app.use(cors({
  origin: process.env.CORS_ORIGIN?.split(',') || ['http://localhost:3001'],
  credentials: true,
}));

// Request logging
app.use(morgan('combined'));

// Body parsing
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ limit: '10mb', extended: true }));

// ============================================================================
// ROUTES
// ============================================================================

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    version: '0.1.0',
  });
});

// API Documentation endpoint
app.get('/api', (req, res) => {
  res.json({
    name: 'Gym Management System API',
    version: '0.1.0',
    status: 'Spec-Driven Development (Phase 1: Foundation)',
    endpoints: {
      auth: '/api/auth (FR-AUTH-01, NFR-REL-01, CON-CACHE-01)',
      booking: '/api/booking (FR-BKG-01, FR-BKG-02, DOM-BKG-01, NFR-PERF-01)',
      roster: '/api/roster (FR-ROST-01)',
      inventory: '/api/inventory (FR-INV-01, FR-INV-02, FR-INV-03, NFR-DATA-01)',
      occupancy: '/api/occupancy (FR-BI-01, NFR-PERF-02)',
      notification: '/api/notification (FR-NOTI-01)',
      linebot: '/api/linebot (FR-LINE-01, FR-LINE-02)',
      network: '/api/network (NFR-SEC-01)',
    },
  });
});

// Auth Routes (Spec 001: FR-AUTH-01, NFR-REL-01, CON-CACHE-01)
app.use('/api/auth', authRoutes);

// Booking Routes (Spec 005: FR-BKG-01, FR-BKG-02, DOM-BKG-01, NFR-PERF-01)
app.use('/api/booking', bookingRoutes);

// Roster Routes (Spec 004: FR-ROST-01) - Phase 2
// app.use('/api/roster', rosterRoutes);

// Inventory Routes (Spec 006: FR-INV-01, FR-INV-02, FR-INV-03, NFR-DATA-01) - Phase 2
// app.use('/api/inventory', inventoryRoutes);

// Occupancy Routes (Spec 007: FR-BI-01, NFR-PERF-02) - Phase 2
// app.use('/api/occupancy', occupancyRoutes);

// Notification Routes (Spec 002: FR-NOTI-01) - Phase 2
// app.use('/api/notification', notificationRoutes);

// LINE Bot Routes (Spec 003: FR-LINE-01, FR-LINE-02) - Phase 2
// app.use('/api/linebot', linebotRoutes);

// Network Routes (Spec 008: NFR-SEC-01) - Phase 2
// app.use('/api/network', networkRoutes);

// ============================================================================
// ERROR HANDLING
// ============================================================================

// 404 handler
app.use((req, res) => {
  res.status(404).json({
    error: 'Not Found',
    path: req.path,
    method: req.method,
  });
});

// Global error handler
app.use((err, req, res, next) => {
  console.error('Error:', err);

  const status = err.status || err.statusCode || 500;
  const message = err.message || 'Internal Server Error';

  res.status(status).json({
    error: message,
    status,
    timestamp: new Date().toISOString(),
    ...(process.env.NODE_ENV === 'development' && { stack: err.stack }),
  });
});

// ============================================================================
// SERVER STARTUP
// ============================================================================

const PORT = process.env.PORT || 3000;

const server = app.listen(PORT, process.env.HOST || '0.0.0.0', () => {
  console.log(`
    ╔═══════════════════════════════════════════════════════════╗
    ║   Gym Management System - Backend API                    ║
    ║   Spec-Driven Development (Phase 1: Foundation)          ║
    ╠═══════════════════════════════════════════════════════════╣
    ║   Server running at http://${process.env.HOST || 'localhost'}:${PORT}        ║
    ║   Environment: ${(process.env.NODE_ENV || 'development').toUpperCase().padEnd(40)}║
    ║   Database: ${process.env.DATABASE_URL?.split('@')[1]?.split('/')[0] || 'Not configured'}${' '.repeat(25)}║
    ║   Redis: ${process.env.REDIS_URL || 'Not configured'}${' '.repeat(28)}║
    ╚═══════════════════════════════════════════════════════════╝
  `);
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down gracefully...');
  server.close(() => {
    console.log('Server closed');
    process.exit(0);
  });
});

module.exports = app;
