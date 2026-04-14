package cmd

import (
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "photo-scraper",
	Short: "Download images from websites with copyright attribution",
	Long: `A CLI tool that downloads images from target websites while preserving
photographer attribution and copyright information.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load .env from the working directory, silently ignore if absent.
		_ = godotenv.Load()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
