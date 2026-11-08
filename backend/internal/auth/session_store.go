// Package auth: session_store.go เก็บ Session (Refresh Token) ใน Redis (CON-CACHE-01).
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// DefaultSessionTTL คืออายุ Refresh Token 7 วันนับจากเวลาเข้าสู่ระบบ ไม่ต่ออายุ (ASM-02).
const DefaultSessionTTL = 7 * 24 * time.Hour

// sessionKeyPrefix คือ prefix ของ key ใน Redis ตาม plan.md §3 (CON-CACHE-01, AC-AUTH-05).
const sessionKeyPrefix = "session:"

// tokenBytes คือจำนวนไบต์สุ่มของ token ก่อน encode (FR-AUTH-01: opaque token).
const tokenBytes = 32

// ErrSessionNotFound แปลว่าไม่พบ Session (ไม่มี/หมดอายุ/ข้อมูลเสีย) ใช้กับ AC-AUTH-03, 04 (FR-AUTH-01).
var ErrSessionNotFound = errors.New("auth: session not found")

// SessionData คือข้อมูลที่เก็บใต้ key ของ Session (FR-AUTH-01, plan.md §3).
type SessionData struct {
	MemberID  int64  `json:"member_id"`
	StudentID string `json:"student_id"`
}

// SessionStore เก็บ/อ่าน Session ใน Redis (CON-CACHE-01); ttl กำหนดจากภายนอกเพื่อให้ test ใช้ค่าสั้นได้.
type SessionStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewSessionStore สร้าง store; production ส่ง DefaultSessionTTL (CON-CACHE-01, ASM-02).
func NewSessionStore(client *redis.Client, ttl time.Duration) *SessionStore {
	return &SessionStore{client: client, ttl: ttl}
}

// SessionKey คืน "session:" + hex(sha256(token)) เก็บเฉพาะ hash ไม่เก็บ token ดิบ (CON-CACHE-01, AC-AUTH-05).
func SessionKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return sessionKeyPrefix + hex.EncodeToString(sum[:])
}

// Create สร้าง token สุ่มใหม่ทุกครั้งและเก็บด้วย SET+EX ครั้งเดียว ไม่ลบ Session เดิม (FR-AUTH-01, ASM-02, ASM-04).
func (s *SessionStore) Create(ctx context.Context, memberID int64, studentID string) (string, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("auth: generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)

	value, err := json.Marshal(SessionData{MemberID: memberID, StudentID: studentID})
	if err != nil {
		return "", fmt.Errorf("auth: marshal session data: %w", err)
	}
	// SET พร้อม EX: ตั้ง TTL ครั้งเดียวตอนสร้าง (ASM-02 "ไม่ต่ออายุ").
	if err := s.client.Set(ctx, SessionKey(token), value, s.ttl).Err(); err != nil {
		return "", fmt.Errorf("auth: store session in redis: %w", err)
	}
	return token, nil
}

// Get อ่าน Session ด้วย GET ธรรมดา (ไม่ต่อ TTL) คืน ErrSessionNotFound เมื่อไม่พบ/ข้อมูลเสีย
// ส่วน error อื่นของ Redis ถูกห่อและแยกจาก not-found เพื่อให้ชั้นบนตอบ 503 (CON-CACHE-01, ASM-02, ASM-05).
func (s *SessionStore) Get(ctx context.Context, token string) (SessionData, error) {
	raw, err := s.client.Get(ctx, SessionKey(token)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return SessionData{}, ErrSessionNotFound
		}
		return SessionData{}, fmt.Errorf("auth: read session from redis: %w", err)
	}
	var data SessionData
	if err := json.Unmarshal(raw, &data); err != nil {
		return SessionData{}, ErrSessionNotFound
	}
	return data, nil
}
