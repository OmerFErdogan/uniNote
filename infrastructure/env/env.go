package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config, uygulama yapılandırmasını içerir
type Config struct {
	// Server
	ServerPort string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret     string
	JWTExpiryHour int

	// Storage
	PDFStoragePath string
}

// LoadConfig, çevre değişkenlerinden yapılandırmayı yükler
func LoadConfig() (*Config, error) {
	// .env dosyasını yükle (opsiyonel hata kontrolü)
	if err := godotenv.Load(); err != nil {
		fmt.Println("UYARI: .env dosyası yüklenemedi. Devam ediliyor...")
	}

	// PDF depolama dizinini oluştur (yoksa)
	if err := os.MkdirAll("./storage/pdfs", 0755); err != nil {
		fmt.Println("UYARI: PDF depolama dizini oluşturulamadı:", err)
	}

	config := &Config{
		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "uninotes"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// JWT
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTExpiryHour: getEnvAsInt("JWT_EXPIRY_HOUR", 24),

		// Storage
		PDFStoragePath: getEnv("PDF_STORAGE_PATH", "./storage/pdfs"),
	}

	return config, nil
}

// getEnv, çevre değişkenini alır veya varsayılan değeri döndürür
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt, çevre değişkenini int olarak alır veya varsayılan değeri döndürür
func getEnvAsInt(key string, defaultValue int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultValue
	}

	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultValue
}
