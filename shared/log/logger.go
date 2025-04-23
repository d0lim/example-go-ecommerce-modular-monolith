package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger는 애플리케이션 로깅을 위한 인터페이스입니다.
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger는 새로운 로거 인스턴스를 생성합니다.
func NewLogger(environment string) *Logger {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// JSON 형태 로깅
	config.Encoding = "json"

	// 로거 생성
	logger, err := config.Build()
	if err != nil {
		// 로거 생성 실패 시 기본 로거로 폴백
		logger = zap.NewExample()
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

// NewLoggerFromEnv는 환경 변수를 기반으로 로거를 생성합니다.
func NewLoggerFromEnv() *Logger {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}
	return NewLogger(environment)
}

// WithRequestID는 요청 ID를 포함한 로거를 반환합니다.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		SugaredLogger: l.With("request_id", requestID),
	}
}

// WithContext는 추가 컨텍스트 정보를 포함한 로거를 반환합니다.
func (l *Logger) WithContext(key string, value interface{}) *Logger {
	return &Logger{
		SugaredLogger: l.With(key, value),
	}
}

// LogRequest는 HTTP 요청을 로깅합니다.
func (l *Logger) LogRequest(method, path string, status int, duration time.Duration) {
	l.Infow("HTTP 요청",
		"method", method,
		"path", path,
		"status", status,
		"duration_ms", duration.Milliseconds(),
	)
}