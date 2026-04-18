package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"brandhunt/api-service/config"
	"brandhunt/api-service/internal/handler"
	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/repository"
	"brandhunt/api-service/internal/service"

	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	cfg := config.Load()

	// Database
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Wait for DB
	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Printf("Ожидание БД (%d/30)...", i+1)
		time.Sleep(2 * time.Second)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("БД не доступна: %v", err)
	}
	log.Println("Подключение к БД установлено")

	// MinIO
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioSecure,
	})
	if err != nil {
		log.Printf("WARN: не удалось подключиться к MinIO: %v (удаление фото будет недоступно)", err)
		minioClient = nil
	}

	// Фото MinIO отдаются через nginx-proxy на /minio/, чтобы браузер
	// не видел внутренний docker-хост minio:9000. Репозиторий переписывает
	// URL в ответах на лету.
	repository.SetMinioEndpoint(cfg.MinioEndpoint)

	// Repositories
	userRepo := repository.NewUserRepo(db)
	listingRepo := repository.NewListingRepo(db)
	favoriteRepo := repository.NewFavoriteRepo(db)
	sourceRepo := repository.NewSourceRepo(db)
	statsRepo := repository.NewStatsRepo(db)
	photoRepo := repository.NewPhotoRepo(db)
	filtersRepo := repository.NewFiltersRepo(db)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	userSvc := service.NewUserService(userRepo)
	listingSvc := service.NewListingService(listingRepo)
	favoriteSvc := service.NewFavoriteService(favoriteRepo, listingRepo)
	sourceSvc := service.NewSourceService(sourceRepo)
	statsSvc := service.NewStatsService(statsRepo)
	photoSvc := service.NewPhotoService(photoRepo, listingRepo, minioClient, cfg.MinioBucket)
	filtersSvc := service.NewFiltersService(filtersRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	listingH := handler.NewListingHandler(listingSvc)
	favoriteH := handler.NewFavoriteHandler(favoriteSvc)
	adminH := handler.NewAdminHandler(listingSvc, sourceSvc, statsSvc, photoSvc)
	filtersH := handler.NewFiltersHandler(filtersSvc)

	// Middleware shortcuts
	E := middleware.ErrorHandler
	auth := middleware.Auth(cfg.JWTSecret)
	admin := middleware.RequireAdmin

	mux := http.NewServeMux()

	// ── Public ──────────────────────────────────────────
	mux.HandleFunc("/api/v1/auth/register", E(authH.Register))
	mux.HandleFunc("/api/v1/auth/login", E(authH.Login))

	// Listings: route /api/v1/listings exactly vs /api/v1/listings/{id}
	mux.HandleFunc("/api/v1/listings", E(listingH.Search))
	mux.HandleFunc("/api/v1/listings/", E(listingH.GetByID))

	// Filters (meta-информация для UI-фильтров)
	mux.HandleFunc("/api/v1/filters/sizes", E(filtersH.Sizes))

	// ── Auth required ───────────────────────────────────
	mux.HandleFunc("/api/v1/auth/logout", E(auth(authH.Logout)))
	mux.HandleFunc("/api/v1/users/me", E(auth(userH.Me)))
	mux.HandleFunc("/api/v1/users/me/favorites", E(auth(favoriteH.Favorites)))
	mux.HandleFunc("/api/v1/users/me/favorites/", E(auth(favoriteH.Remove)))

	// ── Admin ───────────────────────────────────────────
	mux.HandleFunc("/api/v1/admin/listings", E(auth(admin(adminH.AdminListings))))
	mux.HandleFunc("/api/v1/admin/sources", E(auth(admin(adminH.Sources))))
	mux.HandleFunc("/api/v1/admin/stats", E(auth(admin(adminH.Stats))))
	mux.HandleFunc("/api/v1/admin/stats/listings-by-day", E(auth(admin(adminH.ListingsByDay))))
	mux.HandleFunc("/api/v1/admin/stats/top-brands", E(auth(admin(adminH.TopBrands))))
	mux.HandleFunc("/api/v1/admin/stats/top-cities", E(auth(admin(adminH.TopCities))))

	// Admin routes with path params — use prefix matching
	mux.HandleFunc("/api/v1/admin/listings/", E(auth(admin(adminRouter(adminH)))))
	mux.HandleFunc("/api/v1/admin/sources/", E(auth(admin(adminH.ToggleSource))))

	// ── Swagger ─────────────────────────────────────────
	mux.HandleFunc("/api/v1/swagger.json", serveSwaggerJSON)
	mux.HandleFunc("/swagger/", serveSwaggerUI)
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply CORS
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      middleware.CORS(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("API-сервис запущен на порту %d", cfg.Port)
		log.Printf("Swagger UI: http://localhost:%d/swagger/", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Завершение работы...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

// adminRouter dispatches /api/v1/admin/listings/{id}/... requests
func adminRouter(h *handler.AdminHandler) middleware.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/visibility"):
			return h.Visibility(w, r)
		case strings.HasSuffix(path, "/text"):
			return h.EditText(w, r)
		case strings.Contains(path, "/photos/"):
			return h.DeletePhoto(w, r)
		}
		// /api/v1/admin/listings/{id} или /api/v1/admin/listings/{id}/ — без суффикса
		rest := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/admin/listings/"), "/")
		if rest != "" && !strings.Contains(rest, "/") {
			return h.GetListing(w, r)
		}
		return middleware.NewAppError(http.StatusNotFound, "маршрут не найден")
	}
}
