package upload

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Result struct {
	UploadURL string `json:"upload_url"`
	ImageKey  string `json:"image_key"`
	ImageURL  string `json:"image_url"`
}
type Generator interface {
	Generate(context.Context, string, string, string) (Result, error)
}
type S3 struct {
	bucket, publicBase string
	client             *s3.PresignClient
}

func New(ctx context.Context, region, bucket, publicBase, endpoint string, usePathStyle bool) (*S3, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = usePathStyle
		if endpoint != "" {
			options.BaseEndpoint = aws.String(endpoint)
		}
	})
	return &S3{bucket: bucket, publicBase: strings.TrimRight(publicBase, "/"), client: s3.NewPresignClient(client)}, nil
}
func (s *S3) Generate(ctx context.Context, userID, filename, contentType string) (Result, error) {
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	ext, ok := extensions[contentType]
	if !ok {
		return Result{}, errors.New("unsupported content type")
	}
	provided := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	if provided == ".jpeg" {
		provided = ".jpg"
	}
	if provided != ext {
		return Result{}, errors.New("filename extension does not match content type")
	}
	key := fmt.Sprintf("products/%s/%s%s", userID, uuid.NewString(), ext)
	presigned, err := s.client.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), ContentType: aws.String(contentType)}, func(o *s3.PresignOptions) { o.Expires = 15 * time.Minute })
	if err != nil {
		return Result{}, err
	}
	return Result{UploadURL: presigned.URL, ImageKey: key, ImageURL: s.publicBase + "/" + key}, nil
}
