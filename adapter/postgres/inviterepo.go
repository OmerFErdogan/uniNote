package postgres

import (
	"fmt"
	"time"

	"github.com/OmerFErdogan/uninote/domain"
	"gorm.io/gorm"
)

// InviteModel, davet bağlantısı için veritabanı modeli
type InviteModel struct {
	ID        uint   `gorm:"primaryKey"`
	ContentID uint   `gorm:"index"`
	Type      string `gorm:"size:10;index"` // "note" veya "pdf"
	Token     string `gorm:"size:100;uniqueIndex"`
	CreatedBy uint   `gorm:"index"`
	ExpiresAt time.Time
	IsActive  bool `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName, tablo adını belirtir
func (InviteModel) TableName() string {
	return "invites"
}

// ToDomain, veritabanı modelini domain modeline dönüştürür
func (m *InviteModel) ToDomain() *domain.Invite {
	return &domain.Invite{
		ID:        m.ID,
		ContentID: m.ContentID,
		Type:      m.Type,
		Token:     m.Token,
		CreatedBy: m.CreatedBy,
		ExpiresAt: m.ExpiresAt,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// FromDomain, domain modelini veritabanı modeline dönüştürür
func (m *InviteModel) FromDomain(invite *domain.Invite) {
	m.ID = invite.ID
	m.ContentID = invite.ContentID
	m.Type = invite.Type
	m.Token = invite.Token
	m.CreatedBy = invite.CreatedBy
	m.ExpiresAt = invite.ExpiresAt
	m.IsActive = invite.IsActive
	m.CreatedAt = invite.CreatedAt
	m.UpdatedAt = invite.UpdatedAt
}

// InviteRepository, davet bağlantısı için veritabanı işlemlerini içerir
type InviteRepository struct {
	db *gorm.DB
}

// NewInviteRepository, yeni bir InviteRepository örneği oluşturur
func NewInviteRepository(db *gorm.DB) *InviteRepository {
	return &InviteRepository{
		db: db,
	}
}

// FindByID, ID'ye göre davet bağlantısını bulur
func (r *InviteRepository) FindByID(id uint) (*domain.Invite, error) {
	var model InviteModel
	if err := r.db.First(&model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("davet bağlantısı arama hatası: %w", err)
	}
	return model.ToDomain(), nil
}

// FindByToken, token'a göre davet bağlantısını bulur
func (r *InviteRepository) FindByToken(token string) (*domain.Invite, error) {
	var model InviteModel
	if err := r.db.Where("token = ?", token).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("davet bağlantısı arama hatası: %w", err)
	}
	return model.ToDomain(), nil
}

// FindByContentID, içerik ID'sine göre davet bağlantılarını bulur
func (r *InviteRepository) FindByContentID(contentID uint, contentType string) ([]*domain.Invite, error) {
	var models []InviteModel
	if err := r.db.Where("content_id = ? AND type = ?", contentID, contentType).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("davet bağlantıları arama hatası: %w", err)
	}

	invites := make([]*domain.Invite, len(models))
	for i, model := range models {
		invites[i] = model.ToDomain()
	}
	return invites, nil
}

// Create, yeni bir davet bağlantısı oluşturur
func (r *InviteRepository) Create(invite *domain.Invite) error {
	model := InviteModel{}
	model.FromDomain(invite)

	if err := r.db.Create(&model).Error; err != nil {
		return fmt.Errorf("davet bağlantısı oluşturma hatası: %w", err)
	}

	// ID'yi güncelle
	invite.ID = model.ID
	return nil
}

// Update, bir davet bağlantısını günceller
func (r *InviteRepository) Update(invite *domain.Invite) error {
	model := InviteModel{}
	model.FromDomain(invite)

	if err := r.db.Save(&model).Error; err != nil {
		return fmt.Errorf("davet bağlantısı güncelleme hatası: %w", err)
	}
	return nil
}

// Delete, bir davet bağlantısını siler
func (r *InviteRepository) Delete(id uint) error {
	if err := r.db.Delete(&InviteModel{}, id).Error; err != nil {
		return fmt.Errorf("davet bağlantısı silme hatası: %w", err)
	}
	return nil
}

// DeleteByContentID, içerik ID'sine göre davet bağlantılarını siler
func (r *InviteRepository) DeleteByContentID(contentID uint, contentType string) error {
	if err := r.db.Where("content_id = ? AND type = ?", contentID, contentType).Delete(&InviteModel{}).Error; err != nil {
		return fmt.Errorf("davet bağlantıları silme hatası: %w", err)
	}
	return nil
}

// Ensure InviteRepository implements domain.InviteRepository
var _ domain.InviteRepository = (*InviteRepository)(nil)
