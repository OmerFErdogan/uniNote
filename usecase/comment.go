package usecase

import (
	"errors"

	"github.com/OmerFErdogan/uninote/domain"
)

var (
	// Yorum servisi için hata değişkenleri
	commentErrNoteNotFound = errors.New("not bulunamadı")
	commentErrPDFNotFound  = errors.New("PDF bulunamadı")
)

// CommentService, yorum ile ilgili iş mantığını içerir
type CommentService struct {
	noteRepo       domain.NoteRepository
	commentRepo    domain.CommentRepository
	pdfRepo        domain.PDFRepository
	pdfCommentRepo domain.PDFCommentRepository
	userRepo       domain.UserRepository
}

// NewCommentService, yeni bir CommentService örneği oluşturur
func NewCommentService(
	noteRepo domain.NoteRepository,
	commentRepo domain.CommentRepository,
	pdfRepo domain.PDFRepository,
	pdfCommentRepo domain.PDFCommentRepository,
	userRepo domain.UserRepository,
) *CommentService {
	return &CommentService{
		noteRepo:       noteRepo,
		commentRepo:    commentRepo,
		pdfRepo:        pdfRepo,
		pdfCommentRepo: pdfCommentRepo,
		userRepo:       userRepo,
	}
}

// EnrichNoteComments, not yorumlarını kullanıcı bilgileriyle zenginleştirir
func (s *CommentService) EnrichNoteComments(comments []*domain.Comment) ([]*domain.CommentResponse, error) {
	var enrichedComments []*domain.CommentResponse

	for _, comment := range comments {
		// Kullanıcı bilgilerini getir
		user, err := s.userRepo.FindByID(comment.UserID)
		if err != nil {
			return nil, err
		}

		// Kullanıcı bulunamadıysa, varsayılan değerler kullan
		username := "Silinmiş Kullanıcı"
		fullName := "Silinmiş Kullanıcı"
		if user != nil {
			username = user.Username
			fullName = user.FirstName + " " + user.LastName
		}

		// Zenginleştirilmiş yorum yanıtı oluştur
		enrichedComment := &domain.CommentResponse{
			ID:        comment.ID,
			ContentID: comment.NoteID,
			UserID:    comment.UserID,
			Username:  username,
			FullName:  fullName,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		}

		enrichedComments = append(enrichedComments, enrichedComment)
	}

	return enrichedComments, nil
}

// EnrichPDFComments, PDF yorumlarını kullanıcı bilgileriyle zenginleştirir
func (s *CommentService) EnrichPDFComments(comments []*domain.PDFComment) ([]*domain.CommentResponse, error) {
	var enrichedComments []*domain.CommentResponse

	for _, comment := range comments {
		// Kullanıcı bilgilerini getir
		user, err := s.userRepo.FindByID(comment.UserID)
		if err != nil {
			return nil, err
		}

		// Kullanıcı bulunamadıysa, varsayılan değerler kullan
		username := "Silinmiş Kullanıcı"
		fullName := "Silinmiş Kullanıcı"
		if user != nil {
			username = user.Username
			fullName = user.FirstName + " " + user.LastName
		}

		// Zenginleştirilmiş yorum yanıtı oluştur
		enrichedComment := &domain.CommentResponse{
			ID:         comment.ID,
			ContentID:  comment.PDFID,
			UserID:     comment.UserID,
			Username:   username,
			FullName:   fullName,
			Content:    comment.Content,
			PageNumber: comment.PageNumber,
			CreatedAt:  comment.CreatedAt,
			UpdatedAt:  comment.UpdatedAt,
		}

		enrichedComments = append(enrichedComments, enrichedComment)
	}

	return enrichedComments, nil
}

// GetNoteComments, bir notun yorumlarını getirir ve kullanıcı bilgileriyle zenginleştirir
func (s *CommentService) GetNoteComments(noteID uint, limit, offset int) ([]*domain.CommentResponse, error) {
	// Notu bul
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, commentErrNoteNotFound
	}

	// Yorumları getir
	comments, err := s.commentRepo.FindByNoteID(noteID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Yorumları zenginleştir
	return s.EnrichNoteComments(comments)
}

// GetPDFComments, bir PDF'in yorumlarını getirir ve kullanıcı bilgileriyle zenginleştirir
func (s *CommentService) GetPDFComments(pdfID uint, limit, offset int) ([]*domain.CommentResponse, error) {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(pdfID)
	if err != nil {
		return nil, err
	}
	if pdf == nil {
		return nil, commentErrPDFNotFound
	}

	// Yorumları getir
	comments, err := s.pdfCommentRepo.FindByPDFID(pdfID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Yorumları zenginleştir
	return s.EnrichPDFComments(comments)
}

// CheckNoteAccess, bir nota erişim izni olup olmadığını kontrol eder
func (s *CommentService) CheckNoteAccess(noteID, userID uint) (bool, error) {
	// Notu bul
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return false, err
	}
	if note == nil {
		return false, commentErrNoteNotFound
	}

	// Not herkese açıksa veya kullanıcı notun sahibiyse erişim izni var
	return note.IsPublic || note.UserID == userID, nil
}

// CheckPDFAccess, bir PDF'e erişim izni olup olmadığını kontrol eder
func (s *CommentService) CheckPDFAccess(pdfID, userID uint) (bool, error) {
	// PDF'i bul
	pdf, err := s.pdfRepo.FindByID(pdfID)
	if err != nil {
		return false, err
	}
	if pdf == nil {
		return false, commentErrPDFNotFound
	}

	// PDF herkese açıksa veya kullanıcı PDF'in sahibiyse erişim izni var
	return pdf.IsPublic || pdf.UserID == userID, nil
}
