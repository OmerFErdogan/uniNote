package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/OmerFErdogan/uninote/infrastructure/http/middleware"
	"github.com/OmerFErdogan/uninote/infrastructure/http/utils"
	"github.com/OmerFErdogan/uninote/infrastructure/logger"
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
		r.Post("/likes/check-bulk", h.CheckBulkLikeStatus)
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
	startTime := time.Now()

	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		logger.Error("[LIKE] LikeContent - Kullanıcı kimliği bulunamadı - IP: %s", r.RemoteAddr)
		return
	}

	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		logger.Error("[LIKE] LikeContent - Geçersiz istek formatı - UserID: %d - IP: %s - Error: %v", userID, r.RemoteAddr, err)
		return
	}

	logger.Debug("[LIKE] LikeContent isteği - UserID: %d - ContentID: %d - Type: %s", userID, req.ContentID, req.Type)

	// İçeriği beğen
	if err := h.likeService.LikeContent(userID, req.ContentID, req.Type); err != nil {
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			logger.LogLikeOperation("LikeContent", fmt.Sprintf("%d", userID), req.ContentID, req.Type, false, err)
			return
		}
		if err == usecase.ErrContentNotFound {
			http.Error(w, "İçerik bulunamadı", http.StatusNotFound)
			logger.LogLikeOperation("LikeContent", fmt.Sprintf("%d", userID), req.ContentID, req.Type, false, err)
			return
		}
		http.Error(w, "İçerik beğenme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		logger.LogLikeOperation("LikeContent", fmt.Sprintf("%d", userID), req.ContentID, req.Type, false, err)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "İçerik başarıyla beğenildi",
	})

	duration := time.Since(startTime)
	logger.LogLikeOperation("LikeContent", fmt.Sprintf("%d", userID), req.ContentID, req.Type, true, nil)
	logger.LogRequest("POST", "/likes", r.RemoteAddr, fmt.Sprintf("%d", userID), http.StatusOK, duration)
}

// UnlikeContent, bir içeriğin beğenisini kaldırır
func (h *LikeHandler) UnlikeContent(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		logger.Error("[LIKE] UnlikeContent - Kullanıcı kimliği bulunamadı - IP: %s", r.RemoteAddr)
		return
	}

	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		logger.Error("[LIKE] UnlikeContent - Geçersiz istek formatı - UserID: %d - IP: %s - Error: %v", userID, r.RemoteAddr, err)
		return
	}

	logger.Debug("[LIKE] UnlikeContent isteği - UserID: %d - ContentID: %d - Type: %s", userID, req.ContentID, req.Type)

	// İçerik beğenisini kaldır
	if err := h.likeService.UnlikeContent(userID, req.ContentID, req.Type); err != nil {
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			logger.LogLikeOperation("UnlikeContent", fmt.Sprintf("%d", userID), req.ContentID, req.Type, false, err)
			return
		}
		if err == usecase.ErrContentNotFound {
			http.Error(w, "İçerik bulunamadı", http.StatusNotFound)
			logger.LogLikeOperation("UnlikeContent", fmt.Sprintf("%d", userID), req.ContentID, req.Type, false, err)
			return
		}
		http.Error(w, "İçerik beğeni kaldırma sırasında hata: "+err.Error(), http.StatusInternalServerError)
		logger.LogLikeOperation("UnlikeContent", fmt.Sprintf("%d", userID), req.ContentID, req.Type, false, err)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "İçerik beğenisi başarıyla kaldırıldı",
	})

	duration := time.Since(startTime)
	logger.LogLikeOperation("UnlikeContent", fmt.Sprintf("%d", userID), req.ContentID, req.Type, true, nil)
	logger.LogRequest("DELETE", "/likes", r.RemoteAddr, fmt.Sprintf("%d", userID), http.StatusOK, duration)
}

// GetUserLikes, kullanıcının beğenilerini getirir
func (h *LikeHandler) GetUserLikes(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		logger.Error("[LIKE] GetUserLikes - Kullanıcı kimliği bulunamadı - IP: %s", r.RemoteAddr)
		return
	}

	// Sayfalama parametrelerini al
	limit, offset := utils.GetPaginationParams(r)

	logger.Debug("[LIKE] GetUserLikes isteği - UserID: %d - Limit: %d - Offset: %d", userID, limit, offset)

	// Beğenileri getir
	likes, err := h.likeService.GetUserLikes(userID, limit, offset)
	if err != nil {
		http.Error(w, "Beğenileri getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		logger.Error("[LIKE] GetUserLikes - Beğenileri getirme hatası - UserID: %d - Error: %v", userID, err)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(likes)

	duration := time.Since(startTime)
	logger.Info("[LIKE] GetUserLikes başarılı - UserID: %d - Sonuç sayısı: %d", userID, len(likes))
	logger.LogRequest("GET", "/likes/my", r.RemoteAddr, fmt.Sprintf("%d", userID), http.StatusOK, duration)
}

// GetContentLikes, bir içeriğin beğenilerini getirir
func (h *LikeHandler) GetContentLikes(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// İçerik ID'sini ve türünü al
	contentIDStr := r.URL.Query().Get("contentId")
	contentType := r.URL.Query().Get("type")

	if contentIDStr == "" || contentType == "" {
		http.Error(w, "İçerik ID'si ve türü gerekli", http.StatusBadRequest)
		logger.Error("[LIKE] GetContentLikes - Eksik parametreler - IP: %s", r.RemoteAddr)
		return
	}

	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz içerik ID'si", http.StatusBadRequest)
		logger.Error("[LIKE] GetContentLikes - Geçersiz içerik ID'si - ContentID: %s - IP: %s", contentIDStr, r.RemoteAddr)
		return
	}

	// Sayfalama parametrelerini al
	limit, offset := utils.GetPaginationParams(r)

	logger.Debug("[LIKE] GetContentLikes isteği - ContentID: %d - Type: %s - Limit: %d - Offset: %d", contentID, contentType, limit, offset)

	// Beğenileri getir
	likes, err := h.likeService.GetContentLikes(uint(contentID), contentType, limit, offset)
	if err != nil {
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			logger.Error("[LIKE] GetContentLikes - Geçersiz içerik türü - ContentID: %d - Type: %s", contentID, contentType)
			return
		}
		http.Error(w, "Beğenileri getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		logger.Error("[LIKE] GetContentLikes - Beğenileri getirme hatası - ContentID: %d - Type: %s - Error: %v", contentID, contentType, err)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(likes)

	duration := time.Since(startTime)
	logger.Info("[LIKE] GetContentLikes başarılı - ContentID: %d - Type: %s - Sonuç sayısı: %d", contentID, contentType, len(likes))
	logger.LogRequest("GET", "/likes", r.RemoteAddr, "anonymous", http.StatusOK, duration)
}

// CheckLikeStatus, bir içeriğin kullanıcı tarafından beğenilip beğenilmediğini kontrol eder
func (h *LikeHandler) CheckLikeStatus(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		logger.Error("[LIKE] CheckLikeStatus - Kullanıcı kimliği bulunamadı - IP: %s", r.RemoteAddr)
		return
	}

	// İçerik ID'sini ve türünü al
	contentIDStr := r.URL.Query().Get("contentId")
	contentType := r.URL.Query().Get("type")

	if contentIDStr == "" || contentType == "" {
		http.Error(w, "İçerik ID'si ve türü gerekli", http.StatusBadRequest)
		logger.Error("[LIKE] CheckLikeStatus - Eksik parametreler - UserID: %d - IP: %s", userID, r.RemoteAddr)
		return
	}

	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz içerik ID'si", http.StatusBadRequest)
		logger.Error("[LIKE] CheckLikeStatus - Geçersiz içerik ID'si - UserID: %d - ContentID: %s - IP: %s", userID, contentIDStr, r.RemoteAddr)
		return
	}

	logger.Debug("[LIKE] CheckLikeStatus isteği - UserID: %d - ContentID: %d - Type: %s", userID, contentID, contentType)

	// Beğeni durumunu kontrol et
	isLiked, err := h.likeService.IsLikedByUser(userID, uint(contentID), contentType)
	if err != nil {
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			logger.Error("[LIKE] CheckLikeStatus - Geçersiz içerik türü - UserID: %d - ContentID: %d - Type: %s", userID, contentID, contentType)
			return
		}
		http.Error(w, "Beğeni durumu kontrol sırasında hata: "+err.Error(), http.StatusInternalServerError)
		logger.Error("[LIKE] CheckLikeStatus - Kontrol hatası - UserID: %d - ContentID: %d - Type: %s - Error: %v", userID, contentID, contentType, err)
		return
	}

	// İstemci tarafı önbelleğe alma için Cache-Control header'ı ekle
	// 5 dakika boyunca önbellekte tutulacak
	w.Header().Set("Cache-Control", "private, max-age=300")

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"isLiked": isLiked,
	})

	duration := time.Since(startTime)
	logger.Info("[LIKE] CheckLikeStatus başarılı - UserID: %d - ContentID: %d - Type: %s - IsLiked: %v", userID, contentID, contentType, isLiked)
	logger.LogRequest("GET", "/likes/check", r.RemoteAddr, fmt.Sprintf("%d", userID), http.StatusOK, duration)
}

// CheckBulkLikeStatus, birden fazla içeriğin kullanıcı tarafından beğenilip beğenilmediğini kontrol eder
func (h *LikeHandler) CheckBulkLikeStatus(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		logger.Error("[LIKE] CheckBulkLikeStatus - Kullanıcı kimliği bulunamadı - IP: %s", r.RemoteAddr)
		return
	}

	// İstek gövdesini ayrıştır
	var req struct {
		Items []struct {
			ContentID uint   `json:"contentId"`
			Type      string `json:"type"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		logger.Error("[LIKE] CheckBulkLikeStatus - Geçersiz istek formatı - UserID: %d - IP: %s - Error: %v", userID, r.RemoteAddr, err)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "En az bir içerik belirtilmelidir", http.StatusBadRequest)
		logger.Error("[LIKE] CheckBulkLikeStatus - Boş items dizisi - UserID: %d - IP: %s", userID, r.RemoteAddr)
		return
	}

	logger.Debug("[LIKE] CheckBulkLikeStatus isteği - UserID: %d - İçerik sayısı: %d", userID, len(req.Items))

	// Her bir içerik için beğeni durumunu kontrol et
	results := make(map[string]bool)
	skippedItems := 0

	for _, item := range req.Items {
		// İçerik türünü kontrol et
		if item.Type != "note" && item.Type != "pdf" {
			logger.Debug("[LIKE] CheckBulkLikeStatus - Geçersiz içerik türü atlandı - UserID: %d - ContentID: %d - Type: %s", userID, item.ContentID, item.Type)
			skippedItems++
			continue // Geçersiz içerik türünü atla
		}

		// Beğeni durumunu kontrol et
		isLiked, err := h.likeService.IsLikedByUser(userID, item.ContentID, item.Type)
		if err != nil {
			logger.Debug("[LIKE] CheckBulkLikeStatus - İçerik kontrolü sırasında hata - UserID: %d - ContentID: %d - Type: %s - Error: %v", userID, item.ContentID, item.Type, err)
			skippedItems++
			continue // Hata durumunda bu içeriği atla
		}

		// Sonucu ekle (contentId_type formatında anahtar kullan)
		key := fmt.Sprintf("%d_%s", item.ContentID, item.Type)
		results[key] = isLiked
	}

	// İstemci tarafı önbelleğe alma için Cache-Control header'ı ekle
	// 5 dakika boyunca önbellekte tutulacak
	w.Header().Set("Cache-Control", "private, max-age=300")

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})

	duration := time.Since(startTime)
	logger.LogBulkOperation("CheckBulkLikeStatus", fmt.Sprintf("%d", userID), len(req.Items), true, nil)
	logger.Info("[LIKE] CheckBulkLikeStatus başarılı - UserID: %d - Toplam içerik: %d - Başarılı: %d - Atlanan: %d",
		userID, len(req.Items), len(results), skippedItems)
	logger.LogRequest("POST", "/likes/check-bulk", r.RemoteAddr, fmt.Sprintf("%d", userID), http.StatusOK, duration)
}
