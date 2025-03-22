package localfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/OmerFErdogan/uninote/domain"
)

// PDFStorage, domain.PDFStorage arayüzünün yerel dosya sistemi implementasyonu
type PDFStorage struct {
	basePath string
}

// NewPDFStorage, yeni bir PDFStorage örneği oluşturur
func NewPDFStorage(basePath string) (*PDFStorage, error) {
	// Dizin yoksa oluştur
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("PDF depolama dizini oluşturulamadı: %w", err)
	}
	return &PDFStorage{basePath: basePath}, nil
}

// Save, PDF dosyasını yerel dosya sistemine kaydeder
func (s *PDFStorage) Save(fileContent []byte, fileName string) (string, error) {
	// Dosya adını güvenli hale getir
	safeFileName := filepath.Clean(fileName)

	// Dosya yolunu oluştur
	filePath := filepath.Join(s.basePath, safeFileName)

	// Dosyayı kaydet
	if err := ioutil.WriteFile(filePath, fileContent, 0644); err != nil {
		return "", fmt.Errorf("PDF dosyası kaydedilemedi: %w", err)
	}

	// Dosya yolunu döndür
	return filePath, nil
}

// Get, PDF dosyasını yerel dosya sisteminden alır
func (s *PDFStorage) Get(filePath string) ([]byte, error) {
	// Dosya yolunu doğrula
	if !s.isPathSafe(filePath) {
		return nil, fmt.Errorf("güvenli olmayan dosya yolu: %s", filePath)
	}

	// Dosyayı oku
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("PDF dosyası okunamadı: %w", err)
	}

	return content, nil
}

// Delete, PDF dosyasını yerel dosya sisteminden siler
func (s *PDFStorage) Delete(filePath string) error {
	// Dosya yolunu doğrula
	if !s.isPathSafe(filePath) {
		return fmt.Errorf("güvenli olmayan dosya yolu: %s", filePath)
	}

	// Dosyayı sil
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("PDF dosyası silinemedi: %w", err)
	}

	return nil
}

// isPathSafe, dosya yolunun güvenli olup olmadığını kontrol eder
func (s *PDFStorage) isPathSafe(filePath string) bool {
	// Dosya yolunu temizle
	cleanPath := filepath.Clean(filePath)

	// Dosya yolunun basePath ile başlayıp başlamadığını kontrol et
	absBasePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return false
	}

	absFilePath, err := filepath.Abs(cleanPath)
	if err != nil {
		return false
	}

	return filepath.HasPrefix(absFilePath, absBasePath)
}

// Ensure PDFStorage implements domain.PDFStorage
var _ domain.PDFStorage = (*PDFStorage)(nil)
