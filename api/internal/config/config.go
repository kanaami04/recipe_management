package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config はアプリケーション全体の設定値を保持する。
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	CORSOrigin  string
	LogLevel    string // debug / info / warn / error
	LogFormat   string // text / json
}

// Load は指定された env ファイル（存在すれば）と環境変数から設定を読み込む。
// envFile が空文字の場合はデフォルトの ".env" を読み込む。
func Load(envFile string) *Config {
	if envFile == "" {
		envFile = ".env"
	}
	_ = godotenv.Load(envFile)

	return &Config{
		Port:        getEnv("PORT", "8000"),
		DatabaseURL: getEnv("DATABASE_URL", "host=localhost user=recipe password=recipe dbname=recipe port=5432 sslmode=disable TimeZone=Asia/Tokyo"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-insecure-secret-change-me"),
		CORSOrigin:  getEnv("CORS_ORIGIN", "http://localhost:5173"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogFormat:   getEnv("LOG_FORMAT", "text"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
