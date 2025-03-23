package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OmerFErdogan/uninote/adapter/localfs"
	"github.com/OmerFErdogan/uninote/adapter/postgres"
	"github.com/OmerFErdogan/uninote/infrastructure/env"
	apphttp "github.com/OmerFErdogan/uninote/infrastructure/http"
	"github.com/OmerFErdogan/uninote/infrastructure/http/handler"
	"github.com/OmerFErdogan/uninote/infrastructure/http/middleware"
	"github.com/OmerFErdogan/uninote/infrastructure/logger"
	"github.com/OmerFErdogan/uninote/usecase"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Logger'ı başlat
	logger.Init()
	logger.Info("UniNotes uygulaması başlatılıyor...")

	// Yapılandırmayı yükle
	config, err := env.LoadConfig()
	if err != nil {
		logger.Error("Yapılandırma yüklenemedi: %v", err)
		log.Fatalf("Yapılandırma yüklenemedi: %v", err)
	}

	// Veritabanı bağlantısını oluştur
	dbConfig := &postgres.Config{
		Host:     config.DBHost,
		Port:     config.DBPort,
		User:     config.DBUser,
		Password: config.DBPassword,
		DBName:   config.DBName,
		SSLMode:  config.DBSSLMode,
	}

	logger.Info("Veritabanına bağlanılıyor: %s:%s/%s", config.DBHost, config.DBPort, config.DBName)
	db, err := postgres.NewConnection(dbConfig)
	if err != nil {
		logger.Error("Veritabanı bağlantısı kurulamadı: %v", err)
		log.Fatalf("Veritabanı bağlantısı kurulamadı: %v", err)
	}

	// Veritabanı modellerini migrate et
	logger.Info("Veritabanı modelleri migrate ediliyor...")
	err = postgres.Migrate(db,
		&postgres.UserModel{},
		&postgres.NoteModel{},
		&postgres.CommentModel{},
		&postgres.PDFModel{},
		&postgres.PDFCommentModel{},
		&postgres.PDFAnnotationModel{},
		&postgres.ContentLikeModel{},
	)
	if err != nil {
		logger.Error("Veritabanı migrasyonu başarısız: %v", err)
		log.Fatalf("Veritabanı migrasyonu başarısız: %v", err)
	}
	logger.Info("Veritabanı migrasyonu başarılı")

	// Repository'leri oluştur
	userRepo := postgres.NewUserRepository(db)
	noteRepo := postgres.NewNoteRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	pdfRepo := postgres.NewPDFRepository(db)
	pdfCommentRepo := postgres.NewPDFCommentRepository(db)
	pdfAnnotationRepo := postgres.NewPDFAnnotationRepository(db)
	likeRepo := postgres.NewLikeRepository(db)

	// PDF depolama servisini oluştur
	pdfStorage, err := localfs.NewPDFStorage(config.PDFStoragePath)
	if err != nil {
		log.Fatalf("PDF depolama servisi oluşturulamadı: %v", err)
	}

	// Servisleri oluştur
	authService := usecase.NewAuthService(userRepo, config.JWTSecret, config.JWTExpiryHour)
	noteService := usecase.NewNoteService(noteRepo, commentRepo)
	pdfService := usecase.NewPDFService(pdfRepo, pdfCommentRepo, pdfAnnotationRepo, pdfStorage)
	likeService := usecase.NewLikeService(likeRepo, noteRepo, pdfRepo)
	commentService := usecase.NewCommentService(noteRepo, commentRepo, pdfRepo, pdfCommentRepo, userRepo)

	// Middleware'leri oluştur
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Handler'ları oluştur
	authHandler := handler.NewAuthHandler(authService)
	noteHandler := handler.NewNoteHandler(noteService, likeService, commentService)
	pdfHandler := handler.NewPDFHandler(pdfService, likeService, commentService)
	likeHandler := handler.NewLikeHandler(likeService)

	// Router'ı oluştur
	router := apphttp.NewRouter()

	// CORS middleware'ini ekleyin
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// OPTIONS istekleri için hemen yanıt döndür
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Temel endpoint
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "UniNotes API'ye Hoş Geldiniz!", "version": "0.1.0"}`))
	})

	// API endpoint'lerini ekle
	router.Route("/api/v1", func(r chi.Router) {
		// Sağlık kontrolü
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status": "ok"}`))
		})

		// Auth endpoint'leri
		authHandler.RegisterRoutes(r, authMiddleware)

		// Not endpoint'leri
		noteHandler.RegisterRoutes(r, authMiddleware)

		// PDF endpoint'leri
		pdfHandler.RegisterRoutes(r, authMiddleware)

		// Beğeni endpoint'leri
		likeHandler.RegisterRoutes(r, authMiddleware)
	})

	// Statik dosyaları web klasöründen sun (isteğe bağlı)
	// Eğer web klasörü yoksa veya dosyalarınızı başka bir şekilde sunmak istiyorsanız bu kısmı kaldırabilirsiniz
	fs := http.FileServer(http.Dir("./web"))
	router.Get("/web/*", http.StripPrefix("/web/", fs).ServeHTTP)

	// Sunucuyu başlat
	port := ":" + config.ServerPort
	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	// Graceful shutdown için sinyal yakalama
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("UniNotes sunucusu port %s üzerinde çalışıyor", port)
		log.Printf("UniNotes sunucusu port %s üzerinde çalışıyor", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Sunucu hatası: %v", err)
			log.Fatalf("Sunucu hatası: %v", err)
		}
	}()

	// Sinyal bekle
	<-stop
	logger.Info("Sunucu kapatılıyor...")
	log.Println("Sunucu kapatılıyor...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Sunucu kapatma hatası: %v", err)
		log.Fatalf("Sunucu kapatma hatası: %v", err)
	}

	// Veritabanı bağlantısını kapat
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Veritabanı bağlantısı alınamadı: %v", err)
		log.Fatalf("Veritabanı bağlantısı alınamadı: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		logger.Error("Veritabanı bağlantısı kapatılamadı: %v", err)
		log.Fatalf("Veritabanı bağlantısı kapatılamadı: %v", err)
	}

	logger.Info("Sunucu başarıyla kapatıldı")
	log.Println("Sunucu başarıyla kapatıldı")
}
