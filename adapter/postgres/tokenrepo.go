package postgres

import (
	"time"

	"github.com/OmerFErdogan/uninote/domain"
	"github.com/OmerFErdogan/uninote/infrastructure/logger"
	"gorm.io/gorm"
)

// RevokedTokenModel, iptal edilmiş token'ların veritabanı modeli
type RevokedTokenModel struct {
	ID        uint      `gorm:"primaryKey"`
	Token     string    `gorm:"type:text;not null;uniqueIndex"`
	UserID    uint      `gorm:"not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt time.Time `gorm:"not null"`
}

// TableName, tablo adını belirtir
func (RevokedTokenModel) TableName() string {
	return "revoked_tokens"
}

// TokenRepository, iptal edilmiş token'lar için repository implementasyonu
type TokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository, yeni bir TokenRepository örneği oluşturur
func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

// RevokeToken, bir token'ı iptal eder
func (r *TokenRepository) RevokeToken(token *domain.RevokedToken) error {
	model := &RevokedTokenModel{
		Token:     token.Token,
		UserID:    token.UserID,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
	}

	result := r.db.Create(model)
	if result.Error != nil {
		logger.Error("Token iptal edilirken hata oluştu: %v", result.Error)
		return result.Error
	}

	token.ID = model.ID
	return nil
}

// IsTokenRevoked, bir token'ın iptal edilip edilmediğini kontrol eder
func (r *TokenRepository) IsTokenRevoked(tokenString string) (bool, error) {
	var count int64
	result := r.db.Model(&RevokedTokenModel{}).Where("token = ?", tokenString).Count(&count)
	if result.Error != nil {
		logger.Error("Token iptal durumu kontrol edilirken hata oluştu: %v", result.Error)
		return false, result.Error
	}

	return count > 0, nil
}

// CleanupExpiredTokens, süresi dolmuş token'ları temizler
func (r *TokenRepository) CleanupExpiredTokens() error {
	now := time.Now()
	result := r.db.Where("expires_at < ?", now).Delete(&RevokedTokenModel{})
	if result.Error != nil {
		logger.Error("Süresi dolmuş token'lar temizlenirken hata oluştu: %v", result.Error)
		return result.Error
	}

	logger.Info("Süresi dolmuş %d token temizlendi", result.RowsAffected)
	return nil
}

// Ensure TokenRepository implements domain.TokenRepository
var _ domain.TokenRepository = (*TokenRepository)(nil)
