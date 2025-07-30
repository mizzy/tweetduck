package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	archivePath string
	dbPath      string
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "tweetduck",
	Short: "Import Twitter archive data into DuckDB",
	Long:  `TweetDuck extracts data from Twitter archive ZIP files and imports it into DuckDB database.`,
	Run: func(cmd *cobra.Command, args []string) {
		if archivePath == "" {
			log.Fatal("Archive path is required. Use -archive flag.")
		}

		importer := NewImporter(archivePath, dbPath, verbose)
		if err := importer.Import(); err != nil {
			log.Fatalf("Import failed: %v", err)
		}

		fmt.Printf("Successfully imported Twitter archive to %s\n", dbPath)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&archivePath, "archive", "a", "", "Path to Twitter archive ZIP file (required)")
	rootCmd.Flags().StringVarP(&dbPath, "db", "d", "tweets.duckdb", "Output DuckDB file path")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.MarkFlagRequired("archive")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}