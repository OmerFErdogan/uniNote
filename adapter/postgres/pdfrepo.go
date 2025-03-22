package postgres

import (
	"errors"

	"github.com/OmerFErdogan/uninote/domain"
	"gorm.io/gorm"
)

// PDFModel, PDF varlığının veritabanı modelini temsil eder
type PDFModel struct {
	gorm.Model
	Title        string     `gorm:"not null"`
	Description  string     `gorm:"type:text"`
	FilePath     string     `gorm:"not null"`
	FileSize     int64      `gorm:"not null"`
	UserID       uint       `gorm:"not null"`
	Tags         []TagModel `gorm:"many2many:pdf_tags;"`
	IsPublic     bool
	ViewCount    int
	LikeCount    int
	CommentCount int
}

// PDFCommentModel, PDFComment varlığının veritabanı modelini temsil eder
type PDFCommentModel struct {
	gorm.Model
	PDFID      uint   `gorm:"not null"`
	UserID     uint   `gorm:"not null"`
	Content    string `gorm:"type:text;not null"`
	PageNumber int
}

// PDFAnnotationModel, PDFAnnotation varlığının veritabanı modelini temsil eder
type PDFAnnotationModel struct {
	gorm.Model
	PDFID      uint    `gorm:"not null"`
	UserID     uint    `gorm:"not null"`
	PageNumber int     `gorm:"not null"`
	Content    string  `gorm:"type:text"`
	X          float64 `gorm:"not null"`
	Y          float64 `gorm:"not null"`
	Width      float64 `gorm:"not null"`
	Height     float64 `gorm:"not null"`
	Type       string  `gorm:"not null"` // highlight, underline, note, etc.
	Color      string  `gorm:"not null"`
}

// ToEntity, veritabanı modelini domain varlığına dönüştürür
func (p *PDFModel) ToEntity() *domain.PDF {
	tags := make([]string, len(p.Tags))
	for i, tag := range p.Tags {
		tags[i] = tag.Name
	}

	return &domain.PDF{
		ID:           uint(p.ID),
		Title:        p.Title,
		Description:  p.Description,
		FilePath:     p.FilePath,
		FileSize:     p.FileSize,
		UserID:       p.UserID,
		Tags:         tags,
		IsPublic:     p.IsPublic,
		ViewCount:    p.ViewCount,
		LikeCount:    p.LikeCount,
		CommentCount: p.CommentCount,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// ToEntity, veritabanı modelini domain varlığına dönüştürür
func (c *PDFCommentModel) ToEntity() *domain.PDFComment {
	return &domain.PDFComment{
		ID:         uint(c.ID),
		PDFID:      c.PDFID,
		UserID:     c.UserID,
		Content:    c.Content,
		PageNumber: c.PageNumber,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

// ToEntity, veritabanı modelini domain varlığına dönüştürür
func (a *PDFAnnotationModel) ToEntity() *domain.PDFAnnotation {
	return &domain.PDFAnnotation{
		ID:         uint(a.ID),
		PDFID:      a.PDFID,
		UserID:     a.UserID,
		PageNumber: a.PageNumber,
		Content:    a.Content,
		X:          a.X,
		Y:          a.Y,
		Width:      a.Width,
		Height:     a.Height,
		Type:       a.Type,
		Color:      a.Color,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}

// PDFRepository, domain.PDFRepository arayüzünün PostgreSQL implementasyonu
type PDFRepository struct {
	db *gorm.DB
}

// NewPDFRepository, yeni bir PDFRepository örneği oluşturur
func NewPDFRepository(db *gorm.DB) *PDFRepository {
	return &PDFRepository{db: db}
}

// FindByID, ID'ye göre PDF bulur
func (r *PDFRepository) FindByID(id uint) (*domain.PDF, error) {
	var pdf PDFModel
	result := r.db.Preload("Tags").First(&pdf, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // PDF bulunamadı
		}
		return nil, result.Error
	}
	return pdf.ToEntity(), nil
}

// FindByUserID, kullanıcı ID'sine göre PDF'leri bulur
func (r *PDFRepository) FindByUserID(userID uint, limit, offset int) ([]*domain.PDF, error) {
	var pdfs []PDFModel
	result := r.db.Preload("Tags").Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&pdfs)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainPDFs []*domain.PDF
	for _, pdf := range pdfs {
		domainPDFs = append(domainPDFs, pdf.ToEntity())
	}
	return domainPDFs, nil
}

// FindPublic, herkese açık PDF'leri bulur
func (r *PDFRepository) FindPublic(limit, offset int) ([]*domain.PDF, error) {
	var pdfs []PDFModel
	result := r.db.Preload("Tags").Where("is_public = ?", true).Limit(limit).Offset(offset).Find(&pdfs)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainPDFs []*domain.PDF
	for _, pdf := range pdfs {
		domainPDFs = append(domainPDFs, pdf.ToEntity())
	}
	return domainPDFs, nil
}

// FindByTag, etikete göre PDF'leri bulur
func (r *PDFRepository) FindByTag(tag string, limit, offset int) ([]*domain.PDF, error) {
	var pdfs []PDFModel
	result := r.db.Preload("Tags").
		Joins("JOIN pdf_tags ON pdf_tags.pdf_model_id = pdf_models.id").
		Joins("JOIN tag_models ON tag_models.id = pdf_tags.tag_model_id").
		Where("tag_models.name = ?", tag).
		Limit(limit).Offset(offset).
		Find(&pdfs)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainPDFs []*domain.PDF
	for _, pdf := range pdfs {
		domainPDFs = append(domainPDFs, pdf.ToEntity())
	}
	return domainPDFs, nil
}

// Search, arama sorgusuna göre PDF'leri bulur
func (r *PDFRepository) Search(query string, limit, offset int) ([]*domain.PDF, error) {
	var pdfs []PDFModel
	result := r.db.Preload("Tags").
		Where("title ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%").
		Limit(limit).Offset(offset).
		Find(&pdfs)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainPDFs []*domain.PDF
	for _, pdf := range pdfs {
		domainPDFs = append(domainPDFs, pdf.ToEntity())
	}
	return domainPDFs, nil
}

// Create, yeni bir PDF oluşturur
func (r *PDFRepository) Create(pdf *domain.PDF) error {
	// PDF modelini oluştur
	pdfModel := PDFModel{
		Title:        pdf.Title,
		Description:  pdf.Description,
		FilePath:     pdf.FilePath,
		FileSize:     pdf.FileSize,
		UserID:       pdf.UserID,
		IsPublic:     pdf.IsPublic,
		ViewCount:    pdf.ViewCount,
		LikeCount:    pdf.LikeCount,
		CommentCount: pdf.CommentCount,
	}

	// Etiketleri işle
	if len(pdf.Tags) > 0 {
		for _, tagName := range pdf.Tags {
			var tag TagModel
			// Etiketi bul veya oluştur
			result := r.db.Where("name = ?", tagName).FirstOrCreate(&tag, TagModel{Name: tagName})
			if result.Error != nil {
				return result.Error
			}
			pdfModel.Tags = append(pdfModel.Tags, tag)
		}
	}

	// PDF'i kaydet
	result := r.db.Create(&pdfModel)
	if result.Error != nil {
		return result.Error
	}

	// ID'yi güncelle
	pdf.ID = uint(pdfModel.ID)
	return nil
}

// Update, bir PDF'i günceller
func (r *PDFRepository) Update(pdf *domain.PDF) error {
	// Mevcut PDF'i bul
	var pdfModel PDFModel
	result := r.db.First(&pdfModel, pdf.ID)
	if result.Error != nil {
		return result.Error
	}

	// PDF'i güncelle
	pdfModel.Title = pdf.Title
	pdfModel.Description = pdf.Description
	pdfModel.IsPublic = pdf.IsPublic

	// Etiketleri temizle
	r.db.Model(&pdfModel).Association("Tags").Clear()

	// Etiketleri işle
	if len(pdf.Tags) > 0 {
		for _, tagName := range pdf.Tags {
			var tag TagModel
			// Etiketi bul veya oluştur
			result := r.db.Where("name = ?", tagName).FirstOrCreate(&tag, TagModel{Name: tagName})
			if result.Error != nil {
				return result.Error
			}
			r.db.Model(&pdfModel).Association("Tags").Append(&tag)
		}
	}

	// PDF'i kaydet
	result = r.db.Save(&pdfModel)
	return result.Error
}

// Delete, bir PDF'i siler
func (r *PDFRepository) Delete(id uint) error {
	// İlişkili yorumları sil
	r.db.Where("pdf_id = ?", id).Delete(&PDFCommentModel{})

	// İlişkili işaretlemeleri sil
	r.db.Where("pdf_id = ?", id).Delete(&PDFAnnotationModel{})

	// İlişkili beğenileri sil
	r.db.Where("content_id = ? AND type = ?", id, "pdf").Delete(&ContentLikeModel{})

	// PDF'i sil
	result := r.db.Delete(&PDFModel{}, id)
	return result.Error
}

// IncrementViewCount, görüntülenme sayısını artırır
func (r *PDFRepository) IncrementViewCount(id uint) error {
	result := r.db.Model(&PDFModel{}).Where("id = ?", id).Update("view_count", gorm.Expr("view_count + 1"))
	return result.Error
}

// IncrementLikeCount, beğeni sayısını artırır
func (r *PDFRepository) IncrementLikeCount(id uint) error {
	result := r.db.Model(&PDFModel{}).Where("id = ?", id).Update("like_count", gorm.Expr("like_count + 1"))
	return result.Error
}

// DecrementLikeCount, beğeni sayısını azaltır
func (r *PDFRepository) DecrementLikeCount(id uint) error {
	result := r.db.Model(&PDFModel{}).Where("id = ?", id).Update("like_count", gorm.Expr("like_count - 1"))
	return result.Error
}

// PDFCommentRepository, domain.PDFCommentRepository arayüzünün PostgreSQL implementasyonu
type PDFCommentRepository struct {
	db *gorm.DB
}

// NewPDFCommentRepository, yeni bir PDFCommentRepository örneği oluşturur
func NewPDFCommentRepository(db *gorm.DB) *PDFCommentRepository {
	return &PDFCommentRepository{db: db}
}

// FindByPDFID, PDF ID'sine göre yorumları bulur
func (r *PDFCommentRepository) FindByPDFID(pdfID uint, limit, offset int) ([]*domain.PDFComment, error) {
	var comments []PDFCommentModel
	result := r.db.Where("pdf_id = ?", pdfID).Limit(limit).Offset(offset).Find(&comments)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainComments []*domain.PDFComment
	for _, comment := range comments {
		domainComments = append(domainComments, comment.ToEntity())
	}
	return domainComments, nil
}

// Create, yeni bir yorum oluşturur
func (r *PDFCommentRepository) Create(comment *domain.PDFComment) error {
	commentModel := PDFCommentModel{
		PDFID:      comment.PDFID,
		UserID:     comment.UserID,
		Content:    comment.Content,
		PageNumber: comment.PageNumber,
	}

	result := r.db.Create(&commentModel)
	if result.Error != nil {
		return result.Error
	}

	// Yorum sayısını artır
	r.db.Model(&PDFModel{}).Where("id = ?", comment.PDFID).Update("comment_count", gorm.Expr("comment_count + 1"))

	// ID'yi güncelle
	comment.ID = uint(commentModel.ID)
	return nil
}

// Update, bir yorumu günceller
func (r *PDFCommentRepository) Update(comment *domain.PDFComment) error {
	commentModel := PDFCommentModel{
		Model: gorm.Model{
			ID: uint(comment.ID),
		},
		Content:    comment.Content,
		PageNumber: comment.PageNumber,
	}

	result := r.db.Model(&commentModel).Updates(map[string]interface{}{
		"content":     comment.Content,
		"page_number": comment.PageNumber,
	})
	return result.Error
}

// Delete, bir yorumu siler
func (r *PDFCommentRepository) Delete(id uint) error {
	// Yorumu bul
	var comment PDFCommentModel
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
	r.db.Model(&PDFModel{}).Where("id = ?", comment.PDFID).Update("comment_count", gorm.Expr("comment_count - 1"))

	return nil
}

// PDFAnnotationRepository, domain.PDFAnnotationRepository arayüzünün PostgreSQL implementasyonu
type PDFAnnotationRepository struct {
	db *gorm.DB
}

// NewPDFAnnotationRepository, yeni bir PDFAnnotationRepository örneği oluşturur
func NewPDFAnnotationRepository(db *gorm.DB) *PDFAnnotationRepository {
	return &PDFAnnotationRepository{db: db}
}

// FindByPDFID, PDF ID'sine göre işaretlemeleri bulur
func (r *PDFAnnotationRepository) FindByPDFID(pdfID uint, limit, offset int) ([]*domain.PDFAnnotation, error) {
	var annotations []PDFAnnotationModel
	result := r.db.Where("pdf_id = ?", pdfID).Limit(limit).Offset(offset).Find(&annotations)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainAnnotations []*domain.PDFAnnotation
	for _, annotation := range annotations {
		domainAnnotations = append(domainAnnotations, annotation.ToEntity())
	}
	return domainAnnotations, nil
}

// FindByPDFIDAndUserID, PDF ID'si ve kullanıcı ID'sine göre işaretlemeleri bulur
func (r *PDFAnnotationRepository) FindByPDFIDAndUserID(pdfID, userID uint) ([]*domain.PDFAnnotation, error) {
	var annotations []PDFAnnotationModel
	result := r.db.Where("pdf_id = ? AND user_id = ?", pdfID, userID).Find(&annotations)
	if result.Error != nil {
		return nil, result.Error
	}

	var domainAnnotations []*domain.PDFAnnotation
	for _, annotation := range annotations {
		domainAnnotations = append(domainAnnotations, annotation.ToEntity())
	}
	return domainAnnotations, nil
}

// Create, yeni bir işaretleme oluşturur
func (r *PDFAnnotationRepository) Create(annotation *domain.PDFAnnotation) error {
	annotationModel := PDFAnnotationModel{
		PDFID:      annotation.PDFID,
		UserID:     annotation.UserID,
		PageNumber: annotation.PageNumber,
		Content:    annotation.Content,
		X:          annotation.X,
		Y:          annotation.Y,
		Width:      annotation.Width,
		Height:     annotation.Height,
		Type:       annotation.Type,
		Color:      annotation.Color,
	}

	result := r.db.Create(&annotationModel)
	if result.Error != nil {
		return result.Error
	}

	// ID'yi güncelle
	annotation.ID = uint(annotationModel.ID)
	return nil
}

// Update, bir işaretlemeyi günceller
func (r *PDFAnnotationRepository) Update(annotation *domain.PDFAnnotation) error {
	annotationModel := PDFAnnotationModel{
		Model: gorm.Model{
			ID: uint(annotation.ID),
		},
		Content:    annotation.Content,
		X:          annotation.X,
		Y:          annotation.Y,
		Width:      annotation.Width,
		Height:     annotation.Height,
		Type:       annotation.Type,
		Color:      annotation.Color,
		PageNumber: annotation.PageNumber,
	}

	result := r.db.Model(&annotationModel).Updates(map[string]interface{}{
		"content":     annotation.Content,
		"x":           annotation.X,
		"y":           annotation.Y,
		"width":       annotation.Width,
		"height":      annotation.Height,
		"type":        annotation.Type,
		"color":       annotation.Color,
		"page_number": annotation.PageNumber,
	})
	return result.Error
}

// Delete, bir işaretlemeyi siler
func (r *PDFAnnotationRepository) Delete(id uint) error {
	result := r.db.Delete(&PDFAnnotationModel{}, id)
	return result.Error
}

// Ensure PDFRepository implements domain.PDFRepository
var _ domain.PDFRepository = (*PDFRepository)(nil)

// Ensure PDFCommentRepository implements domain.PDFCommentRepository
var _ domain.PDFCommentRepository = (*PDFCommentRepository)(nil)

// Ensure PDFAnnotationRepository implements domain.PDFAnnotationRepository
var _ domain.PDFAnnotationRepository = (*PDFAnnotationRepository)(nil)
