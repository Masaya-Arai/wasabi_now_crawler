package aws

import (
	"config"
	"io"
	"log"
	"os"
	"util"
	"net/http"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Storage struct {
	s3       *s3.S3
	uploader *s3manager.Uploader
}

type s3Info struct {
	accessKeyId     string
	secretAccessKey string
	region          string
}

func (ss *S3Storage) Connect() {
	cred := credentials.NewStaticCredentials(config.AWS_ACCESS_KEY_ID, config.AWS_SECRET_ACCESS_KEY, "")
	ss.s3 = s3.New(session.New(), &aws.Config{Credentials: cred, Region: aws.String(config.AWS_REGION)})
	ss.uploader = s3manager.NewUploader(session.New(&aws.Config{Credentials: cred, Region: aws.String(config.AWS_REGION)}))
}

func (ss *S3Storage) PutFile(localFilePath string, uploadingPath string) {
	file, err := os.Open(localFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = ss.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(config.AWS_BUCKET_NAME),
		Key:    aws.String(uploadingPath),
		Body:   file,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (ss *S3Storage) PutTmpFile(tmpFile io.Reader, uploadingPath string) error {
	params := &s3manager.UploadInput{
		Bucket: aws.String(config.AWS_BUCKET_NAME),
		Key:    aws.String(uploadingPath),
		Body:   tmpFile,
	}
	_, err := ss.uploader.Upload(params)

	return err
}

func (ss *S3Storage) ReplaceWithS3(ext string, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fileName := util.SecureRandom(16)+"."+ext
	if err = ss.PutTmpFile(resp.Body, fileName); err != nil {
		return "", err
	}

	return fileName, nil
}
