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

	// --- プロフィール画像(S3 互換オブジェクトストレージ) ---
	// AvatarBucket はアバター画像を置くバケット名。
	AvatarBucket string
	// S3Region は SDK が要求するリージョン値。MinIO 互換ストレージでは値そのものは
	// 使われないが指定必須のため、本番と同じ ap-northeast-1 を既定にする。
	S3Region string
	// S3Endpoint が空なら実 AWS の S3 エンドポイントを使う(本番)。
	// 値があればそこへ向ける(ローカルの pgsty/minio 等)。
	S3Endpoint string
	// S3ForcePathStyle は virtual-hosted style ではなく path-style
	// (http://host/bucket/key)でアクセスするか。MinIO 互換ストレージでは true が必要。
	S3ForcePathStyle bool
	// S3AccessKeyID / S3SecretAccessKey が空なら SDK のデフォルト認証チェーン
	// (本番は Lambda 実行ロール)を使う。値があれば静的認証情報として使う(ローカル向け)。
	S3AccessKeyID     string
	S3SecretAccessKey string
	// AvatarPublicBaseURL が空ならアバター URL は相対パス "/{key}" になる
	// (本番: CloudFront 同一オリジン)。値があれば "{AvatarPublicBaseURL}/{key}" になる
	// (ローカル: pgsty/minio への絶対 URL)。
	AvatarPublicBaseURL string
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

		// 既定値はローカルの pgsty/minio(docker-compose)にそのまま繋がるようにする。
		// 本番は CDK が全て上書きする。S3Endpoint / 認証情報は "明示的に空文字" を渡して
		// 実 AWS + Lambda 実行ロールを使うため、これらは未設定と空を区別する lookupOr を使う。
		AvatarBucket:      getEnv("AVATAR_BUCKET", "recipe-avatars"),
		S3Region:          getEnv("S3_REGION", "ap-northeast-1"),
		S3Endpoint:        lookupOr("S3_ENDPOINT", "http://localhost:9000"),
		S3ForcePathStyle:  getEnvBool("S3_FORCE_PATH_STYLE", true),
		S3AccessKeyID:     lookupOr("S3_ACCESS_KEY_ID", "minioadmin"),
		S3SecretAccessKey: lookupOr("S3_SECRET_ACCESS_KEY", "minioadmin"),
		// 本番は "" を明示指定して相対パス "/avatars/{key}"(同一オリジンの CloudFront)にする。
		AvatarPublicBaseURL: lookupOr("AVATAR_PUBLIC_BASE_URL", "http://localhost:9000/recipe-avatars"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// lookupOr は環境変数が設定されていればその値(空文字でも)を、未設定なら fallback を返す。
// 本番で「実 AWS を使う」意図の空文字指定を、未設定によるローカル既定値と区別するために使う。
func lookupOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
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
