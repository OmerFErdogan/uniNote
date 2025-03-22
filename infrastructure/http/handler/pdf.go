package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/OmerFErdogan/uninote/domain"
	"github.com/OmerFErdogan/uninote/infrastructure/http/middleware"
	"github.com/OmerFErdogan/uninote/usecase"
	"github.com/go-chi/chi/v5"
)

// PDFHandler, PDF işlemlerini yönetir
type PDFHandler struct {
	pdfService  *usecase.PDFService
	likeService *usecase.LikeService
}

// NewPDFHandler, yeni bir PDFHandler örneği oluşturur
func NewPDFHandler(pdfService *usecase.PDFService, likeService *usecase.LikeService) *PDFHandler {
	return &PDFHandler{
		pdfService:  pdfService,
		likeService: likeService,
	}
}

// RegisterRoutes, yönlendirmeleri kaydeder
func (h *PDFHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	// Kimlik doğrulama gerektiren rotalar
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)
		r.Post("/pdfs", h.UploadPDF)
		r.Put("/pdfs/{id}", h.UpdatePDF)
		r.Delete("/pdfs/{id}", h.DeletePDF)
		r.Get("/pdfs/my", h.GetUserPDFs)
		r.Post("/pdfs/{id}/comments", h.AddComment)
		r.Post("/pdfs/{id}/annotations", h.AddAnnotation)
		r.Get("/pdfs/{id}/annotations", h.GetAnnotations)
		r.Post("/pdfs/{id}/like", h.LikePDF)
		r.Delete("/pdfs/{id}/like", h.UnlikePDF)
		r.Get("/pdfs/liked", h.GetLikedPDFs)
	})

	// Kimlik doğrulama gerektirmeyen rotalar
	r.Get("/pdfs", h.GetPublicPDFs)
	r.Get("/pdfs/{id}", func(w http.ResponseWriter, r *http.Request) {
		middleware.OptionalAuth(authMiddleware, h.GetPDF).ServeHTTP(w, r)
	})
	r.Get("/pdfs/{id}/content", h.GetPDFContent)
	r.Get("/pdfs/{id}/comments", h.GetComments)
	r.Get("/pdfs/search", h.SearchPDFs)
	r.Get("/pdfs/tag/{tag}", h.GetPDFsByTag)
}

// UploadPDFRequest, PDF yükleme isteği
type UploadPDFRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	IsPublic    bool     `json:"isPublic"`
}

// UpdatePDFRequest, PDF güncelleme isteği
type UpdatePDFRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	IsPublic    bool     `json:"isPublic"`
}

// CommentRequest, yorum isteği
type PDFCommentRequest struct {
	Content    string `json:"content"`
	PageNumber int    `json:"pageNumber"`
}

// AnnotationRequest, işaretleme isteği
type AnnotationRequest struct {
	PageNumber int     `json:"pageNumber"`
	Content    string  `json:"content"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	Type       string  `json:"type"`
	Color      string  `json:"color"`
}

// UploadPDF, yeni bir PDF yükler
func (h *PDFHandler) UploadPDF(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Multipart form'u parse et
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Form parse hatası: "+err.Error(), http.StatusBadRequest)
		return
	}

	// PDF dosyasını al
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Dosya yükleme hatası: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Dosya içeriğini oku
	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Dosya okuma hatası: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Form verilerini al
	title := r.FormValue("title")
	description := r.FormValue("description")
	tagsStr := r.FormValue("tags")
	isPublicStr := r.FormValue("isPublic")

	// Etiketleri parse et
	var tags []string
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
			http.Error(w, "Geçersiz etiket formatı", http.StatusBadRequest)
			return
		}
	}

	// IsPublic'i parse et
	isPublic := false
	if isPublicStr == "true" {
		isPublic = true
	}

	// PDF oluştur
	pdf := &domain.PDF{
		Title:       title,
		Description: description,
		UserID:      userID,
		Tags:        tags,
		IsPublic:    isPublic,
	}

	// PDF'i yükle
	if err := h.pdfService.UploadPDF(pdf, fileContent); err != nil {
		if err == usecase.ErrInvalidParameters {
			http.Error(w, "Geçersiz parametreler", http.StatusBadRequest)
			return
		}
		http.Error(w, "PDF yükleme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pdf)
}

// UpdatePDF, bir PDF'i günceller
func (h *PDFHandler) UpdatePDF(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// PDF ID'sini al
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz PDF ID'si", http.StatusBadRequest)
		return
	}

	var req UpdatePDFRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// PDF güncelle
	pdf := &domain.PDF{
		ID:          uint(id),
		Title:       req.Title,
		Description: req.Description,
		UserID:      userID,
		Tags:        req.Tags,
		IsPublic:    req.IsPublic,
	}

	if err := h.pdfService.UpdatePDF(pdf); err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrNotAuthorized {
			http.Error(w, "Bu işlem için yetkiniz yok", http.StatusForbidden)
			return
		}
		http.Error(w, "PDF güncelleme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pdf)
}

// DeletePDF, bir PDF'i siler
func (h *PDFHandler) DeletePDF(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// PDF ID'sini al
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz PDF ID'si", http.StatusBadRequest)
		return
	}

	// PDF'i sil
	if err := h.pdfService.DeletePDF(uint(id), userID); err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		if err == usecase.ErrNotAuthorized {
			http.Error(w, "Bu işlem için yetkiniz yok", http.StatusForbidden)
			return
		}
		http.Error(w, "PDF silme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "PDF başarıyla silindi",
	})
}

// GetPDF, bir PDF'i getirir
func (h *PDFHandler) GetPDF(w http.ResponseWriter, r *http.Request) {
	// PDF ID'sini al
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz PDF ID'si", http.StatusBadRequest)
		return
	}

	// PDF'i getir
	pdf, err := h.pdfService.GetPDF(uint(id))
	if err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "PDF getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kullanıcı ID'sini al (opsiyonel)
	userID, ok := middleware.GetUserID(r)

	// Eğer PDF herkese açık değilse ve kullanıcı PDF'i yükleyen değilse, erişimi reddet
	if !pdf.IsPublic && (!ok || pdf.UserID != userID) {
		http.Error(w, "Bu PDF'e erişim izniniz yok", http.StatusForbidden)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pdf)
}

// GetPDFContent, bir PDF'in içeriğini getirir
func (h *PDFHandler) GetPDFContent(w http.ResponseWriter, r *http.Request) {
	// PDF ID'sini al
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz PDF ID'si", http.StatusBadRequest)
		return
	}

	// PDF'i getir
	pdf, err := h.pdfService.GetPDF(uint(id))
	if err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "PDF getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kullanıcı ID'sini al (opsiyonel)
	userID, ok := middleware.GetUserID(r)

	// Eğer PDF herkese açık değilse ve kullanıcı PDF'i yükleyen değilse, erişimi reddet
	if !pdf.IsPublic && (!ok || pdf.UserID != userID) {
		http.Error(w, "Bu PDF'e erişim izniniz yok", http.StatusForbidden)
		return
	}

	// PDF içeriğini getir
	content, err := h.pdfService.GetPDFContent(uint(id))
	if err != nil {
		http.Error(w, "PDF içeriği getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// PDF yanıtı
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename="+pdf.Title+".pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// GetUserPDFs, kullanıcının PDF'lerini getirir
func (h *PDFHandler) GetUserPDFs(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Sayfalama parametrelerini al
	limit, offset := getPaginationParams(r)

	// PDF'leri getir
	pdfs, err := h.pdfService.GetUserPDFs(userID, limit, offset)
	if err != nil {
		http.Error(w, "PDF'leri getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pdfs)
}

// GetPublicPDFs, herkese açık PDF'leri getirir
func (h *PDFHandler) GetPublicPDFs(w http.ResponseWriter, r *http.Request) {
	// Sayfalama parametrelerini al
	limit, offset := getPaginationParams(r)

	// PDF'leri getir
	pdfs, err := h.pdfService.GetPublicPDFs(limit, offset)
	if err != nil {
		http.Error(w, "PDF'leri getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pdfs)
}

// SearchPDFs, PDF'leri arar
func (h *PDFHandler) SearchPDFs(w http.ResponseWriter, r *http.Request) {
	// Arama sorgusunu al
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Arama sorgusu gerekli", http.StatusBadRequest)
		return
	}

	// Sayfalama parametrelerini al
	limit, offset := getPaginationParams(r)

	// PDF'leri ara
	pdfs, err := h.pdfService.SearchPDFs(query, limit, offset)
	if err != nil {
		if err == usecase.ErrInvalidParameters {
			http.Error(w, "Geçersiz parametreler", http.StatusBadRequest)
			return
		}
		http.Error(w, "PDF arama sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pdfs)
}

// GetPDFsByTag, etikete göre PDF'leri getirir
func (h *PDFHandler) GetPDFsByTag(w http.ResponseWriter, r *http.Request) {
	// Etiketi al
	tag := chi.URLParam(r, "tag")
	if tag == "" {
		http.Error(w, "Etiket gerekli", http.StatusBadRequest)
		return
	}

	// Sayfalama parametrelerini al
	limit, offset := getPaginationParams(r)

	// PDF'leri getir
	pdfs, err := h.pdfService.SearchPDFs(tag, limit, offset)
	if err != nil {
		http.Error(w, "PDF'leri getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pdfs)
}

// AddComment, bir PDF'e yorum ekler
func (h *PDFHandler) AddComment(w http.ResponseWriter, r *http.Request) {
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

	var req PDFCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Yorum oluştur
	comment := &domain.PDFComment{
		PDFID:      uint(pdfID),
		UserID:     userID,
		Content:    req.Content,
		PageNumber: req.PageNumber,
	}

	// Yorumu ekle
	if err := h.pdfService.AddComment(comment); err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Yorum ekleme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GetComments, bir PDF'in yorumlarını getirir
func (h *PDFHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	// PDF ID'sini al
	idStr := chi.URLParam(r, "id")
	pdfID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Geçersiz PDF ID'si", http.StatusBadRequest)
		return
	}

	// Sayfalama parametrelerini al
	limit, offset := getPaginationParams(r)

	// Yorumları getir
	comments, err := h.pdfService.GetComments(uint(pdfID), limit, offset)
	if err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "Yorumları getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comments)
}

// AddAnnotation, bir PDF'e işaretleme ekler
func (h *PDFHandler) AddAnnotation(w http.ResponseWriter, r *http.Request) {
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

	var req AnnotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// İşaretleme oluştur
	annotation := &domain.PDFAnnotation{
		PDFID:      uint(pdfID),
		UserID:     userID,
		PageNumber: req.PageNumber,
		Content:    req.Content,
		X:          req.X,
		Y:          req.Y,
		Width:      req.Width,
		Height:     req.Height,
		Type:       req.Type,
		Color:      req.Color,
	}

	// İşaretlemeyi ekle
	if err := h.pdfService.AddAnnotation(annotation); err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "İşaretleme ekleme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(annotation)
}

// GetAnnotations, bir PDF'in işaretlemelerini getirir
func (h *PDFHandler) GetAnnotations(w http.ResponseWriter, r *http.Request) {
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

	// İşaretlemeleri getir
	annotations, err := h.pdfService.GetAnnotations(uint(pdfID), userID)
	if err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "İşaretlemeleri getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(annotations)
}

// LikePDF, bir PDF'i beğenir
func (h *PDFHandler) LikePDF(w http.ResponseWriter, r *http.Request) {
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

	// PDF'i beğen
	if err := h.pdfService.LikePDF(uint(pdfID), userID); err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "PDF beğenme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "PDF başarıyla beğenildi",
	})
}

// UnlikePDF, bir PDF'in beğenisini kaldırır
func (h *PDFHandler) UnlikePDF(w http.ResponseWriter, r *http.Request) {
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

	// PDF beğenisini kaldır
	if err := h.pdfService.UnlikePDF(uint(pdfID), userID); err != nil {
		if err == usecase.ErrPDFNotFound {
			http.Error(w, "PDF bulunamadı", http.StatusNotFound)
			return
		}
		http.Error(w, "PDF beğeni kaldırma sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "PDF beğenisi başarıyla kaldırıldı",
	})
}

// GetLikedPDFs, kullanıcının beğendiği PDF'leri getirir
func (h *PDFHandler) GetLikedPDFs(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini al
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Kullanıcı kimliği bulunamadı", http.StatusUnauthorized)
		return
	}

	// Sayfalama parametrelerini al
	limit, offset := getPaginationParams(r)

	// Kullanıcının beğenilerini getir
	likes, err := h.likeService.GetUserLikes(userID, limit, offset)
	if err != nil {
		http.Error(w, "Beğenileri getirme sırasında hata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// PDF türündeki beğenileri filtrele
	var pdfIDs []uint
	for _, like := range likes {
		if like.Type == "pdf" {
			pdfIDs = append(pdfIDs, like.ContentID)
		}
	}

	// Beğenilen PDF yoksa boş dizi döndür
	if len(pdfIDs) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]*domain.PDF{})
		return
	}

	// Beğenilen PDF'leri getir
	pdfs := make([]*domain.PDF, 0, len(pdfIDs))
	for _, pdfID := range pdfIDs {
		pdf, err := h.pdfService.GetPDF(pdfID)
		if err == nil && pdf != nil {
			pdfs = append(pdfs, pdf)
		}
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pdfs)
}

// getPaginationParams, sayfalama parametrelerini alır
func getPaginationParams(r *http.Request) (int, int) {
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

	return limit, offset
}
