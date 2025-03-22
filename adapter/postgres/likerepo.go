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

// Create, yeni bir beğeni oluşturur
func (r *LikeRepository) Create(like *domain.Like) error {
	likeModel := ContentLikeModel{
		UserID:    like.UserID,
		ContentID: like.ContentID,
		Type:      like.Type,
	}

	// İçerik türüne göre beğeni sayısını artır
	if like.Type == "note" {
		r.db.Model(&NoteModel{}).Where("id = ?", like.ContentID).Update("like_count", gorm.Expr("like_count + 1"))
	} else if like.Type == "pdf" {
		r.db.Model(&PDFModel{}).Where("id = ?", like.ContentID).Update("like_count", gorm.Expr("like_count + 1"))
	}

	result := r.db.Create(&likeModel)
	if result.Error != nil {
		return result.Error
	}

	// ID'yi güncelle
	like.ID = uint(likeModel.ID)
	return nil
}

// Delete, bir beğeniyi siler
func (r *LikeRepository) Delete(id uint) error {
	// Beğeniyi bul
	var like ContentLikeModel
	result := r.db.First(&like, id)
	if result.Error != nil {
		return result.Error
	}

	// İçerik türüne göre beğeni sayısını azalt
	if like.Type == "note" {
		r.db.Model(&NoteModel{}).Where("id = ?", like.ContentID).Update("like_count", gorm.Expr("like_count - 1"))
	} else if like.Type == "pdf" {
		r.db.Model(&PDFModel{}).Where("id = ?", like.ContentID).Update("like_count", gorm.Expr("like_count - 1"))
	}

	// Beğeniyi sil
	return r.db.Delete(&like).Error
}

// DeleteByUserIDAndContent, kullanıcı ID'si ve içerik bilgisine göre beğeniyi siler
func (r *LikeRepository) DeleteByUserIDAndContent(userID, contentID uint, contentType string) error {
	// Beğeniyi bul
	var like ContentLikeModel
	result := r.db.Where("user_id = ? AND content_id = ? AND type = ?", userID, contentID, contentType).First(&like)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil // Beğeni zaten yok
		}
		return result.Error
	}

	// İçerik türüne göre beğeni sayısını azalt
	if contentType == "note" {
		r.db.Model(&NoteModel{}).Where("id = ?", contentID).Update("like_count", gorm.Expr("like_count - 1"))
	} else if contentType == "pdf" {
		r.db.Model(&PDFModel{}).Where("id = ?", contentID).Update("like_count", gorm.Expr("like_count - 1"))
	}

	// Beğeniyi sil
	return r.db.Delete(&like).Error
}

// Ensure LikeRepository implements domain.LikeRepository
var _ domain.LikeRepository = (*LikeRepository)(nil)
