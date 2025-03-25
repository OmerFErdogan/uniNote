package domain

import (
	"errors"
	"time"
)

// Hata türleri
var (
	ErrNotFound           = errors.New("kayıt bulunamadı")
	ErrInvalidCredentials = errors.New("geçersiz kimlik bilgileri")
	ErrInvalidToken       = errors.New("geçersiz token")
	ErrExpiredToken       = errors.New("süresi dolmuş token")
	ErrUnauthorized       = errors.New("yetkisiz erişim")
	ErrForbidden          = errors.New("erişim engellendi")
	ErrInvalidContentType = errors.New("geçersiz içerik türü")
	ErrInvalidInput       = errors.New("geçersiz girdi")
	ErrInternalServer     = errors.New("sunucu hatası")
	ErrDuplicateEntry     = errors.New("kayıt zaten mevcut")
)

// Now, şu anki zamanı döndürür (test edilebilirlik için)
func Now() time.Time {
	return time.Now()
}
