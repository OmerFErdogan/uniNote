package usecase

import (
	"github.com/OmerFErdogan/uninote/domain"
	"github.com/OmerFErdogan/uninote/infrastructure/logger"
)

// ViewService, görüntüleme işlemlerini yöneten servistir
type ViewService struct {
	viewRepo domain.ViewRepository
	userRepo domain.UserRepository
	noteRepo domain.NoteRepository
	pdfRepo  domain.PDFRepository
	logger   *logger.Logger
}

// NewViewService, yeni bir ViewService oluşturur
func NewViewService(
	viewRepo domain.ViewRepository,
	userRepo domain.UserRepository,
	noteRepo domain.NoteRepository,
	pdfRepo domain.PDFRepository,
	logger *logger.Logger,
) *ViewService {
	return &ViewService{
		viewRepo: viewRepo,
		userRepo: userRepo,
		noteRepo: noteRepo,
		pdfRepo:  pdfRepo,
		logger:   logger,
	}
}

// RecordView, bir kullanıcının bir içeriği görüntülemesini kaydeder
func (s *ViewService) RecordView(userID, contentID uint, contentType string) error {
	// İçerik türüne göre içeriğin var olup olmadığını kontrol et
	if contentType == "note" {
		note, err := s.noteRepo.FindByID(contentID)
		if err != nil {
			s.logger.Error("Note bulunamadı", "error", err, "noteID", contentID)
			return err
		}
		if note == nil {
			s.logger.Error("Note bulunamadı", "noteID", contentID)
			return domain.ErrNotFound
		}

		// Not sahibi kendi notunu görüntülediğinde kaydetme
		if note.UserID == userID {
			return nil
		}

		// Not özel ise, kullanıcının erişim izni olup olmadığını kontrol et
		if !note.IsPublic {
			// Burada davet bağlantısı kontrolü yapılabilir
			// Şimdilik sadece özel notları görüntüleme kaydı tutmuyoruz
			return nil
		}

		// Görüntüleme sayısını artır
		if err := s.noteRepo.IncrementViewCount(contentID); err != nil {
			s.logger.Error("Not görüntüleme sayısı artırılamadı", "error", err, "noteID", contentID)
			return err
		}
	} else if contentType == "pdf" {
		pdf, err := s.pdfRepo.FindByID(contentID)
		if err != nil {
			s.logger.Error("PDF bulunamadı", "error", err, "pdfID", contentID)
			return err
		}
		if pdf == nil {
			s.logger.Error("PDF bulunamadı", "pdfID", contentID)
			return domain.ErrNotFound
		}

		// PDF sahibi kendi PDF'ini görüntülediğinde kaydetme
		if pdf.UserID == userID {
			return nil
		}

		// PDF özel ise, kullanıcının erişim izni olup olmadığını kontrol et
		if !pdf.IsPublic {
			// Burada davet bağlantısı kontrolü yapılabilir
			// Şimdilik sadece özel PDF'leri görüntüleme kaydı tutmuyoruz
			return nil
		}

		// Görüntüleme sayısını artır
		if err := s.pdfRepo.IncrementViewCount(contentID); err != nil {
			s.logger.Error("PDF görüntüleme sayısı artırılamadı", "error", err, "pdfID", contentID)
			return err
		}
	} else {
		s.logger.Error("Geçersiz içerik türü", "contentType", contentType)
		return domain.ErrInvalidContentType
	}

	// Kullanıcının daha önce bu içeriği görüntüleyip görüntülemediğini kontrol et
	existingView, err := s.viewRepo.FindByUserIDAndContent(userID, contentID, contentType)
	if err != nil {
		s.logger.Error("Görüntüleme kaydı kontrol edilemedi", "error", err, "userID", userID, "contentID", contentID, "contentType", contentType)
		return err
	}

	// Eğer daha önce görüntülemişse, görüntüleme zamanını güncelle
	if existingView != nil {
		existingView.ViewedAt = domain.Now()
		return s.viewRepo.Update(existingView)
	}

	// Yeni görüntüleme kaydı oluştur
	view := &domain.View{
		UserID:    userID,
		ContentID: contentID,
		Type:      contentType,
		ViewedAt:  domain.Now(),
	}

	s.logger.Info("Görüntüleme kaydı oluşturuluyor", "userID", userID, "contentID", contentID, "contentType", contentType)
	return s.viewRepo.Create(view)
}

// GetContentViews, bir içeriğin görüntüleme kayıtlarını kullanıcı bilgileriyle birlikte döndürür
func (s *ViewService) GetContentViews(contentID uint, contentType string, limit, offset int) ([]*domain.ViewResponse, error) {
	// İçerik türüne göre içeriğin var olup olmadığını ve kullanıcının erişim izni olup olmadığını kontrol et
	if contentType == "note" {
		note, err := s.noteRepo.FindByID(contentID)
		if err != nil {
			s.logger.Error("Note bulunamadı", "error", err, "noteID", contentID)
			return nil, err
		}
		if note == nil {
			s.logger.Error("Note bulunamadı", "noteID", contentID)
			return nil, domain.ErrNotFound
		}
	} else if contentType == "pdf" {
		pdf, err := s.pdfRepo.FindByID(contentID)
		if err != nil {
			s.logger.Error("PDF bulunamadı", "error", err, "pdfID", contentID)
			return nil, err
		}
		if pdf == nil {
			s.logger.Error("PDF bulunamadı", "pdfID", contentID)
			return nil, domain.ErrNotFound
		}
	} else {
		s.logger.Error("Geçersiz içerik türü", "contentType", contentType)
		return nil, domain.ErrInvalidContentType
	}

	// İçeriğin görüntüleme kayıtlarını getir
	views, err := s.viewRepo.FindByContentID(contentID, contentType, limit, offset)
	if err != nil {
		s.logger.Error("Görüntüleme kayıtları getirilemedi", "error", err, "contentID", contentID, "contentType", contentType)
		return nil, err
	}

	// Görüntüleme kayıtlarını kullanıcı bilgileriyle zenginleştir
	viewResponses := make([]*domain.ViewResponse, 0, len(views))
	for _, view := range views {
		user, err := s.userRepo.FindByID(view.UserID)
		if err != nil {
			s.logger.Error("Kullanıcı bulunamadı", "error", err, "userID", view.UserID)
			continue
		}
		if user == nil {
			s.logger.Error("Kullanıcı bulunamadı", "userID", view.UserID)
			continue
		}

		viewResponse := &domain.ViewResponse{
			ID:        view.ID,
			UserID:    view.UserID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			ContentID: view.ContentID,
			Type:      view.Type,
			ViewedAt:  view.ViewedAt,
		}
		viewResponses = append(viewResponses, viewResponse)
	}

	return viewResponses, nil
}

// GetUserViews, bir kullanıcının görüntüleme kayıtlarını döndürür
func (s *ViewService) GetUserViews(userID uint, limit, offset int) ([]*domain.View, error) {
	// Kullanıcının var olup olmadığını kontrol et
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		s.logger.Error("Kullanıcı bulunamadı", "error", err, "userID", userID)
		return nil, err
	}
	if user == nil {
		s.logger.Error("Kullanıcı bulunamadı", "userID", userID)
		return nil, domain.ErrNotFound
	}

	// Kullanıcının görüntüleme kayıtlarını getir
	return s.viewRepo.FindByUserID(userID, limit, offset)
}

// HasUserViewed, bir kullanıcının bir içeriği görüntüleyip görüntülemediğini kontrol eder
func (s *ViewService) HasUserViewed(userID, contentID uint, contentType string) (bool, error) {
	view, err := s.viewRepo.FindByUserIDAndContent(userID, contentID, contentType)
	if err != nil {
		s.logger.Error("Görüntüleme kaydı kontrol edilemedi", "error", err, "userID", userID, "contentID", contentID, "contentType", contentType)
		return false, err
	}
	return view != nil, nil
}

// Ensure ViewService implements domain.ViewService
var _ domain.ViewService = (*ViewService)(nil)
