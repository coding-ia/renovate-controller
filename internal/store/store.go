package store

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

func GetS3Object(bucketName string, key string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	getObjectOutput, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("Failed to get object from S3, %v", err)
	}
	defer getObjectOutput.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, getObjectOutput.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read object data, %v", err)
	}

	return buf.String(), nil
}
