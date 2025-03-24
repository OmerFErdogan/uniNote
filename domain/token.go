package domain

import (
	"time"
)

// RevokedToken, iptal edilmiş bir JWT token'ı temsil eder
type RevokedToken struct {
	ID        uint      `json:"id"`
	Token     string    `json:"token"`
	UserID    uint      `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	RevokedAt time.Time `json:"revokedAt"`
}

// TokenRepository, iptal edilmiş token'ların saklanması ve alınması için bir arayüz tanımlar
type TokenRepository interface {
	// RevokeToken, bir token'ı iptal eder
	RevokeToken(token *RevokedToken) error

	// IsTokenRevoked, bir token'ın iptal edilip edilmediğini kontrol eder
	IsTokenRevoked(tokenString string) (bool, error)

	// CleanupExpiredTokens, süresi dolmuş token'ları temizler
	CleanupExpiredTokens() error
}

// LoginAttempt, bir kullanıcının giriş denemesini temsil eder
type LoginAttempt struct {
	ID         uint      `json:"id"`
	IP         string    `json:"ip"`
	Email      string    `json:"email"`
	Successful bool      `json:"successful"`
	CreatedAt  time.Time `json:"createdAt"`
}

// LoginAttemptRepository, giriş denemelerinin saklanması ve alınması için bir arayüz tanımlar
type LoginAttemptRepository interface {
	// RecordAttempt, bir giriş denemesini kaydeder
	RecordAttempt(attempt *LoginAttempt) error

	// GetRecentAttempts, belirli bir IP veya e-posta için son giriş denemelerini getirir
	GetRecentAttempts(ip, email string, since time.Time) ([]*LoginAttempt, error)

	// CleanupOldAttempts, eski giriş denemelerini temizler
	CleanupOldAttempts(before time.Time) error
}
