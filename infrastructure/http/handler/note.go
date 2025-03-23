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

// NoteHandler, not işlemlerini yönetir
type NoteHandler struct {
	noteService *usecase.NoteService
	likeService *usecase.LikeService
}

// NewNoteHandler, yeni bir NoteHandler örneği oluşturur
func NewNoteHandler(noteService *usecase.NoteService, likeService *usecase.LikeService) *NoteHandler {
	return &NoteHandler{
		noteService: noteService,
		likeService: likeService,
	}
}

// RegisterRoutes, yönlendirmeleri kaydeder
func (h *NoteHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	// Kimlik doğrulama gerektiren rotalar
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)
		r.Post("/notes", h.CreateNote)
		r.Put("/notes/{id}", h.UpdateNote)
		r.Delete("/notes/{id}", h.DeleteNote)
		r.Get("/notes/my", h.GetUserNotes)
		r.Post("/notes/{id}/comments", h.AddComment)
		r.Post("/notes/{id}/like", h.LikeNote)
		r.Delete("/notes/{id}/like", h.UnlikeNote)
		r.Get("/notes/liked", h.GetLikedNotes)
	})

	// Kimlik doğrulama gerektirmeyen rotalar
	r.Get("/notes", h.GetPublicNotes)
	r.Get("/notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		middleware.OptionalAuth(authMiddleware, h.GetNote).ServeHTTP(w, r)
	})
	r.Get("/notes/{id}/comments", h.GetComments)
	r.Get("/notes/search", h.SearchNotes)
	r.Get("/notes/tag/{tag}", h.GetNotesByTag)
}

// CreateNoteRequest, not oluşturma isteği
type CreateNoteRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	IsPublic bool     `json:"isPublic"`
}

// UpdateNoteRequest, not güncelleme isteği
type UpdateNoteRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	IsPublic bool     `json:"isPublic"`
}

// CommentRequest, yorum isteği
type CommentRequest struct {
	Content string `json:"content"`
}

// CreateNote, yeni bir not oluşturur
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Not oluştur
	note := &domain.Note{
		Title:    req.Title,
		Content:  req.Content,
		UserID:   userID,
		Tags:     req.Tags,
		IsPublic: req.IsPublic,
	}

	// Notu kaydet
	if err := h.noteService.CreateNote(note); err != nil {
		if err == usecase.ErrInvalidParameters {
			http.Error(w, "Geçersiz parametreler", http.StatusBadRequest)
			return
		}
		http.Error(w, "Not oluşturma sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

// UpdateNote, bir notu günceller
func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Not ID'sini al
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz not ID'si", http.StatusBadRequest)
		return
	}

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Not güncelle
	note := &domain.Note{
		ID:       uint(id),
		Title:    req.Title,
		Content:  req.Content,
		UserID:   userID,
		Tags:     req.Tags,
		IsPublic: req.IsPublic,
	}

	if err := h.noteService.UpdateNote(note); err != nil {
		if err == usecase.ErrNoteNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrNotAuthorized {
			http.Error(w, "Bu işlem için yetkiniz yok", http.StatusForbidden)
			return
		}
		http.Error(w, "Not güncelleme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(note)
}

// DeleteNote, bir notu siler
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Not ID'sini al
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz not ID'si", http.StatusBadRequest)
		return
	}

	// Notu sil
	if err := h.noteService.DeleteNote(uint(id), userID); err != nil {
		if err == usecase.ErrNoteNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrNotAuthorized {
			http.Error(w, "Bu işlem için yetkiniz yok", http.StatusForbidden)
			return
		}
		http.Error(w, "Not silme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Not başarıyla silindi",
	})
}

// GetNote, bir notu getirir
func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	// Not ID'sini al
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz not ID'si", http.StatusBadRequest)
		return
	}

	// Notu getir
	note, err := h.noteService.GetNote(uint(id))
	if err != nil {
		if err == usecase.ErrNoteNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Not getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kullanıcı ID'sini al (opsiyonel)
	userID, ok := middleware.GetUserID(r)

	// Eğer not herkese açık değilse ve kullanıcı notu oluşturan değilse, erişimi reddet
	if !note.IsPublic && (!ok || note.UserID != userID) {
		http.Error(w, "Bu nota erişim izniniz yok", http.StatusForbidden)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(note)
}

// GetUserNotes, kullanıcının notlarını getirir
func (h *NoteHandler) GetUserNotes(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
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

	// Notları getir
	notes, err := h.noteService.GetUserNotes(userID, limit, offset)
	if err != nil {
		http.Error(w, "Notları getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

// GetPublicNotes, herkese açık notları getirir
func (h *NoteHandler) GetPublicNotes(w http.ResponseWriter, r *http.Request) {
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

	// Notları getir
	notes, err := h.noteService.GetPublicNotes(limit, offset)
	if err != nil {
		http.Error(w, "Notları getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

// SearchNotes, notları arar
func (h *NoteHandler) SearchNotes(w http.ResponseWriter, r *http.Request) {
	// Arama sorgusunu al
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Arama sorgusu gerekli", http.StatusBadRequest)
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

	// Notları ara
	notes, err := h.noteService.SearchNotes(query, limit, offset)
	if err != nil {
		if err == usecase.ErrInvalidParameters {
			http.Error(w, "Geçersiz parametreler", http.StatusBadRequest)
			return
		}
		http.Error(w, "Not arama sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

// GetNotesByTag, etikete göre notları getirir
func (h *NoteHandler) GetNotesByTag(w http.ResponseWriter, r *http.Request) {
	// Etiketi al
	tag := chi.URLParam(r, "tag")
	if tag == "" {
		http.Error(w, "Etiket gerekli", http.StatusBadRequest)
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

	// Notları getir
	notes, err := h.noteService.SearchNotes(tag, limit, offset)
	if err != nil {
		http.Error(w, "Notları getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

// AddComment, bir nota yorum ekler
func (h *NoteHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Not ID'sini al
	idStr := chi.URLParam(r, "id")
	noteID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz not ID'si", http.StatusBadRequest)
		return
	}

	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Yorum oluştur
	comment := &domain.Comment{
		NoteID:  uint(noteID),
		UserID:  userID,
		Content: req.Content,
	}

	// Yorumu ekle
	if err := h.noteService.AddComment(comment); err != nil {
		if err == usecase.ErrNoteNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Yorum ekleme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GetComments, bir notun yorumlarını getirir
func (h *NoteHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	// Not ID'sini al
	idStr := chi.URLParam(r, "id")
	noteID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz not ID'si", http.StatusBadRequest)
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

	// Yorumları getir
	comments, err := h.noteService.GetComments(uint(noteID), limit, offset)
	if err != nil {
		if err == usecase.ErrNoteNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Yorumları getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comments)
}

// LikeNote, bir notu beğenir
func (h *NoteHandler) LikeNote(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Not ID'sini al
	idStr := chi.URLParam(r, "id")
	noteID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz not ID'si", http.StatusBadRequest)
		return
	}

	// Notu beğen (like count'u artırır)
	if err := h.noteService.LikeNote(uint(noteID), userID); err != nil {
		if err == usecase.ErrNoteNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Not beğenme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Like record'u oluştur (beğeniler listesinde göstermek için)
	if err := h.likeService.LikeContent(userID, uint(noteID), "note"); err != nil {
		// Hata durumunda log yaz ama kullanıcıya hata gösterme
		// çünkü note count zaten artırıldı
		logger.Error("Beğeni kaydı oluşturulurken hata: %v", err)
	} else {
		logger.Info("Not beğeni kaydı oluşturuldu - UserID: %d, NoteID: %d", userID, noteID)
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Not başarıyla beğenildi",
	})
}

// UnlikeNote, bir notun beğenisini kaldırır
func (h *NoteHandler) UnlikeNote(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Not ID'sini al
	idStr := chi.URLParam(r, "id")
	noteID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz not ID'si", http.StatusBadRequest)
		return
	}

	// Not beğenisini kaldır (like count'u azaltır)
	if err := h.noteService.UnlikeNote(uint(noteID), userID); err != nil {
		if err == usecase.ErrNoteNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Not beğeni kaldırma sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Like record'unu sil (beğeniler listesinden kaldırmak için)
	if err := h.likeService.UnlikeContent(userID, uint(noteID), "note"); err != nil {
		// Hata durumunda log yaz ama kullanıcıya hata gösterme
		// çünkü note count zaten azaltıldı
		logger.Error("Beğeni kaydı silinirken hata: %v", err)
	} else {
		logger.Info("Not beğeni kaydı silindi - UserID: %d, NoteID: %d", userID, noteID)
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Not beğenisi başarıyla kaldırıldı",
	})
}

// GetLikedNotes, kullanıcının beğendiği notları getirir
func (h *NoteHandler) GetLikedNotes(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
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

	// Kullanıcının beğendiği notları doğrudan getir
	notes, err := h.likeService.GetLikedNotes(userID, limit, offset)
	if err != nil {
		http.Error(w, "Beğenilen notları getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}
