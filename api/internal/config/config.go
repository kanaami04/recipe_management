package config

import (
	"os"
	"strconv"

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
	// CookieSecure は refresh Cookie に Secure 属性を付けるか。
	// 本番(HTTPS)は true。dev は http の Vite proxy 経由のため false。
	CookieSecure bool
	// AutoMigrate は起動時に AutoMigrate を実行するか。
	// dev は true(現状維持)。Lambda はコールドスタート毎に走らせないため false。
	AutoMigrate bool
	// OriginVerifySecret は CloudFront が付与する X-Origin-Verify ヘッダの期待値
	//。空なら検証しない(dev 向け)。
	OriginVerifySecret string
}

// Load は指定された env ファイル（存在すれば）と環境変数から設定を読み込む。
// envFile が空文字の場合はデフォルトの ".env" を読み込む。
func Load(envFile string) *Config {
	if envFile == "" {
		envFile = ".env"
	}
	_ = godotenv.Load(envFile)

	return &Config{
		Port:               getEnv("PORT", "8000"),
		DatabaseURL:        getEnv("DATABASE_URL", "host=localhost user=recipe password=recipe dbname=recipe port=5433 sslmode=disable TimeZone=Asia/Tokyo"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-insecure-secret-change-me"),
		CORSOrigin:         getEnv("CORS_ORIGIN", "http://localhost:5273"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		LogFormat:          getEnv("LOG_FORMAT", "text"),
		CookieSecure:       getEnvBool("COOKIE_SECURE", false),
		AutoMigrate:        getEnvBool("AUTO_MIGRATE", true),
		OriginVerifySecret: getEnv("ORIGIN_VERIFY_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}
