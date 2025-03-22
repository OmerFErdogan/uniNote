package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/OmerFErdogan/uninote/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("geçersiz kimlik bilgileri")
	ErrUserAlreadyExists  = errors.New("kullanıcı zaten mevcut")
	ErrInvalidToken       = errors.New("geçersiz token")
)

// AuthService, kullanıcı kimlik doğrulama işlemlerini yönetir
type AuthService struct {
	userRepo    domain.UserRepository
	jwtSecret   string
	jwtExpiry   time.Duration
	hashingCost int
}

// NewAuthService, yeni bir AuthService örneği oluşturur
func NewAuthService(userRepo domain.UserRepository, jwtSecret string, jwtExpiryHours int) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		jwtExpiry:   time.Duration(jwtExpiryHours) * time.Hour,
		hashingCost: 10, // bcrypt için maliyet faktörü
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
func (s *AuthService) Login(email, password string) (string, error) {
	// Kullanıcıyı bul
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", fmt.Errorf("kullanıcı arama sırasında hata: %w", err)
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	// Şifreyi doğrula
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// JWT token oluştur
	token, err := s.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("token oluşturma sırasında hata: %w", err)
	}

	return token, nil
}

// ValidateToken, JWT token'ı doğrular ve kullanıcı ID'sini döndürür
func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
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
var _ domain.UserService = (*AuthService)(nil)
