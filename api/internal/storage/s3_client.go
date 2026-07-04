package storage

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewS3Client は S3(または互換オブジェクトストレージ)クライアントを作る。
//
// endpoint が空なら実 AWS のエンドポイントを使い、認証は Lambda 実行ロールなど
// 環境のデフォルト認証チェーンに任せる(本番)。
// endpoint を指定するとそこへ向け、accessKeyID/secretAccessKey を静的認証情報として使う
// (ローカルの pgsty/minio 等、MinIO 互換ストレージ向け)。
func NewS3Client(ctx context.Context, region, endpoint, accessKeyID, secretAccessKey string, forcePathStyle bool) (*s3.Client, error) {
	optsFuncs := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(region)}
	if accessKeyID != "" {
		optsFuncs = append(optsFuncs, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, optsFuncs...)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = forcePathStyle
	}), nil
}
