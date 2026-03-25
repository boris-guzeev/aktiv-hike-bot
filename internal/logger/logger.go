package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var log *logrus.Entry
var once sync.Once

type Fields map[string]any

type Logger interface {
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Debug(args ...any)
	Debugf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	StructuredError(msg string, err error)
	Printf(format string, args ...any)
	WithFields(fields Fields) Logger
	WithField(key string, value any) Logger
	WithError(err error) Logger
}

type structuredLogger struct {
	base *logrus.Entry
}

func (l *structuredLogger) Info(args ...any)                  { l.base.Info(args...) }
func (l *structuredLogger) Infof(format string, args ...any)  { l.base.Infof(format, args...) }
func (l *structuredLogger) Warn(args ...any)                  { l.base.Warn(args...) }
func (l *structuredLogger) Warnf(format string, args ...any)  { l.base.Warnf(format, args...) }
func (l *structuredLogger) Error(args ...any)                 { l.base.Error(args...) }
func (l *structuredLogger) Errorf(format string, args ...any) { l.base.Errorf(format, args...) }
func (l *structuredLogger) Debug(args ...any)                 { l.base.Debug(args...) }
func (l *structuredLogger) Debugf(format string, args ...any) { l.base.Debugf(format, args...) }
func (l *structuredLogger) Printf(format string, args ...any) { l.base.Printf(format, args...) }

func (l *structuredLogger) Fatal(args ...any) {
	_, file, line, ok := runtime.Caller(1)
	recorded := "unknown"
	if ok {
		recorded = fmt.Sprintf("%s:%d", getRelativePath(file), line)
	}

	l.base.WithFields(logrus.Fields{
		"recorded": recorded,
	}).Fatal(args...)
}

func (l *structuredLogger) Fatalf(format string, args ...any) {
	_, file, line, ok := runtime.Caller(1)
	recorded := "unknown"
	if ok {
		recorded = fmt.Sprintf("%s:%d", getRelativePath(file), line)
	}

	l.base.WithFields(logrus.Fields{
		"recorded": recorded,
	}).Fatalf(format, args...)
}

func (l *structuredLogger) WithFields(fields Fields) Logger {
	logrusFields := make(logrus.Fields, len(fields))
	for k, v := range fields {
		logrusFields[k] = v
	}

	return &structuredLogger{base: l.base.WithFields(logrusFields)}
}

func (l *structuredLogger) WithField(key string, value any) Logger {
	return &structuredLogger{base: l.base.WithField(key, value)}
}

func (l *structuredLogger) WithError(err error) Logger {
	return &structuredLogger{base: l.base.WithError(err)}
}

func (l *structuredLogger) StructuredError(msg string, err error) {
	recorded := "unknown"
	if _, file, line, ok := runtime.Caller(1); ok {
		recorded = fmt.Sprintf("%s:%d", getRelativePath(file), line)
	}

	raised := "unknown"
	var rich RichError
	if errors.As(err, &rich) {
		raised = fmt.Sprintf("%s:%d", rich.File, rich.Line)
	}

	l.base.WithFields(logrus.Fields{
		"recorded": recorded,
		"raised":   raised,
	}).Errorf("%s: %v", msg, err)
}

type RichError struct {
	Err  error
	File string
	Line int
}

func (e RichError) Error() string {
	//TODO: вариант с сместом ошибки в сообщении (проверить)
	// return fmt.Sprintf("%s (%s:%d)", e.Err.Error(), e.File, e.Line)
	return e.Err.Error()
}

func (e RichError) Unwrap() error {
	return e.Err
}

func WrapError(err error) error {
	if err == nil {
		return nil
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return err
	}

	return RichError{Err: err, File: getRelativePath(file), Line: line}
}

func InitLogger() Logger {
	logger := logrus.New()
	log = logrus.NewEntry(logger)
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006.01.02 15:04:05.000",
	})

	return &structuredLogger{base: log}
}

func getRelativePath(absPath string) string {
	wd, err := os.Getwd()
	if err != nil {
		return absPath
	}

	projectRoot := filepath.Dir(wd)
	relPath, err := filepath.Rel(projectRoot, absPath)
	if err != nil {
		return absPath
	}

	cleanPath := strings.TrimPrefix(relPath, "app/")
	return "./" + cleanPath
}
