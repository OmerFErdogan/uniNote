package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/OmerFErdogan/uninote/usecase"
)

// AuthMiddleware, kimlik doğrulama middleware'i
type AuthMiddleware struct {
	authService *usecase.AuthService
}

// NewAuthMiddleware, yeni bir AuthMiddleware örneği oluşturur
func NewAuthMiddleware(authService *usecase.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Middleware, HTTP isteklerini işler ve kimlik doğrulama yapar
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization header'ını al
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Yetkilendirme başlığı eksik", http.StatusUnauthorized)
			return
		}

		// Bearer token'ı ayıkla
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Geçersiz yetkilendirme formatı", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// Token'ı doğrula
		userID, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Geçersiz token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Kullanıcı ID'sini context'e ekle
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID, context'ten kullanıcı ID'sini alır
func GetUserID(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value("userID").(uint)
	return userID, ok
}

// RequireAuth, kimlik doğrulama gerektiren handler'lar için bir yardımcı fonksiyon
func RequireAuth(authMiddleware *AuthMiddleware, handler http.HandlerFunc) http.Handler {
	return authMiddleware.Middleware(handler)
}

// OptionalAuth, kimlik doğrulama gerektirmeyen ancak kimlik doğrulama bilgilerini kullanan handler'lar için bir yardımcı fonksiyon
func OptionalAuth(authMiddleware *AuthMiddleware, handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization header'ını al
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Kimlik doğrulama yok, normal devam et
			handler.ServeHTTP(w, r)
			return
		}

		// Bearer token'ı ayıkla
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Geçersiz format, normal devam et
			handler.ServeHTTP(w, r)
			return
		}
		tokenString := parts[1]

		// Token'ı doğrula
		userID, err := authMiddleware.authService.ValidateToken(tokenString)
		if err != nil {
			// Geçersiz token, normal devam et
			handler.ServeHTTP(w, r)
			return
		}

		// Kullanıcı ID'sini context'e ekle
		ctx := context.WithValue(r.Context(), "userID", userID)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
