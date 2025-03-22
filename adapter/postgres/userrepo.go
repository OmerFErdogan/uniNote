package postgres

import (
	"errors"

	"github.com/OmerFErdogan/uninote/domain"
	"gorm.io/gorm"
)

// UserModel, User varlığının veritabanı modelini temsil eder
type UserModel struct {
	gorm.Model
	Username   string `gorm:"uniqueIndex;not null"`
	Email      string `gorm:"uniqueIndex;not null"`
	Password   string `gorm:"not null"`
	FirstName  string
	LastName   string
	University string
	Department string
	Class      string
}

// ToEntity, veritabanı modelini domain varlığına dönüştürür
func (u *UserModel) ToEntity() *domain.User {
	return &domain.User{
		ID:         uint(u.ID),
		Username:   u.Username,
		Email:      u.Email,
		Password:   u.Password,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		University: u.University,
		Department: u.Department,
		Class:      u.Class,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

// FromEntity, domain varlığını veritabanı modeline dönüştürür
func (u *UserModel) FromEntity(user *domain.User) {
	u.ID = uint(user.ID)
	u.Username = user.Username
	u.Email = user.Email
	u.Password = user.Password
	u.FirstName = user.FirstName
	u.LastName = user.LastName
	u.University = user.University
	u.Department = user.Department
	u.Class = user.Class
	// CreatedAt ve UpdatedAt alanları GORM tarafından otomatik olarak yönetilir
}

// UserRepository, domain.UserRepository arayüzünün PostgreSQL implementasyonu
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository, yeni bir UserRepository örneği oluşturur
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID, ID'ye göre kullanıcı bulur
func (r *UserRepository) FindByID(id uint) (*domain.User, error) {
	var user UserModel
	result := r.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Kullanıcı bulunamadı
		}
		return nil, result.Error
	}
	return user.ToEntity(), nil
}

// FindByEmail, e-posta adresine göre kullanıcı bulur
func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	var user UserModel
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Kullanıcı bulunamadı
		}
		return nil, result.Error
	}
	return user.ToEntity(), nil
}

// FindByUsername, kullanıcı adına göre kullanıcı bulur
func (r *UserRepository) FindByUsername(username string) (*domain.User, error) {
	var user UserModel
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Kullanıcı bulunamadı
		}
		return nil, result.Error
	}
	return user.ToEntity(), nil
}

// Create, yeni bir kullanıcı oluşturur
func (r *UserRepository) Create(user *domain.User) error {
	var userModel UserModel
	userModel.FromEntity(user)
	result := r.db.Create(&userModel)
	if result.Error != nil {
		return result.Error
	}
	// ID'yi güncelle
	user.ID = uint(userModel.ID)
	return nil
}

// Update, bir kullanıcıyı günceller
func (r *UserRepository) Update(user *domain.User) error {
	var userModel UserModel
	userModel.FromEntity(user)
	result := r.db.Save(&userModel)
	return result.Error
}

// Delete, bir kullanıcıyı siler
func (r *UserRepository) Delete(id uint) error {
	result := r.db.Delete(&UserModel{}, id)
	return result.Error
}

// List, kullanıcıları listeler
func (r *UserRepository) List(limit, offset int) ([]*domain.User, error) {
	var users []UserModel
	result := r.db.Limit(limit).Offset(offset).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainUsers []*domain.User
	for _, user := range users {
		domainUsers = append(domainUsers, user.ToEntity())
	}
	return domainUsers, nil
}
