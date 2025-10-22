package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	CacheVersion    = "2.0"
	CacheFilePath   = "imager-cache.txt"
	CacheBackupPath = "imager-cache.txt.bak"
)

// CacheEntry represents a single cached image with metadata
type CacheEntry struct {
	Filename  string
	Hash      string
	Timestamp int64
	Width     int
	Height    int
	GCSPaths  []string
}

// ImageCache manages the text-based cache with enhanced metadata
type ImageCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	version string
	dirty   bool
}

// NewImageCache creates a new cache instance
func NewImageCache() *ImageCache {
	return &ImageCache{
		entries: make(map[string]*CacheEntry),
		version: CacheVersion,
		dirty:   false,
	}
}

// Load reads the cache from disk
func (c *ImageCache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(CacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No cache file yet, start fresh
			return nil
		}
		return fmt.Errorf("failed to open cache: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			// Check version comment
			if strings.HasPrefix(line, "# Version:") {
				c.version = strings.TrimSpace(strings.TrimPrefix(line, "# Version:"))
			}
			continue
		}

		// Parse cache entry
		entry, err := c.parseCacheEntry(line)
		if err != nil {
			// If parsing fails, might be old format - try legacy parse
			entry = c.parseLegacyEntry(line)
			if entry == nil {
				fmt.Printf("Warning: skipping invalid cache line %d: %s\n", lineNum, err)
				continue
			}
			c.dirty = true // Mark dirty to trigger save in new format
		}

		c.entries[entry.Filename] = entry
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading cache: %w", err)
	}

	fmt.Printf("Loaded %d entries from cache (version %s)\n", len(c.entries), c.version)

	// Auto-upgrade old format
	if c.dirty {
		fmt.Println("Cache format upgraded, will save in new format")
	}

	return nil
}

// parseCacheEntry parses a v2.0 format cache line
// Format: filename|hash|timestamp|width|height|gcs_0|gcs_1|gcs_2|gcs_3
func (c *ImageCache) parseCacheEntry(line string) (*CacheEntry, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid format: expected at least 5 fields, got %d", len(parts))
	}

	timestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	width, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid width: %w", err)
	}

	height, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid height: %w", err)
	}

	gcsPaths := []string{}
	if len(parts) > 5 {
		gcsPaths = parts[5:]
	}

	return &CacheEntry{
		Filename:  parts[0],
		Hash:      parts[1],
		Timestamp: timestamp,
		Width:     width,
		Height:    height,
		GCSPaths:  gcsPaths,
	}, nil
}

// parseLegacyEntry parses a v1.0 format cache line (just filename)
func (c *ImageCache) parseLegacyEntry(line string) *CacheEntry {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// Legacy format is just a filename
	return &CacheEntry{
		Filename:  line,
		Hash:      "", // No hash in legacy format
		Timestamp: 0,  // Unknown timestamp
		Width:     0,  // Unknown dimensions
		Height:    0,
		GCSPaths:  []string{},
	}
}

// Save writes the cache to disk in the v2.0 format
func (c *ImageCache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create backup of existing cache
	if _, err := os.Stat(CacheFilePath); err == nil {
		if err := os.Rename(CacheFilePath, CacheBackupPath); err != nil {
			fmt.Printf("Warning: failed to create cache backup: %v\n", err)
		}
	}

	file, err := os.Create(CacheFilePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write header
	fmt.Fprintf(writer, "# Version: %s\n", CacheVersion)
	fmt.Fprintf(writer, "# Format: filename|hash|timestamp|width|height|gcs_0|gcs_1|gcs_2|gcs_3\n")
	fmt.Fprintln(writer)

	// Sort entries for consistent output
	filenames := make([]string, 0, len(c.entries))
	for filename := range c.entries {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	// Write entries
	for _, filename := range filenames {
		entry := c.entries[filename]
		line := fmt.Sprintf("%s|%s|%d|%d|%d",
			entry.Filename,
			entry.Hash,
			entry.Timestamp,
			entry.Width,
			entry.Height,
		)

		if len(entry.GCSPaths) > 0 {
			line += "|" + strings.Join(entry.GCSPaths, "|")
		}

		fmt.Fprintln(writer, line)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush cache: %w", err)
	}

	c.dirty = false
	return nil
}

// Get retrieves a cache entry (thread-safe)
func (c *ImageCache) Get(filename string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[filename]
	return entry, ok
}

// Has checks if a filename exists in cache (thread-safe)
func (c *ImageCache) Has(filename string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.entries[filename]
	return ok
}

// Add adds or updates a cache entry (thread-safe)
func (c *ImageCache) Add(entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[entry.Filename] = entry
	c.dirty = true
}

// Remove removes a cache entry (thread-safe)
func (c *ImageCache) Remove(filename string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, filename)
	c.dirty = true
}

// Size returns the number of cached entries
func (c *ImageCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// ValidateEntry checks if a cached entry is still valid by comparing content hash
func (c *ImageCache) ValidateEntry(filename string, currentHash string) bool {
	entry, ok := c.Get(filename)
	if !ok {
		return false
	}

	// If no hash stored (legacy entry), consider it invalid
	if entry.Hash == "" {
		return false
	}

	return entry.Hash == currentHash
}

// ComputeHash calculates SHA256 hash of image content from a reader
func ComputeHash(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Stats returns cache statistics
func (c *ImageCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]interface{}{
		"version":       c.version,
		"total_entries": len(c.entries),
		"dirty":         c.dirty,
	}

	// Count entries with/without hash
	withHash := 0
	withoutHash := 0
	for _, entry := range c.entries {
		if entry.Hash != "" {
			withHash++
		} else {
			withoutHash++
		}
	}

	stats["entries_with_hash"] = withHash
	stats["entries_without_hash"] = withoutHash

	return stats
}
