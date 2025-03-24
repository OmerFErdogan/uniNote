package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/OmerFErdogan/uninote/domain"
	"github.com/OmerFErdogan/uninote/infrastructure/logger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("geçersiz kimlik bilgileri")
	ErrUserAlreadyExists  = errors.New("kullanıcı zaten mevcut")
	ErrInvalidToken       = errors.New("geçersiz token")
	ErrTokenRevoked       = errors.New("token iptal edilmiş")
	ErrTooManyAttempts    = errors.New("çok fazla başarısız giriş denemesi, lütfen daha sonra tekrar deneyin")
)

// AuthService, kullanıcı kimlik doğrulama işlemlerini yönetir
type AuthService struct {
	userRepo         domain.UserRepository
	tokenRepo        domain.TokenRepository
	loginAttemptRepo domain.LoginAttemptRepository
	jwtSecret        string
	jwtExpiry        time.Duration
	hashingCost      int
	maxLoginAttempts int
	loginWindowMins  int
}

// NewAuthService, yeni bir AuthService örneği oluşturur
func NewAuthService(
	userRepo domain.UserRepository,
	tokenRepo domain.TokenRepository,
	loginAttemptRepo domain.LoginAttemptRepository,
	jwtSecret string,
	jwtExpiryHours int,
	maxLoginAttempts int,
	loginWindowMins int,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		tokenRepo:        tokenRepo,
		loginAttemptRepo: loginAttemptRepo,
		jwtSecret:        jwtSecret,
		jwtExpiry:        time.Duration(jwtExpiryHours) * time.Hour,
		hashingCost:      10, // bcrypt için maliyet faktörü
		maxLoginAttempts: maxLoginAttempts,
		loginWindowMins:  loginWindowMins,
	}
}

// Register, yeni bir kullanıcı kaydeder
func (s *AuthService) Register(user *domain.User) error {
	// Kullanıcı adı veya e-posta zaten kullanılıyor mu kontrol et
	existingUser, err := s.userRepo.FindByEmail(user.Email)
	if err != nil {
		return fmt.Errorf("kullanıcı kontrolü sırasında hata: %w", err)
	}
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	existingUser, err = s.userRepo.FindByUsername(user.Username)
	if err != nil {
		return fmt.Errorf("kullanıcı kontrolü sırasında hata: %w", err)
	}
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	// Şifreyi hash'le
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), s.hashingCost)
	if err != nil {
		return fmt.Errorf("şifre hash'leme sırasında hata: %w", err)
	}
	user.Password = string(hashedPassword)

	// Kullanıcıyı kaydet
	if err := s.userRepo.Create(user); err != nil {
		return fmt.Errorf("kullanıcı oluşturma sırasında hata: %w", err)
	}

	return nil
}

// Login, kullanıcı girişi yapar ve JWT token döndürür
func (s *AuthService) Login(email, password string, ip string) (string, error) {
	// Rate limiting kontrolü
	if err := s.checkLoginAttempts(ip, email); err != nil {
		return "", err
	}

	// Kullanıcıyı bul
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		s.recordLoginAttempt(ip, email, false)
		return "", fmt.Errorf("kullanıcı arama sırasında hata: %w", err)
	}
	if user == nil {
		s.recordLoginAttempt(ip, email, false)
		return "", ErrInvalidCredentials
	}

	// Şifreyi doğrula
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		s.recordLoginAttempt(ip, email, false)
		return "", ErrInvalidCredentials
	}

	// Başarılı giriş denemesini kaydet
	s.recordLoginAttempt(ip, email, true)

	// JWT token oluştur
	token, err := s.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("token oluşturma sırasında hata: %w", err)
	}

	return token, nil
}

// checkLoginAttempts, belirli bir IP veya e-posta için giriş denemelerini kontrol eder
func (s *AuthService) checkLoginAttempts(ip, email string) error {
	// Son x dakika içindeki başarısız giriş denemelerini getir
	since := time.Now().Add(-time.Duration(s.loginWindowMins) * time.Minute)
	attempts, err := s.loginAttemptRepo.GetRecentAttempts(ip, email, since)
	if err != nil {
		logger.Error("Giriş denemeleri kontrol edilirken hata oluştu: %v", err)
		return fmt.Errorf("giriş denemeleri kontrol edilirken hata: %w", err)
	}

	// Başarısız giriş denemesi sayısını kontrol et
	if len(attempts) >= s.maxLoginAttempts {
		logger.Error("Çok fazla başarısız giriş denemesi. IP: %s, Email: %s", ip, email)
		return ErrTooManyAttempts
	}

	return nil
}

// recordLoginAttempt, bir giriş denemesini kaydeder
func (s *AuthService) recordLoginAttempt(ip, email string, successful bool) {
	attempt := &domain.LoginAttempt{
		IP:         ip,
		Email:      email,
		Successful: successful,
		CreatedAt:  time.Now(),
	}

	if err := s.loginAttemptRepo.RecordAttempt(attempt); err != nil {
		logger.Error("Giriş denemesi kaydedilirken hata oluştu: %v", err)
	}
}

// ValidateToken, JWT token'ı doğrular ve kullanıcı ID'sini döndürür
func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
	// Token'ın iptal edilip edilmediğini kontrol et
	isRevoked, err := s.tokenRepo.IsTokenRevoked(tokenString)
	if err != nil {
		logger.Error("Token iptal durumu kontrol edilirken hata oluştu: %v", err)
		return 0, fmt.Errorf("token iptal durumu kontrol hatası: %w", err)
	}
	if isRevoked {
		return 0, ErrTokenRevoked
	}

	// Token'ı parse et
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Algoritma kontrolü
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("beklenmeyen imzalama metodu: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, fmt.Errorf("token parse hatası: %w", err)
	}

	// Token geçerli mi kontrol et
	if !token.Valid {
		return 0, ErrInvalidToken
	}

	// Claims'i al
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidToken
	}

	// Kullanıcı ID'sini al
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, ErrInvalidToken
	}

	return uint(userID), nil
}

// RevokeToken, bir token'ı iptal eder
func (s *AuthService) RevokeToken(tokenString string) error {
	// Token'ı parse et ve claims'i al
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return fmt.Errorf("token parse hatası: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ErrInvalidToken
	}

	// Kullanıcı ID'sini al
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return ErrInvalidToken
	}
	userID := uint(userIDFloat)

	// Sona erme zamanını al
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return ErrInvalidToken
	}
	expiresAt := time.Unix(int64(expFloat), 0)

	// Token'ı iptal et
	revokedToken := &domain.RevokedToken{
		Token:     tokenString,
		UserID:    userID,
		ExpiresAt: expiresAt,
		RevokedAt: time.Now(),
	}

	if err := s.tokenRepo.RevokeToken(revokedToken); err != nil {
		logger.Error("Token iptal edilirken hata oluştu: %v", err)
		return fmt.Errorf("token iptal hatası: %w", err)
	}

	return nil
}

// CleanupExpiredTokens, süresi dolmuş token'ları temizler
func (s *AuthService) CleanupExpiredTokens() error {
	if err := s.tokenRepo.CleanupExpiredTokens(); err != nil {
		logger.Error("Süresi dolmuş token'lar temizlenirken hata oluştu: %v", err)
		return fmt.Errorf("token temizleme hatası: %w", err)
	}
	return nil
}

// CleanupOldLoginAttempts, eski giriş denemelerini temizler
func (s *AuthService) CleanupOldLoginAttempts() error {
	// 30 günden eski giriş denemelerini temizle
	before := time.Now().AddDate(0, 0, -30)
	if err := s.loginAttemptRepo.CleanupOldAttempts(before); err != nil {
		logger.Error("Eski giriş denemeleri temizlenirken hata oluştu: %v", err)
		return fmt.Errorf("giriş denemesi temizleme hatası: %w", err)
	}
	return nil
}

// generateJWT, JWT token oluşturur
func (s *AuthService) generateJWT(user *domain.User) (string, error) {
	// Token oluşturma zamanı
	now := time.Now()

	// Claims oluştur
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"iat":      now.Unix(),
		"exp":      now.Add(s.jwtExpiry).Unix(),
	}

	// Token oluştur
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Token'ı imzala
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ChangePassword, kullanıcının şifresini değiştirir
func (s *AuthService) ChangePassword(id uint, oldPassword, newPassword string) error {
	// Kullanıcıyı bul
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("kullanıcı arama sırasında hata: %w", err)
	}
	if user == nil {
		return ErrInvalidCredentials
	}

	// Eski şifreyi doğrula
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Yeni şifreyi hash'le
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.hashingCost)
	if err != nil {
		return fmt.Errorf("şifre hash'leme sırasında hata: %w", err)
	}

	// Şifreyi güncelle
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("kullanıcı güncelleme sırasında hata: %w", err)
	}

	return nil
}

// GetProfile, kullanıcı profilini getirir
func (s *AuthService) GetProfile(id uint) (*domain.User, error) {
	return s.userRepo.FindByID(id)
}

// UpdateProfile, kullanıcı profilini günceller
func (s *AuthService) UpdateProfile(user *domain.User) error {
	// Mevcut kullanıcıyı bul
	existingUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return fmt.Errorf("kullanıcı arama sırasında hata: %w", err)
	}
	if existingUser == nil {
		return errors.New("kullanıcı bulunamadı")
	}

	// Şifreyi korumak için mevcut şifreyi kullan
	user.Password = existingUser.Password

	// Kullanıcıyı güncelle
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("kullanıcı güncelleme sırasında hata: %w", err)
	}

	return nil
}

// Ensure AuthService implements domain.UserService
// var _ domain.UserService = (*AuthService)(nil)
