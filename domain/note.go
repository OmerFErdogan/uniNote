package domain

import (
	"time"
)

// Note, bir kullanıcının oluşturduğu notu temsil eder
type Note struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	UserID       uint      `json:"userId"`
	Tags         []string  `json:"tags"`
	IsPublic     bool      `json:"isPublic"`
	ViewCount    int       `json:"viewCount"`
	LikeCount    int       `json:"likeCount"`
	CommentCount int       `json:"commentCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Comment, bir not üzerindeki yorumu temsil eder
type Comment struct {
	ID        uint      `json:"id"`
	NoteID    uint      `json:"noteId"`
	UserID    uint      `json:"userId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NoteRepository, not verilerinin saklanması ve alınması için bir arayüz tanımlar
type NoteRepository interface {
	FindByID(id uint) (*Note, error)
	FindByUserID(userID uint, limit, offset int) ([]*Note, error)
	FindPublic(limit, offset int) ([]*Note, error)
	FindByTag(tag string, limit, offset int) ([]*Note, error)
	Search(query string, limit, offset int) ([]*Note, error)
	Create(note *Note) error
	Update(note *Note) error
	Delete(id uint) error
	IncrementViewCount(id uint) error
	IncrementLikeCount(id uint) error
	DecrementLikeCount(id uint) error
}

// CommentRepository, yorum verilerinin saklanması ve alınması için bir arayüz tanımlar
type CommentRepository interface {
	FindByNoteID(noteID uint, limit, offset int) ([]*Comment, error)
	Create(comment *Comment) error
	Update(comment *Comment) error
	Delete(id uint) error
}

// NoteService, not ile ilgili iş mantığını içerir
type NoteService interface {
	CreateNote(note *Note) error
	UpdateNote(note *Note) error
	DeleteNote(id uint, userID uint) error
	GetNote(id uint) (*Note, error)
	GetUserNotes(userID uint, limit, offset int) ([]*Note, error)
	GetPublicNotes(limit, offset int) ([]*Note, error)
	SearchNotes(query string, limit, offset int) ([]*Note, error)
	AddComment(comment *Comment) error
	GetComments(noteID uint, limit, offset int) ([]*Comment, error)
	LikeNote(noteID uint, userID uint) error
	UnlikeNote(noteID uint, userID uint) error
}
