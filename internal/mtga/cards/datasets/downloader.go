package datasets

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// S3 bucket URL for 17Lands public datasets
	PublicDatasetsBaseURL = "https://17lands-public.s3.amazonaws.com/analysis_data/game_data"

	// Default timeout for HTTP requests
	DownloadTimeout = 5 * time.Minute

	// Default cache directory (relative to user home)
	DefaultCacheDir = ".mtga-companion/datasets"
)

// Downloader handles downloading 17Lands datasets from S3.
type Downloader struct {
	httpClient *http.Client
	cacheDir   string
}

// DownloaderOptions configures the dataset downloader.
type DownloaderOptions struct {
	// Timeout for HTTP requests (default: 5 minutes)
	Timeout time.Duration

	// CacheDir is the local directory to store downloaded datasets
	// Default: ~/.mtga-companion/datasets
	CacheDir string

	// HTTPClient allows custom HTTP client
	HTTPClient *http.Client
}

// DefaultDownloaderOptions returns default downloader options.
func DefaultDownloaderOptions() DownloaderOptions {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return DownloaderOptions{
		Timeout:  DownloadTimeout,
		CacheDir: filepath.Join(homeDir, DefaultCacheDir),
	}
}

// NewDownloader creates a new dataset downloader.
func NewDownloader(options DownloaderOptions) (*Downloader, error) {
	if options.Timeout == 0 {
		options.Timeout = DownloadTimeout
	}
	if options.CacheDir == "" {
		opts := DefaultDownloaderOptions()
		options.CacheDir = opts.CacheDir
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(options.CacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: options.Timeout,
		}
	}

	return &Downloader{
		httpClient: httpClient,
		cacheDir:   options.CacheDir,
	}, nil
}

// DownloadDataset downloads a 17Lands dataset from S3 and returns the path to the decompressed CSV file.
// If the dataset is already cached and fresh, returns the cached file path.
func (d *Downloader) DownloadDataset(ctx context.Context, setCode, format string) (string, error) {
	// Build S3 URL
	// Format: game_data_public.{SET}.{FORMAT}.csv.gz
	filename := fmt.Sprintf("game_data_public.%s.%s.csv.gz", setCode, format)
	url := fmt.Sprintf("%s/%s", PublicDatasetsBaseURL, filename)

	// Check if file is already cached
	csvPath := filepath.Join(d.cacheDir, fmt.Sprintf("%s_%s.csv", setCode, format))
	gzPath := filepath.Join(d.cacheDir, filename)

	// Check if CSV file exists and is recent (less than 24 hours old)
	if info, err := os.Stat(csvPath); err == nil {
		age := time.Since(info.ModTime())
		if age < 24*time.Hour {
			log.Printf("[Downloader] Using cached dataset: %s (age: %v)", csvPath, age)
			return csvPath, nil
		}
		log.Printf("[Downloader] Cached dataset is stale (age: %v), re-downloading", age)
	}

	log.Printf("[Downloader] Downloading dataset from: %s", url)

	// Download gzipped file
	if err := d.downloadFile(ctx, url, gzPath); err != nil {
		return "", fmt.Errorf("failed to download dataset: %w", err)
	}

	// Decompress to CSV
	if err := d.decompressGzip(gzPath, csvPath); err != nil {
		return "", fmt.Errorf("failed to decompress dataset: %w", err)
	}

	// Clean up gzipped file (optional - keep it for verification)
	// os.Remove(gzPath)

	log.Printf("[Downloader] Dataset ready: %s", csvPath)
	return csvPath, nil
}

// downloadFile downloads a file from a URL to a local path.
func (d *Downloader) downloadFile(ctx context.Context, url, destPath string) error {
	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "MTGA-Companion/1.0")

	// Execute request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = out.Close() }()

	// Copy response body to file
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("[Downloader] Downloaded %d bytes to %s", written, destPath)
	return nil
}

// decompressGzip decompresses a gzipped file to a destination path.
func (d *Downloader) decompressGzip(gzPath, csvPath string) error {
	// Open gzipped file
	gzFile, err := os.Open(gzPath)
	if err != nil {
		return fmt.Errorf("failed to open gzipped file: %w", err)
	}
	defer func() { _ = gzFile.Close() }()

	// Create gzip reader
	gr, err := gzip.NewReader(gzFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gr.Close() }()

	// Create CSV file
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer func() { _ = csvFile.Close() }()

	// Decompress
	written, err := io.Copy(csvFile, gr)
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}

	log.Printf("[Downloader] Decompressed %d bytes to %s", written, csvPath)
	return nil
}

// GetCachedDatasetPath returns the path to a cached dataset if it exists and is fresh.
// Returns empty string if not cached or stale.
func (d *Downloader) GetCachedDatasetPath(setCode, format string) (string, error) {
	csvPath := filepath.Join(d.cacheDir, fmt.Sprintf("%s_%s.csv", setCode, format))

	// Check if file exists and is recent (less than 24 hours old)
	if info, err := os.Stat(csvPath); err == nil {
		age := time.Since(info.ModTime())
		if age < 24*time.Hour {
			return csvPath, nil
		}
	}

	return "", nil
}

// ClearCache removes all cached datasets.
func (d *Downloader) ClearCache() error {
	return os.RemoveAll(d.cacheDir)
}
