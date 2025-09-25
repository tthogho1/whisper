package main

import (
	"fmt"
	"os"

	"gladia-transcribe/client"
	"gladia-transcribe/processor"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// .env file not found, continue without error
	}

	var (
		apiKey           string
		language         string
		output           string
		noDetectLanguage bool
		noSubtitles      bool
		pollingInterval  int
		maxWait          int
		debug            bool
	)

	var rootCmd = &cobra.Command{
		Use:   "gladia-transcribe [file|batch]",
		Short: "Audio transcription using Gladia API",
		Long: `Transcribe MP4/audio files using Gladia API.

Usage:
  # Single file processing
  # Set API key in .env file
  echo "GLADIA_API_KEY=your_api_key_here" > .env
  gladia-transcribe input.mp4
  
  # S3 batch processing
  # Set AWS configuration and API key in .env file
  echo "GLADIA_API_KEY=your_api_key_here" >> .env
  echo "AWS_ACCESS_KEY_ID=your_access_key" >> .env
  echo "AWS_SECRET_ACCESS_KEY=your_secret_key" >> .env
  echo "AWS_REGION=us-east-1" >> .env
  echo "AWS_INPUT_BUCKET=input-bucket" >> .env
  echo "AWS_OUTPUT_BUCKET=output-bucket" >> .env
  gladia-transcribe batch
  
  # Or specify with environment variables
  export GLADIA_API_KEY=your_api_key_here
  gladia-transcribe input.mp4
  
  # Or specify with command line arguments
  gladia-transcribe input.mp4 --api-key YOUR_API_KEY`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputArg := args[0]

			// Check if this is batch processing
			if inputArg == "batch" {
				fmt.Println("🚀 Starting S3 batch processing mode...")
				fmt.Println()
				return processor.ProcessS3Files()
			}

			// Single file processing
			inputFile := inputArg

			// Get API key (priority: command line > environment variable)
			if apiKey == "" {
				apiKey = os.Getenv("GLADIA_API_KEY")
			}

			if apiKey == "" {
				fmt.Println("❌ Error: API key not specified")
				fmt.Println("Please specify the API key using one of the following methods:")
				fmt.Println("1. Command line argument: --api-key YOUR_API_KEY")
				fmt.Println("2. Environment variable: export GLADIA_API_KEY=YOUR_API_KEY")
				fmt.Println("3. .env file: echo \"GLADIA_API_KEY=YOUR_API_KEY\" > .env")
				return fmt.Errorf("API key required")
			}

			// Check if file exists
			if _, err := os.Stat(inputFile); os.IsNotExist(err) {
				return fmt.Errorf("❌ File not found: %s", inputFile)
			}

			// Display configuration
			fmt.Println("🚀 Starting Gladia API transcription")
			fmt.Printf("📁 Input file: %s\n", inputFile)
			fmt.Printf("🌐 Language setting: %s\n", language)
			if !noDetectLanguage {
				fmt.Println("🔍 Auto language detection: Enabled")
			}
			if !noSubtitles {
				fmt.Println("📝 Subtitle generation: Enabled")
			}
			fmt.Println()

			// Create Gladia service
			gladiaService := client.NewGladiaService(apiKey)

			// Validate API key
			fmt.Println("🔑 Validating API key...")
			if !gladiaService.ValidateAPIKey() {
				return fmt.Errorf("❌ Invalid API key. Please check your API key")
			}
			fmt.Println("✅ API key validated successfully")
			fmt.Println()

			// Start transcription
			fmt.Println("🚀 Starting transcription with Gladia API...")

			err := gladiaService.TranscribeFile(
				inputFile,
				language,
				output,
				!noDetectLanguage,
				!noSubtitles,
			)

			if err != nil {
				return fmt.Errorf("❌ Processing failed: %v", err)
			}

			fmt.Println()
			fmt.Println("🎉 All processing completed successfully!")
			return nil
		},
	}

	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "Gladia API key (can also be set via GLADIA_API_KEY environment variable)")
	rootCmd.Flags().StringVar(&language, "language", "ja", "Language setting (ja|en|zh|ko|es|fr|de|auto)")
	rootCmd.Flags().StringVar(&output, "output", "", "Output filename (without extension)")
	rootCmd.Flags().BoolVar(&noDetectLanguage, "no-detect-language", false, "Disable automatic language detection")
	rootCmd.Flags().BoolVar(&noSubtitles, "no-subtitles", false, "Disable subtitle file generation")
	rootCmd.Flags().IntVar(&pollingInterval, "polling-interval", 5, "Polling interval (seconds)")
	rootCmd.Flags().IntVar(&maxWait, "max-wait", 1800, "Maximum wait time (seconds)")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}