package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// env สร้าง getenv ปลอมจาก map เพื่อใช้กับ seam load (CON-TECH-01).
func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

// base คืน env ที่ครบตามที่จำเป็น (CON-DB-01, CON-CACHE-01) แล้วผสมค่าเพิ่ม.
func base(extra map[string]string) map[string]string {
	m := map[string]string{"DATABASE_URL": "postgres://u:p@h/db", "REDIS_ADDR": "r:6379"}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

// ครอบคลุม ASM-09 (AC-AUTH-07, AC-AUTH-08): ไม่ตั้ง DEPENDENCY_TIMEOUT -> ค่าเริ่มต้น 5 วินาที.
func TestLoad_DependencyTimeout_DefaultsTo5s(t *testing.T) {
	cfg, err := load(env(base(nil)))
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, cfg.DependencyTimeout)
	assert.Equal(t, ":8080", cfg.HTTPAddr)
}

// ครอบคลุม ASM-09: ค่า override ที่ถูกต้องถูกอ่านตามที่ตั้ง.
func TestLoad_DependencyTimeout_ValidOverride(t *testing.T) {
	for raw, want := range map[string]time.Duration{
		"2s":     2 * time.Second,
		"1500ms": 1500 * time.Millisecond,
		"1m":     time.Minute,
	} {
		t.Run(raw, func(t *testing.T) {
			cfg, err := load(env(base(map[string]string{"DEPENDENCY_TIMEOUT": raw})))
			require.NoError(t, err)
			assert.Equal(t, want, cfg.DependencyTimeout)
		})
	}
}

// ครอบคลุม ASM-09: ค่าผิดรูปหรือไม่เป็นบวกต้อง error ชัดเจน ไม่ถอยไปใช้ค่าเริ่มต้นเงียบ ๆ.
func TestLoad_DependencyTimeout_InvalidRejected(t *testing.T) {
	for _, raw := range []string{"abc", "0", "-1s", "0s", "5", "1.5.2s"} {
		t.Run(raw, func(t *testing.T) {
			cfg, err := load(env(base(map[string]string{"DEPENDENCY_TIMEOUT": raw})))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "DEPENDENCY_TIMEOUT")
			assert.Equal(t, Config{}, cfg)
		})
	}
}

// ครอบคลุม CON-DB-01, CON-CACHE-01 (พฤติกรรมเดิม): ขาด DATABASE_URL หรือ REDIS_ADDR ยังต้อง error และรายงานครบทุกข้อ.
func TestLoad_RequiredSettingsUnchanged(t *testing.T) {
	_, err := load(env(map[string]string{"REDIS_ADDR": "r:6379"}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL is required")

	_, err = load(env(map[string]string{"DATABASE_URL": "postgres://x"}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REDIS_ADDR is required")

	_, err = load(env(map[string]string{"DEPENDENCY_TIMEOUT": "abc"}))
	require.Error(t, err)
	for _, want := range []string{"DATABASE_URL is required", "REDIS_ADDR is required", "DEPENDENCY_TIMEOUT"} {
		assert.Contains(t, err.Error(), want)
	}

	cfg, err := load(env(base(map[string]string{"HTTP_ADDR": ":9999"})))
	require.NoError(t, err)
	assert.Equal(t, ":9999", cfg.HTTPAddr)
	assert.Equal(t, "postgres://u:p@h/db", cfg.DatabaseURL)
	assert.Equal(t, "r:6379", cfg.RedisAddr)
}
