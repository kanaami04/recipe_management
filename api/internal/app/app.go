package app

import (
	"log/slog"
	"net/http"

	"recipe-backend/internal/config"
	"recipe-backend/internal/handler"
	appmw "recipe-backend/internal/middleware"
	jwtpkg "recipe-backend/internal/pkg/jwt"
	"recipe-backend/internal/repository"
	"recipe-backend/internal/router"
	"recipe-backend/internal/service"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

// New は合成ルート（Composition Root）。各層を配線し、ミドルウェアを適用した Echo を返す。
// ルート定義は router に委譲する。main からはこの関数だけを呼べばよい。
func New(cfg *config.Config, db *gorm.DB, logger *slog.Logger) *echo.Echo {
	jwtManager := jwtpkg.NewManager(cfg.JWTSecret)

	// 各層をそれぞれの New で配線（下位層 → 上位層）。
	repos := repository.New(db)
	services := service.New(repos.User, repos.Label, repos.Recipe, jwtManager)
	handlers := handler.New(services, cfg.CookieSecure)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = newValidator()

	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())         // X-Request-Id を発行
	e.Use(appmw.RequestIDToContext()) // request_id を context へ伝播
	e.Use(appmw.RequestLogger(logger))
	// CloudFront 経由以外の直接アクセスを遮断する (adr/infra/0001)。dev は未設定で素通し。
	e.Use(appmw.RequireOriginVerify(cfg.OriginVerifySecret))
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{cfg.CORSOrigin},
		AllowMethods: []string{
			http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodDelete, http.MethodOptions,
		},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, appmw.CSRFHeaderName},
		// refresh Cookie をクロスオリジンでも送受信するため credentials を許可する
		// (api ADR-0004)。AllowOrigins はワイルドカード不可・特定オリジンのみ。
		AllowCredentials: true,
	}))

	router.Register(e, handlers, jwtManager)
	return e
}
