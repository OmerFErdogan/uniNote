package handler

import (
	"encoding/json"
	"net/http"

	"github.com/OmerFErdogan/uninote/domain"
	"github.com/OmerFErdogan/uninote/infrastructure/http/middleware"
	"github.com/OmerFErdogan/uninote/usecase"
	"github.com/go-chi/chi/v5"
)

// AuthHandler, kimlik doğrulama işlemlerini yönetir
type AuthHandler struct {
	authService *usecase.AuthService
}

// NewAuthHandler, yeni bir AuthHandler örneği oluşturur
func NewAuthHandler(authService *usecase.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRoutes, yönlendirmeleri kaydeder
func (h *AuthHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)
		r.Get("/profile", h.GetProfile)
		r.Put("/profile", h.UpdateProfile)
		r.Post("/change-password", h.ChangePassword)
	})
}

// RegisterRequest, kayıt isteği
type RegisterRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	University string `json:"university"`
	Department string `json:"department"`
	Class      string `json:"class"`
}

// LoginRequest, giriş isteği
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ChangePasswordRequest, şifre değiştirme isteği
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// TokenResponse, token yanıtı
type TokenResponse struct {
	Token string `json:"token"`
}

// Register, yeni bir kullanıcı kaydeder
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Kullanıcı oluştur
	user := &domain.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   req.Password,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		University: req.University,
		Department: req.Department,
		Class:      req.Class,
	}

	// Kullanıcıyı kaydet
	if err := h.authService.Register(user); err != nil {
		if err == usecase.ErrUserAlreadyExists {
			http.Error(w, "Kullanıcı zaten mevcut", http.StatusConflict)
			return
		}
		http.Error(w, "Kayıt sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Kullanıcı başarıyla kaydedildi",
	})
}

// Login, kullanıcı girişi yapar
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Giriş yap
	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			http.Error(w, "Geçersiz kimlik bilgileri", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Giriş sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Token yanıtı
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TokenResponse{
		Token: token,
	})
}

// GetProfile, kullanıcı profilini getirir
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Profili getir
	user, err := h.authService.GetProfile(userID)
	if err != nil {
		http.Error(w, "Profil getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Profil yanıtı
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// UpdateProfile, kullanıcı profilini günceller
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Kullanıcı ID'sini ayarla
	user.ID = userID

	// Profili güncelle
	if err := h.authService.UpdateProfile(&user); err != nil {
		http.Error(w, "Profil güncelleme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profil başarıyla güncellendi",
	})
}

// ChangePassword, kullanıcı şifresini değiştirir
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Şifreyi değiştir
	if err := h.authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		if err == usecase.ErrInvalidCredentials {
			http.Error(w, "Geçersiz şifre", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Şifre değiştirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Şifre başarıyla değiştirildi",
	})
}
