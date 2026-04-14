package main

import (
	"os"

	"github.com/pranjaldubey/photo-scraper/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
