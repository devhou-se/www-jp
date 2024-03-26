package utils

import (
	"os"
	"regexp"
	"strings"
)

const (
	ContentDirectory = "site/content"
)

var (
	MarkdownImageR = regexp.MustCompile("!\\[([^]]*)]\\(((https?://[^)]*)|([^)]*))\\)") // 1: alt 2: location 3: web location 4: local location
)

// Image represents the data that can be found in a markdown image tag
type Image struct {
	Location      string
	WebLocation   string
	LocalLocation string
	Alt           string

	FullMarkdown string
	InFile       string
}

// Images finds all images an all content markdown files
func Images() ([]Image, error) {
	markdowns, err := Markdowns()
	if err != nil {
		return nil, err
	}
	var images []Image
	for _, markdown := range markdowns {
		fileBytes, err := os.ReadFile(markdown)
		if err != nil {
			return nil, err
		}
		for _, findings := range MarkdownImageR.FindAllStringSubmatch(string(fileBytes), -1) {
			images = append(images, Image{
				Location:      findings[2],
				WebLocation:   findings[3],
				LocalLocation: findings[4],
				Alt:           findings[1],
				FullMarkdown:  findings[0],
				InFile:        markdown,
			})
		}
	}
	return images, nil
}

// WebImages finds all images on all content markdown files that are from the web
func WebImages() ([]Image, error) {
	images, err := Images()
	if err != nil {
		return nil, err
	}
	return Filter(images, func(image Image) bool {
		return image.WebLocation != ""
	}), nil
}

// LocalImages finds all images on all content markdown files that are local
func LocalImages() ([]Image, error) {
	images, err := Images()
	if err != nil {
		return nil, err
	}
	return Filter(images, func(image Image) bool {
		return image.LocalLocation != ""
	}), nil
}

// Markdowns returns all markdown files in the content directory
func Markdowns() ([]string, error) {
	files, err := FlatFiles(ContentDirectory)
	return FilterFiletype(files, "md"), err
}

// FilterFiletype filters a list of files based on their extension
func FilterFiletype(s []string, suffix string) []string {
	return Filter(s, func(s string) bool {
		return strings.HasSuffix(s, "."+suffix) || strings.HasSuffix(s, "."+strings.ToUpper(suffix))
	})
}

// Filter returns back a list of items that match a predicate function
func Filter[S ~[]E, E any](s S, predicate func(E) bool) S {
	var s2 S
	for _, v := range s {
		if predicate(v) {
			s2 = append(s2, v)
		}
	}
	return s2
}

// Map maps a slice of items using a mapping function
func Map[S1 ~[]E1, E1 any, E2 any](in S1, f func(E1) E2) []E2 {
	out := make([]E2, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

// FlatFiles returns a list of all files inside directory and all subdirectories
func FlatFiles(name string) ([]string, error) {
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
			nextRead, err := FlatFiles(name + "/" + readEntry.Name())
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
