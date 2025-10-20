package main

import (
	"bufio"
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
	"sort"
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
	cacheFilePath = "imager-cache.txt"
)

var httpClient = &http.Client{
	Timeout: 120 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     90 * time.Second,
	},
}

// imageCache tracks which base filenames have been uploaded to GCS
type imageCache struct {
	mu    sync.Mutex
	cache map[string]bool
}

// loadCache reads the cache file and returns a populated imageCache
func loadCache() (*imageCache, error) {
	ic := &imageCache{
		cache: make(map[string]bool),
	}

	file, err := os.Open(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Cache file doesn't exist yet, return empty cache
			return ic, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			ic.cache[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ic, nil
}

// saveCache writes the cache to disk, sorted alphabetically
func (ic *imageCache) saveCache() error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	// Get all cache entries and sort them
	entries := make([]string, 0, len(ic.cache))
	for filename := range ic.cache {
		entries = append(entries, filename)
	}
	sort.Strings(entries)

	// Write to file
	file, err := os.Create(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, entry := range entries {
		_, err := writer.WriteString(entry + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

// isInCache checks if a base filename is in the cache
func (ic *imageCache) isInCache(baseFilename string) bool {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	return ic.cache[baseFilename]
}

// addToCache adds a base filename to the cache and saves to disk
func (ic *imageCache) addToCache(baseFilename string) error {
	ic.mu.Lock()
	ic.cache[baseFilename] = true
	ic.mu.Unlock()

	return ic.saveCache()
}

// rebuildCache queries GCS and rebuilds the cache from what's actually uploaded
func rebuildCache(ctx context.Context, bucket *storage.BucketHandle) (*imageCache, error) {
	ic := &imageCache{
		cache: make(map[string]bool),
	}

	fmt.Println("Rebuilding cache from GCS...")

	// List all objects in the images path
	query := &storage.Query{Prefix: gcsImagePath + "/"}
	it := bucket.Objects(ctx, query)

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// Extract filename from path (remove "images/" prefix)
		fullPath := attrs.Name
		if !strings.HasPrefix(fullPath, gcsImagePath+"/") {
			continue
		}

		filename := strings.TrimPrefix(fullPath, gcsImagePath+"/")

		// Extract base filename (remove _0, _1, _2, _3 suffixes)
		baseFilename := filename
		for _, suffix := range []string{"_0.jpeg", "_1.jpeg", "_2.jpeg", "_3.jpeg"} {
			if strings.HasSuffix(filename, suffix) {
				baseFilename = strings.TrimSuffix(filename, suffix) + ".jpeg"
				break
			}
		}

		ic.cache[baseFilename] = true
	}

	fmt.Printf("Found %d unique images in GCS\n", len(ic.cache))

	// Save the cache
	if err := ic.saveCache(); err != nil {
		return nil, err
	}

	fmt.Println("Cache rebuilt successfully")
	return ic, nil
}

func main() {
	// Parse command-line flags
	rebuildCacheFlag := flag.Bool("rebuild-cache", false, "Rebuild the cache from GCS")
	flag.Parse()

	ctx := context.Background()

	// Initialize GCS client
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to create GCS client: %v", err))
	}
	defer gcsClient.Close()

	bucket := gcsClient.Bucket(gcsBucketName)

	// Load or rebuild cache
	var cache *imageCache
	if *rebuildCacheFlag {
		cache, err = rebuildCache(ctx, bucket)
		if err != nil {
			panic(fmt.Sprintf("Failed to rebuild cache: %v", err))
		}
		fmt.Println("Cache rebuild complete")
		return
	} else {
		cache, err = loadCache()
		if err != nil {
			panic(fmt.Sprintf("Failed to load cache: %v", err))
		}
		fmt.Printf("Loaded cache with %d entries\n", len(cache.cache))
	}

	images, err := utils.WebImages()
	if err != nil {
		panic(err.Error())
	}

	imageWidths := []int{240, 480, 960, 0}

	// Filter out cached images before spawning goroutines
	uncachedImages := make([]utils.Image, 0, len(images))
	cachedCount := 0
	for _, image := range images {
		webLocationParts := strings.Split(image.WebLocation, "/")
		filenameBase := webLocationParts[len(webLocationParts)-1]

		if cache.isInCache(filenameBase) {
			cachedCount++
			continue
		}
		uncachedImages = append(uncachedImages, image)
	}

	fmt.Printf("Skipping %d cached images, processing %d new images\n", cachedCount, len(uncachedImages))

	fl := &fileLocker{fl: make(map[string]*sync.Mutex)}
	wg := sync.WaitGroup{}

	sem := semaphore.NewWeighted(20)

	for _, image := range uncachedImages {
		wg.Add(1)
		image := image // Capture loop variable

		go func() {
			defer wg.Done()

			err := sem.Acquire(context.Background(), 1)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer sem.Release(1)

			webLocationParts := strings.Split(image.WebLocation, "/")
			filenameBase := webLocationParts[len(webLocationParts)-1]

			_, _, err = resizeAndUpload(ctx, bucket, image.WebLocation, filenameBase, imageWidths, fl, cache)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			fmt.Printf("Uploaded %s\n", image.Location)
		}()
	}

	wg.Wait()
}

// fetchWithRetry attempts to fetch a URL with exponential backoff retry logic
func fetchWithRetry(url string, maxRetries int) (*http.Response, error) {
	var response *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err = httpClient.Get(url)
		if err == nil && response.StatusCode == http.StatusOK {
			return response, nil
		}

		if response != nil {
			response.Body.Close()
		}

		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			fmt.Printf("Retry %d/%d for %s after %v (error: %v)\n", attempt+1, maxRetries, url, backoff, err)
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

// resizeAndUpload downloads an image from a url and uploads resized versions to GCS
// for each of the widths defined. A width of 0 will keep the original width.
func resizeAndUpload(ctx context.Context, bucket *storage.BucketHandle, url, filenameBase string, widths []int, fl *fileLocker, cache *imageCache) (int, int, error) {
	// Fetch image with retry
	response, err := fetchWithRetry(url, 3)
	if err != nil {
		return 0, 0, err
	}
	defer response.Body.Close()

	exifReader := &bytes.Buffer{}
	imageReader := io.TeeReader(response.Body, exifReader)

	// Load image data
	img, format, err := image.Decode(imageReader)
	if err != nil {
		return 0, 0, err
	}

	// Load exif data (only for JPEG)
	var eb *exif.IfdBuilder
	if format == "jpeg" {
		mc, err := jis.NewJpegMediaParser().ParseBytes(exifReader.Bytes())
		if err != nil {
			return 0, 0, err
		}
		sl := mc.(*jis.SegmentList)

		eb, err = sl.ConstructExifBuilder()
		if err != nil {
			return 0, 0, err
		}
	}

	// Find original image dimensions
	x1 := img.Bounds().Size().X
	y1 := img.Bounds().Size().Y

	// Process all width variants in parallel
	widthWg := sync.WaitGroup{}
	errChan := make(chan error, len(widths))

	for i, width := range widths {
		widthWg.Add(1)
		go func(width, index int) {
			defer widthWg.Done()

			// Calculate new dimensions
			x2 := width
			if x2 == 0 {
				x2 = x1
			}
			y2 := newY(x1, y1, x2)

			// Resize image
			resized := resize.Resize(uint(x2), uint(y2), img, resize.Lanczos3)

			// Create GCS object path
			suffix := ""
			if width > 0 {
				suffix = fmt.Sprintf("_%d", index)
			}

			objectPath := fmt.Sprintf("%s/%s%s.jpeg", gcsImagePath, filenameBase, suffix)

			fl.Lock(objectPath)

			// Check if object already exists in GCS
			obj := bucket.Object(objectPath)
			_, existsErr := obj.Attrs(ctx)
			if existsErr == nil {
				// Object exists, skip upload
				fl.Unlock(objectPath)
				return
			}

			// Encode to buffer first
			buf := &bytes.Buffer{}
			err := jpeg.Encode(buf, resized, nil)
			if err != nil {
				fl.Unlock(objectPath)
				errChan <- err
				return
			}

			// Prepare final buffer with or without EXIF
			var finalBuf *bytes.Buffer
			if eb != nil {
				// Extract new exif data from buffer
				mc2, err := jis.NewJpegMediaParser().ParseBytes(buf.Bytes())
				if err != nil {
					fl.Unlock(objectPath)
					errChan <- err
					return
				}
				sl2 := mc2.(*jis.SegmentList)

				// Replace new exif data with previous
				err = sl2.SetExif(eb)
				if err != nil {
					fl.Unlock(objectPath)
					errChan <- err
					return
				}

				// Write with EXIF to buffer
				finalBuf = &bytes.Buffer{}
				err = sl2.Write(finalBuf)
				if err != nil {
					fl.Unlock(objectPath)
					errChan <- err
					return
				}
			} else {
				// No EXIF, use the original buffer
				finalBuf = buf
			}

			// Upload to GCS
			writer := obj.NewWriter(ctx)
			writer.ContentType = "image/jpeg"
			writer.CacheControl = "public, max-age=31536000, immutable"

			_, err = writer.Write(finalBuf.Bytes())
			if err != nil {
				writer.Close()
				fl.Unlock(objectPath)
				errChan <- err
				return
			}

			err = writer.Close()
			fl.Unlock(objectPath)

			if err != nil {
				errChan <- err
			}
		}(width, i)
	}

	widthWg.Wait()
	close(errChan)

	// Check for any errors
	if err := <-errChan; err != nil {
		return 0, 0, err
	}

	// All variants uploaded successfully, add to cache
	if err := cache.addToCache(filenameBase); err != nil {
		fmt.Printf("Warning: Failed to update cache for %s: %v\n", filenameBase, err)
	}

	return x1, y1, nil
}

// newY calculates the new height for an image, maintaining aspect ratio
func newY(oldX, oldY, newX int) int {
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
