package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
	"golang.org/x/sync/errgroup"

	"github.com/nfnt/resize"

	"github.com/devhou-se/www-jp/go/utils"
)

const (
	imageStorePath = utils.ContentDirectory + "/images"
)

var (
	// Find filename in url
	rFilename = regexp.MustCompile("[^/]*$")
)

const imgtemplate = `{{< lazyimage %s 425 >}}`

func main() {
	// Find all files in content directory
	files, err := utils.Markdowns()
	if err != nil {
		panic(err.Error())
	}

	eg := &errgroup.Group{}
	for _, file := range files {
		// Only check markdown files
		if !strings.Contains(file, ".md") {
			continue
		}

		fileBytes, err := os.ReadFile(file)
		if err != nil {
			panic(err.Error())
		}

		var images []string

		// Find all markdown images on page
		fImages := utils.MarkdownImageR.FindAllSubmatch(fileBytes, -1)
		for _, fImage := range fImages {
			images = append(images, string(fImage[1]))
		}

		fLock := &sync.Mutex{}
		fWait := &sync.WaitGroup{}

		for i, image := range images {
			fWait.Add(1)
			local := rFilename.FindString(image)

			eg.Go(func() error {
				defer fWait.Done()

				// Resize and store locally each image
				_, _, err = resizeAndStore(image, imageStorePath+"/"+local, []int{240, 480, 960, 0})
				if err != nil {
					fmt.Println(err.Error())
					return nil
				}

				// y := newY(width, height, 425)
				// todo: aspect ratio is inverted for exif rotated images

				fLock.Lock()
				// Replace image tags with lazyimage shortcode
				fileBytes = bytes.ReplaceAll(fileBytes, fImages[i][0], []byte(fmt.Sprintf(imgtemplate, local)))
				fLock.Unlock()

				return nil
			})
		}

		eg.Go(func() error {
			// Write final markdown file back to disk after all images have been processed
			fWait.Wait()
			return os.WriteFile(file, fileBytes, os.ModePerm)
		})
	}

	err = eg.Wait()
	if err != nil {
		panic(err.Error())
	}
}

// resizeAndStore downloads an image from a url and stores a resized version for
// each of the widths defined. A width of 0 will keep the original width.
func resizeAndStore(url, filenameBase string, widths []int) (int, int, error) {
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
		file, err := os.Create(filenameBase + suffix + ".jpeg")
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
