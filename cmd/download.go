package cmd

import (
	"fmt"
	"os"

	"github.com/pranjaldubey/photo-scraper/internal"
	"github.com/spf13/cobra"
)

var (
	query   string
	count   int
	output  string
	quality string
	apiKey  string
	dryRun  bool
)

func init() {
	downloadCmd.Flags().StringVarP(&query, "query", "q", "", "search query (required)")
	downloadCmd.Flags().IntVarP(&count, "count", "c", 10, "number of images to download")
	downloadCmd.Flags().StringVarP(&output, "output", "o", "./photos", "output directory")
	downloadCmd.Flags().StringVar(&quality, "quality", "regular", "image quality: raw, full, regular, small, thumb")
	downloadCmd.Flags().StringVar(&apiKey, "api-key", "", "Unsplash API access key (or set UNSPLASH_ACCESS_KEY)")
	downloadCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be downloaded without downloading")

	downloadCmd.MarkFlagRequired("query")

	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Search and download photos from Unsplash",
	Long: `Search Unsplash for photos matching a query and download them along with
photographer attribution and metadata. Each photo is saved as an image file
with a companion .json file containing copyright and EXIF information.`,
	Example: `  photo-scraper download -q "mountains" -c 20 -o ./mountains
  photo-scraper download -q "street photography" --quality full --dry-run
  UNSPLASH_ACCESS_KEY=xxx photo-scraper download -q "nature"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := apiKey
		if key == "" {
			key = os.Getenv("UNSPLASH_ACCESS_KEY")
		}
		if key == "" {
			return fmt.Errorf("Unsplash API key required: use --api-key or set UNSPLASH_ACCESS_KEY env var.\n" +
				"Get one at https://unsplash.com/developers")
		}

		return internal.NewUnsplashClient(key).Download(cmd.Context(), internal.DownloadConfig{
			Query:   query,
			Count:   count,
			Output:  output,
			Quality: quality,
			DryRun:  dryRun,
		})
	},
}
