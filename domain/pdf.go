package domain

import (
	"time"
)

// PDF, bir kullanıcının yüklediği PDF dosyasını temsil eder
type PDF struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	FilePath     string    `json:"filePath"`
	FileSize     int64     `json:"fileSize"`
	UserID       uint      `json:"userId"`
	Tags         []string  `json:"tags"`
	IsPublic     bool      `json:"isPublic"`
	ViewCount    int       `json:"viewCount"`
	LikeCount    int       `json:"likeCount"`
	CommentCount int       `json:"commentCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// PDFComment, bir PDF üzerindeki yorumu temsil eder
type PDFComment struct {
	ID         uint      `json:"id"`
	PDFID      uint      `json:"pdfId"`
	UserID     uint      `json:"userId"`
	Content    string    `json:"content"`
	PageNumber int       `json:"pageNumber"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// PDFAnnotation, bir PDF üzerindeki işaretlemeyi temsil eder
type PDFAnnotation struct {
	ID         uint      `json:"id"`
	PDFID      uint      `json:"pdfId"`
	UserID     uint      `json:"userId"`
	PageNumber int       `json:"pageNumber"`
	Content    string    `json:"content"`
	X          float64   `json:"x"`
	Y          float64   `json:"y"`
	Width      float64   `json:"width"`
	Height     float64   `json:"height"`
	Type       string    `json:"type"` // highlight, underline, note, etc.
	Color      string    `json:"color"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// PDFRepository, PDF verilerinin saklanması ve alınması için bir arayüz tanımlar
type PDFRepository interface {
	FindByID(id uint) (*PDF, error)
	FindByUserID(userID uint, limit, offset int) ([]*PDF, error)
	FindPublic(limit, offset int) ([]*PDF, error)
	FindByTag(tag string, limit, offset int) ([]*PDF, error)
	Search(query string, limit, offset int) ([]*PDF, error)
	Create(pdf *PDF) error
	Update(pdf *PDF) error
	Delete(id uint) error
	IncrementViewCount(id uint) error
	IncrementLikeCount(id uint) error
	DecrementLikeCount(id uint) error
}

// PDFCommentRepository, PDF yorumlarının saklanması ve alınması için bir arayüz tanımlar
type PDFCommentRepository interface {
	FindByPDFID(pdfID uint, limit, offset int) ([]*PDFComment, error)
	Create(comment *PDFComment) error
	Update(comment *PDFComment) error
	Delete(id uint) error
}

// PDFAnnotationRepository, PDF işaretlemelerinin saklanması ve alınması için bir arayüz tanımlar
type PDFAnnotationRepository interface {
	FindByPDFID(pdfID uint, limit, offset int) ([]*PDFAnnotation, error)
	FindByPDFIDAndUserID(pdfID, userID uint) ([]*PDFAnnotation, error)
	Create(annotation *PDFAnnotation) error
	Update(annotation *PDFAnnotation) error
	Delete(id uint) error
}

// PDFService, PDF ile ilgili iş mantığını içerir
type PDFService interface {
	UploadPDF(pdf *PDF, fileContent []byte) error
	UpdatePDF(pdf *PDF) error
	DeletePDF(id uint, userID uint) error
	GetPDF(id uint) (*PDF, error)
	GetPDFContent(id uint) ([]byte, error)
	GetUserPDFs(userID uint, limit, offset int) ([]*PDF, error)
	GetPublicPDFs(limit, offset int) ([]*PDF, error)
	SearchPDFs(query string, limit, offset int) ([]*PDF, error)
	AddComment(comment *PDFComment) error
	GetComments(pdfID uint, limit, offset int) ([]*PDFComment, error)
	AddAnnotation(annotation *PDFAnnotation) error
	GetAnnotations(pdfID uint, userID uint) ([]*PDFAnnotation, error)
	LikePDF(pdfID uint, userID uint) error
	UnlikePDF(pdfID uint, userID uint) error
}

// PDFStorage, PDF dosyalarının saklanması ve alınması için bir arayüz tanımlar
type PDFStorage interface {
	Save(fileContent []byte, fileName string) (string, error)
	Get(filePath string) ([]byte, error)
	Delete(filePath string) error
}
