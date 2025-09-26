package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"gladia-transcribe/client"
	"gladia-transcribe/models"
)

// ProcessResult represents the result of processing a single file
type ProcessResult struct {
	FileName    string
	Success     bool
	Error       error
	OutputKey   string
	ProcessTime time.Duration
}

// ProcessS3Files processes all MP4 files in the S3 input bucket with parallel processing
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

	// Get worker count from environment or use default
	maxWorkers := getMaxWorkers()
	fmt.Printf("🔧 Using %d parallel workers\n", maxWorkers)
	fmt.Println()

	// Process files with parallel workers
	return processFilesParallel(files, s3Service, gladiaService, maxWorkers)
}

// processFilesParallel processes files using multiple goroutines
func processFilesParallel(files []models.S3FileInfo, s3Service *client.S3Service, gladiaService *client.GladiaService, maxWorkers int) error {
	// Create file queue channel
	fileQueue := make(chan models.S3FileInfo, len(files))
	
	// Result tracking
	results := make(chan ProcessResult, len(files))
	
	// Add all files to the queue
	for _, file := range files {
		fileQueue <- file
	}
	close(fileQueue) // Close channel to signal no more files

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(i+1, fileQueue, results, s3Service, gladiaService, &wg)
	}

	// Wait for all workers to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(results) // Close results channel when all workers are done
	}()

	// Collect and display results
	return collectResults(results, len(files))
}

// worker processes files from the queue
func worker(workerID int, fileQueue <-chan models.S3FileInfo, results chan<- ProcessResult, s3Service *client.S3Service, gladiaService *client.GladiaService, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range fileQueue {
		startTime := time.Now()
		fmt.Printf("🔄 Worker %d: Starting %s (%.2fMB)\n", workerID, file.Key, float64(file.Size)/(1024*1024))

		result := ProcessResult{
			FileName: file.Key,
			Success:  false,
		}

		// Process the file
		outputKey, err := processFile(file, s3Service, gladiaService)
		if err != nil {
			fmt.Printf("❌ Worker %d: Failed %s - %v\n", workerID, file.Key, err)
			result.Error = err
		} else {
			fmt.Printf("✅ Worker %d: Completed %s → %s\n", workerID, file.Key, outputKey)
			result.Success = true
			result.OutputKey = outputKey
		}

		result.ProcessTime = time.Since(startTime)
		results <- result
	}

	fmt.Printf("👷 Worker %d: Finished\n", workerID)
}

// processFile processes a single file using memory streaming (no disk I/O)
func processFile(file models.S3FileInfo, s3Service *client.S3Service, gladiaService *client.GladiaService) (string, error) {
	// Download file from S3 directly to memory
	buffer, contentType, err := s3Service.DownloadFileStream(file.Key)
	if err != nil {
		return "", fmt.Errorf("stream download failed: %v", err)
	}

	// Upload buffer directly to Gladia (no disk I/O)
	audioURL, err := gladiaService.UploadFileStream(buffer, file.Key, contentType)
	if err != nil {
		return "", fmt.Errorf("gladia stream upload failed: %v", err)
	}

	// Start transcription (use empty language for auto-detection)
	resultURL, err := gladiaService.StartTranscription(audioURL, "", true, false)
	if err != nil {
		return "", fmt.Errorf("transcription start failed: %v", err)
	}

	// Poll for results
	result, err := gladiaService.PollResult(resultURL, 5, 1800)
	if err != nil {
		return "", fmt.Errorf("result retrieval failed: %v", err)
	}

	// Convert result to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json conversion failed: %v", err)
	}

	// Generate output key (replace .mp4 with .json)
	outputKey := strings.TrimSuffix(file.Key, ".mp4") + "_transcription.json"

	// Upload JSON to S3
	err = s3Service.UploadJSON(outputKey, jsonData)
	if err != nil {
		return "", fmt.Errorf("result upload failed: %v", err)
	}

	return outputKey, nil
}

// collectResults collects and displays processing results
func collectResults(results <-chan ProcessResult, totalFiles int) error {
	var (
		successCount int
		errorCount   int
		totalTime    time.Duration
		errors       []string
	)

	fmt.Println("\n📊 Processing Results:")
	fmt.Println("─────────────────────")

	for result := range results {
		totalTime += result.ProcessTime

		if result.Success {
			successCount++
			fmt.Printf("✅ %s (%.1fs)\n", result.FileName, result.ProcessTime.Seconds())
		} else {
			errorCount++
			fmt.Printf("❌ %s (%.1fs) - %v\n", result.FileName, result.ProcessTime.Seconds(), result.Error)
			errors = append(errors, fmt.Sprintf("%s: %v", result.FileName, result.Error))
		}
	}

	// Summary
	fmt.Println("\n🎉 Batch processing completed!")
	fmt.Printf("• Success: %d files\n", successCount)
	fmt.Printf("• Failed: %d files\n", errorCount)
	fmt.Printf("• Total: %d files\n", totalFiles)
	fmt.Printf("• Average time per file: %.1fs\n", totalTime.Seconds()/float64(totalFiles))

	if len(errors) > 0 {
		fmt.Println("\n❌ Error details:")
		for _, err := range errors {
			fmt.Printf("  • %s\n", err)
		}
	}

	return nil
}

// getMaxWorkers gets the maximum number of workers from environment or returns default
func getMaxWorkers() int {
	maxWorkers := 3 // Default value
	if workerEnv := os.Getenv("MAX_WORKERS"); workerEnv != "" {
		if parsed, err := fmt.Sscanf(workerEnv, "%d", &maxWorkers); err != nil || parsed != 1 {
			fmt.Printf("⚠️  Invalid MAX_WORKERS value, using default: %d\n", maxWorkers)
		}
	}
	
	// Safety limits
	if maxWorkers < 1 {
		maxWorkers = 1
	} else if maxWorkers > 10 {
		maxWorkers = 10 // Prevent too many concurrent API calls
	}
	
	return maxWorkers
}
