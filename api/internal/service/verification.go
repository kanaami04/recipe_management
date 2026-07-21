package service

import (
	"context"
	"net/url"

	"recipe-backend/internal/domain"
	jwtpkg "recipe-backend/internal/pkg/jwt"
)

// sendEmailVerification は user 宛にメール確認リンクを送る。
// 署名付きトークン(24h)を発行し、baseURL に token クエリを付けたリンクを組み立てる。
// 登録・再送・メール変更で共通に使う。
func sendEmailVerification(ctx context.Context, jwt *jwtpkg.Manager, mailer domain.Mailer, baseURL string, user *domain.User) error {
	token, err := jwt.GenerateEmailVerify(user.ID, user.Email)
	if err != nil {
		return err
	}
	return mailer.SendEmailVerification(ctx, user.Email, tokenLink(baseURL, token))
}

// sendPasswordReset は user 宛にパスワードリセットリンクを送る(トークン 1h)。
func sendPasswordReset(ctx context.Context, jwt *jwtpkg.Manager, mailer domain.Mailer, baseURL string, user *domain.User) error {
	token, err := jwt.GeneratePasswordReset(user.ID)
	if err != nil {
		return err
	}
	return mailer.SendPasswordReset(ctx, user.Email, tokenLink(baseURL, token))
}

// tokenLink は baseURL に token クエリを付けたリンクを返す。
// baseURL に既存のクエリがあっても壊さないよう url.Values で組み立てる。
func tokenLink(baseURL, token string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		// baseURL は設定値で通常壊れない。万一パースできなければ素朴に連結する。
		return baseURL + "?token=" + url.QueryEscape(token)
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String()
}
