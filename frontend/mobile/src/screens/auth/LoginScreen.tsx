/**
 * LoginScreen - Member Authentication
 * Spec: 001-auth-session
 * Requirements: FR-AUTH-01, UI-AUTH-01
 *
 * UI-AUTH-01: User can log in with persistent session
 * On app launch within session lifetime, should not require re-login
 */

import React, { useState } from 'react';
import {
  View,
  TextInput,
  TouchableOpacity,
  Text,
  ActivityIndicator,
  StyleSheet,
  Alert,
} from 'react-native';
import { login } from '@gym-mgmt/shared';

interface LoginScreenProps {
  onLoginSuccess?: () => void;
}

export const LoginScreen: React.FC<LoginScreenProps> = ({ onLoginSuccess }) => {
  const [studentId, setStudentId] = useState('STU001'); // Test credentials
  const [password, setPassword] = useState('demo123');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleLogin = async () => {
    // FR-AUTH-01: Login with persistent session
    // ASSUMPTION: student_id + password (verified with backend contract)
    if (!studentId.trim() || !password.trim()) {
      setError('Please enter both student ID and password');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // Call real backend API (POST /auth/login)
      // Backend returns: { access_token, refresh_token, user }
      // saveSession() stores tokens in localStorage for FR-AUTH-01 persistence
      const user = await login(studentId, password);

      Alert.alert('Success', `Logged in as ${user.name}`);
      onLoginSuccess?.();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed';
      setError(message);
      Alert.alert('Login Error', message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Gym Management</Text>
      <Text style={styles.subtitle}>Member Login</Text>

      {error && <Text style={styles.errorText}>{error}</Text>}

      <TextInput
        style={styles.input}
        placeholder="Student ID"
        placeholderTextColor="#999"
        value={studentId}
        onChangeText={setStudentId}
        editable={!isLoading}
        accessibilityLabel="Student ID input"
      />

      <TextInput
        style={styles.input}
        placeholder="Password"
        placeholderTextColor="#999"
        secureTextEntry
        value={password}
        onChangeText={setPassword}
        editable={!isLoading}
        accessibilityLabel="Password input"
      />

      <TouchableOpacity
        style={[styles.button, isLoading && styles.buttonDisabled]}
        onPress={handleLogin}
        disabled={isLoading}
        accessibilityRole="button"
        accessibilityLabel="Login button"
      >
        {isLoading ? (
          <ActivityIndicator color="#fff" />
        ) : (
          <Text style={styles.buttonText}>Login</Text>
        )}
      </TouchableOpacity>

      {/* TODO: Add registration link when FR-REG-01/02 implemented */}
      <Text style={styles.footerText}>
        • Session persists for 24 hours (Q2: TBD)
        {'\n'}• UI-AUTH-01: No re-login on app reopen
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    paddingHorizontal: 20,
    backgroundColor: '#f5f5f5',
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    marginBottom: 8,
    textAlign: 'center',
    color: '#000',
  },
  subtitle: {
    fontSize: 16,
    color: '#666',
    marginBottom: 24,
    textAlign: 'center',
  },
  input: {
    backgroundColor: '#fff',
    borderColor: '#ddd',
    borderWidth: 1,
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    marginBottom: 16,
    fontSize: 16,
  },
  button: {
    backgroundColor: '#007AFF',
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: 'center',
    marginTop: 8,
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  buttonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  errorText: {
    color: '#d32f2f',
    marginBottom: 12,
    textAlign: 'center',
  },
  footerText: {
    fontSize: 12,
    color: '#999',
    marginTop: 24,
    textAlign: 'center',
    fontStyle: 'italic',
  },
});
