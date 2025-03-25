package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/OmerFErdogan/uninote/domain"
	"github.com/OmerFErdogan/uninote/infrastructure/http/middleware"
	"github.com/OmerFErdogan/uninote/infrastructure/logger"
	"github.com/OmerFErdogan/uninote/usecase"
	"github.com/go-chi/chi/v5"
)

// ViewHandler, görüntüleme işlemlerini yöneten HTTP handler'ıdır
type ViewHandler struct {
	viewService domain.ViewService
	noteService *usecase.NoteService
	pdfService  *usecase.PDFService
	logger      *logger.Logger
}

// NewViewHandler, yeni bir ViewHandler oluşturur
func NewViewHandler(viewService domain.ViewService, noteService *usecase.NoteService, pdfService *usecase.PDFService, logger *logger.Logger) *ViewHandler {
	return &ViewHandler{
		viewService: viewService,
		noteService: noteService,
		pdfService:  pdfService,
		logger:      logger,
	}
}

// RegisterRoutes, görüntüleme ile ilgili rotaları kaydeder
func (h *ViewHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	// Kimlik doğrulama gerektiren rotalar
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)
		r.Get("/views/content/{type}/{id}", h.GetContentViews)
		r.Get("/views/user", h.GetUserViews)
		r.Get("/views/check", h.CheckUserViewed)
	})

	// Not görüntüleme endpoint'i (görüntüleme kaydı oluşturur)
	r.Get("/notes/{id}/view", func(w http.ResponseWriter, r *http.Request) {
		middleware.OptionalAuth(authMiddleware, h.ViewNote).ServeHTTP(w, r)
	})

	// PDF görüntüleme endpoint'i (görüntüleme kaydı oluşturur)
	r.Get("/pdfs/{id}/view", func(w http.ResponseWriter, r *http.Request) {
		middleware.OptionalAuth(authMiddleware, h.ViewPDF).ServeHTTP(w, r)
	})
}

// GetContentViews, bir içeriğin görüntüleme kayıtlarını döndürür
func (h *ViewHandler) GetContentViews(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı kimliğini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// İçerik türünü ve ID'sini al
	contentType := chi.URLParam(r, "type")
	contentIDStr := chi.URLParam(r, "id")
	contentID, err := strconv.ParseUint(contentIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Geçersiz içerik ID'si", "error", err, "contentID", contentIDStr)
		http.Error(w, "Geçersiz içerik ID'si", http.StatusBadRequest)
		return
	}

	// İçerik türünü doğrula
	if contentType != "note" && contentType != "pdf" {
		h.logger.Error("Geçersiz içerik türü", "contentType", contentType)
		http.Error(w, "Geçersiz içerik türü. 'note' veya 'pdf' olmalıdır.", http.StatusBadRequest)
		return
	}

	// Sayfalama parametrelerini al
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Varsayılan limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Varsayılan offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// İçerik sahibi olup olmadığını kontrol et
	isOwner := false
	if contentType == "note" {
		// Not repository'den notu al ve sahibini kontrol et
		note, err := h.noteService.GetNote(uint(contentID))
		if err != nil {
			h.logger.Error("Not bulunamadı", "error", err, "noteID", contentID)
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		isOwner = note.UserID == userID
	} else if contentType == "pdf" {
		// PDF repository'den PDF'i al ve sahibini kontrol et
		pdf, err := h.pdfService.GetPDF(uint(contentID))
		if err != nil {
			h.logger.Error("PDF bulunamadı", "error", err, "pdfID", contentID)
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		isOwner = pdf.UserID == userID
	}

	// Eğer içerik sahibi değilse, erişimi reddet
	if !isOwner {
		h.logger.Error("Yetkisiz erişim", "userID", userID, "contentID", contentID, "contentType", contentType)
		http.Error(w, "Bu içeriğin görüntüleme kayıtlarına erişim izniniz yok", http.StatusForbidden)
		return
	}

	// İçeriğin görüntüleme kayıtlarını getir
	views, err := h.viewService.GetContentViews(uint(contentID), contentType, limit, offset)
	if err != nil {
		h.logger.Error("Görüntüleme kayıtları getirilemedi", "error", err, "contentID", contentID, "contentType", contentType)
		http.Error(w, "Görüntüleme kayıtları getirilemedi", http.StatusInternalServerError)
		return
	}

	// Görüntüleme kayıtlarını döndür
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"views": views,
		"pagination": map[string]int{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUserViews, kullanıcının görüntüleme kayıtlarını döndürür
func (h *ViewHandler) GetUserViews(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı kimliğini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Sayfalama parametrelerini al
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Varsayılan limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Varsayılan offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Kullanıcının görüntüleme kayıtlarını getir
	views, err := h.viewService.GetUserViews(userID, limit, offset)
	if err != nil {
		h.logger.Error("Kullanıcı görüntüleme kayıtları getirilemedi", "error", err, "userID", userID)
		http.Error(w, "Görüntüleme kayıtları getirilemedi", http.StatusInternalServerError)
		return
	}

	// Görüntüleme kayıtlarını döndür
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"views": views,
		"pagination": map[string]int{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// CheckUserViewed, kullanıcının bir içeriği görüntüleyip görüntülemediğini kontrol eder
func (h *ViewHandler) CheckUserViewed(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı kimliğini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// İçerik türünü ve ID'sini al
	contentType := r.URL.Query().Get("type")
	contentIDStr := r.URL.Query().Get("contentId")
	contentID, err := strconv.ParseUint(contentIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Geçersiz içerik ID'si", "error", err, "contentID", contentIDStr)
		http.Error(w, "Geçersiz içerik ID'si", http.StatusBadRequest)
		return
	}

	// İçerik türünü doğrula
	if contentType != "note" && contentType != "pdf" {
		h.logger.Error("Geçersiz içerik türü", "contentType", contentType)
		http.Error(w, "Geçersiz içerik türü. 'note' veya 'pdf' olmalıdır.", http.StatusBadRequest)
		return
	}

	// Kullanıcının içeriği görüntüleyip görüntülemediğini kontrol et
	viewed, err := h.viewService.HasUserViewed(userID, uint(contentID), contentType)
	if err != nil {
		h.logger.Error("Görüntüleme durumu kontrol edilemedi", "error", err, "userID", userID, "contentID", contentID, "contentType", contentType)
		http.Error(w, "Görüntüleme durumu kontrol edilemedi", http.StatusInternalServerError)
		return
	}

	// Görüntüleme durumunu döndür
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"viewed": viewed,
	})
}

// ViewNote, bir notu görüntüler ve görüntüleme kaydı oluşturur
func (h *ViewHandler) ViewNote(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı kimliğini al (opsiyonel)
	userID, ok := middleware.GetUserID(r)

	// Not ID'sini al
	idStr := chi.URLParam(r, "id")
	noteID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Error("Geçersiz not ID'si", "error", err, "noteID", idStr)
		http.Error(w, "Geçersiz not ID'si", http.StatusBadRequest)
		return
	}

	// Eğer kullanıcı giriş yapmışsa, görüntüleme kaydı oluştur
	if ok && userID > 0 {
		err := h.viewService.RecordView(userID, uint(noteID), "note")
		if err != nil {
			h.logger.Error("Görüntüleme kaydı oluşturulamadı", "error", err, "userID", userID, "noteID", noteID)
			// Görüntüleme kaydı oluşturulamazsa bile, notu göstermeye devam et
		}
	}

	// Not içeriğini getir ve göster
	// Bu kısım gerçek implementasyonda doldurulmalıdır
	// Şimdilik sadece başarılı yanıt döndürüyoruz
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Not görüntülendi",
	})
}

// ViewPDF, bir PDF'i görüntüler ve görüntüleme kaydı oluşturur
func (h *ViewHandler) ViewPDF(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı kimliğini al (opsiyonel)
	userID, ok := middleware.GetUserID(r)

	// PDF ID'sini al
	idStr := chi.URLParam(r, "id")
	pdfID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Error("Geçersiz PDF ID'si", "error", err, "pdfID", idStr)
		http.Error(w, "Geçersiz PDF ID'si", http.StatusBadRequest)
		return
	}

	// Eğer kullanıcı giriş yapmışsa, görüntüleme kaydı oluştur
	if ok && userID > 0 {
		err := h.viewService.RecordView(userID, uint(pdfID), "pdf")
		if err != nil {
			h.logger.Error("Görüntüleme kaydı oluşturulamadı", "error", err, "userID", userID, "pdfID", pdfID)
			// Görüntüleme kaydı oluşturulamazsa bile, PDF'i göstermeye devam et
		}
	}

	// PDF içeriğini getir ve göster
	// Bu kısım gerçek implementasyonda doldurulmalıdır
	// Şimdilik sadece başarılı yanıt döndürüyoruz
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "PDF görüntülendi",
	})
}
