package storage

import (
	"context"
	"time"

	"recipe-backend/internal/domain"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// presignExpiry はアップロード用署名付き URL の有効期限。
const presignExpiry = 5 * time.Minute

type avatarStorage struct {
	client     *s3.Client
	presign    *s3.PresignClient
	bucket     string
	publicBase string // 空なら相対パス "/{key}" を返す(CloudFront 同一オリジン向け)
}

// NewAvatarStorage は S3 互換クライアントから domain.AvatarStorage を作る。
// publicBaseURL が空文字なら、PublicURL は相対パスを返す(本番: CloudFront 同一オリジン)。
// 値があれば "{publicBaseURL}/{key}" を返す(ローカル: pgsty/minio への絶対 URL)。
func NewAvatarStorage(client *s3.Client, bucket, publicBaseURL string) domain.AvatarStorage {
	return &avatarStorage{
		client:     client,
		presign:    s3.NewPresignClient(client),
		bucket:     bucket,
		publicBase: publicBaseURL,
	}
}

func (s *avatarStorage) PresignUpload(ctx context.Context, key, contentType string) (string, error) {
	req, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		ContentType: &contentType,
	}, s3.WithPresignExpires(presignExpiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *avatarStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &s.bucket, Key: &key})
	return err
}

func (s *avatarStorage) PublicURL(key string) string {
	if s.publicBase == "" {
		return "/" + key
	}
	return s.publicBase + "/" + key
}
