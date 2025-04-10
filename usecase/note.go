package usecase

import (
	"errors"
	"fmt"

	"github.com/OmerFErdogan/uninote/domain"
)

var (
	ErrNoteNotFound      = errors.New("not bulunamadı")
	ErrNotAuthorized     = errors.New("bu işlem için yetkiniz yok")
	ErrCommentNotFound   = errors.New("yorum bulunamadı")
	ErrInvalidParameters = errors.New("geçersiz parametreler")
)

// NoteService, not ile ilgili iş mantığını içerir
type NoteService struct {
	noteRepo    domain.NoteRepository
	commentRepo domain.CommentRepository
}

// NewNoteService, yeni bir NoteService örneği oluşturur
func NewNoteService(noteRepo domain.NoteRepository, commentRepo domain.CommentRepository) *NoteService {
	return &NoteService{
		noteRepo:    noteRepo,
		commentRepo: commentRepo,
	}
}

// CreateNote, yeni bir not oluşturur
func (s *NoteService) CreateNote(note *domain.Note) error {
	if note.Title == "" {
		return ErrInvalidParameters
	}

	return s.noteRepo.Create(note)
}

// UpdateNote, bir notu günceller
func (s *NoteService) UpdateNote(note *domain.Note) error {
	// Notu bul
	existingNote, err := s.noteRepo.FindByID(note.ID)
	if err != nil {
		return fmt.Errorf("not arama sırasında hata: %w", err)
	}
	if existingNote == nil {
		return ErrNoteNotFound
	}

	// Kullanıcı yetkisi kontrol et
	if existingNote.UserID != note.UserID {
		return ErrNotAuthorized
	}

	// Notu güncelle
	return s.noteRepo.Update(note)
}

// DeleteNote, bir notu siler
func (s *NoteService) DeleteNote(id uint, userID uint) error {
	// Notu bul
	note, err := s.noteRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("not arama sırasında hata: %w", err)
	}
	if note == nil {
		return ErrNoteNotFound
	}

	// Kullanıcı yetkisi kontrol et
	if note.UserID != userID {
		return ErrNotAuthorized
	}

	// Notu sil
	return s.noteRepo.Delete(id)
}

// GetNote, bir notu getirir
func (s *NoteService) GetNote(id uint) (*domain.Note, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("not arama sırasında hata: %w", err)
	}
	if note == nil {
		return nil, ErrNoteNotFound
	}

	// Görüntülenme sayısını artır
	s.noteRepo.IncrementViewCount(id)

	return note, nil
}

// GetUserNotes, bir kullanıcının notlarını getirir
func (s *NoteService) GetUserNotes(userID uint, limit, offset int) ([]*domain.Note, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.noteRepo.FindByUserID(userID, limit, offset)
}

// GetPublicNotes, herkese açık notları getirir
func (s *NoteService) GetPublicNotes(limit, offset int) ([]*domain.Note, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.noteRepo.FindPublic(limit, offset)
}

// SearchNotes, notları arar
func (s *NoteService) SearchNotes(query string, limit, offset int) ([]*domain.Note, error) {
	if query == "" {
		return nil, ErrInvalidParameters
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.noteRepo.Search(query, limit, offset)
}

// AddComment, bir nota yorum ekler
func (s *NoteService) AddComment(comment *domain.Comment) error {
	// Notu bul
	note, err := s.noteRepo.FindByID(comment.NoteID)
	if err != nil {
		return fmt.Errorf("not arama sırasında hata: %w", err)
	}
	if note == nil {
		return ErrNoteNotFound
	}

	// Yorumu ekle
	return s.commentRepo.Create(comment)
}

// GetComments, bir notun yorumlarını getirir
func (s *NoteService) GetComments(noteID uint, limit, offset int) ([]*domain.Comment, error) {
	// Notu bul
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, fmt.Errorf("not arama sırasında hata: %w", err)
	}
	if note == nil {
		return nil, ErrNoteNotFound
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.commentRepo.FindByNoteID(noteID, limit, offset)
}

// LikeNote, bir notu beğenir
func (s *NoteService) LikeNote(noteID uint, userID uint) error {
	// Notu bul
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return fmt.Errorf("not arama sırasında hata: %w", err)
	}
	if note == nil {
		return ErrNoteNotFound
	}

	// Beğeni sayısını artır
	return s.noteRepo.IncrementLikeCount(noteID)
}

// UnlikeNote, bir notun beğenisini kaldırır
func (s *NoteService) UnlikeNote(noteID uint, userID uint) error {
	// Notu bul
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return fmt.Errorf("not arama sırasında hata: %w", err)
	}
	if note == nil {
		return ErrNoteNotFound
	}

	// Beğeni sayısını azalt
	return s.noteRepo.DecrementLikeCount(noteID)
}

// GetNoteByInviteToken, davet bağlantısı ile erişim için bir notu getirir
// Bu fonksiyon erişim kontrolü yapmadan doğrudan notu getirir
func (s *NoteService) GetNoteByInviteToken(id uint) (*domain.Note, error) {
	note, err := s.noteRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("not arama sırasında hata: %w", err)
	}
	if note == nil {
		return nil, ErrNoteNotFound
	}

	// Görüntülenme sayısını artır
	s.noteRepo.IncrementViewCount(id)

	return note, nil
}

// Ensure NoteService implements domain.NoteService
var _ domain.NoteService = (*NoteService)(nil)
