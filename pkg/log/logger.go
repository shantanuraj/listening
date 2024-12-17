package log

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[41m"
	colorGreen  = "\033[42m"
	colorYellow = "\033[43m"
)

type Logger struct {
	logger *log.Logger
	debug  *log.Logger
	info   *log.Logger
	warn   *log.Logger
	error  *log.Logger
}

var defaultLogger = New()

func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		debug:  log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		info:   log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		warn:   log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime),
		error:  log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) colorizeStatus(status int) string {
	code := fmt.Sprintf(" %d ", status)
	switch {
	case status >= 500:
		return fmt.Sprintf("%s%s%s", colorRed, code, colorReset)
	case status >= 400:
		return fmt.Sprintf("%s%s%s", colorYellow, code, colorReset)
	default:
		return fmt.Sprintf("%s%s%s", colorGreen, code, colorReset)
	}
}

func (l *Logger) Debugf(format string, args ...any) {
	l.debug.Printf(format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.info.Printf(format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.warn.Printf(format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.error.Printf(format, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.error.Fatalf(format, args...)
}

func (l *Logger) LogRequest(method, path string, status int, duration time.Duration) {
	l.Infof("%s %s %s %v", l.colorizeStatus(status), method, path, duration)
}

func Debugf(format string, args ...any) {
	defaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	defaultLogger.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	defaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...any) {
	defaultLogger.Fatalf(format, args...)
}
