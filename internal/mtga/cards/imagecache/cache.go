package imagecache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ImageSize represents different sizes of card images.
type ImageSize string

const (
	ImageSizeSmall   ImageSize = "small"
	ImageSizeNormal  ImageSize = "normal"
	ImageSizeLarge   ImageSize = "large"
	ImageSizeArtCrop ImageSize = "art_crop"
)

// Cache manages local caching of card images.
type Cache struct {
	cacheDir   string
	maxSize    int64 // Maximum cache size in bytes
	mu         sync.RWMutex
	sizes      map[string]int64     // Map of file path to file size
	lastUsed   map[string]time.Time // LRU tracking
	httpClient *http.Client
}

// CacheOptions configures the image cache.
type CacheOptions struct {
	CacheDir string        // Directory to store cached images
	MaxSize  int64         // Maximum cache size in bytes (0 = unlimited)
	Timeout  time.Duration // HTTP request timeout
}

// DefaultCacheOptions returns sensible default cache options.
func DefaultCacheOptions() CacheOptions {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".mtga-companion", "image-cache")

	return CacheOptions{
		CacheDir: cacheDir,
		MaxSize:  500 * 1024 * 1024, // 500 MB default
		Timeout:  30 * time.Second,
	}
}

// NewCache creates a new image cache.
func NewCache(options CacheOptions) (*Cache, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(options.CacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &Cache{
		cacheDir: options.CacheDir,
		maxSize:  options.MaxSize,
		sizes:    make(map[string]int64),
		lastUsed: make(map[string]time.Time),
		httpClient: &http.Client{
			Timeout: options.Timeout,
		},
	}

	// Initialize cache metadata by scanning existing files
	if err := cache.scan(); err != nil {
		return nil, fmt.Errorf("failed to scan cache directory: %w", err)
	}

	return cache, nil
}

// GetImage retrieves an image from cache or downloads it if not cached.
// Returns the path to the cached image file.
func (c *Cache) GetImage(imageURL string, size ImageSize) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("image URL is empty")
	}

	// Generate cache key from URL
	cacheKey := c.generateCacheKey(imageURL, size)
	cachePath := filepath.Join(c.cacheDir, cacheKey)

	// Check if image is already cached
	c.mu.RLock()
	if _, exists := c.sizes[cachePath]; exists {
		// Update last used time
		c.mu.RUnlock()
		c.mu.Lock()
		c.lastUsed[cachePath] = time.Now()
		c.mu.Unlock()
		return cachePath, nil
	}
	c.mu.RUnlock()

	// Download and cache the image
	return c.downloadAndCache(imageURL, cachePath)
}

// downloadAndCache downloads an image and stores it in the cache.
func (c *Cache) downloadAndCache(imageURL, cachePath string) (string, error) {
	// Download image
	resp, err := c.httpClient.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp(c.cacheDir, "download-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Download to temp file
	size, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Ensure cache has space
	c.mu.Lock()
	if err := c.ensureSpace(size); err != nil {
		c.mu.Unlock()
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to ensure cache space: %w", err)
	}

	// Move temp file to final location
	if err := os.Rename(tempPath, cachePath); err != nil {
		c.mu.Unlock()
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to move cached file: %w", err)
	}

	// Update cache metadata
	c.sizes[cachePath] = size
	c.lastUsed[cachePath] = time.Now()
	c.mu.Unlock()

	return cachePath, nil
}

// ensureSpace evicts old files if necessary to make room for a new file.
// Must be called with c.mu locked.
func (c *Cache) ensureSpace(neededSize int64) error {
	if c.maxSize == 0 {
		return nil // Unlimited cache size
	}

	// Calculate current cache size
	var currentSize int64
	for _, size := range c.sizes {
		currentSize += size
	}

	// Check if we need to evict files
	if currentSize+neededSize <= c.maxSize {
		return nil
	}

	// Build list of files sorted by last used time
	type fileEntry struct {
		path     string
		lastUsed time.Time
		size     int64
	}

	files := make([]fileEntry, 0, len(c.sizes))
	for path, size := range c.sizes {
		files = append(files, fileEntry{
			path:     path,
			lastUsed: c.lastUsed[path],
			size:     size,
		})
	}

	// Sort by last used time (oldest first)
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].lastUsed.After(files[j].lastUsed) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Evict oldest files until we have enough space
	for _, file := range files {
		if currentSize+neededSize <= c.maxSize {
			break
		}

		// Remove file
		if err := os.Remove(file.path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to evict cached file: %w", err)
		}

		// Update metadata
		delete(c.sizes, file.path)
		delete(c.lastUsed, file.path)
		currentSize -= file.size
	}

	return nil
}

// Clear removes all cached images.
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all files
	for path := range c.sizes {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove cached file: %w", err)
		}
	}

	// Clear metadata
	c.sizes = make(map[string]int64)
	c.lastUsed = make(map[string]time.Time)

	return nil
}

// GetCacheStats returns statistics about the cache.
func (c *Cache) GetCacheStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalSize int64
	for _, size := range c.sizes {
		totalSize += size
	}

	return CacheStats{
		TotalFiles: len(c.sizes),
		TotalSize:  totalSize,
		MaxSize:    c.maxSize,
		CacheDir:   c.cacheDir,
	}
}

// CacheStats contains statistics about the cache.
type CacheStats struct {
	TotalFiles int
	TotalSize  int64
	MaxSize    int64
	CacheDir   string
}

// scan initializes cache metadata by scanning the cache directory.
func (c *Cache) scan() error {
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) == ".tmp" {
			continue
		}

		path := filepath.Join(c.cacheDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		c.sizes[path] = info.Size()
		c.lastUsed[path] = info.ModTime()
	}

	return nil
}

// generateCacheKey generates a unique cache key for an image URL.
func (c *Cache) generateCacheKey(imageURL string, size ImageSize) string {
	// Use SHA256 hash of URL + size as filename
	hash := sha256.Sum256([]byte(imageURL + string(size)))
	return hex.EncodeToString(hash[:]) + ".jpg"
}
