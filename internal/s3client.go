package internal

import (
	"github.com/minio/minio-go"
	"github.com/paysuper/paysuper-management-api/config"
)

type S3ClientInterface interface {
	Put(string, string, PutObjectOptions) (int64, error)
	Get(string, string, GetObjectOptions) error
	RemoveObject(string) error
}

type S3Client struct {
	config   *config.S3
	s3client *minio.Client
}

type GetObjectOptions minio.GetObjectOptions
type PutObjectOptions minio.PutObjectOptions

func NewS3Client(config *config.S3) (S3ClientInterface, error) {
	client := S3Client{config: config}
	mClt, err := minio.New(
		config.Endpoint,
		config.AccessKeyId,
		config.SecretKey,
		config.Secure,
	)

	if err != nil {
		return nil, err
	}

	err = mClt.MakeBucket(config.BucketName, config.Region)

	if err != nil {
		return nil, err
	}

	client.s3client = mClt

	return client, nil
}

func (c S3Client) Get(name string, filePath string, options GetObjectOptions) error {
	return c.s3client.FGetObject(c.config.BucketName, name, filePath, minio.GetObjectOptions{
		ServerSideEncryption: options.ServerSideEncryption,
	})
}

func (c S3Client) Put(name string, filePath string, options PutObjectOptions) (int64, error) {
	return c.s3client.FPutObject(c.config.BucketName, name, filePath, minio.PutObjectOptions{
		UserMetadata:            options.UserMetadata,
		ContentType:             options.ContentType,
		ContentEncoding:         options.ContentEncoding,
		ContentDisposition:      options.ContentDisposition,
		ContentLanguage:         options.ContentLanguage,
		CacheControl:            options.CacheControl,
		Progress:                options.Progress,
		ServerSideEncryption:    options.ServerSideEncryption,
		NumThreads:              options.NumThreads,
		StorageClass:            options.StorageClass,
		WebsiteRedirectLocation: options.WebsiteRedirectLocation,
	})
}

func (c S3Client) RemoveObject(name string) error {
	return c.s3client.RemoveObject(c.config.BucketName, name)
}
