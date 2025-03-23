package usecase

import (
	"errors"
	"fmt"

	"github.com/OmerFErdogan/uninote/domain"
)

var (
	ErrContentNotFound = errors.New("içerik bulunamadı")
	ErrInvalidType     = errors.New("geçersiz içerik türü")
)

// LikeService, beğeni ile ilgili iş mantığını içerir
type LikeService struct {
	likeRepo domain.LikeRepository
	noteRepo domain.NoteRepository
	pdfRepo  domain.PDFRepository
}

// NewLikeService, yeni bir LikeService örneği oluşturur
func NewLikeService(likeRepo domain.LikeRepository, noteRepo domain.NoteRepository, pdfRepo domain.PDFRepository) *LikeService {
	return &LikeService{
		likeRepo: likeRepo,
		noteRepo: noteRepo,
		pdfRepo:  pdfRepo,
	}
}

// LikeContent, bir içeriği beğenir
func (s *LikeService) LikeContent(userID, contentID uint, contentType string) error {
	// İçerik türünü kontrol et
	if contentType != "note" && contentType != "pdf" {
		return ErrInvalidType
	}

	// İçeriğin var olup olmadığını kontrol et
	if contentType == "note" {
		note, err := s.noteRepo.FindByID(contentID)
		if err != nil {
			return fmt.Errorf("not arama sırasında hata: %w", err)
		}
		if note == nil {
			return ErrContentNotFound
		}
	} else if contentType == "pdf" {
		pdf, err := s.pdfRepo.FindByID(contentID)
		if err != nil {
			return fmt.Errorf("PDF arama sırasında hata: %w", err)
		}
		if pdf == nil {
			return ErrContentNotFound
		}
	}

	// Kullanıcının içeriği daha önce beğenip beğenmediğini kontrol et
	existingLike, err := s.likeRepo.FindByUserIDAndContent(userID, contentID, contentType)
	if err != nil {
		return fmt.Errorf("beğeni arama sırasında hata: %w", err)
	}
	if existingLike != nil {
		// Kullanıcı zaten içeriği beğenmiş
		return nil
	}

	// Beğeni oluştur
	like := &domain.Like{
		UserID:    userID,
		ContentID: contentID,
		Type:      contentType,
	}

	return s.likeRepo.Create(like)
}

// UnlikeContent, bir içeriğin beğenisini kaldırır
func (s *LikeService) UnlikeContent(userID, contentID uint, contentType string) error {
	// İçerik türünü kontrol et
	if contentType != "note" && contentType != "pdf" {
		return ErrInvalidType
	}

	// İçeriğin var olup olmadığını kontrol et
	if contentType == "note" {
		note, err := s.noteRepo.FindByID(contentID)
		if err != nil {
			return fmt.Errorf("not arama sırasında hata: %w", err)
		}
		if note == nil {
			return ErrContentNotFound
		}
	} else if contentType == "pdf" {
		pdf, err := s.pdfRepo.FindByID(contentID)
		if err != nil {
			return fmt.Errorf("PDF arama sırasında hata: %w", err)
		}
		if pdf == nil {
			return ErrContentNotFound
		}
	}

	// Beğeniyi kaldır
	return s.likeRepo.DeleteByUserIDAndContent(userID, contentID, contentType)
}

// GetUserLikes, bir kullanıcının beğenilerini getirir
func (s *LikeService) GetUserLikes(userID uint, limit, offset int) ([]*domain.Like, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.likeRepo.FindByUserID(userID, limit, offset)
}

// GetContentLikes, bir içeriğin beğenilerini getirir
func (s *LikeService) GetContentLikes(contentID uint, contentType string, limit, offset int) ([]*domain.Like, error) {
	// İçerik türünü kontrol et
	if contentType != "note" && contentType != "pdf" {
		return nil, ErrInvalidType
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.likeRepo.FindByContentID(contentID, contentType, limit, offset)
}

// IsLikedByUser, bir içeriğin kullanıcı tarafından beğenilip beğenilmediğini kontrol eder
func (s *LikeService) IsLikedByUser(userID, contentID uint, contentType string) (bool, error) {
	// İçerik türünü kontrol et
	if contentType != "note" && contentType != "pdf" {
		return false, ErrInvalidType
	}

	like, err := s.likeRepo.FindByUserIDAndContent(userID, contentID, contentType)
	if err != nil {
		return false, fmt.Errorf("beğeni arama sırasında hata: %w", err)
	}

	return like != nil, nil
}

// GetLikedNotes, kullanıcının beğendiği notları getirir
func (s *LikeService) GetLikedNotes(userID uint, limit, offset int) ([]*domain.Note, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.likeRepo.FindLikedNotesByUserID(userID, limit, offset)
}

// GetLikedPDFs, kullanıcının beğendiği PDF'leri getirir
func (s *LikeService) GetLikedPDFs(userID uint, limit, offset int) ([]*domain.PDF, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.likeRepo.FindLikedPDFsByUserID(userID, limit, offset)
}

// Ensure LikeService implements domain.LikeService
var _ domain.LikeService = (*LikeService)(nil)
