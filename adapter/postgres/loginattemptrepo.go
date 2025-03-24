package postgres

import (
	"time"

	"github.com/OmerFErdogan/uninote/domain"
	"github.com/OmerFErdogan/uninote/infrastructure/logger"
	"gorm.io/gorm"
)

// LoginAttemptModel, giriş denemelerinin veritabanı modeli
type LoginAttemptModel struct {
	ID         uint      `gorm:"primaryKey"`
	IP         string    `gorm:"type:varchar(45);not null;index"`
	Email      string    `gorm:"type:varchar(255);not null;index"`
	Successful bool      `gorm:"not null"`
	CreatedAt  time.Time `gorm:"not null;index"`
}

// TableName, tablo adını belirtir
func (LoginAttemptModel) TableName() string {
	return "login_attempts"
}

// LoginAttemptRepository, giriş denemeleri için repository implementasyonu
type LoginAttemptRepository struct {
	db *gorm.DB
}

// NewLoginAttemptRepository, yeni bir LoginAttemptRepository örneği oluşturur
func NewLoginAttemptRepository(db *gorm.DB) *LoginAttemptRepository {
	return &LoginAttemptRepository{
		db: db,
	}
}

// RecordAttempt, bir giriş denemesini kaydeder
func (r *LoginAttemptRepository) RecordAttempt(attempt *domain.LoginAttempt) error {
	model := &LoginAttemptModel{
		IP:         attempt.IP,
		Email:      attempt.Email,
		Successful: attempt.Successful,
		CreatedAt:  attempt.CreatedAt,
	}

	result := r.db.Create(model)
	if result.Error != nil {
		logger.Error("Giriş denemesi kaydedilirken hata oluştu: %v", result.Error)
		return result.Error
	}

	attempt.ID = model.ID
	return nil
}

// GetRecentAttempts, belirli bir IP veya e-posta için son giriş denemelerini getirir
func (r *LoginAttemptRepository) GetRecentAttempts(ip, email string, since time.Time) ([]*domain.LoginAttempt, error) {
	var models []*LoginAttemptModel
	query := r.db.Where("created_at > ?", since)

	// IP veya e-posta filtreleri ekle
	if ip != "" {
		query = query.Where("ip = ?", ip)
	}
	if email != "" {
		query = query.Where("email = ?", email)
	}

	// Başarısız denemeleri getir
	query = query.Where("successful = ?", false)

	// Sırala ve getir
	result := query.Order("created_at DESC").Find(&models)
	if result.Error != nil {
		logger.Error("Giriş denemeleri getirilirken hata oluştu: %v", result.Error)
		return nil, result.Error
	}

	// Domain modellerine dönüştür
	attempts := make([]*domain.LoginAttempt, len(models))
	for i, model := range models {
		attempts[i] = &domain.LoginAttempt{
			ID:         model.ID,
			IP:         model.IP,
			Email:      model.Email,
			Successful: model.Successful,
			CreatedAt:  model.CreatedAt,
		}
	}

	return attempts, nil
}

// CleanupOldAttempts, eski giriş denemelerini temizler
func (r *LoginAttemptRepository) CleanupOldAttempts(before time.Time) error {
	result := r.db.Where("created_at < ?", before).Delete(&LoginAttemptModel{})
	if result.Error != nil {
		logger.Error("Eski giriş denemeleri temizlenirken hata oluştu: %v", result.Error)
		return result.Error
	}

	logger.Info("Eski %d giriş denemesi temizlendi", result.RowsAffected)
	return nil
}

// Ensure LoginAttemptRepository implements domain.LoginAttemptRepository
var _ domain.LoginAttemptRepository = (*LoginAttemptRepository)(nil)
