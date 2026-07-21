// Package mail はメール送信(本番は AWS SES v2)の薄いラッパを提供する。
// domain.Mailer を満たし、確認メール・パスワードリセットメールの本文を組み立てて送る。
package mail

import (
	"context"
	"fmt"
	"log/slog"

	"recipe-backend/internal/domain"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// New はメール送信の実装(domain.Mailer)を返す。storage.NewAvatarStorage と同じく、
// 外部 IO の抽象を実装して domain のインターフェースを返す。
// fromAddress が空なら送信基盤(SES)を用意せず、送信内容をログに出すだけの logMailer を
// 返す(ローカル開発向け。SES に相当するエミュレータが無いため、確認リンクはログで拾う)。
// 値があれば SES クライアントを作る。
func New(ctx context.Context, region, fromAddress string) (domain.Mailer, error) {
	if fromAddress == "" {
		slog.Info("mail: SES_FROM_ADDRESS 未設定のためログ出力のみのメーラーを使う(dev)")
		return &logMailer{}, nil
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &sesMailer{client: sesv2.NewFromConfig(cfg), from: fromAddress}, nil
}

// sesMailer は AWS SES v2 でメールを送る。
type sesMailer struct {
	client *sesv2.Client
	from   string
}

func (m *sesMailer) SendEmailVerification(ctx context.Context, toEmail, link string) error {
	return m.send(ctx, toEmail,
		"【Cookience】メールアドレスの確認",
		fmt.Sprintf(
			"Cookience にご登録ありがとうございます。\n\n"+
				"以下のリンクを開いてメールアドレスの確認を完了してください(有効期限 24 時間)。\n\n%s\n\n"+
				"心当たりがない場合はこのメールを破棄してください。", link),
	)
}

func (m *sesMailer) SendPasswordReset(ctx context.Context, toEmail, link string) error {
	return m.send(ctx, toEmail,
		"【Cookience】パスワードの再設定",
		fmt.Sprintf(
			"パスワード再設定のリクエストを受け付けました。\n\n"+
				"以下のリンクを開いて新しいパスワードを設定してください(有効期限 1 時間)。\n\n%s\n\n"+
				"心当たりがない場合はこのメールを破棄してください。パスワードは変更されません。", link),
	)
}

// send は SES で 1 通送る。本文はプレーンテキスト・UTF-8。
func (m *sesMailer) send(ctx context.Context, toEmail, subject, body string) error {
	utf8 := func(s string) *types.Content {
		return &types.Content{Data: &s, Charset: strPtr("UTF-8")}
	}
	_, err := m.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: &m.from,
		Destination:      &types.Destination{ToAddresses: []string{toEmail}},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: utf8(subject),
				Body:    &types.Body{Text: utf8(body)},
			},
		},
	})
	return err
}

// logMailer は送信せず、宛先とリンクをログに出す(ローカル開発向け)。
type logMailer struct{}

func (m *logMailer) SendEmailVerification(_ context.Context, toEmail, link string) error {
	slog.Info("mail(dev): email verification", "to", toEmail, "link", link)
	return nil
}

func (m *logMailer) SendPasswordReset(_ context.Context, toEmail, link string) error {
	slog.Info("mail(dev): password reset", "to", toEmail, "link", link)
	return nil
}

func strPtr(s string) *string { return &s }
