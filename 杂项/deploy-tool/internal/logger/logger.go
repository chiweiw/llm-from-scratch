package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var (
	logger       *Logger
	loggerMutex  sync.Mutex
	defaultPath  string
	eventEmitter func(level string, message string, ts string, line string)
)

func SetEventEmitter(emitter func(level string, message string, ts string, line string)) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	eventEmitter = emitter
}

type Logger struct {
	mu         sync.Mutex
	level      Level
	file       *os.File
	fileWriter io.Writer
	console    *log.Logger
	fileLog    *log.Logger
	logPath    string
	logLines   []string
	maxLines   int
}

func init() {
	Init(InfoLevel)
}

func Init(level Level) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if logger != nil && logger.file != nil {
		logger.file.Close()
	}

	scriptDir := getScriptDir()
	if scriptDir == "" {
		scriptDir = os.TempDir()
	}

	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		scriptDir = os.TempDir()
	}

	timestamp := time.Now().Format("20060102_150405")
	logPath := filepath.Join(scriptDir, fmt.Sprintf("deploy_%s.log", timestamp))
	defaultPath = logPath

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}

	var consoleWriter io.Writer = os.Stdout
	// Suppress console output when running wailsbindings to avoid interfering with Wails CLI output
	if strings.Contains(strings.ToLower(filepath.Base(os.Args[0])), "wailsbindings") {
		consoleWriter = io.Discard
	}

	logger = &Logger{
		level:      level,
		file:       file,
		fileWriter: io.MultiWriter(consoleWriter, file),
		console:    log.New(consoleWriter, "", 0),
		fileLog:    log.New(file, "", 0),
		logPath:    logPath,
		logLines:   []string{},
		maxLines:   1000,
	}

	logger.log(InfoLevel, "日志文件: %s", logPath)
}

func getScriptDir() string {
	exePath, err := os.Executable()
	if err != nil {
		// Fallback silently
		return ""
	}
	dir := filepath.Dir(exePath)
	return dir
}

func GetLogPath() string {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if logger != nil {
		return logger.logPath
	}
	return defaultPath
}

func SetLevel(level Level) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if logger != nil {
		logger.level = level
	}
}

func Debug(format string, v ...interface{}) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if logger != nil && logger.level <= DebugLevel {
		logger.log(DebugLevel, format, v...)
	}
}

func Info(format string, v ...interface{}) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if logger != nil && logger.level <= InfoLevel {
		logger.log(InfoLevel, format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if logger != nil && logger.level <= WarnLevel {
		logger.log(WarnLevel, format, v...)
	}
}

func Error(format string, v ...interface{}) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if logger != nil {
		logger.log(ErrorLevel, format, v...)
	}
}

func Fatal(format string, v ...interface{}) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if logger != nil {
		logger.log(ErrorLevel, format, v...)
	}
	if logger != nil && logger.file != nil {
		logger.file.Close()
	}
	os.Exit(1)
}

func (l *Logger) log(level Level, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	levelStr := "DEBUG"
	switch level {
	case InfoLevel:
		levelStr = "INFO"
	case WarnLevel:
		levelStr = "WARN"
	case ErrorLevel:
		levelStr = "ERROR"
	}

	logLine := fmt.Sprintf("%s - %s - %s", timestamp, levelStr, msg)

	l.console.Println(logLine)
	l.fileLog.Println(logLine)

	l.mu.Lock()
	l.logLines = append(l.logLines, logLine)
	if len(l.logLines) > l.maxLines {
		l.logLines = l.logLines[len(l.logLines)-l.maxLines:]
	}
	l.mu.Unlock()

	if eventEmitter != nil {
		eventEmitter(levelStr, msg, timestamp, logLine)
	}
}

func Close() {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if logger != nil && logger.file != nil {
		logger.file.Close()
		logger.file = nil
	}
}

func GetRecentLogs(count int) []string {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if logger == nil {
		return []string{}
	}
	if count <= 0 || count > len(logger.logLines) {
		return logger.logLines
	}
	return logger.logLines[len(logger.logLines)-count:]
}
