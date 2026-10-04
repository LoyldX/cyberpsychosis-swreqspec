/**
 * BookingCalendarScreen - Class Booking Interface
 * Spec: 005-class-booking
 * Requirements: FR-BKG-01, FR-BKG-02, DOM-BKG-01, UI-BKG-01, UI-BKG-02
 *
 * UI-BKG-01: Display available classes and block unavailable time slots
 * UI-BKG-02: Show error feedback when booking fails (trainer unavailable)
 *
 * DOM-BKG-01: System prevents booking during trainer off-shift or overlapping bookings
 * NFR-PERF-01: Real-time concurrency check (latency TBD)
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  ScrollView,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Alert,
  FlatList,
} from 'react-native';
import { Booking, getStoredSession } from '@gym-mgmt/shared';
import { apiClient } from '@gym-mgmt/shared';

interface BookingCalendarScreenProps {
  onBookingSuccess?: (booking: Booking) => void;
}

interface ClassDisplay {
  id: string;
  trainer_id: string;
  trainer_name: string;
  start_time: string;
  end_time: string;
  class_type: string;
  capacity: number;
  booked_count: number;
}

/**
 * MOCK DATA - Available classes for demo
 * TODO: Fetch from GET /api/classes when backend endpoint available
 */
const MOCK_CLASSES: ClassDisplay[] = [
  {
    id: 'class_001',
    trainer_id: 'trainer_001',
    trainer_name: 'Alex Coach',
    start_time: '2026-10-07T09:00:00Z',
    end_time: '2026-10-07T10:00:00Z',
    class_type: 'CrossFit',
    capacity: 1, // 1:1 trainer:student (MVP assumption)
    booked_count: 0,
  },
  {
    id: 'class_002',
    trainer_id: 'trainer_001',
    trainer_name: 'Alex Coach',
    start_time: '2026-10-07T10:00:00Z',
    end_time: '2026-10-07T11:00:00Z',
    class_type: 'CrossFit',
    capacity: 1,
    booked_count: 1, // BLOCKED - trainer busy (FR-BKG-02)
  },
  {
    id: 'class_003',
    trainer_id: 'trainer_002',
    trainer_name: 'Sarah Trainer',
    start_time: '2026-10-07T14:00:00Z',
    end_time: '2026-10-07T15:00:00Z',
    class_type: 'Yoga',
    capacity: 1,
    booked_count: 0,
  },
];

export const BookingCalendarScreen: React.FC<BookingCalendarScreenProps> = ({
  onBookingSuccess,
}) => {
  const [classes, setClasses] = useState<ClassDisplay[]>(MOCK_CLASSES);
  const [selectedClassId, setSelectedClassId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingBookings, setIsLoadingBookings] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    // FR-BKG-01: Fetch user's existing bookings on load
    const fetchBookings = async () => {
      try {
        const session = getStoredSession();
        if (!session.user) {
          setIsLoadingBookings(false);
          return;
        }

        // GET /api/booking/member/:memberId
        const bookings = await apiClient.getMyBookings(session.user.id);

        // Mark classes that user has already booked
        setClasses(prev => prev.map(cls => {
          const hasUserBooked = bookings.some(
            b => b.class_id === cls.id && b.status === 'confirmed'
          );
          return {
            ...cls,
            booked_count: hasUserBooked ? cls.capacity : cls.booked_count,
          };
        }));
      } catch (err) {
        console.warn('Failed to fetch bookings:', err);
        // Continue with mock classes on error
      } finally {
        setIsLoadingBookings(false);
      }
    };

    fetchBookings();
  }, []);

  const isClassAvailable = (cls: ClassDisplay): boolean => {
    // DOM-BKG-01: Check if slots are available
    return cls.booked_count < cls.capacity;
  };

  const handleBookClass = async (classId: string) => {
    const selectedClass = classes.find((c) => c.id === classId);
    if (!selectedClass || !isClassAvailable(selectedClass)) {
      // UI-BKG-02: Show error when trainer unavailable (FR-BKG-02)
      Alert.alert(
        'Booking Failed',
        'This time slot is no longer available. The trainer is unavailable or this class is full.',
        [{ text: 'OK' }],
      );
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // FR-BKG-01: Call real booking API
      // POST /api/booking with { class_id }
      // FR-BKG-02: Backend validates DOM-BKG-01 (trainer availability)
      const booking = await apiClient.bookClass(classId);

      // AC-005_no_duplicate_bookings: System prevents duplicate bookings
      setSuccessMessage(`Booked ${selectedClass.class_type} with ${selectedClass.trainer_name}`);
      setSelectedClassId(null);
      onBookingSuccess?.(booking);

      Alert.alert('Success', `Booking confirmed for ${selectedClass.trainer_name}'s class`);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Booking failed';
      setError(message);
      // FR-BKG-02: Error feedback on failed booking
      Alert.alert('Booking Error', message);
    } finally {
      setIsLoading(false);
    }
  };

  const renderClassItem = ({ item }: { item: Class }) => {
    const isAvailable = isClassAvailable(item);
    const startTime = new Date(item.start_time).toLocaleTimeString();
    const endTime = new Date(item.end_time).toLocaleTimeString();

    return (
      <TouchableOpacity
        style={[
          styles.classCard,
          !isAvailable && styles.classCardDisabled,
          selectedClassId === item.id && styles.classCardSelected,
        ]}
        onPress={() => {
          if (isAvailable) {
            setSelectedClassId(item.id);
          }
        }}
        disabled={!isAvailable}
        accessibilityRole="button"
        accessibilityLabel={`${item.class_type} with ${item.trainer_name} at ${startTime}`}
        accessibilityHint={
          isAvailable
            ? 'Double tap to book this class'
            : 'This class is no longer available'
        }
      >
        <View style={styles.classInfo}>
          <Text style={styles.className}>{item.class_type}</Text>
          <Text style={styles.trainerName}>Trainer: {item.trainer_name}</Text>
          <Text style={styles.time}>
            {startTime} - {endTime}
          </Text>
          <Text style={styles.capacity}>
            {item.booked_count}/{item.capacity} booked
          </Text>
        </View>

        {!isAvailable && (
          <View style={styles.blockedBadge}>
            <Text style={styles.blockedText}>UNAVAILABLE</Text>
          </View>
        )}
      </TouchableOpacity>
    );
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Book a Class</Text>

      {error && <Text style={styles.errorText}>{error}</Text>}
      {successMessage && <Text style={styles.successText}>{successMessage}</Text>}

      <FlatList
        data={classes}
        renderItem={renderClassItem}
        keyExtractor={(item) => item.id}
        scrollEnabled={false}
        contentContainerStyle={styles.listContainer}
      />

      {selectedClassId && (
        <TouchableOpacity
          style={[styles.bookButton, isLoading && styles.bookButtonDisabled]}
          onPress={() => handleBookClass(selectedClassId)}
          disabled={isLoading}
        >
          {isLoading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={styles.bookButtonText}>Confirm Booking</Text>
          )}
        </TouchableOpacity>
      )}

      <View style={styles.infoBox}>
        <Text style={styles.infoTitle}>About Bookings</Text>
        <Text style={styles.infoText}>
          • FR-BKG-01: Select available classes to book{'\n'}
          • DOM-BKG-01: Blocked times = trainer unavailable or off-shift{'\n'}
          • NFR-PERF-01: Real-time availability (latency TBD)
        </Text>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
    paddingHorizontal: 16,
    paddingTop: 16,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 16,
    color: '#000',
  },
  listContainer: {
    paddingBottom: 16,
  },
  classCard: {
    backgroundColor: '#fff',
    borderRadius: 8,
    padding: 16,
    marginBottom: 12,
    borderWidth: 2,
    borderColor: '#e0e0e0',
  },
  classCardDisabled: {
    opacity: 0.5,
    backgroundColor: '#f9f9f9',
  },
  classCardSelected: {
    borderColor: '#007AFF',
    backgroundColor: '#f0f8ff',
  },
  classInfo: {
    marginBottom: 8,
  },
  className: {
    fontSize: 18,
    fontWeight: '600',
    color: '#000',
    marginBottom: 4,
  },
  trainerName: {
    fontSize: 14,
    color: '#666',
    marginBottom: 4,
  },
  time: {
    fontSize: 14,
    color: '#666',
    marginBottom: 4,
  },
  capacity: {
    fontSize: 12,
    color: '#999',
  },
  blockedBadge: {
    backgroundColor: '#ffebee',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 4,
    alignSelf: 'flex-start',
  },
  blockedText: {
    color: '#d32f2f',
    fontSize: 12,
    fontWeight: '600',
  },
  bookButton: {
    backgroundColor: '#007AFF',
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: 'center',
    marginBottom: 16,
  },
  bookButtonDisabled: {
    opacity: 0.6,
  },
  bookButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  errorText: {
    color: '#d32f2f',
    marginBottom: 12,
    textAlign: 'center',
  },
  successText: {
    color: '#388e3c',
    marginBottom: 12,
    textAlign: 'center',
  },
  infoBox: {
    backgroundColor: '#e3f2fd',
    borderRadius: 8,
    padding: 12,
    marginBottom: 16,
  },
  infoTitle: {
    fontWeight: '600',
    color: '#1976d2',
    marginBottom: 8,
  },
  infoText: {
    fontSize: 12,
    color: '#0d47a1',
    lineHeight: 18,
  },
});
