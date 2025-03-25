package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger, uygulama loglaması için kullanılan yapıdır
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

var (
	// InfoLogger, bilgi mesajları için logger
	InfoLogger *log.Logger
	// ErrorLogger, hata mesajları için logger
	ErrorLogger *log.Logger
	// DebugLogger, hata ayıklama mesajları için logger
	DebugLogger *log.Logger
)

// NewLogger, yeni bir Logger örneği oluşturur
func NewLogger() *Logger {
	return &Logger{
		infoLogger:  InfoLogger,
		errorLogger: ErrorLogger,
		debugLogger: DebugLogger,
	}
}

// Info, bilgi mesajı loglar
func (l *Logger) Info(message string, keysAndValues ...interface{}) {
	if l.infoLogger != nil {
		l.logWithKeyValues(l.infoLogger, message, keysAndValues...)
	}
}

// Error, hata mesajı loglar
func (l *Logger) Error(message string, keysAndValues ...interface{}) {
	if l.errorLogger != nil {
		l.logWithKeyValues(l.errorLogger, message, keysAndValues...)
	}
}

// Debug, hata ayıklama mesajı loglar
func (l *Logger) Debug(message string, keysAndValues ...interface{}) {
	if l.debugLogger != nil {
		l.logWithKeyValues(l.debugLogger, message, keysAndValues...)
	}
}

// logWithKeyValues, anahtar-değer çiftleriyle log mesajı oluşturur
func (l *Logger) logWithKeyValues(logger *log.Logger, message string, keysAndValues ...interface{}) {
	if len(keysAndValues) == 0 {
		logger.Println(message)
		return
	}

	// Anahtar-değer çiftlerini formatlı şekilde ekle
	logMessage := message
	for i := 0; i < len(keysAndValues); i += 2 {
		key := keysAndValues[i]
		var value interface{} = "<?>"
		if i+1 < len(keysAndValues) {
			value = keysAndValues[i+1]
		}
		logMessage += fmt.Sprintf(" %v=%v", key, value)
	}
	logger.Println(logMessage)
}

// Init, logger'ları başlatır
func Init() {
	// Log dosyaları için dizin oluştur
	if err := os.MkdirAll("./logs", 0755); err != nil {
		fmt.Println("UYARI: Log dizini oluşturulamadı:", err)
	}

	// Günlük log dosyası oluştur
	currentDate := time.Now().Format("2006-01-02")
	logFile, err := os.OpenFile(fmt.Sprintf("./logs/app_%s.log", currentDate), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("UYARI: Log dosyası açılamadı:", err)
		// Dosya açılamazsa stdout'a log yaz
		InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		ErrorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
		DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		return
	}

	// Logger'ları yapılandır
	InfoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(logFile, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Ayrıca stdout'a da yazdır
	InfoLogger.SetOutput(os.Stdout)
	ErrorLogger.SetOutput(os.Stderr)
	DebugLogger.SetOutput(os.Stdout)
}

// Info, bilgi mesajı loglar
func Info(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, v...)
	}
}

// Error, hata mesajı loglar
func Error(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, v...)
	}
}

// Debug, hata ayıklama mesajı loglar
func Debug(format string, v ...interface{}) {
	if DebugLogger != nil {
		DebugLogger.Printf(format, v...)
	}
}

// LogRequest, HTTP isteğini loglar
func LogRequest(method, path, ip, userID string, statusCode int, duration time.Duration) {
	Info("[%s] %s %s - UserID: %s - Status: %d - Duration: %v", method, path, ip, userID, statusCode, duration)
}

// LogLikeOperation, beğeni işlemlerini loglar
func LogLikeOperation(operation, userID string, contentID uint, contentType string, success bool, err error) {
	if success {
		Info("[LIKE] %s - UserID: %s - ContentID: %d - Type: %s - Success: true", operation, userID, contentID, contentType)
	} else {
		Error("[LIKE] %s - UserID: %s - ContentID: %d - Type: %s - Success: false - Error: %v", operation, userID, contentID, contentType, err)
	}
}

// LogBulkOperation, toplu işlemleri loglar
func LogBulkOperation(operation, userID string, itemCount int, success bool, err error) {
	if success {
		Info("[BULK] %s - UserID: %s - ItemCount: %d - Success: true", operation, userID, itemCount)
	} else {
		Error("[BULK] %s - UserID: %s - ItemCount: %d - Success: false - Error: %v", operation, userID, itemCount, err)
	}
}
