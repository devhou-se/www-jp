package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"regexp"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/nfnt/resize"
)

func main() {
	files, err := flatFiles("site/content")
	if err != nil {
		panic(err.Error())
	}

	rImage, err := regexp.Compile("!\\[.*]\\((.*)\\)")
	if err != nil {
		panic(err.Error())
	}

	rFilename, err := regexp.Compile("[^/]*$")
	if err != nil {
		panic(err.Error())
	}

	eg := &errgroup.Group{}

	for _, file := range files {
		fileBytes, err := os.ReadFile(file)
		if err != nil {
			panic(err.Error())
		}

		var images []string

		fImages := rImage.FindAllSubmatch(fileBytes, -1)
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

				err = cacheimage(image, "site/content/static/"+local)
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

const imgtemplate = `{{< lazyimage %s >}}`

func cacheimage(from, local string) error {
	response, err := http.Get(from)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	image, _, err := image.Decode(response.Body)
	if err != nil {
		return err
	}
	s := sizes(image)

	for i, size := range s {
		resized := resize.Resize(size[0], size[1], image, resize.Lanczos3)

		err = os.MkdirAll("site/content/static", os.ModePerm)
		if err != nil {
			return err
		}
		file, err := os.Create(local + []string{"_0", "_1", "_2", ""}[i] + ".jpeg")
		if err != nil {
			return err
		}

		err = jpeg.Encode(file, resized, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func sizes(img image.Image) [][2]uint {
	var s [][2]uint

	width := img.Bounds().Size().X
	height := img.Bounds().Size().Y

	aspectRatio := float64(width) / float64(height)

	widths := []int{240, 480, 960, width}

	for _, w := range widths {
		y := int(float64(w) / aspectRatio)
		s = append(s, [2]uint{uint(w), uint(y)})
	}

	return s
}

// return all files in directory and all subdirectories
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
