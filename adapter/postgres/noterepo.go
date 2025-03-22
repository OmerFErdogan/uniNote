package postgres

import (
	"errors"

	"github.com/OmerFErdogan/uninote/domain"
	"gorm.io/gorm"
)

// NoteModel, Note varlığının veritabanı modelini temsil eder
type NoteModel struct {
	gorm.Model
	Title        string     `gorm:"not null"`
	Content      string     `gorm:"type:text"`
	UserID       uint       `gorm:"not null"`
	Tags         []TagModel `gorm:"many2many:note_tags;"`
	IsPublic     bool
	ViewCount    int
	LikeCount    int
	CommentCount int
}

// CommentModel, Comment varlığının veritabanı modelini temsil eder
type CommentModel struct {
	gorm.Model
	NoteID  uint   `gorm:"not null"`
	UserID  uint   `gorm:"not null"`
	Content string `gorm:"type:text;not null"`
}

// TagModel, etiketleri temsil eder
type TagModel struct {
	gorm.Model
	Name  string      `gorm:"uniqueIndex;not null"`
	Notes []NoteModel `gorm:"many2many:note_tags;"`
}

// ToEntity, veritabanı modelini domain varlığına dönüştürür
func (n *NoteModel) ToEntity() *domain.Note {
	tags := make([]string, len(n.Tags))
	for i, tag := range n.Tags {
		tags[i] = tag.Name
	}

	return &domain.Note{
		ID:           uint(n.ID),
		Title:        n.Title,
		Content:      n.Content,
		UserID:       n.UserID,
		Tags:         tags,
		IsPublic:     n.IsPublic,
		ViewCount:    n.ViewCount,
		LikeCount:    n.LikeCount,
		CommentCount: n.CommentCount,
		CreatedAt:    n.CreatedAt,
		UpdatedAt:    n.UpdatedAt,
	}
}

// ToEntity, veritabanı modelini domain varlığına dönüştürür
func (c *CommentModel) ToEntity() *domain.Comment {
	return &domain.Comment{
		ID:        uint(c.ID),
		NoteID:    c.NoteID,
		UserID:    c.UserID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// NoteRepository, domain.NoteRepository arayüzünün PostgreSQL implementasyonu
type NoteRepository struct {
	db *gorm.DB
}

// NewNoteRepository, yeni bir NoteRepository örneği oluşturur
func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// FindByID, ID'ye göre not bulur
func (r *NoteRepository) FindByID(id uint) (*domain.Note, error) {
	var note NoteModel
	result := r.db.Preload("Tags").First(&note, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not bulunamadı
		}
		return nil, result.Error
	}
	return note.ToEntity(), nil
}

// FindByUserID, kullanıcı ID'sine göre notları bulur
func (r *NoteRepository) FindByUserID(userID uint, limit, offset int) ([]*domain.Note, error) {
	var notes []NoteModel
	result := r.db.Preload("Tags").Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&notes)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainNotes []*domain.Note
	for _, note := range notes {
		domainNotes = append(domainNotes, note.ToEntity())
	}
	return domainNotes, nil
}

// FindPublic, herkese açık notları bulur
func (r *NoteRepository) FindPublic(limit, offset int) ([]*domain.Note, error) {
	var notes []NoteModel
	result := r.db.Preload("Tags").Where("is_public = ?", true).Limit(limit).Offset(offset).Find(&notes)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainNotes []*domain.Note
	for _, note := range notes {
		domainNotes = append(domainNotes, note.ToEntity())
	}
	return domainNotes, nil
}

// FindByTag, etikete göre notları bulur
func (r *NoteRepository) FindByTag(tag string, limit, offset int) ([]*domain.Note, error) {
	var notes []NoteModel
	result := r.db.Preload("Tags").
		Joins("JOIN note_tags ON note_tags.note_model_id = note_models.id").
		Joins("JOIN tag_models ON tag_models.id = note_tags.tag_model_id").
		Where("tag_models.name = ?", tag).
		Limit(limit).Offset(offset).
		Find(&notes)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainNotes []*domain.Note
	for _, note := range notes {
		domainNotes = append(domainNotes, note.ToEntity())
	}
	return domainNotes, nil
}

// Search, arama sorgusuna göre notları bulur
func (r *NoteRepository) Search(query string, limit, offset int) ([]*domain.Note, error) {
	var notes []NoteModel
	result := r.db.Preload("Tags").
		Where("title ILIKE ? OR content ILIKE ?", "%"+query+"%", "%"+query+"%").
		Limit(limit).Offset(offset).
		Find(&notes)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainNotes []*domain.Note
	for _, note := range notes {
		domainNotes = append(domainNotes, note.ToEntity())
	}
	return domainNotes, nil
}

// Create, yeni bir not oluşturur
func (r *NoteRepository) Create(note *domain.Note) error {
	// Not modelini oluştur
	noteModel := NoteModel{
		Title:        note.Title,
		Content:      note.Content,
		UserID:       note.UserID,
		IsPublic:     note.IsPublic,
		ViewCount:    note.ViewCount,
		LikeCount:    note.LikeCount,
		CommentCount: note.CommentCount,
	}

	// Etiketleri işle
	if len(note.Tags) > 0 {
		for _, tagName := range note.Tags {
			var tag TagModel
			// Etiketi bul veya oluştur
			result := r.db.Where("name = ?", tagName).FirstOrCreate(&tag, TagModel{Name: tagName})
			if result.Error != nil {
				return result.Error
			}
			noteModel.Tags = append(noteModel.Tags, tag)
		}
	}

	// Notu kaydet
	result := r.db.Create(&noteModel)
	if result.Error != nil {
		return result.Error
	}

	// ID'yi güncelle
	note.ID = uint(noteModel.ID)
	return nil
}

// Update, bir notu günceller
func (r *NoteRepository) Update(note *domain.Note) error {
	// Mevcut notu bul
	var noteModel NoteModel
	result := r.db.First(&noteModel, note.ID)
	if result.Error != nil {
		return result.Error
	}

	// Notu güncelle
	noteModel.Title = note.Title
	noteModel.Content = note.Content
	noteModel.IsPublic = note.IsPublic

	// Etiketleri temizle
	r.db.Model(&noteModel).Association("Tags").Clear()

	// Etiketleri işle
	if len(note.Tags) > 0 {
		for _, tagName := range note.Tags {
			var tag TagModel
			// Etiketi bul veya oluştur
			result := r.db.Where("name = ?", tagName).FirstOrCreate(&tag, TagModel{Name: tagName})
			if result.Error != nil {
				return result.Error
			}
			r.db.Model(&noteModel).Association("Tags").Append(&tag)
		}
	}

	// Notu kaydet
	result = r.db.Save(&noteModel)
	return result.Error
}

// Delete, bir notu siler
func (r *NoteRepository) Delete(id uint) error {
	// İlişkili yorumları sil
	r.db.Where("note_id = ?", id).Delete(&CommentModel{})

	// İlişkili beğenileri sil
	r.db.Where("content_id = ? AND type = ?", id, "note").Delete(&ContentLikeModel{})

	// Notu sil
	result := r.db.Delete(&NoteModel{}, id)
	return result.Error
}

// IncrementViewCount, görüntülenme sayısını artırır
func (r *NoteRepository) IncrementViewCount(id uint) error {
	result := r.db.Model(&NoteModel{}).Where("id = ?", id).Update("view_count", gorm.Expr("view_count + 1"))
	return result.Error
}

// IncrementLikeCount, beğeni sayısını artırır
func (r *NoteRepository) IncrementLikeCount(id uint) error {
	result := r.db.Model(&NoteModel{}).Where("id = ?", id).Update("like_count", gorm.Expr("like_count + 1"))
	return result.Error
}

// DecrementLikeCount, beğeni sayısını azaltır
func (r *NoteRepository) DecrementLikeCount(id uint) error {
	result := r.db.Model(&NoteModel{}).Where("id = ?", id).Update("like_count", gorm.Expr("like_count - 1"))
	return result.Error
}

// CommentRepository, domain.CommentRepository arayüzünün PostgreSQL implementasyonu
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository, yeni bir CommentRepository örneği oluşturur
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// FindByNoteID, not ID'sine göre yorumları bulur
func (r *CommentRepository) FindByNoteID(noteID uint, limit, offset int) ([]*domain.Comment, error) {
	var comments []CommentModel
	result := r.db.Where("note_id = ?", noteID).Limit(limit).Offset(offset).Find(&comments)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainComments []*domain.Comment
	for _, comment := range comments {
		domainComments = append(domainComments, comment.ToEntity())
	}
	return domainComments, nil
}

// Create, yeni bir yorum oluşturur
func (r *CommentRepository) Create(comment *domain.Comment) error {
	commentModel := CommentModel{
		NoteID:  comment.NoteID,
		UserID:  comment.UserID,
		Content: comment.Content,
	}

	result := r.db.Create(&commentModel)
	if result.Error != nil {
		return result.Error
	}

	// Yorum sayısını artır
	r.db.Model(&NoteModel{}).Where("id = ?", comment.NoteID).Update("comment_count", gorm.Expr("comment_count + 1"))

	// ID'yi güncelle
	comment.ID = uint(commentModel.ID)
	return nil
}

// Update, bir yorumu günceller
func (r *CommentRepository) Update(comment *domain.Comment) error {
	commentModel := CommentModel{
		Model: gorm.Model{
			ID: uint(comment.ID),
		},
		Content: comment.Content,
	}

	result := r.db.Model(&commentModel).Updates(map[string]interface{}{
		"content": comment.Content,
	})
	return result.Error
}

// Delete, bir yorumu siler
func (r *CommentRepository) Delete(id uint) error {
	// Yorumu bul
	var comment CommentModel
	result := r.db.First(&comment, id)
	if result.Error != nil {
		return result.Error
	}

	// Yorumu sil
	result = r.db.Delete(&comment)
	if result.Error != nil {
		return result.Error
	}

	// Yorum sayısını azalt
	r.db.Model(&NoteModel{}).Where("id = ?", comment.NoteID).Update("comment_count", gorm.Expr("comment_count - 1"))

	return nil
}

// Ensure NoteRepository implements domain.NoteRepository
var _ domain.NoteRepository = (*NoteRepository)(nil)

// Ensure CommentRepository implements domain.CommentRepository
var _ domain.CommentRepository = (*CommentRepository)(nil)
