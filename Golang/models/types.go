package models

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// GladiaClient represents the Gladia API client
type GladiaClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// S3Client represents the AWS S3 client
type S3Client struct {
	Session      *session.Session
	S3Service    *s3.S3
	Uploader     *s3manager.Uploader
	Downloader   *s3manager.Downloader
	InputBucket  string
	OutputBucket string
}

// S3FileInfo represents information about an S3 file
type S3FileInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
}

// UploadResponse represents the response from the upload endpoint
type UploadResponse struct {
	AudioURL string `json:"audio_url"`
}

// TranscriptionRequest represents the transcription request payload
type TranscriptionRequest struct {
	AudioURL        string                 `json:"audio_url"`
	Language        string                 `json:"language"`
	DetectLanguage  bool                   `json:"detect_language"`
	Subtitles       bool                   `json:"subtitles"`
	SubtitlesConfig map[string]interface{} `json:"subtitles_config"`
}

// TranscriptionResponse represents the response from the transcription endpoint
type TranscriptionResponse struct {
	ResultURL string `json:"result_url"`
}

// ResultResponse represents the final transcription result
type ResultResponse struct {
	Status string `json:"status"`
	Result struct {
		Language      string  `json:"language"`
		Confidence    float64 `json:"confidence"`
		Transcription struct {
			FullTranscript string `json:"full_transcript"`
			Utterances     []struct {
				Speaker string  `json:"speaker"`
				Start   float64 `json:"start"`
				End     float64 `json:"end"`
				Text    string  `json:"text"`
			} `json:"utterances"`
		} `json:"transcription"`
		Metadata struct {
			Duration float64 `json:"audio_duration"`
		} `json:"metadata"`
	} `json:"result"`
}