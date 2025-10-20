package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Register PNG decoder
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dsoprea/go-exif/v3"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
	"github.com/nfnt/resize"
	"golang.org/x/sync/semaphore"

	"github.com/devhou-se/www-jp/go/utils"
)

const (
	gcsBucketName = "static.devh.se"
	gcsImagePath  = "images"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     90 * time.Second,
	},
}

func main() {
	ctx := context.Background()

	// Initialize GCS client
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to create GCS client: %v", err))
	}
	defer gcsClient.Close()

	bucket := gcsClient.Bucket(gcsBucketName)

	images, err := utils.WebImages()
	if err != nil {
		panic(err.Error())
	}

	imageWidths := []int{240, 480, 960, 0}

	fl := &fileLocker{fl: make(map[string]*sync.Mutex)}
	wg := sync.WaitGroup{}

	sem := semaphore.NewWeighted(150)

	for _, image := range images {
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

			_, _, err = resizeAndUpload(ctx, bucket, image.WebLocation, filenameBase, imageWidths, fl)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			fmt.Printf("Uploaded %s\n", image.Location)
		}()
	}

	wg.Wait()
}

// resizeAndUpload downloads an image from a url and uploads resized versions to GCS
// for each of the widths defined. A width of 0 will keep the original width.
func resizeAndUpload(ctx context.Context, bucket *storage.BucketHandle, url, filenameBase string, widths []int, fl *fileLocker) (int, int, error) {
	// Fetch image
	response, err := httpClient.Get(url)
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

			// Check if object already exists
			obj := bucket.Object(objectPath)
			if _, err := obj.Attrs(ctx); err == nil {
				fmt.Printf("Already completed %s\n", objectPath)
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
