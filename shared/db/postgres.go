package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Database는 PostgreSQL 데이터베이스 연결을 관리합니다.
type Database struct {
	Pool *pgxpool.Pool
}

// Config는 데이터베이스 연결 설정을 정의합니다.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabase는 PostgreSQL 데이터베이스 연결을 생성합니다.
func NewDatabase(config Config) (*Database, error) {
	// DB 연결 문자열 생성
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	// 연결 풀 설정
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection string: %w", err)
	}

	// 연결 제한 및 타임아웃 설정
	poolConfig.MaxConns = 10
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// 풀 생성
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 연결 확인
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{Pool: pool}, nil
}

// NewDatabaseFromEnv는 환경 변수에서 설정을 읽어 PostgreSQL 데이터베이스 연결을 생성합니다.
func NewDatabaseFromEnv() (*Database, error) {
	config := Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "myapp"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	return NewDatabase(config)
}

// Close는 데이터베이스 연결을 닫습니다.
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}