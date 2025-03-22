package domain

import (
	"time"
)

// Like, bir kullanıcının bir içeriği beğenmesini temsil eder
type Like struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	ContentID uint      `json:"contentId"`
	Type      string    `json:"type"` // "note" veya "pdf"
	CreatedAt time.Time `json:"createdAt"`
}

// LikeRepository, beğeni verilerinin saklanması ve alınması için bir arayüz tanımlar
type LikeRepository interface {
	FindByID(id uint) (*Like, error)
	FindByUserIDAndContent(userID, contentID uint, contentType string) (*Like, error)
	FindByContentID(contentID uint, contentType string, limit, offset int) ([]*Like, error)
	FindByUserID(userID uint, limit, offset int) ([]*Like, error)
	Create(like *Like) error
	Delete(id uint) error
	DeleteByUserIDAndContent(userID, contentID uint, contentType string) error
}

// LikeService, beğeni ile ilgili iş mantığını içerir
type LikeService interface {
	LikeContent(userID, contentID uint, contentType string) error
	UnlikeContent(userID, contentID uint, contentType string) error
	GetUserLikes(userID uint, limit, offset int) ([]*Like, error)
	GetContentLikes(contentID uint, contentType string, limit, offset int) ([]*Like, error)
	IsLikedByUser(userID, contentID uint, contentType string) (bool, error)
}
