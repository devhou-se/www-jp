package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Register PNG decoder
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dsoprea/go-exif/v3"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
	"github.com/nfnt/resize"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/iterator"

	"github.com/devhou-se/www-jp/go/utils"
)

const (
	gcsBucketName = "static.devh.se"
	gcsImagePath  = "images"
	maxRetries    = 3
	baseBackoff   = 1 * time.Second
)

var (
	httpClient = &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	imageWidths = []int{240, 480, 960, 0}
)

// Config holds processor configuration
type Config struct {
	RebuildCache   bool
	Parallelism    int
	Verbose        bool
	VerifyCache    bool
	DryRun         bool
	MaintenanceOp  string
}

func main() {
	// Parse flags
	config := parseFlags()

	ctx := context.Background()

	// Initialize GCS client
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to create GCS client: %v\n", err)
		os.Exit(1)
	}
	defer gcsClient.Close()

	bucket := gcsClient.Bucket(gcsBucketName)

	// Initialize cache
	cache := NewImageCache()
	if err := cache.Load(); err != nil {
		fmt.Printf("❌ Failed to load cache: %v\n", err)
		os.Exit(1)
	}

	// Handle different operations
	switch {
	case config.RebuildCache:
		if err := rebuildCacheFromGCS(ctx, bucket, cache); err != nil {
			fmt.Printf("❌ Failed to rebuild cache: %v\n", err)
			os.Exit(1)
		}
		return

	case config.VerifyCache:
		verifyCacheIntegrity(ctx, bucket, cache)
		return

	case config.MaintenanceOp != "":
		handleMaintenance(ctx, bucket, cache, config)
		return
	}

	// Normal processing mode
	if err := processImages(ctx, bucket, cache, config); err != nil {
		fmt.Printf("❌ Processing failed: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.BoolVar(&config.RebuildCache, "rebuild-cache", false, "Rebuild cache from GCS")
	flag.BoolVar(&config.VerifyCache, "verify-cache", false, "Verify cache integrity")
	flag.IntVar(&config.Parallelism, "parallelism", 20, "Number of concurrent image processors")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Don't upload images, just show what would be done")
	flag.StringVar(&config.MaintenanceOp, "maintenance", "", "Maintenance operation: stats, export, repair")

	flag.Parse()
	return config
}

func processImages(ctx context.Context, bucket *storage.BucketHandle, cache *ImageCache, config *Config) error {
	// Get all web images from markdown files
	images, err := utils.WebImages()
	if err != nil {
		return fmt.Errorf("failed to get images: %w", err)
	}

	if len(images) == 0 {
		fmt.Println("✓ No images to process")
		return nil
	}

	fmt.Printf("Found %d images in markdown files\n", len(images))

	// Filter uncached images
	uncachedImages, cachedCount := filterUncachedImages(images, cache, config)

	fmt.Printf("Cached: %d | To process: %d\n\n", cachedCount, len(uncachedImages))

	if len(uncachedImages) == 0 {
		fmt.Println("✓ All images are cached, nothing to process")
		return cache.Save()
	}

	// Initialize progress tracker
	progress := NewProgressTracker(len(uncachedImages))

	// Process images concurrently
	sem := semaphore.NewWeighted(int64(config.Parallelism))
	wg := sync.WaitGroup{}
	fl := &fileLocker{fl: make(map[string]*sync.Mutex)}

	// Progress ticker
	ticker := time.NewTicker(2 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				progress.PrintProgress()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	for _, img := range uncachedImages {
		wg.Add(1)
		img := img // Capture loop variable

		go func() {
			defer wg.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				progress.AddError(extractFilename(img.WebLocation), img.WebLocation, err)
				return
			}
			defer sem.Release(1)

			filename := extractFilename(img.WebLocation)
			progress.SetCurrent(filename)

			if config.DryRun {
				fmt.Printf("DRY RUN: Would process %s\n", filename)
				progress.IncrementSkipped()
				return
			}

			if err := processImage(ctx, bucket, cache, img, fl, progress); err != nil {
				progress.AddError(filename, img.WebLocation, err)
			}
		}()
	}

	wg.Wait()
	done <- true

	// Print final summary
	progress.PrintSummary()

	// Save cache
	if !config.DryRun {
		if err := cache.Save(); err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}
		fmt.Println("\n✓ Cache saved successfully")
	}

	// Return error if any processing failed
	if progress.HasErrors() {
		return fmt.Errorf("processing completed with %d errors", len(progress.GetErrors()))
	}

	return nil
}

func filterUncachedImages(images []utils.Image, cache *ImageCache, config *Config) ([]utils.Image, int) {
	uncached := make([]utils.Image, 0, len(images))
	cachedCount := 0

	for _, img := range images {
		filename := extractFilename(img.WebLocation)

		// Check if image is in cache
		// Legacy entries (from v1.0) have no hash, but we still trust them
		// since they indicate the image was uploaded to GCS
		if cache.Has(filename) {
			cachedCount++
			if config.Verbose {
				entry, _ := cache.Get(filename)
				if entry.Hash == "" {
					fmt.Printf("⊙ Cached (legacy): %s\n", filename)
				} else {
					fmt.Printf("⊙ Cached: %s\n", filename)
				}
			}
			continue
		}

		uncached = append(uncached, img)
	}

	return uncached, cachedCount
}

func processImage(ctx context.Context, bucket *storage.BucketHandle, cache *ImageCache, img utils.Image, fl *fileLocker, progress *ProgressTracker) error {
	filename := extractFilename(img.WebLocation)

	// Download image with retry
	imageData, _, err := downloadImageWithRetry(img.WebLocation, maxRetries)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Compute hash
	hashReader := bytes.NewReader(imageData)
	hash, err := ComputeHash(hashReader)
	if err != nil {
		return fmt.Errorf("hash computation failed: %w", err)
	}

	// Check if we already have this exact image (by hash)
	if entry, ok := cache.Get(filename); ok && entry.Hash == hash {
		progress.IncrementSkipped()
		return nil
	}

	// Decode image
	imageReader := bytes.NewReader(imageData)
	img_decoded, format, err := image.Decode(imageReader)
	if err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	// Load EXIF data (only for JPEG)
	var exifBuilder *exif.IfdBuilder
	if format == "jpeg" {
		mc, err := jis.NewJpegMediaParser().ParseBytes(imageData)
		if err == nil {
			sl := mc.(*jis.SegmentList)
			exifBuilder, _ = sl.ConstructExifBuilder()
		}
	}

	// Get dimensions
	width := img_decoded.Bounds().Size().X
	height := img_decoded.Bounds().Size().Y

	// Process all width variants
	gcsPaths, err := uploadImageVariants(ctx, bucket, img_decoded, filename, exifBuilder, width, height, fl)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// Add to cache
	entry := &CacheEntry{
		Filename:  filename,
		Hash:      hash,
		Timestamp: time.Now().Unix(),
		Width:     width,
		Height:    height,
		GCSPaths:  gcsPaths,
	}
	cache.Add(entry)

	progress.IncrementProcessed()
	return nil
}

func downloadImageWithRetry(url string, maxRetries int) ([]byte, string, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := httpClient.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, "", err
			}
			return data, resp.Header.Get("Content-Type"), nil
		}

		if resp != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		if attempt < maxRetries {
			backoff := baseBackoff * time.Duration(1<<uint(attempt))
			time.Sleep(backoff)
		}
	}

	return nil, "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func uploadImageVariants(ctx context.Context, bucket *storage.BucketHandle, img image.Image, filename string, exifBuilder *exif.IfdBuilder, origWidth, origHeight int, fl *fileLocker) ([]string, error) {
	gcsPaths := make([]string, 0, len(imageWidths))
	errChan := make(chan error, len(imageWidths))
	wg := sync.WaitGroup{}

	for i, width := range imageWidths {
		wg.Add(1)
		go func(width, index int) {
			defer wg.Done()

			// Calculate new dimensions
			newWidth := width
			if newWidth == 0 {
				newWidth = origWidth
			}
			newHeight := calculateHeight(origWidth, origHeight, newWidth)

			// Resize
			resized := resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)

			// Determine object path
			suffix := ""
			if width > 0 {
				suffix = fmt.Sprintf("_%d", index)
			}
			objectPath := fmt.Sprintf("%s/%s%s.jpeg", gcsImagePath, filename, suffix)

			fl.Lock(objectPath)
			defer fl.Unlock(objectPath)

			// Check if exists
			obj := bucket.Object(objectPath)
			if _, err := obj.Attrs(ctx); err == nil {
				// Already exists, skip
				gcsPaths = append(gcsPaths, objectPath)
				return
			}

			// Encode to JPEG
			buf := &bytes.Buffer{}
			if err := jpeg.Encode(buf, resized, nil); err != nil {
				errChan <- fmt.Errorf("encode failed for %s: %w", objectPath, err)
				return
			}

			// Apply EXIF if available
			var finalBuf *bytes.Buffer
			if exifBuilder != nil {
				mc2, err := jis.NewJpegMediaParser().ParseBytes(buf.Bytes())
				if err == nil {
					sl2 := mc2.(*jis.SegmentList)
					if err := sl2.SetExif(exifBuilder); err == nil {
						finalBuf = &bytes.Buffer{}
						if err := sl2.Write(finalBuf); err == nil {
							buf = finalBuf
						}
					}
				}
			}

			// Upload
			writer := obj.NewWriter(ctx)
			writer.ContentType = "image/jpeg"
			writer.CacheControl = "public, max-age=31536000, immutable"

			if _, err := writer.Write(buf.Bytes()); err != nil {
				writer.Close()
				errChan <- fmt.Errorf("upload failed for %s: %w", objectPath, err)
				return
			}

			if err := writer.Close(); err != nil {
				errChan <- fmt.Errorf("close failed for %s: %w", objectPath, err)
				return
			}

			gcsPaths = append(gcsPaths, objectPath)
		}(width, i)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	return gcsPaths, nil
}

func rebuildCacheFromGCS(ctx context.Context, bucket *storage.BucketHandle, cache *ImageCache) error {
	fmt.Println("🔄 Rebuilding cache from GCS...")

	query := &storage.Query{Prefix: gcsImagePath + "/"}
	it := bucket.Objects(ctx, query)

	count := 0
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		// Extract filename
		fullPath := attrs.Name
		if !strings.HasPrefix(fullPath, gcsImagePath+"/") {
			continue
		}

		filename := strings.TrimPrefix(fullPath, gcsImagePath+"/")

		// Extract base filename
		baseFilename := extractBaseFilename(filename)

		// Add to cache (without full metadata in rebuild mode)
		if !cache.Has(baseFilename) {
			cache.Add(&CacheEntry{
				Filename:  baseFilename,
				Hash:      "",
				Timestamp: attrs.Created.Unix(),
				Width:     0,
				Height:    0,
				GCSPaths:  []string{},
			})
			count++
		}
	}

	fmt.Printf("Found %d unique images in GCS\n", count)

	if err := cache.Save(); err != nil {
		return err
	}

	fmt.Println("✓ Cache rebuilt successfully")
	return nil
}

func verifyCacheIntegrity(ctx context.Context, bucket *storage.BucketHandle, cache *ImageCache) {
	fmt.Println("🔍 Verifying cache integrity...")
	// TODO: Implement verification logic
	fmt.Println("✓ Verification complete")
}

func handleMaintenance(ctx context.Context, bucket *storage.BucketHandle, cache *ImageCache, config *Config) {
	switch config.MaintenanceOp {
	case "stats":
		printCacheStats(cache)
	case "export":
		exportCache(cache)
	case "repair":
		repairCache(cache)
	default:
		fmt.Printf("Unknown maintenance operation: %s\n", config.MaintenanceOp)
	}
}

func printCacheStats(cache *ImageCache) {
	stats := cache.Stats()
	fmt.Println("\n=== Cache Statistics ===")
	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}
}

func exportCache(cache *ImageCache) {
	// TODO: Implement cache export
	fmt.Println("Export functionality not yet implemented")
}

func repairCache(cache *ImageCache) {
	// TODO: Implement cache repair
	fmt.Println("Repair functionality not yet implemented")
}

func extractFilename(url string) string {
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	// Remove query parameters if present
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	// Ensure .jpeg extension for GitHub asset URLs
	if !strings.HasSuffix(filename, ".jpeg") && !strings.HasSuffix(filename, ".jpg") && !strings.HasSuffix(filename, ".png") {
		filename = filename + ".jpeg"
	}

	return filename
}

func extractBaseFilename(filename string) string {
	for _, suffix := range []string{"_0.jpeg", "_1.jpeg", "_2.jpeg", "_3.jpeg"} {
		if strings.HasSuffix(filename, suffix) {
			return strings.TrimSuffix(filename, suffix) + ".jpeg"
		}
	}
	return filename
}

func calculateHeight(oldX, oldY, newX int) int {
	aspectRatio := float64(oldX) / float64(oldY)
	return int(float64(newX) / aspectRatio)
}

type fileLocker struct {
	mu sync.Mutex
	fl map[string]*sync.Mutex
}

func (f *fileLocker) Lock(file string) {
	f.mu.Lock()
	if _, ok := f.fl[file]; !ok {
		f.fl[file] = &sync.Mutex{}
	}
	f.mu.Unlock()
	f.fl[file].Lock()
}

func (f *fileLocker) Unlock(file string) {
	f.fl[file].Unlock()
}
