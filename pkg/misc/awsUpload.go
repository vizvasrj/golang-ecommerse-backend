package misc

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"src/pkg/conf"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func S3Upload(file *multipart.FileHeader, config *conf.Config) (string, string, error) {
	if config.Env.AWSAccessKeyID == "" {
		log.Println("Missing AWS keys")
		return "", "", fmt.Errorf("missing AWS keys")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Env.AWSRegion),
		Endpoint:    aws.String(config.Env.AWSEndpoint),
		Credentials: credentials.NewStaticCredentials(config.Env.AWSAccessKeyID, config.Env.AWSSecretAccessKey, ""),
	})
	if err != nil {
		return "", "", err
	}

	s3Client := s3.New(sess)

	fileContent, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer fileContent.Close()

	buffer := bytes.NewBuffer(nil)
	if _, err := buffer.ReadFrom(fileContent); err != nil {
		return "", "", err
	}

	params := &s3.PutObjectInput{
		Bucket:      aws.String(config.Env.AWSBucketName),
		Key:         aws.String(file.Filename),
		Body:        bytes.NewReader(buffer.Bytes()),
		ContentType: aws.String(file.Header.Get("Content-Type")),
	}

	result, err := s3Client.PutObject(params)
	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", config.Env.AWSBucketName, file.Filename), *result.ETag, nil
}
