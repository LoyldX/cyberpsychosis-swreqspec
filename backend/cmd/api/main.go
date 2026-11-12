// Command api เปิด HTTP server ของ backend (CON-TECH-01).
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"gymapp/backend/internal/app"
	"gymapp/backend/internal/auth"
	"gymapp/backend/internal/config"
	"gymapp/backend/internal/inventory"
	"gymapp/backend/internal/occupancy"
	"gymapp/backend/internal/roster"
)

// main เรียก run แล้วจบด้วย log.Fatalf เมื่อมี error (CON-TECH-01); แยก run เพื่อให้ defer ปิด pool/client ทำงาน.
func main() {
	if err := run(); err != nil {
		log.Fatalf("api: %v", err)
	}
}

// run โหลด config ต่อ PostgreSQL (CON-DB-01) และ Redis (CON-CACHE-01) ประกอบ auth, inventory, occupancy และ roster เข้ากับ Gin
// (CON-TECH-01, FR-AUTH-01, FR-INV-01, FR-BI-01, FR-ROST-01) แล้วเปิดเซิร์ฟเวอร์จนได้รับสัญญาณหยุด ไม่รัน migration;
// ตั้ง timeout ของ dependency จาก config (AC-AUTH-07, AC-AUTH-08, ASM-09).
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// pgxpool.New ยังไม่เชื่อมต่อจริงจนกว่าจะมี query (CON-DB-01)
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer func() { _ = redisClient.Close() }()

	members := auth.NewMemberStore(pool)
	sessions := auth.NewSessionStore(redisClient, auth.DefaultSessionTTL) // ASM-02: อายุ 7 วัน
	// timeout ของ dependency มาจาก config (AC-AUTH-07, AC-AUTH-08, ASM-09)
	handler := auth.NewHandler(auth.NewService(members, sessions, auth.DefaultSessionTTL)).WithSessionReader(sessions).WithDependencyTimeout(cfg.DependencyTimeout)
	inventoryHandler := inventory.NewHandler(inventory.NewStore(pool))
	// FR-BI-01, NFR-PERF-02: occupancy ใช้ PostgreSQL ตาม ASM-02 และนาฬิกาจริงของเครื่อง
	occupancyService := occupancy.NewService(occupancy.NewStore(pool), occupancy.NewBroadcaster(), nil)
	occupancyHandler := occupancy.NewHandler(occupancyService)
	// FR-ROST-01: ตารางงานเทรนเนอร์ใช้ PostgreSQL และนาฬิกาจริง ไม่ต้องล็อกอิน (004 ASM-06)
	rosterHandler := roster.NewHandler(roster.NewStore(pool), nil)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: app.NewRouter(handler.Register, inventoryHandler.Register, occupancyHandler.Register, rosterHandler.Register),
		// context ของทุกคำขอมาจาก ctx: เมื่อได้รับสัญญาณหยุด การเชื่อมต่อ SSE ที่เปิดค้างจะจบเอง Shutdown จึงไม่ต้องรอจนหมดเวลา (NFR-PERF-02)
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	// ASM-03, AC-BI-06: กระจายจำนวนของวันใหม่ตอน 00:00 Asia/Bangkok จนกว่าเซิร์ฟเวอร์จะหยุด
	go occupancyService.RunMidnightReset(ctx)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", cfg.HTTPAddr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
