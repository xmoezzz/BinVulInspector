package minio

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"github.com/minio/minio-go/v7"
)

type MinIO struct {
	client *minio.Client
	bucket string
}

type Option func(*MinIO)

func WithBucket(bucket string) Option {
	return func(o *MinIO) {
		o.bucket = bucket
	}
}

func New(client *minio.Client, opts ...Option) *MinIO {
	o := &MinIO{
		client: client,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.bucket == "" {
		o.bucket = Bucket
	}

	return o
}

func (c *MinIO) FPutObject(ctx context.Context, objectName, filePath string) (minio.UploadInfo, error) {
	ct, err := c.detectFileType(filePath)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	opts := minio.PutObjectOptions{ContentType: ct, DisableMultipart: true}
	return c.client.FPutObject(ctx, c.bucket, objectName, filePath, opts)
}

func (c *MinIO) PutDir(ctx context.Context, dst string, src string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 上传文件
		var objectName string
		{
			var rel string
			if rel, err = filepath.Rel(src, path); err != nil {
				return err
			}
			objectName = filepath.Join(dst, rel)
		}
		_, err = c.FPutObject(ctx, objectName, path)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (c *MinIO) FGetObject(ctx context.Context, objectName, filePath string) error {
	opts := minio.GetObjectOptions{}
	return c.client.FGetObject(ctx, c.bucket, objectName, filePath, opts)
}

func (c *MinIO) GetObjectBytes(ctx context.Context, objectName string) ([]byte, error) {
	opts := minio.GetObjectOptions{}
	object, err := c.client.GetObject(ctx, c.bucket, objectName, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = object.Close() }()

	var objectBytes []byte
	if objectBytes, err = io.ReadAll(object); err != nil {
		return nil, err
	}

	return objectBytes, nil
}

func (c *MinIO) GetObjectToWriter(ctx context.Context, objectName string, dst io.Writer) error {
	opts := minio.GetObjectOptions{}
	object, err := c.client.GetObject(ctx, c.bucket, objectName, opts)
	if err != nil {
		return err
	}
	defer func() { _ = object.Close() }()

	if _, err = io.Copy(dst, object); err != nil {
		return err
	}

	return nil
}

func (c *MinIO) CopyObject(ctx context.Context, dst string, src string) (minio.UploadInfo, error) {
	return c.client.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: c.bucket,
			Object: dst,
		},
		minio.CopySrcOptions{
			Bucket: c.bucket,
			Object: src,
		},
	)
}

func (c *MinIO) RemoveObject(ctx context.Context, objectName string) error {
	opts := minio.RemoveObjectOptions{}

	return c.client.RemoveObject(ctx, c.bucket, objectName, opts)
}

func (c *MinIO) ListObjects(ctx context.Context, dir string) <-chan minio.ObjectInfo {
	opts := minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: true,
	}
	return c.client.ListObjects(ctx, c.bucket, opts)
}

func (c *MinIO) StatObject(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	opts := minio.StatObjectOptions{}
	return c.client.StatObject(ctx, c.bucket, objectName, opts)
}

func (c *MinIO) Rename(ctx context.Context, newObjectName string, oldObjectName string) error {
	if oldObjectName == newObjectName {
		return nil
	}

	var err error
	if _, err = c.CopyObject(ctx, newObjectName, oldObjectName); err != nil {
		return err
	}
	return c.RemoveObject(ctx, oldObjectName)
}

// detectFile return mime type of file
func (c *MinIO) detectFileType(filePath string) (string, error) {
	ct, err := mimetype.DetectFile(filePath)
	if err != nil {
		return "", err
	}
	return ct.String(), nil
}
