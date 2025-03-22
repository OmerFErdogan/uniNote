package postgres

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config, veritabanı bağlantı ayarlarını içerir
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConnection, PostgreSQL veritabanına yeni bir bağlantı oluşturur
func NewConnection(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("veritabani bağlantisi kurulamadi: %w", err)
	}

	log.Println("PostgreSQL veritabanına başarıyla bağlanıldı")
	return db, nil
}

// Close, veritabanı bağlantısını kapatır
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate, veritabanı şemasını oluşturur veya günceller
func Migrate(db *gorm.DB, models ...interface{}) error {
	err := db.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("veritabanı migrasyonu başarısız: %w", err)
	}
	log.Println("Veritabanı şeması başarıyla oluşturuldu/güncellendi")
	return nil
}
