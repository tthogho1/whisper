package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gladia-transcribe/client"
)

// ProcessS3Files processes all MP4 files in the S3 input bucket
func ProcessS3Files() error {
	fmt.Println("🔧 Starting AWS S3 batch processing...")

	// Create S3 service
	s3Service, err := client.NewS3Service()
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %v", err)
	}

	// Create Gladia service
	apiKey := os.Getenv("GLADIA_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GLADIA_API_KEY not found in environment")
	}
	gladiaService := client.NewGladiaService(apiKey)

	// List MP4 files in S3
	fmt.Println("📋 Retrieving MP4 file list from S3...")
	files, err := s3Service.ListMP4Files()
	if err != nil {
		return fmt.Errorf("failed to retrieve S3 file list: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("⚠️  No MP4 files found for processing")
		return nil
	}

	fmt.Printf("✅ Found %d MP4 files\n", len(files))
	fmt.Println()

	// Process each file
	successCount := 0
	errorCount := 0

	for i, file := range files {
		fmt.Printf("📁 [%d/%d] Processing: %s (%.2fMB)\n", i+1, len(files), file.Key, float64(file.Size)/(1024*1024))

		// Download file from S3
		localPath, err := s3Service.DownloadFile(file.Key)
		if err != nil {
			fmt.Printf("❌ Download failed: %v\n", err)
			errorCount++
			continue
		}

		// Clean up local file when done
		defer os.Remove(localPath)

		// Upload to Gladia
		audioURL, err := gladiaService.UploadFile(localPath)
		if err != nil {
			fmt.Printf("❌ Gladia upload failed: %v\n", err)
			errorCount++
			continue
		}

		// Start transcription
		resultURL, err := gladiaService.StartTranscription(audioURL, "auto", true, false)
		if err != nil {
			fmt.Printf("❌ Transcription start failed: %v\n", err)
			errorCount++
			continue
		}

		// Poll for results
		result, err := gladiaService.PollResult(resultURL, 5, 1800)
		if err != nil {
			fmt.Printf("❌ Result retrieval failed: %v\n", err)
			errorCount++
			continue
		}

		// Convert result to JSON
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("❌ JSON conversion failed: %v\n", err)
			errorCount++
			continue
		}

		// Generate output key (replace .mp4 with .json)
		outputKey := strings.TrimSuffix(file.Key, ".mp4") + "_transcription.json"

		// Upload JSON to S3
		err = s3Service.UploadJSON(outputKey, jsonData)
		if err != nil {
			fmt.Printf("❌ Result upload failed: %v\n", err)
			errorCount++
			continue
		}

		fmt.Printf("✅ Completed: %s → %s\n", file.Key, outputKey)
		successCount++
		fmt.Println()
	}

	// Summary
	fmt.Printf("🎉 Batch processing completed!\n")
	fmt.Printf("• Success: %d files\n", successCount)
	fmt.Printf("• Failed: %d files\n", errorCount)
	fmt.Printf("• Total: %d files\n", len(files))

	return nil
}