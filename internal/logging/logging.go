package logging

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	logger  *log.Logger
	logChan = make(chan string, 1000) // Buffer size as needed
)

// InitLogger initializes a single log file
func InitLogger() error {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %v", err)
	}
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	logger = log.New(file, "", 0)
	return nil
}

func logMessage(level, message string, details ...interface{}) {
	if logger == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if len(details) > 0 {
		logger.Printf("[%s] %s - %s | %v", level, timestamp, message, details)
	} else {
		logger.Printf("[%s] %s - %s", level, timestamp, message)
	}
}

func Info(message string, details ...interface{}) {
	logMessage("INFO", message, details...)
	select {
	case logChan <- "[INFO] " + message:
	default:
		// Drop log if channel is full to avoid blocking
	}
}

func Warn(message string, details ...interface{}) {
	logMessage("WARN", message, details...)
}

func Error(message string, details ...interface{}) {
	logMessage("ERROR", message, details...)
}

func Debug(message string, details ...interface{}) {
	logMessage("DEBUG", message, details...)
}

func Errorf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logger.Printf("[ERROR] %s - %s", timestamp, message)
}

// InitAsyncLogger initializes the asynchronous logger
func InitAsyncLogger(logFile string) {
    // Ensure the directory exists
    dir := "logs"
    if err := os.MkdirAll(dir, 0755); err != nil {
        log.Fatal("Failed to create logs directory:", err)
    }
    f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    go func() {
        for msg := range logChan {
            timestamp := time.Now().Format("2006-01-02 15:04:05")
            f.WriteString(fmt.Sprintf("%s %s\n", timestamp, msg))
        }
    }()
}

// GinLogger returns a Gin middleware for logging endpoint info and errors
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		logMsg := fmt.Sprintf("%s %s | Status: %d | IP: %s | Duration: %v", method, path, statusCode, clientIP, duration)
		if statusCode >= 400 {
			Error(logMsg)
		} else {
			Info(logMsg)
		}
	}
}

// LogRequest logs endpoint info
func LogRequest(method, endpoint, clientIP string, statusCode int) {
	msg := fmt.Sprintf("Request: %s %s from %s - Status: %d", method, endpoint, clientIP, statusCode)
	Info(msg)
}

// LogError logs error with context
func LogError(err error, context string) {
	msg := fmt.Sprintf("Error in %s: %v", context, err)
	Error(msg)
}
