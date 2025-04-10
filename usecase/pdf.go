package usecase

import (
	"errors"
	"fmt"

	"github.com/OmerFErdogan/uninote/domain"
)

var (
	ErrPDFNotFound = errors.New("PDF bulunamadı")
	ErrFileStorage = errors.New("dosya depolama hatası")
)

// PDFService, PDF ile ilgili iş mantığını içerir
type PDFService struct {
	pdfRepo        domain.PDFRepository
	pdfCommentRepo domain.PDFCommentRepository
	pdfAnnotRepo   domain.PDFAnnotationRepository
	pdfStorage     domain.PDFStorage
}

// NewPDFService, yeni bir PDFService örneği oluşturur
func NewPDFService(
	pdfRepo domain.PDFRepository,
	pdfCommentRepo domain.PDFCommentRepository,
	pdfAnnotRepo domain.PDFAnnotationRepository,
	pdfStorage domain.PDFStorage,
) *PDFService {
	return &PDFService{
		pdfRepo:        pdfRepo,
		pdfCommentRepo: pdfCommentRepo,
		pdfAnnotRepo:   pdfAnnotRepo,
		pdfStorage:     pdfStorage,
	}
}

// UploadPDF, yeni bir PDF yükler
func (s *PDFService) UploadPDF(pdf *domain.PDF, fileContent []byte) error {
	if pdf.Title == "" || len(fileContent) == 0 {
		return ErrInvalidParameters
	}

	// Dosya boyutunu ayarla
	pdf.FileSize = int64(len(fileContent))

	// Dosya adını oluştur (ID henüz yok, bu yüzden timestamp ve kullanıcı ID'si kullanılabilir)
	fileName := fmt.Sprintf("%d_%s.pdf", pdf.UserID, pdf.Title)

	// Dosyayı kaydet
	filePath, err := s.pdfStorage.Save(fileContent, fileName)
	if err != nil {
		return fmt.Errorf("dosya kaydetme hatası: %w", err)
	}

	// Dosya yolunu PDF nesnesine ekle
	pdf.FilePath = filePath

	// PDF'i veritabanına kaydet
	return s.pdfRepo.Create(pdf)
}

// UpdatePDF, bir PDF'i günceller
func (s *PDFService) UpdatePDF(pdf *domain.PDF) error {
	// PDF'i bul
	existingPDF, err := s.pdfRepo.FindByID(pdf.ID)
	if err != nil {
		return fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if existingPDF == nil {
		return ErrPDFNotFound
	}

	// Kullanıcı yetkisi kontrol et
	if existingPDF.UserID != pdf.UserID {
		return ErrNotAuthorized
	}

	// Dosya yolunu koru
	pdf.FilePath = existingPDF.FilePath
	pdf.FileSize = existingPDF.FileSize

	// PDF'i güncelle
	return s.pdfRepo.Update(pdf)
}

// DeletePDF, bir PDF'i siler
func (s *PDFService) DeletePDF(id uint, userID uint) error {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return ErrPDFNotFound
	}

	// Kullanıcı yetkisi kontrol et
	if pdf.UserID != userID {
		return ErrNotAuthorized
	}

	// Dosyayı sil
	if err := s.pdfStorage.Delete(pdf.FilePath); err != nil {
		return fmt.Errorf("dosya silme hatası: %w", err)
	}

	// PDF'i veritabanından sil
	return s.pdfRepo.Delete(id)
}

// GetPDF, bir PDF'i getirir
func (s *PDFService) GetPDF(id uint) (*domain.PDF, error) {
	pdf, err := s.pdfRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return nil, ErrPDFNotFound
	}

	// Görüntülenme sayısını artır
	s.pdfRepo.IncrementViewCount(id)

	return pdf, nil
}

// GetPDFContent, bir PDF'in içeriğini getirir
func (s *PDFService) GetPDFContent(id uint) ([]byte, error) {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return nil, ErrPDFNotFound
	}

	// Dosyayı oku
	content, err := s.pdfStorage.Get(pdf.FilePath)
	if err != nil {
		return nil, fmt.Errorf("dosya okuma hatası: %w", err)
	}

	return content, nil
}

// GetUserPDFs, bir kullanıcının PDF'lerini getirir
func (s *PDFService) GetUserPDFs(userID uint, limit, offset int) ([]*domain.PDF, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.pdfRepo.FindByUserID(userID, limit, offset)
}

// GetPublicPDFs, herkese açık PDF'leri getirir
func (s *PDFService) GetPublicPDFs(limit, offset int) ([]*domain.PDF, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.pdfRepo.FindPublic(limit, offset)
}

// SearchPDFs, PDF'leri arar
func (s *PDFService) SearchPDFs(query string, limit, offset int) ([]*domain.PDF, error) {
	if query == "" {
		return nil, ErrInvalidParameters
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.pdfRepo.Search(query, limit, offset)
}

// AddComment, bir PDF'e yorum ekler
func (s *PDFService) AddComment(comment *domain.PDFComment) error {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(comment.PDFID)
	if err != nil {
		return fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return ErrPDFNotFound
	}

	// Yorumu ekle
	return s.pdfCommentRepo.Create(comment)
}

// GetComments, bir PDF'in yorumlarını getirir
func (s *PDFService) GetComments(pdfID uint, limit, offset int) ([]*domain.PDFComment, error) {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(pdfID)
	if err != nil {
		return nil, fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return nil, ErrPDFNotFound
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.pdfCommentRepo.FindByPDFID(pdfID, limit, offset)
}

// AddAnnotation, bir PDF'e işaretleme ekler
func (s *PDFService) AddAnnotation(annotation *domain.PDFAnnotation) error {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(annotation.PDFID)
	if err != nil {
		return fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return ErrPDFNotFound
	}

	// İşaretlemeyi ekle
	return s.pdfAnnotRepo.Create(annotation)
}

// GetAnnotations, bir PDF'in işaretlemelerini getirir
func (s *PDFService) GetAnnotations(pdfID uint, userID uint) ([]*domain.PDFAnnotation, error) {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(pdfID)
	if err != nil {
		return nil, fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return nil, ErrPDFNotFound
	}

	return s.pdfAnnotRepo.FindByPDFIDAndUserID(pdfID, userID)
}

// LikePDF, bir PDF'i beğenir
func (s *PDFService) LikePDF(pdfID uint, userID uint) error {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(pdfID)
	if err != nil {
		return fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return ErrPDFNotFound
	}

	// Beğeni sayısını artır
	return s.pdfRepo.IncrementLikeCount(pdfID)
}

// UnlikePDF, bir PDF'in beğenisini kaldırır
func (s *PDFService) UnlikePDF(pdfID uint, userID uint) error {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(pdfID)
	if err != nil {
		return fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return ErrPDFNotFound
	}

	// Beğeni sayısını azalt
	return s.pdfRepo.DecrementLikeCount(pdfID)
}

// GetPDFByInviteToken, davet bağlantısı ile erişim için bir PDF'i getirir
// Bu fonksiyon erişim kontrolü yapmadan doğrudan PDF'i getirir
func (s *PDFService) GetPDFByInviteToken(id uint) (*domain.PDF, error) {
	pdf, err := s.pdfRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("PDF arama sırasında hata: %w", err)
	}
	if pdf == nil {
		return nil, ErrPDFNotFound
	}

	// Görüntülenme sayısını artır
	s.pdfRepo.IncrementViewCount(id)

	return pdf, nil
}

// Ensure PDFService implements domain.PDFService
var _ domain.PDFService = (*PDFService)(nil)
