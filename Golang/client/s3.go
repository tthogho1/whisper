package client

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"gladia-transcribe/models"
)

// S3Service represents the AWS S3 service
type S3Service struct {
	client *models.S3Client
}

// NewS3Service creates a new S3 service with AWS credentials
func NewS3Service() (*S3Service, error) {
	// Get AWS credentials from environment
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	inputBucket := os.Getenv("AWS_INPUT_BUCKET")
	outputBucket := os.Getenv("AWS_OUTPUT_BUCKET")

	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("AWS credentials not found in environment variables")
	}
	if region == "" {
		region = "us-east-1" // Default region
	}
	if inputBucket == "" || outputBucket == "" {
		return nil, fmt.Errorf("AWS_INPUT_BUCKET and AWS_OUTPUT_BUCKET must be set")
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"", // token
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	return &S3Service{
		client: &models.S3Client{
			Session:      sess,
			S3Service:    s3.New(sess),
			Uploader:     s3manager.NewUploader(sess),
			Downloader:   s3manager.NewDownloader(sess),
			InputBucket:  inputBucket,
			OutputBucket: outputBucket,
		},
	}, nil
}

// ListMP4Files lists all MP4 files in the input S3 bucket
func (s3s *S3Service) ListMP4Files() ([]models.S3FileInfo, error) {
	var files []models.S3FileInfo

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s3s.client.InputBucket),
	}

	err := s3s.client.S3Service.ListObjectsV2Pages(input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			// Check if file has MP4 extension
			if strings.HasSuffix(strings.ToLower(*obj.Key), ".mp4") {
				files = append(files, models.S3FileInfo{
					Key:          *obj.Key,
					Size:         *obj.Size,
					LastModified: *obj.LastModified,
				})
			}
		}
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %v", err)
	}

	return files, nil
}

// DownloadFile downloads a file from S3 to local temporary directory
func (s3s *S3Service) DownloadFile(key string) (string, error) {
	// Create temporary file
	tempDir := os.TempDir()
	localPath := filepath.Join(tempDir, filepath.Base(key))

	// Create the file
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer file.Close()

	// Download from S3
	input := &s3.GetObjectInput{
		Bucket: aws.String(s3s.client.InputBucket),
		Key:    aws.String(key),
	}

	_, err = s3s.client.Downloader.Download(file, input)
	if err != nil {
		os.Remove(localPath) // Clean up on error
		return "", fmt.Errorf("failed to download file from S3: %v", err)
	}

	log.Printf("Downloaded %s to %s", key, localPath)
	return localPath, nil
}

// DownloadFileStream downloads a file from S3 directly to memory
func (s3s *S3Service) DownloadFileStream(key string) (*bytes.Buffer, string, error) {
	// Get object from S3
	input := &s3.GetObjectInput{
		Bucket: aws.String(s3s.client.InputBucket),
		Key:    aws.String(key),
	}

	result, err := s3s.client.S3Service.GetObject(input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object from S3: %v", err)
	}
	defer result.Body.Close()

	// Create buffer and read data
	buffer := &bytes.Buffer{}
	_, err = buffer.ReadFrom(result.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read object data: %v", err)
	}

	// Determine content type based on file extension
	contentType := "video/mp4" // Default for MP4 files
	if strings.HasSuffix(strings.ToLower(key), ".mp4") {
		contentType = "video/mp4"
	} else if strings.HasSuffix(strings.ToLower(key), ".wav") {
		contentType = "audio/wav"
	} else if strings.HasSuffix(strings.ToLower(key), ".mp3") {
		contentType = "audio/mpeg"
	}

	log.Printf("Downloaded to memory: %s (%.2fMB)", key, float64(buffer.Len())/(1024*1024))
	return buffer, contentType, nil
}

// UploadJSON uploads a JSON result to the output S3 bucket
func (s3s *S3Service) UploadJSON(key string, jsonData []byte) error {
	input := &s3manager.UploadInput{
		Bucket:      aws.String(s3s.client.OutputBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(jsonData),
		ContentType: aws.String("application/json"),
	}

	_, err := s3s.client.Uploader.Upload(input)
	if err != nil {
		return fmt.Errorf("failed to upload JSON to S3: %v", err)
	}

	log.Printf("Uploaded result to s3://%s/%s", s3s.client.OutputBucket, key)
	return nil
}