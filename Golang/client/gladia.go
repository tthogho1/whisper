package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gladia-transcribe/models"
)

// GladiaService represents the Gladia API service
type GladiaService struct {
	client *models.GladiaClient
}

// NewGladiaService creates a new Gladia API service
func NewGladiaService(apiKey string) *GladiaService {
	return &GladiaService{
		client: &models.GladiaClient{
			APIKey:  apiKey,
			BaseURL: "https://api.gladia.io/v2",
			Client:  &http.Client{Timeout: 300 * time.Second},
		},
	}
}

// ValidateAPIKey checks if the API key is valid
func (gs *GladiaService) ValidateAPIKey() bool {
	if len(gs.client.APIKey) < 10 {
		log.Println("API key is too short or not set")
		return false
	}
	
	// Basic format check
	log.Printf("API key format check completed: %s...", gs.client.APIKey[:10])
	return true
}

// UploadFile uploads a file to Gladia API
func (gs *GladiaService) UploadFile(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get file information: %v", err)
	}

	fileSizeMB := float64(fileInfo.Size()) / (1024 * 1024)
	log.Printf("Uploading file: %s (%.2fMB)", filepath.Base(filePath), fileSizeMB)

	// Check file size limit (50MB)
	if fileSizeMB > 50 {
		return "", fmt.Errorf("file size too large: %.2fMB (maximum 50MB)", fileSizeMB)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file to form
	part, err := writer.CreateFormFile("audio", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}

	// Copy file content to form
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file data: %v", err)
	}

	// Close the writer to finalize the form
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize form: %v", err)
	}

	// Create request
	url := gs.client.BaseURL + "/upload"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("x-gladia-key", gs.client.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Debug log
	log.Printf("Upload request: %s", url)

	// Send request
	resp, err := gs.client.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Check status code
	if resp.StatusCode != 200 {
		log.Printf("Error response: %s", string(body))
		
		switch resp.StatusCode {
		case 400:
			return "", fmt.Errorf("400 Bad Request: Invalid request. Please check API key or file format")
		case 401:
			return "", fmt.Errorf("401 Unauthorized: Invalid or expired API key")
		case 413:
			return "", fmt.Errorf("413 Payload Too Large: File size too large")
		case 415:
			return "", fmt.Errorf("415 Unsupported Media Type: Unsupported file format")
		default:
			return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
	}

	// Parse response
	var uploadResp models.UploadResponse
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	log.Printf("Upload completed: %s", uploadResp.AudioURL)
	return uploadResp.AudioURL, nil
}

// UploadFileStream uploads file data from memory buffer to Gladia
func (gs *GladiaService) UploadFileStream(buffer *bytes.Buffer, filename, contentType string) (string, error) {
	fileSizeMB := float64(buffer.Len()) / (1024 * 1024)
	log.Printf("Uploading file from memory: %s (%.2fMB)", filename, fileSizeMB)

	// Check file size limit (50MB)
	// if fileSizeMB > 50 {
	// 	return "", fmt.Errorf("file size too large: %.2fMB (maximum 50MB)", fileSizeMB)
	// }

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file to form
	part, err := writer.CreateFormFile("audio", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}

	// Copy buffer content to form
	if _, err := io.Copy(part, buffer); err != nil {
		return "", fmt.Errorf("failed to copy buffer data: %v", err)
	}

	// Close the writer to finalize the form
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize form: %v", err)
	}

	// Create request
	url := gs.client.BaseURL + "/upload"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("x-gladia-key", gs.client.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Debug log
	log.Printf("Stream upload request: %s", url)

	// Send request
	resp, err := gs.client.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("stream upload request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Check status code
	if resp.StatusCode != 200 {
		log.Printf("Error response: %s", string(body))
		
		switch resp.StatusCode {
		case 400:
			return "", fmt.Errorf("400 Bad Request: Invalid request. Please check API key or file format")
		case 401:
			return "", fmt.Errorf("401 Unauthorized: Invalid or expired API key")
		case 413:
			return "", fmt.Errorf("413 Payload Too Large: File size too large")
		case 415:
			return "", fmt.Errorf("415 Unsupported Media Type: Unsupported file format")
		default:
			return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
	}

	// Parse response
	var uploadResp models.UploadResponse
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	log.Printf("Stream upload completed: %s", uploadResp.AudioURL)
	return uploadResp.AudioURL, nil
}

// StartTranscription starts the transcription process
func (gs *GladiaService) StartTranscription(audioURL, language string, detectLanguage, enableSubtitles bool) (string, error) {
	// Prepare request payload
	subtitlesConfig := map[string]interface{}{
		"formats": []string{"srt", "vtt"},
	}

	// Create base request data
	reqData := map[string]interface{}{
		"audio_url":        audioURL,
		"detect_language":  detectLanguage,
		"subtitles":        enableSubtitles,
		"subtitles_config": subtitlesConfig,
	}

	// Only add language if it's not empty and detect_language is false
	if language != "" && !detectLanguage {
		reqData["language"] = language
	} else if detectLanguage {
		// For auto-detection, don't include language field
		log.Printf("Using automatic language detection")
	} else {
		// Default to Japanese if no language specified and auto-detect is off
		reqData["language"] = "ja"
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("failed to create request data: %v", err)
	}

	// Create request
	url := gs.client.BaseURL + "/pre-recorded"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("x-gladia-key", gs.client.APIKey)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Starting transcription: %s", url)

	// Send request
	resp, err := gs.client.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("transcription request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Check status code
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", fmt.Errorf("failed to start transcription: HTTP %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var transcriptionResp models.TranscriptionResponse
	if err := json.Unmarshal(body, &transcriptionResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	log.Printf("Result URL: %s", transcriptionResp.ResultURL)
	return transcriptionResp.ResultURL, nil
}

// PollResult polls for transcription results
func (gs *GladiaService) PollResult(resultURL string, intervalSecs, maxWaitSecs int) (*models.ResultResponse, error) {
	log.Printf("Waiting for results (max %d seconds, checking every %d seconds)...", maxWaitSecs, intervalSecs)

	startTime := time.Now()
	maxWaitDuration := time.Duration(maxWaitSecs) * time.Second
	interval := time.Duration(intervalSecs) * time.Second

	for {
		// Check elapsed time
		if time.Since(startTime) > maxWaitDuration {
			return nil, fmt.Errorf("processing timed out (%d seconds)", maxWaitSecs)
		}

		// Create request
		req, err := http.NewRequest("GET", resultURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create result request: %v", err)
		}

		// Set headers
		req.Header.Set("x-gladia-key", gs.client.APIKey)

		// Send request
		resp, err := gs.client.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("result request failed: %v", err)
		}

		// Read response
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %v", err)
		}

		// Check status code
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("failed to get results: HTTP %d - %s", resp.StatusCode, string(body))
		}

		// Parse response
		var result models.ResultResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("failed to parse results: %v", err)
		}

		// Check if processing is complete
		switch result.Status {
		case "done":
			log.Printf("✅ Processing completed!")
			return &result, nil
		case "error":
			return nil, fmt.Errorf("❌ Error occurred during processing")
		case "processing", "queued":
			elapsed := int(time.Since(startTime).Seconds())
			log.Printf("⏳ Processing... (elapsed: %d seconds)", elapsed)
			time.Sleep(interval)
		default:
			log.Printf("⚠️  Unknown status: %s", result.Status)
			time.Sleep(interval)
		}
	}
}

// SaveResults saves the transcription results to files
func (gs *GladiaService) SaveResults(result *models.ResultResponse, outputPath string) (map[string]string, error) {
	savedFiles := make(map[string]string)
	
	// Determine base name
	baseName := outputPath
	if baseName == "" {
		baseName = fmt.Sprintf("gladia_transcript_%s", time.Now().Format("20060102_150405"))
	}

	// Get transcription data
	fullTranscript := result.Result.Transcription.FullTranscript
	language := result.Result.Language
	confidence := result.Result.Confidence
	duration := result.Result.Metadata.Duration

	// Save text file
	txtPath := baseName + ".txt"
	txtContent := fmt.Sprintf("=== Gladia API Transcription Results ===\n")
	txtContent += fmt.Sprintf("Processing Date: %s\n", time.Now().Format("2006-01-02T15:04:05"))
	txtContent += fmt.Sprintf("Language: %s\n", language)
	txtContent += fmt.Sprintf("Confidence: %.2f\n", confidence)
	txtContent += fmt.Sprintf("Audio Duration: %.2f seconds\n", duration)
	txtContent += fmt.Sprintf("%s\n\n", strings.Repeat("=", 50))
	txtContent += fullTranscript

	// Add segment details if available
	if len(result.Result.Transcription.Utterances) > 0 {
		txtContent += "\n\n=== Segment Details ===\n"
		for _, utterance := range result.Result.Transcription.Utterances {
			txtContent += fmt.Sprintf("[%.2fs - %.2fs] %s\n", utterance.Start, utterance.End, utterance.Text)
		}
	}

	if err := os.WriteFile(txtPath, []byte(txtContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to save text file: %v", err)
	}
	savedFiles["txt"] = txtPath
	log.Printf("Text file saved: %s", txtPath)

	// Save JSON file
	jsonPath := baseName + ".json"
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON data: %v", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return nil, fmt.Errorf("failed to save JSON file: %v", err)
	}
	savedFiles["json"] = jsonPath
	log.Printf("JSON file saved: %s", jsonPath)

	return savedFiles, nil
}

// TranscribeFile is the main function to transcribe a single file
func (gs *GladiaService) TranscribeFile(filePath, language, outputPath string, detectLanguage, enableSubtitles bool) error {
	// Upload file
	log.Println("📁 ファイルをGladia APIにアップロード中...")
	audioURL, err := gs.UploadFile(filePath)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}
	
	// Prepare output path
	if outputPath == "" {
		ext := filepath.Ext(filePath)
		outputPath = strings.TrimSuffix(filePath, ext) + "_transcription"
	}

	// Start transcription
	log.Println("🎤 Starting transcription...")
	resultURL, err := gs.StartTranscription(audioURL, language, detectLanguage, enableSubtitles)
	if err != nil {
		return fmt.Errorf("failed to start transcription: %v", err)
	}

	// Poll for results
	log.Println("⏳ Waiting for results...")
	result, err := gs.PollResult(resultURL, 5, 1800)
	if err != nil {
		return fmt.Errorf("failed to get results: %v", err)
	}

	// Save results
	log.Println("💾 Saving results...")
	savedFiles, err := gs.SaveResults(result, outputPath)
	if err != nil {
		return fmt.Errorf("failed to save results: %v", err)
	}

	// Display results
	fmt.Println("✅ Transcription completed successfully!")
	fmt.Printf("• Detected language: %s\n", result.Result.Language)
	fmt.Printf("• Confidence: %.2f\n", result.Result.Confidence)
	fmt.Printf("• Audio duration: %.2f seconds\n", result.Result.Metadata.Duration)
	fmt.Println()

	// Preview text
	previewText := result.Result.Transcription.FullTranscript
	if len(previewText) > 500 {
		previewText = previewText[:500] + "..."
	}

	fmt.Println("📄 Transcription preview:")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println(previewText)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	// Show saved files
	fmt.Println("💾 Saved files:")
	for fileType, filePath := range savedFiles {
		fmt.Printf("• %s: %s\n", strings.ToUpper(fileType), filePath)
	}

	return nil
}