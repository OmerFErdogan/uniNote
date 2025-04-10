package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/OmerFErdogan/uninote/domain"
	"github.com/OmerFErdogan/uninote/infrastructure/http/middleware"
	"github.com/OmerFErdogan/uninote/infrastructure/logger"
	"github.com/OmerFErdogan/uninote/usecase"
	"github.com/go-chi/chi/v5"
)

// InviteHandler, davet bağlantısı işlemlerini yönetir
type InviteHandler struct {
	inviteService *usecase.InviteService
	noteService   *usecase.NoteService
	pdfService    *usecase.PDFService
}

// NewInviteHandler, yeni bir InviteHandler örneği oluşturur
func NewInviteHandler(inviteService *usecase.InviteService, noteService *usecase.NoteService, pdfService *usecase.PDFService) *InviteHandler {
	return &InviteHandler{
		inviteService: inviteService,
		noteService:   noteService,
		pdfService:    pdfService,
	}
}

// RegisterRoutes, yönlendirmeleri kaydeder
func (h *InviteHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	// Kimlik doğrulama gerektiren rotalar
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)
		r.Post("/notes/{id}/invites", h.CreateNoteInvite)
		r.Post("/pdfs/{id}/invites", h.CreatePDFInvite)
		r.Get("/notes/{id}/invites", h.GetNoteInvites)
		r.Get("/pdfs/{id}/invites", h.GetPDFInvites)
		r.Delete("/invites/{id}", h.DeactivateInvite)
	})

	// Kimlik doğrulama gerektirmeyen rotalar
	r.Get("/invites/{token}", h.ValidateInvite)
	r.Get("/notes/invite/{token}", h.GetNoteByInvite)
	r.Get("/pdfs/invite/{token}", h.GetPDFByInvite)
}

// CreateInviteRequest, davet bağlantısı oluşturma isteği
type CreateInviteRequest struct {
	ExpiresAt *time.Time `json:"expiresAt,omitempty"` // Opsiyonel, belirtilmezse 7 gün sonra sona erer
}

// InviteResponse, davet bağlantısı yanıtı
type InviteResponse struct {
	ID        uint      `json:"id"`
	ContentID uint      `json:"contentId"`
	Type      string    `json:"type"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateNoteInvite, bir not için davet bağlantısı oluşturur
func (h *InviteHandler) CreateNoteInvite(w http.ResponseWriter, r *http.Request) {
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

	var req CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Davet bağlantısı oluştur
	invite := &domain.Invite{
		ContentID: uint(noteID),
		Type:      "note",
		CreatedBy: userID,
	}

	// Opsiyonel sona erme tarihi
	if req.ExpiresAt != nil {
		invite.ExpiresAt = *req.ExpiresAt
	}

	// Daveti kaydet
	if err := h.inviteService.CreateInvite(invite); err != nil {
		if err == usecase.ErrContentNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrNotAuthorized {
			http.Error(w, "Bu işlem için yetkiniz yok", http.StatusForbidden)
			return
		}
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			return
		}
		http.Error(w, "Davet bağlantısı oluşturma sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	response := InviteResponse{
		ID:        invite.ID,
		ContentID: invite.ContentID,
		Type:      invite.Type,
		Token:     invite.Token,
		ExpiresAt: invite.ExpiresAt,
		IsActive:  invite.IsActive,
		CreatedAt: invite.CreatedAt,
	}

	logger.Info("Not için davet bağlantısı oluşturuldu - UserID: %d, NoteID: %d, Token: %s", userID, noteID, invite.Token)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CreatePDFInvite, bir PDF için davet bağlantısı oluşturur
func (h *InviteHandler) CreatePDFInvite(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// PDF ID'sini al
	idStr := chi.URLParam(r, "id")
	pdfID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz PDF ID'si", http.StatusBadRequest)
		return
	}

	var req CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Davet bağlantısı oluştur
	invite := &domain.Invite{
		ContentID: uint(pdfID),
		Type:      "pdf",
		CreatedBy: userID,
	}

	// Opsiyonel sona erme tarihi
	if req.ExpiresAt != nil {
		invite.ExpiresAt = *req.ExpiresAt
	}

	// Daveti kaydet
	if err := h.inviteService.CreateInvite(invite); err != nil {
		if err == usecase.ErrContentNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrNotAuthorized {
			http.Error(w, "Bu işlem için yetkiniz yok", http.StatusForbidden)
			return
		}
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			return
		}
		http.Error(w, "Davet bağlantısı oluşturma sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	response := InviteResponse{
		ID:        invite.ID,
		ContentID: invite.ContentID,
		Type:      invite.Type,
		Token:     invite.Token,
		ExpiresAt: invite.ExpiresAt,
		IsActive:  invite.IsActive,
		CreatedAt: invite.CreatedAt,
	}

	logger.Info("PDF için davet bağlantısı oluşturuldu - UserID: %d, PDFID: %d, Token: %s", userID, pdfID, invite.Token)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetNoteInvites, bir notun davet bağlantılarını getirir
func (h *InviteHandler) GetNoteInvites(w http.ResponseWriter, r *http.Request) {
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

	// Notun sahibi olup olmadığını kontrol et
	note, err := h.noteService.GetNote(uint(noteID))
	if err != nil {
		if err == usecase.ErrNoteNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Not getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if note.UserID != userID {
		http.Error(w, "Bu işlem için yetkiniz yok", http.StatusForbidden)
		return
	}

	// Davet bağlantılarını getir
	invites, err := h.inviteService.GetInvitesByContent(uint(noteID), "note")
	if err != nil {
		if err == usecase.ErrContentNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			return
		}
		http.Error(w, "Davet bağlantıları getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Yanıtı oluştur
	responses := make([]InviteResponse, len(invites))
	for i, invite := range invites {
		responses[i] = InviteResponse{
			ID:        invite.ID,
			ContentID: invite.ContentID,
			Type:      invite.Type,
			Token:     invite.Token,
			ExpiresAt: invite.ExpiresAt,
			IsActive:  invite.IsActive,
			CreatedAt: invite.CreatedAt,
		}
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

// GetPDFInvites, bir PDF'in davet bağlantılarını getirir
func (h *InviteHandler) GetPDFInvites(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// PDF ID'sini al
	idStr := chi.URLParam(r, "id")
	pdfID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz PDF ID'si", http.StatusBadRequest)
		return
	}

	// PDF'in sahibi olup olmadığını kontrol et
	pdf, err := h.pdfService.GetPDF(uint(pdfID))
	if err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "PDF getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if pdf.UserID != userID {
		http.Error(w, "Bu işlem için yetkiniz yok", http.StatusForbidden)
		return
	}

	// Davet bağlantılarını getir
	invites, err := h.inviteService.GetInvitesByContent(uint(pdfID), "pdf")
	if err != nil {
		if err == usecase.ErrContentNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrInvalidType {
			http.Error(w, "Geçersiz içerik türü", http.StatusBadRequest)
			return
		}
		http.Error(w, "Davet bağlantıları getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Yanıtı oluştur
	responses := make([]InviteResponse, len(invites))
	for i, invite := range invites {
		responses[i] = InviteResponse{
			ID:        invite.ID,
			ContentID: invite.ContentID,
			Type:      invite.Type,
			Token:     invite.Token,
			ExpiresAt: invite.ExpiresAt,
			IsActive:  invite.IsActive,
			CreatedAt: invite.CreatedAt,
		}
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

// DeactivateInvite, bir davet bağlantısını devre dışı bırakır
func (h *InviteHandler) DeactivateInvite(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Davet ID'sini al
	idStr := chi.URLParam(r, "id")
	inviteID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz davet ID'si", http.StatusBadRequest)
		return
	}

	// Daveti devre dışı bırak
	if err := h.inviteService.DeactivateInvite(uint(inviteID), userID); err != nil {
		if err == usecase.ErrInviteNotFound {
			http.Error(w, "Davet bağlantısı bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrNotAuthorized {
			http.Error(w, "Bu işlem için yetkiniz yok", http.StatusForbidden)
			return
		}
		http.Error(w, "Davet bağlantısı devre dışı bırakma sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("Davet bağlantısı devre dışı bırakıldı - UserID: %d, InviteID: %d", userID, inviteID)

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Davet bağlantısı başarıyla devre dışı bırakıldı",
	})
}

// ValidateInvite, bir davet bağlantısının geçerli olup olmadığını kontrol eder
func (h *InviteHandler) ValidateInvite(w http.ResponseWriter, r *http.Request) {
	// Token'ı al - önce URL parametresinden dene
	token := chi.URLParam(r, "token")

	// Eğer URL'de yoksa, header'dan kontrol et
	if token == "" {
		token = r.Header.Get("X-Invite-Token")
	}

	if token == "" {
		http.Error(w, "Geçersiz token", http.StatusBadRequest)
		return
	}

	// CORS header'larını ekle
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Invite-Token")

	// OPTIONS isteği ise hemen yanıt ver
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Daveti doğrula
	valid, invite, err := h.inviteService.ValidateInvite(token)
	if err != nil {
		if err == usecase.ErrInviteNotFound {
			http.Error(w, "Davet bağlantısı bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrInviteNotActive {
			http.Error(w, "Davet bağlantısı aktif değil", http.StatusForbidden)
			return
		}
		if err == usecase.ErrInviteExpired {
			http.Error(w, "Davet bağlantısı süresi dolmuş", http.StatusForbidden)
			return
		}
		if err == usecase.ErrContentNotFound {
			http.Error(w, "İçerik bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Davet bağlantısı doğrulama sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Content-Type header'ını ayarla
	w.Header().Set("Content-Type", "application/json")

	// Başarılı yanıt
	response := map[string]interface{}{
		"valid":     valid,
		"contentId": invite.ContentID,
		"type":      invite.Type,
		"expiresAt": invite.ExpiresAt,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetNoteByInvite, davet bağlantısı ile bir notu getirir
func (h *InviteHandler) GetNoteByInvite(w http.ResponseWriter, r *http.Request) {
	// Token'ı al - önce URL parametresinden dene
	token := chi.URLParam(r, "token")

	// Eğer URL'de yoksa, header'dan kontrol et
	if token == "" {
		token = r.Header.Get("X-Invite-Token")
	}

	if token == "" {
		http.Error(w, "Geçersiz token", http.StatusBadRequest)
		return
	}

	// CORS header'larını ekle
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Invite-Token")

	// OPTIONS isteği ise hemen yanıt ver
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Daveti doğrula
	valid, invite, err := h.inviteService.ValidateInvite(token)
	if err != nil {
		if err == usecase.ErrInviteNotFound {
			http.Error(w, "Davet bağlantısı bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrInviteNotActive {
			http.Error(w, "Davet bağlantısı aktif değil", http.StatusForbidden)
			return
		}
		if err == usecase.ErrInviteExpired {
			http.Error(w, "Davet bağlantısı süresi dolmuş", http.StatusForbidden)
			return
		}
		if err == usecase.ErrContentNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Davet bağlantısı doğrulama sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Geçersiz davet bağlantısı", http.StatusForbidden)
		return
	}

	if invite.Type != "note" {
		http.Error(w, "Bu davet bağlantısı bir not için değil", http.StatusBadRequest)
		return
	}

	// Davet bağlantısı ile erişim için özel fonksiyonu kullan
	// Bu fonksiyon erişim kontrolü yapmadan doğrudan notu getirir
	note, err := h.noteService.GetNoteByInviteToken(invite.ContentID)
	if err != nil {
		if err == usecase.ErrNoteNotFound {
			http.Error(w, "Not bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Not getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Content-Type header'ını ayarla
	w.Header().Set("Content-Type", "application/json")

	// Davet bağlantısı ile erişim sağlandığı için, özel notlara da erişim izni var
	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(note)
}

// GetPDFByInvite, davet bağlantısı ile bir PDF'i getirir
func (h *InviteHandler) GetPDFByInvite(w http.ResponseWriter, r *http.Request) {
	// Token'ı al - önce URL parametresinden dene
	token := chi.URLParam(r, "token")

	// Eğer URL'de yoksa, header'dan kontrol et
	if token == "" {
		token = r.Header.Get("X-Invite-Token")
	}

	if token == "" {
		http.Error(w, "Geçersiz token", http.StatusBadRequest)
		return
	}

	// CORS header'larını ekle
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Invite-Token")

	// OPTIONS isteği ise hemen yanıt ver
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Daveti doğrula
	valid, invite, err := h.inviteService.ValidateInvite(token)
	if err != nil {
		if err == usecase.ErrInviteNotFound {
			http.Error(w, "Davet bağlantısı bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrInviteNotActive {
			http.Error(w, "Davet bağlantısı aktif değil", http.StatusForbidden)
			return
		}
		if err == usecase.ErrInviteExpired {
			http.Error(w, "Davet bağlantısı süresi dolmuş", http.StatusForbidden)
			return
		}
		if err == usecase.ErrContentNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Davet bağlantısı doğrulama sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Geçersiz davet bağlantısı", http.StatusForbidden)
		return
	}

	if invite.Type != "pdf" {
		http.Error(w, "Bu davet bağlantısı bir PDF için değil", http.StatusBadRequest)
		return
	}

	// Davet bağlantısı ile erişim için özel fonksiyonu kullan
	// Bu fonksiyon erişim kontrolü yapmadan doğrudan PDF'i getirir
	pdf, err := h.pdfService.GetPDFByInviteToken(invite.ContentID)
	if err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "PDF getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Content-Type header'ını ayarla
	w.Header().Set("Content-Type", "application/json")

	// Davet bağlantısı ile erişim sağlandığı için, özel PDF'lere de erişim izni var
	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pdf)
}
