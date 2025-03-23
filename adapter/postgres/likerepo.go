package postgres

import (
	"errors"

	"github.com/OmerFErdogan/uninote/domain"
	"gorm.io/gorm"
)

// ContentLikeModel, Like varlığının veritabanı modelini temsil eder
type ContentLikeModel struct {
	gorm.Model
	UserID    uint   `gorm:"not null;uniqueIndex:idx_user_content_type"`
	ContentID uint   `gorm:"not null;uniqueIndex:idx_user_content_type"`
	Type      string `gorm:"not null;uniqueIndex:idx_user_content_type"` // "note" veya "pdf"
}

// ToEntity, veritabanı modelini domain varlığına dönüştürür
func (l *ContentLikeModel) ToEntity() *domain.Like {
	return &domain.Like{
		ID:        uint(l.ID),
		UserID:    l.UserID,
		ContentID: l.ContentID,
		Type:      l.Type,
		CreatedAt: l.CreatedAt,
	}
}

// LikeRepository, domain.LikeRepository arayüzünün PostgreSQL implementasyonu
type LikeRepository struct {
	db *gorm.DB
}

// NewLikeRepository, yeni bir LikeRepository örneği oluşturur
func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

// FindByID, ID'ye göre beğeni bulur
func (r *LikeRepository) FindByID(id uint) (*domain.Like, error) {
	var like ContentLikeModel
	result := r.db.First(&like, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Beğeni bulunamadı
		}
		return nil, result.Error
	}
	return like.ToEntity(), nil
}

// FindByUserIDAndContent, kullanıcı ID'si ve içerik bilgisine göre beğeni bulur
func (r *LikeRepository) FindByUserIDAndContent(userID, contentID uint, contentType string) (*domain.Like, error) {
	var like ContentLikeModel
	result := r.db.Where("user_id = ? AND content_id = ? AND type = ?", userID, contentID, contentType).First(&like)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Beğeni bulunamadı
		}
		return nil, result.Error
	}
	return like.ToEntity(), nil
}

// FindByContentID, içerik ID'sine göre beğenileri bulur
func (r *LikeRepository) FindByContentID(contentID uint, contentType string, limit, offset int) ([]*domain.Like, error) {
	var likes []ContentLikeModel
	result := r.db.Where("content_id = ? AND type = ?", contentID, contentType).
		Limit(limit).Offset(offset).
		Find(&likes)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainLikes []*domain.Like
	for _, like := range likes {
		domainLikes = append(domainLikes, like.ToEntity())
	}
	return domainLikes, nil
}

// FindByUserID, kullanıcı ID'sine göre beğenileri bulur
func (r *LikeRepository) FindByUserID(userID uint, limit, offset int) ([]*domain.Like, error) {
	var likes []ContentLikeModel
	result := r.db.Where("user_id = ?", userID).
		Limit(limit).Offset(offset).
		Find(&likes)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainLikes []*domain.Like
	for _, like := range likes {
		domainLikes = append(domainLikes, like.ToEntity())
	}
	return domainLikes, nil
}

// FindLikedNotesByUserID, kullanıcının beğendiği notları doğrudan veritabanından getirir
func (r *LikeRepository) FindLikedNotesByUserID(userID uint, limit, offset int) ([]*domain.Note, error) {
	var notes []NoteModel
	result := r.db.Table("notes").
		Joins("JOIN content_like_models ON notes.id = content_like_models.content_id").
		Where("content_like_models.user_id = ? AND content_like_models.type = ?", userID, "note").
		Limit(limit).Offset(offset).
		Find(&notes)

	if result.Error != nil {
		return nil, result.Error
	}

	var domainNotes []*domain.Note
	for _, note := range notes {
		domainNotes = append(domainNotes, note.ToEntity())
	}
	return domainNotes, nil
}

// FindLikedPDFsByUserID, kullanıcının beğendiği PDF'leri doğrudan veritabanından getirir
func (r *LikeRepository) FindLikedPDFsByUserID(userID uint, limit, offset int) ([]*domain.PDF, error) {
	var pdfs []PDFModel
	result := r.db.Table("pdfs").
		Joins("JOIN content_like_models ON pdfs.id = content_like_models.content_id").
		Where("content_like_models.user_id = ? AND content_like_models.type = ?", userID, "pdf").
		Limit(limit).Offset(offset).
		Find(&pdfs)

	if result.Error != nil {
		return nil, result.Error
	}

	var domainPDFs []*domain.PDF
	for _, pdf := range pdfs {
		domainPDFs = append(domainPDFs, pdf.ToEntity())
	}
	return domainPDFs, nil
}

// Create, yeni bir beğeni oluşturur
func (r *LikeRepository) Create(like *domain.Like) error {
	likeModel := ContentLikeModel{
		UserID:    like.UserID,
		ContentID: like.ContentID,
		Type:      like.Type,
	}

	// Önce kullanıcının bu içeriği daha önce beğenip beğenmediğini kontrol et
	var existingLike ContentLikeModel
	result := r.db.Where("user_id = ? AND content_id = ? AND type = ?", like.UserID, like.ContentID, like.Type).First(&existingLike)
	if result.Error == nil {
		// Kullanıcı zaten bu içeriği beğenmiş, başarılı olarak dön
		like.ID = uint(existingLike.ID)
		return nil
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Başka bir hata oluştu
		return result.Error
	}

	// Transaction başlat
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Beğeniyi oluşturmayı dene
		err := tx.Create(&likeModel).Error
		if err != nil {
			// Eğer benzersiz indeks ihlali nedeniyle hata oluştuysa, başarılı olarak dön
			if err.Error() == "ERROR: duplicate key value violates unique constraint \"idx_user_content_type\" (SQLSTATE 23505)" {
				// Kullanıcı zaten bu içeriği beğenmiş, başarılı olarak dön
				return nil
			}
			return err
		}

		// Beğeni başarıyla oluşturuldu, şimdi içerik türüne göre beğeni sayısını artır
		if like.Type == "note" {
			if err := tx.Model(&NoteModel{}).Where("id = ?", like.ContentID).Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
				return err
			}
		} else if like.Type == "pdf" {
			if err := tx.Model(&PDFModel{}).Where("id = ?", like.ContentID).Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
				return err
			}
		}

		// ID'yi güncelle
		like.ID = uint(likeModel.ID)
		return nil
	})
}

// Delete, bir beğeniyi siler
func (r *LikeRepository) Delete(id uint) error {
	// Transaction başlat
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Beğeniyi bul
		var like ContentLikeModel
		if err := tx.First(&like, id).Error; err != nil {
			return err
		}

		// İçerik türüne göre beğeni sayısını azalt
		if like.Type == "note" {
			if err := tx.Model(&NoteModel{}).Where("id = ?", like.ContentID).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
				return err
			}
		} else if like.Type == "pdf" {
			if err := tx.Model(&PDFModel{}).Where("id = ?", like.ContentID).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
				return err
			}
		}

		// Beğeniyi sil
		return tx.Delete(&like).Error
	})
}

// DeleteByUserIDAndContent, kullanıcı ID'si ve içerik bilgisine göre beğeniyi siler
func (r *LikeRepository) DeleteByUserIDAndContent(userID, contentID uint, contentType string) error {
	// Transaction başlat
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Beğeniyi bul
		var like ContentLikeModel
		result := tx.Where("user_id = ? AND content_id = ? AND type = ?", userID, contentID, contentType).First(&like)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil // Beğeni zaten yok
			}
			return result.Error
		}

		// İçerik türüne göre beğeni sayısını azalt
		if contentType == "note" {
			if err := tx.Model(&NoteModel{}).Where("id = ?", contentID).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
				return err
			}
		} else if contentType == "pdf" {
			if err := tx.Model(&PDFModel{}).Where("id = ?", contentID).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
				return err
			}
		}

		// Beğeniyi sil
		return tx.Delete(&like).Error
	})
}

// Ensure LikeRepository implements domain.LikeRepository
var _ domain.LikeRepository = (*LikeRepository)(nil)
