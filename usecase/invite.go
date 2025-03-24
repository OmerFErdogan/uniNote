package usecase

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/OmerFErdogan/uninote/domain"
)

// Diğer paketlerden hataları import et
// ErrContentNotFound ve ErrInvalidType usecase/like.go'dan
// ErrNotAuthorized usecase/note.go'dan

var (
	ErrInviteNotFound  = errors.New("davet bağlantısı bulunamadı")
	ErrInviteExpired   = errors.New("davet bağlantısı süresi dolmuş")
	ErrInviteNotActive = errors.New("davet bağlantısı aktif değil")
)

// InviteService, davet bağlantısı ile ilgili iş mantığını içerir
type InviteService struct {
	inviteRepo domain.InviteRepository
	noteRepo   domain.NoteRepository
	pdfRepo    domain.PDFRepository
}

// NewInviteService, yeni bir InviteService örneği oluşturur
func NewInviteService(
	inviteRepo domain.InviteRepository,
	noteRepo domain.NoteRepository,
	pdfRepo domain.PDFRepository,
) *InviteService {
	return &InviteService{
		inviteRepo: inviteRepo,
		noteRepo:   noteRepo,
		pdfRepo:    pdfRepo,
	}
}

// CreateInvite, yeni bir davet bağlantısı oluşturur
func (s *InviteService) CreateInvite(invite *domain.Invite) error {
	// İçerik tipini kontrol et
	if invite.Type != "note" && invite.Type != "pdf" {
		return ErrInvalidType
	}

	// İçeriğin var olduğunu ve kullanıcının sahibi olduğunu kontrol et
	if invite.Type == "note" {
		note, err := s.noteRepo.FindByID(invite.ContentID)
		if err != nil {
			return fmt.Errorf("not arama sırasında hata: %w", err)
		}
		if note == nil {
			return ErrContentNotFound
		}
		if note.UserID != invite.CreatedBy {
			return ErrNotAuthorized
		}
	} else if invite.Type == "pdf" {
		pdf, err := s.pdfRepo.FindByID(invite.ContentID)
		if err != nil {
			return fmt.Errorf("PDF arama sırasında hata: %w", err)
		}
		if pdf == nil {
			return ErrContentNotFound
		}
		if pdf.UserID != invite.CreatedBy {
			return ErrNotAuthorized
		}
	}

	// Benzersiz token oluştur
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("token oluşturma hatası: %w", err)
	}
	invite.Token = token

	// Varsayılan değerleri ayarla
	if invite.ExpiresAt.IsZero() {
		// Varsayılan olarak 7 gün sonra sona erer
		invite.ExpiresAt = time.Now().AddDate(0, 0, 7)
	}
	invite.IsActive = true
	invite.CreatedAt = time.Now()
	invite.UpdatedAt = time.Now()

	// Daveti kaydet
	return s.inviteRepo.Create(invite)
}

// GetInvite, bir davet bağlantısını getirir
func (s *InviteService) GetInvite(token string) (*domain.Invite, error) {
	invite, err := s.inviteRepo.FindByToken(token)
	if err != nil {
		return nil, fmt.Errorf("davet bağlantısı arama sırasında hata: %w", err)
	}
	if invite == nil {
		return nil, ErrInviteNotFound
	}

	return invite, nil
}

// GetInvitesByContent, bir içeriğin davet bağlantılarını getirir
func (s *InviteService) GetInvitesByContent(contentID uint, contentType string) ([]*domain.Invite, error) {
	// İçerik tipini kontrol et
	if contentType != "note" && contentType != "pdf" {
		return nil, ErrInvalidType
	}

	// İçeriğin var olduğunu kontrol et
	if contentType == "note" {
		note, err := s.noteRepo.FindByID(contentID)
		if err != nil {
			return nil, fmt.Errorf("not arama sırasında hata: %w", err)
		}
		if note == nil {
			return nil, ErrContentNotFound
		}
	} else if contentType == "pdf" {
		pdf, err := s.pdfRepo.FindByID(contentID)
		if err != nil {
			return nil, fmt.Errorf("PDF arama sırasında hata: %w", err)
		}
		if pdf == nil {
			return nil, ErrContentNotFound
		}
	}

	// Davet bağlantılarını getir
	return s.inviteRepo.FindByContentID(contentID, contentType)
}

// DeactivateInvite, bir davet bağlantısını devre dışı bırakır
func (s *InviteService) DeactivateInvite(id uint, userID uint) error {
	// Daveti bul
	invite, err := s.inviteRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("davet bağlantısı arama sırasında hata: %w", err)
	}
	if invite == nil {
		return ErrInviteNotFound
	}

	// Kullanıcı yetkisi kontrol et
	if invite.CreatedBy != userID {
		return ErrNotAuthorized
	}

	// Daveti devre dışı bırak
	invite.IsActive = false
	invite.UpdatedAt = time.Now()

	return s.inviteRepo.Update(invite)
}

// ValidateInvite, bir davet bağlantısının geçerli olup olmadığını kontrol eder
func (s *InviteService) ValidateInvite(token string) (bool, *domain.Invite, error) {
	// Daveti bul
	invite, err := s.inviteRepo.FindByToken(token)
	if err != nil {
		return false, nil, fmt.Errorf("davet bağlantısı arama sırasında hata: %w", err)
	}
	if invite == nil {
		return false, nil, ErrInviteNotFound
	}

	// Aktif olup olmadığını kontrol et
	if !invite.IsActive {
		return false, invite, ErrInviteNotActive
	}

	// Süresinin dolup dolmadığını kontrol et
	if time.Now().After(invite.ExpiresAt) {
		return false, invite, ErrInviteExpired
	}

	// İçeriğin var olduğunu kontrol et
	if invite.Type == "note" {
		note, err := s.noteRepo.FindByID(invite.ContentID)
		if err != nil {
			return false, invite, fmt.Errorf("not arama sırasında hata: %w", err)
		}
		if note == nil {
			return false, invite, ErrContentNotFound
		}
	} else if invite.Type == "pdf" {
		pdf, err := s.pdfRepo.FindByID(invite.ContentID)
		if err != nil {
			return false, invite, fmt.Errorf("PDF arama sırasında hata: %w", err)
		}
		if pdf == nil {
			return false, invite, ErrContentNotFound
		}
	}

	return true, invite, nil
}

// generateToken, benzersiz bir token oluşturur
func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Ensure InviteService implements domain.InviteService
var _ domain.InviteService = (*InviteService)(nil)
