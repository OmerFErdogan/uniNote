package domain

import (
	"time"
)

// View, bir kullanıcının bir içeriği görüntülemesini temsil eder
type View struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	ContentID uint      `json:"contentId"`
	Type      string    `json:"type"` // "note" veya "pdf"
	ViewedAt  time.Time `json:"viewedAt"`
}

// ViewResponse, görüntüleme bilgilerini kullanıcı detaylarıyla birlikte döndürmek için kullanılır
type ViewResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Username  string    `json:"username"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	ContentID uint      `json:"contentId"`
	Type      string    `json:"type"`
	ViewedAt  time.Time `json:"viewedAt"`
}

// ViewRepository, görüntüleme verilerinin saklanması ve alınması için bir arayüz tanımlar
type ViewRepository interface {
	FindByID(id uint) (*View, error)
	FindByUserIDAndContent(userID, contentID uint, contentType string) (*View, error)
	FindByContentID(contentID uint, contentType string, limit, offset int) ([]*View, error)
	FindByUserID(userID uint, limit, offset int) ([]*View, error)
	Create(view *View) error
	Update(view *View) error
	Delete(id uint) error
	DeleteByUserIDAndContent(userID, contentID uint, contentType string) error
}

// ViewService, görüntüleme ile ilgili iş mantığını içerir
type ViewService interface {
	RecordView(userID, contentID uint, contentType string) error
	GetContentViews(contentID uint, contentType string, limit, offset int) ([]*ViewResponse, error)
	GetUserViews(userID uint, limit, offset int) ([]*View, error)
	HasUserViewed(userID, contentID uint, contentType string) (bool, error)
}
