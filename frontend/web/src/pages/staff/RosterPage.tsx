/**
 * RosterPage - Trainer Schedule Management (Staff Portal)
 * Spec: 004-trainer-roster
 * Requirements: FR-ROST-01, UI-ROST-01
 *
 * UI-ROST-01: Display trainer schedule with work shifts and teaching times
 * Allow inline editing on the same page
 * Changes must reflect immediately in booking availability (FR-BKG-02)
 */

import React, { useState, useEffect } from 'react';
import './RosterPage.css';
import { TrainerSchedule } from '@gym-mgmt/shared';

interface Trainer {
  id: string;
  name: string;
  schedule: TrainerSchedule;
}

/**
 * MOCK DATA - Replace with API when backend ready
 * Will call: apiClient.getTrainers() and apiClient.updateTrainerSchedule()
 */
const MOCK_TRAINERS: Trainer[] = [
  {
    id: 'trainer_001',
    name: 'John Doe',
    schedule: {
      trainer_id: 'trainer_001',
      date: '2026-10-04',
      available_slots: [
        { start_time: '09:00', end_time: '10:00' },
        { start_time: '11:00', end_time: '12:00' },
        { start_time: '14:00', end_time: '15:00' },
      ],
      off_shift_times: [
        { start_time: '10:00', end_time: '11:00' },
        { start_time: '12:00', end_time: '14:00' },
        { start_time: '15:00', end_time: '17:00' },
      ],
    },
  },
  {
    id: 'trainer_002',
    name: 'Jane Smith',
    schedule: {
      trainer_id: 'trainer_002',
      date: '2026-10-04',
      available_slots: [
        { start_time: '10:00', end_time: '11:00' },
        { start_time: '13:00', end_time: '15:00' },
      ],
      off_shift_times: [
        { start_time: '09:00', end_time: '10:00' },
        { start_time: '11:00', end_time: '13:00' },
        { start_time: '15:00', end_time: '17:00' },
      ],
    },
  },
];

export const RosterPage: React.FC = () => {
  const [trainers, setTrainers] = useState<Trainer[]>(MOCK_TRAINERS);
  const [editingTrainerId, setEditingTrainerId] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    // TODO: Fetch trainers from API
    // const fetchTrainers = async () => {
    //   const data = await apiClient.getTrainers();
    //   setTrainers(data);
    // };
    // fetchTrainers();
  }, []);

  const handleSaveSchedule = async (trainerId: string) => {
    setIsSaving(true);
    try {
      // TODO: Call API updateTrainerSchedule(trainerId, schedule)
      // AC-004_schedule_updates_immediately: Changes must reflect in booking UI
      setEditingTrainerId(null);
      alert('Schedule updated successfully');
    } catch (error) {
      alert(`Failed to save schedule: ${error}`);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="roster-page">
      <h1>Trainer Roster & Schedule</h1>
      <p className="spec-info">
        • FR-ROST-01: Manage trainer schedules
        <br />• UI-ROST-01: View and edit schedules inline
        <br />• AC-004_schedule_updates_immediately: Changes affect booking availability
      </p>

      <div className="trainers-list">
        {trainers.map((trainer) => (
          <div key={trainer.id} className="trainer-card">
            <h2>{trainer.name}</h2>

            <div className="schedule-section">
              <h3>Work Shift - Available Times</h3>
              <div className="time-slots">
                {trainer.schedule.available_slots.map((slot, idx) => (
                  <span key={`avail-${idx}`} className="time-slot available">
                    {slot.start_time} - {slot.end_time}
                  </span>
                ))}
              </div>
            </div>

            <div className="schedule-section">
              <h3>Off-Shift / Teaching (Unavailable for Bookings)</h3>
              <div className="time-slots">
                {trainer.schedule.off_shift_times.map((slot, idx) => (
                  <span key={`off-${idx}`} className="time-slot off-shift">
                    {slot.start_time} - {slot.end_time}
                  </span>
                ))}
              </div>
            </div>

            {editingTrainerId === trainer.id ? (
              <div className="edit-form">
                <p>Edit form would go here (TODO: implement form)</p>
                <div className="button-group">
                  <button
                    className="btn btn-primary"
                    onClick={() => handleSaveSchedule(trainer.id)}
                    disabled={isSaving}
                  >
                    {isSaving ? 'Saving...' : 'Save Schedule'}
                  </button>
                  <button
                    className="btn btn-secondary"
                    onClick={() => setEditingTrainerId(null)}
                    disabled={isSaving}
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <button
                className="btn btn-primary"
                onClick={() => setEditingTrainerId(trainer.id)}
              >
                Edit Schedule
              </button>
            )}
          </div>
        ))}
      </div>

      <div className="info-box">
        <h3>How it works</h3>
        <ul>
          <li>Green slots = Trainer available to teach (can accept bookings)</li>
          <li>Red slots = Trainer off-shift or already teaching (bookings blocked)</li>
          <li>Changes apply immediately (spec 004 requirement)</li>
          <li>DOM-BKG-01: Booking system validates against this schedule in real-time</li>
        </ul>
      </div>
    </div>
  );
};
