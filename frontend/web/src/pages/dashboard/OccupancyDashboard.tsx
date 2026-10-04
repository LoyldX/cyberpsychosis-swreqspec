/**
 * OccupancyDashboard - Real-Time User Count Dashboard
 * Spec: 007-occupancy-dashboard
 * Requirements: FR-BI-01, NFR-PERF-02, DOM-CAP-01, UI-BI-01, UI-BI-02
 *
 * UI-BI-01: Display current gym occupancy prominently
 * UI-BI-02: Auto-refresh without manual action (WebSocket-based)
 *
 * DOM-CAP-01: Calculate occupancy from entry/exit signals
 * IF-GATE-01: Webhook from door sensor
 */

import React, { useState, useEffect } from 'react';
import './OccupancyDashboard.css';
import { OccupancyData } from '@gym-mgmt/shared';

/**
 * MOCK DATA - Replace with WebSocket when backend ready
 * Will subscribe to: Socket.io channel 'occupancy:update'
 * Backend sends: { current_count, timestamp } on door sensor webhook
 */
const MOCK_INITIAL_OCCUPANCY: OccupancyData = {
  current_count: 14,
  timestamp: new Date().toISOString(),
};

export const OccupancyDashboard: React.FC = () => {
  const [occupancy, setOccupancy] = useState<OccupancyData>(MOCK_INITIAL_OCCUPANCY);
  const [isConnected, setIsConnected] = useState(false);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());
  const [history, setHistory] = useState<OccupancyData[]>([MOCK_INITIAL_OCCUPANCY]);

  useEffect(() => {
    // TODO: Connect to WebSocket when backend ready
    // Simulate updates for now
    const interval = setInterval(() => {
      const randomDelta = Math.random() > 0.5 ? 1 : -1;
      const newCount = Math.max(0, Math.min(50, occupancy.current_count + randomDelta));

      const newData: OccupancyData = {
        current_count: newCount,
        timestamp: new Date().toISOString(),
      };

      setOccupancy(newData);
      setLastUpdate(new Date());
      setHistory((prev) => [...prev.slice(-119), newData]); // Keep last 2 minutes
    }, 30000); // Update every 30 seconds (NFR-PERF-02: TBD refresh rate)

    // Mock WebSocket connection
    setIsConnected(true);

    return () => clearInterval(interval);
  }, [occupancy.current_count]);

  const maxCapacity = 50; // Gym max capacity
  const occupancyPercent = (occupancy.current_count / maxCapacity) * 100;
  const statusColor =
    occupancyPercent > 80 ? '#d32f2f' : occupancyPercent > 50 ? '#ff9800' : '#4caf50';

  return (
    <div className="dashboard">
      <h1>Gym Occupancy Dashboard</h1>
      <p className="spec-info">
        • FR-BI-01: Display current gym occupancy
        <br />• UI-BI-01: User count is primary metric on dashboard
        <br />• UI-BI-02: Auto-refresh via WebSocket (NFR-PERF-02: refresh rate TBD)
        <br />• DOM-CAP-01: Calculated from door sensor (+1 entry, -1 exit)
        <br />• IF-GATE-01: Webhook from turnstile/access control system
      </p>

      {/* Main Occupancy Card */}
      <div className="dashboard-grid">
        <div className="occupancy-card">
          <h2>Current Occupancy</h2>
          <div className="occupancy-number" style={{ color: statusColor }}>
            {occupancy.current_count}
          </div>
          <div className="occupancy-label">/ {maxCapacity} Capacity</div>

          <div className="progress-bar">
            <div
              className="progress-fill"
              style={{
                width: `${occupancyPercent}%`,
                backgroundColor: statusColor,
              }}
            />
          </div>

          <div className="occupancy-percent">{occupancyPercent.toFixed(0)}% Full</div>

          <div className="connection-status">
            <span
              className={`status-dot ${isConnected ? 'connected' : 'disconnected'}`}
            />
            {isConnected ? 'Live' : 'Offline'}
          </div>

          <div className="last-update">
            Last update: {lastUpdate.toLocaleTimeString()}
          </div>
        </div>

        {/* Activity Log */}
        <div className="activity-card">
          <h2>Recent Activity</h2>
          <p className="activity-hint">
            (Live updates would show entries/exits here when WebSocket is connected)
          </p>
          <div className="activity-log">
            <div className="activity-item">
              <span className="time">10:45 AM</span>
              <span className="action entry">ENTRY</span>
              <span className="count">+1 → 14</span>
            </div>
            <div className="activity-item">
              <span className="time">10:40 AM</span>
              <span className="action exit">EXIT</span>
              <span className="count">-1 → 13</span>
            </div>
            <div className="activity-item">
              <span className="time">10:35 AM</span>
              <span className="action entry">ENTRY</span>
              <span className="count">+1 → 14</span>
            </div>
          </div>
        </div>
      </div>

      {/* Capacity Timeline Chart (Simple) */}
      <div className="timeline-card">
        <h2>Occupancy Trend (Last 2 Minutes)</h2>
        <div className="chart-container">
          <div className="chart">
            {history.map((dataPoint, idx) => (
              <div
                key={idx}
                className="chart-bar"
                style={{
                  height: `${(dataPoint.current_count / maxCapacity) * 100}%`,
                  backgroundColor: `hsl(${(dataPoint.current_count / maxCapacity) * 120}, 70%, 50%)`,
                }}
                title={`${dataPoint.current_count} people at ${new Date(dataPoint.timestamp).toLocaleTimeString()}`}
              />
            ))}
          </div>
          <div className="y-axis">
            <div>{maxCapacity}</div>
            <div>{maxCapacity / 2}</div>
            <div>0</div>
          </div>
        </div>
      </div>

      {/* Status Cards */}
      <div className="status-cards">
        <div className="status-card">
          <h3>Capacity Status</h3>
          <p className={occupancyPercent > 80 ? 'critical' : occupancyPercent > 50 ? 'warning' : 'normal'}>
            {occupancyPercent > 80
              ? '🔴 CROWDED'
              : occupancyPercent > 50
                ? '🟡 MODERATELY BUSY'
                : '🟢 COMFORTABLE'}
          </p>
        </div>

        <div className="status-card">
          <h3>Available Capacity</h3>
          <p className="available-count">{maxCapacity - occupancy.current_count}</p>
          <p className="available-label">Spots available</p>
        </div>

        <div className="status-card">
          <h3>WebSocket Connection</h3>
          <p className={isConnected ? 'connected-text' : 'disconnected-text'}>
            {isConnected ? '✓ Connected' : '✗ Disconnected'}
          </p>
        </div>
      </div>

      {/* Implementation Notes */}
      <div className="notes-box">
        <h3>Implementation Notes</h3>
        <ul>
          <li>
            <strong>NFR-PERF-02 Refresh Rate:</strong> Currently simulated at 30-second
            intervals. When backend ready, use actual door sensor webhook and WebSocket
            subscription.
          </li>
          <li>
            <strong>DOM-CAP-01:</strong> Occupancy = sum of all +1 entries and -1 exits from
            IF-GATE-01 webhook
          </li>
          <li>
            <strong>WebSocket TODO:</strong> Replace simulated updates with actual Socket.io
            subscription to &apos;occupancy:update&apos;
          </li>
          <li>
            <strong>Auth:</strong> Manager/staff only access (implement via ProtectedRoute)
          </li>
        </ul>
      </div>
    </div>
  );
};
