package domain

import (
	"time"
)

// User, sistemdeki bir kullanıcıyı temsil eder
type User struct {
	ID         uint      `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   string    `json:"-"` // JSON dönüşlerinde gösterilmez
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	University string    `json:"university"`
	Department string    `json:"department"`
	Class      string    `json:"class"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// UserRepository, kullanıcı verilerinin saklanması ve alınması için bir arayüz tanımlar
type UserRepository interface {
	FindByID(id uint) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByUsername(username string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
	List(limit, offset int) ([]*User, error)
}

// UserService, kullanıcı ile ilgili iş mantığını içerir
type UserService interface {
	Register(user *User) error
	Login(email, password string) (string, error) // JWT token döner
	GetProfile(id uint) (*User, error)
	UpdateProfile(user *User) error
	ChangePassword(id uint, oldPassword, newPassword string) error
}
