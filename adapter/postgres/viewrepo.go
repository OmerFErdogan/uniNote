package postgres

import (
	"errors"
	"time"

	"github.com/OmerFErdogan/uninote/domain"
	"gorm.io/gorm"
)

// ViewModel, görüntüleme verilerinin veritabanında saklanması için kullanılan GORM modelidir
type ViewModel struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index:idx_view_user"`
	ContentID uint      `gorm:"not null;index:idx_view_content"`
	Type      string    `gorm:"not null;index:idx_view_type"` // "note" veya "pdf"
	ViewedAt  time.Time `gorm:"not null"`
}

// TableName, GORM için tablo adını belirtir
func (ViewModel) TableName() string {
	return "views"
}

// ViewRepository, görüntüleme verilerinin PostgreSQL veritabanında saklanması ve alınması için implementasyondur
type ViewRepository struct {
	db *gorm.DB
}

// NewViewRepository, yeni bir ViewRepository oluşturur
func NewViewRepository(db *gorm.DB) *ViewRepository {
	return &ViewRepository{db: db}
}

// ToEntity, ViewModel'i domain.View'e dönüştürür
func (m *ViewModel) ToEntity() *domain.View {
	return &domain.View{
		ID:        uint(m.ID),
		UserID:    m.UserID,
		ContentID: m.ContentID,
		Type:      m.Type,
		ViewedAt:  m.ViewedAt,
	}
}

// toModel, domain.View'i ViewModel'e dönüştürür
func toViewModel(d *domain.View) *ViewModel {
	model := &ViewModel{
		UserID:    d.UserID,
		ContentID: d.ContentID,
		Type:      d.Type,
		ViewedAt:  d.ViewedAt,
	}
	model.ID = uint(d.ID)
	return model
}

// FindByID, belirtilen ID'ye sahip görüntülemeyi bulur
func (r *ViewRepository) FindByID(id uint) (*domain.View, error) {
	var model ViewModel
	if err := r.db.Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToEntity(), nil
}

// FindByUserIDAndContent, belirtilen kullanıcı ve içerik için görüntülemeyi bulur
func (r *ViewRepository) FindByUserIDAndContent(userID, contentID uint, contentType string) (*domain.View, error) {
	var model ViewModel
	if err := r.db.Where("user_id = ? AND content_id = ? AND type = ?", userID, contentID, contentType).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToEntity(), nil
}

// FindByContentID, belirtilen içerik için görüntülemeleri bulur
func (r *ViewRepository) FindByContentID(contentID uint, contentType string, limit, offset int) ([]*domain.View, error) {
	var models []ViewModel
	if err := r.db.Where("content_id = ? AND type = ?", contentID, contentType).
		Order("viewed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, err
	}

	views := make([]*domain.View, len(models))
	for i, model := range models {
		views[i] = model.ToEntity()
	}
	return views, nil
}

// FindByUserID, belirtilen kullanıcı için görüntülemeleri bulur
func (r *ViewRepository) FindByUserID(userID uint, limit, offset int) ([]*domain.View, error) {
	var models []ViewModel
	if err := r.db.Where("user_id = ?", userID).
		Order("viewed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, err
	}

	views := make([]*domain.View, len(models))
	for i, model := range models {
		views[i] = model.ToEntity()
	}
	return views, nil
}

// Create, yeni bir görüntüleme kaydı oluşturur
func (r *ViewRepository) Create(view *domain.View) error {
	model := toViewModel(view)
	model.ViewedAt = time.Now()
	return r.db.Create(model).Error
}

// Update, mevcut bir görüntüleme kaydını günceller
func (r *ViewRepository) Update(view *domain.View) error {
	model := toViewModel(view)
	return r.db.Save(model).Error
}

// Delete, belirtilen ID'ye sahip görüntülemeyi siler
func (r *ViewRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&ViewModel{}).Error
}

// DeleteByUserIDAndContent, belirtilen kullanıcı ve içerik için görüntülemeyi siler
func (r *ViewRepository) DeleteByUserIDAndContent(userID, contentID uint, contentType string) error {
	return r.db.Where("user_id = ? AND content_id = ? AND type = ?", userID, contentID, contentType).Delete(&ViewModel{}).Error
}
