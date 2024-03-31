package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
	"github.com/nfnt/resize"

	"github.com/devhou-se/www-jp/go/utils"
)

const (
	imageStorePath = utils.SiteDirectory + "/content/images"
)

const imgtemplate = `{{< lazyimage %s 425 >}}`

func main() {
	images, err := utils.WebImages()
	if err != nil {
		panic(err.Error())
	}

	imageWidths := []int{240, 480, 960, 0}

	fl := &fileLocker{fl: make(map[string]*sync.Mutex)}
	wg := sync.WaitGroup{}

	for _, image := range images {
		wg.Add(1)

		go func() {
			defer wg.Done()

			webLocationParts := strings.Split(image.WebLocation, "/")
			filenameBase := webLocationParts[len(webLocationParts)-1]

			_, _, err = resizeAndStore(image.WebLocation, filenameBase, imageWidths, fl)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			// y := newY(width, height, 425)
			// todo: aspect ratio is inverted for exif rotated images

			fl.Lock(image.InFile)
			defer fl.Unlock(image.InFile)

			fileBytes, err := os.ReadFile(image.InFile)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			oldImage := image.FullMarkdown
			newImage := fmt.Sprintf(imgtemplate, filenameBase)

			fileBytes = bytes.ReplaceAll(fileBytes, []byte(oldImage), []byte(newImage))

			err = os.WriteFile(image.InFile, fileBytes, os.ModePerm)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			fmt.Printf("Updated %s in %s\n", image.Location, image.InFile)
		}()
	}

	wg.Wait()
}

// resizeAndStore downloads an image from a url and stores a resized version for
// each of the widths defined. A width of 0 will keep the original width.
func resizeAndStore(url, filenameBase string, widths []int, fl *fileLocker) (int, int, error) {
	// Fetch image
	response, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer response.Body.Close()

	exifReader := &bytes.Buffer{}
	imageReader := io.TeeReader(response.Body, exifReader)

	// Load image data
	img, _, err := image.Decode(imageReader)
	if err != nil {
		return 0, 0, err
	}

	// Load exif data
	mc, err := jis.NewJpegMediaParser().ParseBytes(exifReader.Bytes())
	if err != nil {
		return 0, 0, err
	}
	sl := mc.(*jis.SegmentList)

	eb, err := sl.ConstructExifBuilder()
	if err != nil {
		return 0, 0, err
	}

	// Ensure images directory exists
	err = os.MkdirAll(imageStorePath, os.ModePerm)
	if err != nil {
		return 0, 0, err
	}

	// Find original image dimensions
	x1 := img.Bounds().Size().X
	y1 := img.Bounds().Size().Y

	for i, width := range widths {
		// Calculate new dimensions
		x2 := width
		if x2 == 0 {
			x2 = x1
		}
		y2 := newY(x1, y1, x2)

		// Resize image
		resized := resize.Resize(uint(x2), uint(y2), img, resize.Lanczos3)

		// Create file for resized image
		suffix := ""
		if width > 0 {
			suffix = fmt.Sprintf("_%d", i)
		}

		filename := imageStorePath + "/" + filenameBase + suffix + ".jpeg"

		fl.Lock(filename)
		defer fl.Unlock(filename)
		if _, err = os.Stat(filename); err == nil {
			fmt.Printf("Already completed %s\n", filename)
			continue
		}

		file, err := os.Create(filename)
		if err != nil {
			return 0, 0, err
		}

		// Write image data to file
		err = jpeg.Encode(file, resized, nil)
		if err != nil {
			return 0, 0, err
		}

		// todo: this could be redundant? can we remove reading back from disk
		// Read new files bytes
		bytes, err := os.ReadFile(file.Name())
		if err != nil {
			return 0, 0, err
		}

		// Extract new exif data
		mc2, err := jis.NewJpegMediaParser().ParseBytes(bytes)
		if err != nil {
			return 0, 0, err
		}
		sl2 := mc2.(*jis.SegmentList)

		// Replace new exif data with previous
		err = sl2.SetExif(eb)
		if err != nil {
			return 0, 0, err
		}

		// Reset write pointer
		_, err = file.Seek(0, 0)
		if err != nil {
			return 0, 0, err
		}

		// Write new exif data to file
		err = sl2.Write(file)
		if err != nil {
			return 0, 0, err
		}
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
