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
)

const (
	markdownRoot   = "site/content"
	imageStorePath = "site/content/images"
)

var (
	// Find image in markdown file
	rMarkdownImage = regexp.MustCompile("!\\[[^]]*]\\(([^)]*)\\)")

	// Find filename in url
	rFilename = regexp.MustCompile("[^/]*$")
)

const imgtemplate = `{{< lazyimage %s >}}`

func main() {
	files, err := flatFiles(markdownRoot)
	if err != nil {
		panic(err.Error())
	}

	eg := &errgroup.Group{}
	for _, file := range files {
		if !strings.Contains(file, ".md") {
			continue
		}

		fileBytes, err := os.ReadFile(file)
		if err != nil {
			panic(err.Error())
		}

		var images []string

		fImages := rMarkdownImage.FindAllSubmatch(fileBytes, -1)
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

				err = resizeAndStore(image, imageStorePath+"/"+local, []int{240, 480, 960, 0})
				if err != nil {
					fmt.Println(err.Error())
					return nil
				}

				fLock.Lock()
				fileBytes = bytes.ReplaceAll(fileBytes, fImages[i][0], []byte(fmt.Sprintf(imgtemplate, local)))
				fLock.Unlock()

				return nil
			})
		}

		eg.Go(func() error {
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
func resizeAndStore(url, filenameBase string, widths []int) error {
	// Fetch image
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	exifReader := &bytes.Buffer{}
	imageReader := io.TeeReader(response.Body, exifReader)

	// Load image data
	img, _, err := image.Decode(imageReader)
	if err != nil {
		return err
	}

	// Load exif data
	mc, err := jis.NewJpegMediaParser().ParseBytes(exifReader.Bytes())
	if err != nil {
		return err
	}
	sl := mc.(*jis.SegmentList)

	// Ensure images directory exists
	err = os.MkdirAll(imageStorePath, os.ModePerm)
	if err != nil {
		return err
	}

	for i, width := range widths {
		x1 := img.Bounds().Size().X
		y1 := img.Bounds().Size().Y

		if width == 0 {
			x1 = width
		}

		x2 := width
		y2 := newY(x1, y1, x2)

		// Resize image
		resized := resize.Resize(uint(x2), uint(y2), img, resize.Lanczos3)

		// Create file for size
		suffix := ""
		if width > 0 {
			suffix = fmt.Sprintf("_%d", i)
		}
		file, err := os.Create(filenameBase + suffix)
		if err != nil {
			return err
		}

		// Write image data to file
		err = jpeg.Encode(file, resized, nil)
		if err != nil {
			return err
		}

		// Reset write pointer
		_, err = file.Seek(0, 0)
		if err != nil {
			return err
		}

		// Write exif data to file
		err = sl.Write(file)
		if err != nil {
			return err
		}
	}

	return nil
}

// newY calculates the new height for an image, maintaining aspect ratio
func newY(oldX, oldY, newX int) int {
	aspectRatio := float64(oldX) / float64(oldY)
	return int(float64(newX) / aspectRatio)
}

// flatFiles returns a list of all files inside directory and all subdirectories
func flatFiles(name string) ([]string, error) {
	var entries []string
	readEntries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}
	for _, readEntry := range readEntries {
		fileinfo, err := readEntry.Info()
		if err != nil {
			return nil, err
		}
		if fileinfo.IsDir() {
			nextRead, err := flatFiles(name + "/" + readEntry.Name())
			if err != nil {
				return nil, err
			}
			entries = append(entries, nextRead...)
			continue
		}
		entries = append(entries, name+"/"+readEntry.Name())
	}
	return entries, nil
}
