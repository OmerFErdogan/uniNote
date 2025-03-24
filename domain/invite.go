package domain

import (
	"time"
)

// Invite, bir içeriğe (not veya PDF) erişim için davet bağlantısını temsil eder
type Invite struct {
	ID        uint      `json:"id"`
	ContentID uint      `json:"contentId"`
	Type      string    `json:"type"` // "note" veya "pdf"
	Token     string    `json:"token"`
	CreatedBy uint      `json:"createdBy"`
	ExpiresAt time.Time `json:"expiresAt"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// InviteRepository, davet bağlantısı verilerinin saklanması ve alınması için bir arayüz tanımlar
type InviteRepository interface {
	FindByID(id uint) (*Invite, error)
	FindByToken(token string) (*Invite, error)
	FindByContentID(contentID uint, contentType string) ([]*Invite, error)
	Create(invite *Invite) error
	Update(invite *Invite) error
	Delete(id uint) error
	DeleteByContentID(contentID uint, contentType string) error
}

// InviteService, davet bağlantısı ile ilgili iş mantığını içerir
type InviteService interface {
	CreateInvite(invite *Invite) error
	GetInvite(token string) (*Invite, error)
	GetInvitesByContent(contentID uint, contentType string) ([]*Invite, error)
	DeactivateInvite(id uint, userID uint) error
	ValidateInvite(token string) (bool, *Invite, error)
}
