package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/OmerFErdogan/uninote/infrastructure/http/middleware"
	"github.com/OmerFErdogan/uninote/infrastructure/http/utils"
	"github.com/OmerFErdogan/uninote/usecase"
	"github.com/go-chi/chi/v5"
)

// LikeHandler, beğeni işlemlerini yönetir
type LikeHandler struct {
	likeService *usecase.LikeService
}

// NewLikeHandler, yeni bir LikeHandler örneği oluşturur
func NewLikeHandler(likeService *usecase.LikeService) *LikeHandler {
	return &LikeHandler{
		likeService: likeService,
	}
}

// RegisterRoutes, yönlendirmeleri kaydeder
func (h *LikeHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	// Kimlik doğrulama gerektiren rotalar
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)
		r.Post("/likes", h.LikeContent)
		r.Delete("/likes", h.UnlikeContent)
		r.Get("/likes/my", h.GetUserLikes)
		r.Get("/likes/check", h.CheckLikeStatus)
	})

	// Kimlik doğrulama gerektirmeyen rotalar
	r.Get("/likes", h.GetContentLikes)
}

// LikeRequest, beğeni isteği
type LikeRequest struct {
	ContentID uint   `json:"contentId"`
	Type      string `json:"type"` // "note" veya "pdf"
}

// LikeContent, bir içeriği beğenir
func (h *LikeHandler) LikeContent(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// İçeriği beğen
	if err := h.likeService.LikeContent(userID, req.ContentID, req.Type); err != nil {
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			return
		}
		if err == usecase.ErrContentNotFound {
			http.Error(w, "İçerik bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "İçerik beğenme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "İçerik başarıyla beğenildi",
	})
}

// UnlikeContent, bir içeriğin beğenisini kaldırır
func (h *LikeHandler) UnlikeContent(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// İçerik beğenisini kaldır
	if err := h.likeService.UnlikeContent(userID, req.ContentID, req.Type); err != nil {
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			return
		}
		if err == usecase.ErrContentNotFound {
			http.Error(w, "İçerik bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "İçerik beğeni kaldırma sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "İçerik beğenisi başarıyla kaldırıldı",
	})
}

// GetUserLikes, kullanıcının beğenilerini getirir
func (h *LikeHandler) GetUserLikes(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Sayfalama parametrelerini al
	limit, offset := utils.GetPaginationParams(r)

	// Beğenileri getir
	likes, err := h.likeService.GetUserLikes(userID, limit, offset)
	if err != nil {
		http.Error(w, "Beğenileri getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(likes)
}

// GetContentLikes, bir içeriğin beğenilerini getirir
func (h *LikeHandler) GetContentLikes(w http.ResponseWriter, r *http.Request) {
	// İçerik ID'sini ve türünü al
	contentIDStr := r.URL.Query().Get("contentId")
	contentType := r.URL.Query().Get("type")

	if contentIDStr == "" || contentType == "" {
		http.Error(w, "İçerik ID'si ve türü gerekli", http.StatusBadRequest)
		return
	}

	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz içerik ID'si", http.StatusBadRequest)
		return
	}

	// Sayfalama parametrelerini al
	limit, offset := utils.GetPaginationParams(r)

	// Beğenileri getir
	likes, err := h.likeService.GetContentLikes(uint(contentID), contentType, limit, offset)
	if err != nil {
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			return
		}
		http.Error(w, "Beğenileri getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(likes)
}

// CheckLikeStatus, bir içeriğin kullanıcı tarafından beğenilip beğenilmediğini kontrol eder
func (h *LikeHandler) CheckLikeStatus(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// İçerik ID'sini ve türünü al
	contentIDStr := r.URL.Query().Get("contentId")
	contentType := r.URL.Query().Get("type")

	if contentIDStr == "" || contentType == "" {
		http.Error(w, "İçerik ID'si ve türü gerekli", http.StatusBadRequest)
		return
	}

	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz içerik ID'si", http.StatusBadRequest)
		return
	}

	// Beğeni durumunu kontrol et
	isLiked, err := h.likeService.IsLikedByUser(userID, uint(contentID), contentType)
	if err != nil {
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			return
		}
		http.Error(w, "Beğeni durumu kontrol sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"isLiked": isLiked,
	})
}
