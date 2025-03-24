package env

import (
	"crypto/rand"
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

	// Security
	MaxLoginAttempts int
	LoginWindowMins  int
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

	// JWT Secret için rastgele değer oluştur (eğer tanımlanmamışsa)
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		var err error
		jwtSecret, err = generateRandomSecret(32)
		if err != nil {
			return nil, fmt.Errorf("JWT secret oluşturulamadı: %v", err)
		}
		fmt.Println("UYARI: Rastgele JWT secret oluşturuldu. Üretim ortamında güvenli bir JWT_SECRET çevre değişkeni tanımlamanız önerilir.")
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
		JWTSecret:     jwtSecret,
		JWTExpiryHour: getEnvAsInt("JWT_EXPIRY_HOUR", 24),

		// Storage
		PDFStoragePath: getEnv("PDF_STORAGE_PATH", "./storage/pdfs"),

		// Security
		MaxLoginAttempts: getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LoginWindowMins:  getEnvAsInt("LOGIN_WINDOW_MINS", 15),
	}

	return config, nil
}

// generateRandomSecret, belirtilen uzunlukta rastgele bir secret oluşturur
func generateRandomSecret(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
	bytes := make([]byte, length)

	// Rastgele değerler oluştur
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Her byte için charset'ten bir karakter seç
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes), nil
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
