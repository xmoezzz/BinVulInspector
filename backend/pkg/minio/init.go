package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
)

const (
	Bucket = "bin-vul-inspector"
)

func InitBucket(ctx context.Context, client *minio.Client) error {
	buckets := []string{Bucket}
	for _, bucket := range buckets {
		exist, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return err
		}

		// if not exist, make it!
		if !exist {
			return client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		}
	}

	return nil
}
