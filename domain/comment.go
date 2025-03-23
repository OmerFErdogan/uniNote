package domain

import (
	"time"
)

// CommentResponse, bir yorum yanıtını temsil eder
type CommentResponse struct {
	ID         uint      `json:"id"`
	ContentID  uint      `json:"contentId"` // Not veya PDF ID'si
	UserID     uint      `json:"userId"`
	Username   string    `json:"username"`
	FullName   string    `json:"fullName"`
	Content    string    `json:"content"`
	PageNumber int       `json:"pageNumber,omitempty"` // Sadece PDF yorumları için
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
