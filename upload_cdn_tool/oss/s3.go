package oss

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3OSS struct {
	s3Client          *s3.Client
	bucketName        string
	downloadURLPrefix string
}

var _ OSS = &s3OSS{}

// https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2/s3

func NewS3OSS(accesskeyID, secretAccessKey, region, bucketName, downloadURLPrefix string) (OSS, error) {
	sdkConfig, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accesskeyID, secretAccessKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	return &s3OSS{
		s3Client:          s3.NewFromConfig(sdkConfig),
		bucketName:        bucketName,
		downloadURLPrefix: downloadURLPrefix,
	}, nil
}

// OpenObject implements OSS.
func (s *s3OSS) OpenObject(ctx context.Context, name string) (io.ReadCloser, error) {
	r, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &name,
	})
	if err != nil {
		return nil, err
	}
	return r.Body, nil
}

// UploadObject implements OSS.
func (s *s3OSS) UploadObject(ctx context.Context, name string, content io.Reader, contentLength int64) (downloadURL string, err error) {
	contentType := "application/x-gzip"
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &s.bucketName,
		Key:           &name,
		Body:          content,
		ContentType:   &contentType,
		ContentLength: &contentLength,
	}, func(o *s3.Options) { o.RetryMaxAttempts = 3 })
	if err != nil {
		return "", err
	}
	return makeDownloadURL(s.downloadURLPrefix, name), nil
}

func makeDownloadURL(prefix string, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "/" + name
}
