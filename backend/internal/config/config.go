// Package config โหลดค่าตั้งค่าของ backend จาก environment (CON-TECH-01).
package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// defaultHTTPAddr คือ address เริ่มต้นของเซิร์ฟเวอร์ (รายละเอียด implementation ทีมเลือกเอง ไม่ได้มาจาก spec)
const defaultHTTPAddr = ":8080"

// defaultDependencyTimeout คือเวลาที่ถือว่า Redis/PostgreSQL "ไม่ตอบสนอง" = 5 วินาที
// ASM-09 ทีมตัดสินใจเอง รอ Tech Lead ยืนยัน (Q-04); ไม่ได้มาจาก spec (AC-AUTH-07, AC-AUTH-08, ASM-05).
const defaultDependencyTimeout = 5 * time.Second

// Config เก็บค่าตั้งค่าที่ backend อ่านจาก environment (CON-TECH-01; DatabaseURL รองรับ CON-DB-01, RedisAddr รองรับ CON-CACHE-01).
// DependencyTimeout รองรับ AC-AUTH-07/08 (ASM-05, ASM-09, Q-04).
type Config struct {
	HTTPAddr          string        // HTTP_ADDR (ค่าเริ่มต้น ":8080")
	DatabaseURL       string        // DATABASE_URL (ไม่มีค่าเริ่มต้น)
	RedisAddr         string        // REDIS_ADDR (ไม่มีค่าเริ่มต้น)
	DependencyTimeout time.Duration // DEPENDENCY_TIMEOUT เช่น "5s" (ค่าเริ่มต้น 5s ตาม ASM-09; ต้องเป็นค่าบวก)
}

// Load อ่าน Config จาก environment และคืน error เมื่อไม่ได้ตั้ง DATABASE_URL หรือ REDIS_ADDR (CON-DB-01, CON-CACHE-01).
func Load() (Config, error) {
	return load(os.Getenv)
}

// load แยกจาก Load เพื่อให้ทดสอบโดยไม่แตะ environment จริง; อ่าน DEPENDENCY_TIMEOUT ด้วย (CON-TECH-01, AC-AUTH-07, AC-AUTH-08, ASM-09).
func load(getenv func(string) string) (Config, error) {
	cfg := Config{
		HTTPAddr:    getenv("HTTP_ADDR"),
		DatabaseURL: getenv("DATABASE_URL"),
		RedisAddr:   getenv("REDIS_ADDR"),
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = defaultHTTPAddr
	}

	var errs []error
	timeout, err := parseDependencyTimeout(getenv("DEPENDENCY_TIMEOUT"))
	if err != nil {
		errs = append(errs, err)
	}
	cfg.DependencyTimeout = timeout
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if cfg.RedisAddr == "" {
		errs = append(errs, errors.New("REDIS_ADDR is required"))
	}
	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// parseDependencyTimeout แปลง DEPENDENCY_TIMEOUT (Go duration เช่น "5s", "1500ms"); ว่าง = ค่าเริ่มต้น 5s; ค่าผิดรูปหรือไม่เป็นบวกคืน error (AC-AUTH-07, AC-AUTH-08, ASM-09).
func parseDependencyTimeout(raw string) (time.Duration, error) {
	if raw == "" {
		return defaultDependencyTimeout, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("DEPENDENCY_TIMEOUT %q is not a valid duration (example: 5s): %w", raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("DEPENDENCY_TIMEOUT %q must be greater than zero", raw)
	}
	return d, nil
}
