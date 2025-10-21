package main

import (
	"fmt"
	"sync"
	"time"
)

// ProgressTracker tracks image processing progress
type ProgressTracker struct {
	mu sync.Mutex

	total     int
	processed int
	skipped   int
	failed    int
	startTime time.Time

	currentImage string
	errors       []ProcessingError
}

// ProcessingError represents a failed image processing attempt
type ProcessingError struct {
	Filename string
	URL      string
	Error    error
	Time     time.Time
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(total int) *ProgressTracker {
	return &ProgressTracker{
		total:     total,
		startTime: time.Now(),
		errors:    make([]ProcessingError, 0),
	}
}

// SetCurrent sets the currently processing image
func (p *ProgressTracker) SetCurrent(filename string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentImage = filename
}

// IncrementProcessed increments the processed counter
func (p *ProgressTracker) IncrementProcessed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.processed++
}

// IncrementSkipped increments the skipped counter
func (p *ProgressTracker) IncrementSkipped() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.skipped++
}

// IncrementFailed increments the failed counter
func (p *ProgressTracker) IncrementFailed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failed++
}

// AddError adds an error to the tracker
func (p *ProgressTracker) AddError(filename, url string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failed++
	p.errors = append(p.errors, ProcessingError{
		Filename: filename,
		URL:      url,
		Error:    err,
		Time:     time.Now(),
	})
}

// GetProgress returns current progress information
func (p *ProgressTracker) GetProgress() (processed, skipped, failed, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.processed, p.skipped, p.failed, p.total
}

// PrintProgress prints a progress update
func (p *ProgressTracker) PrintProgress() {
	p.mu.Lock()
	defer p.mu.Unlock()

	completed := p.processed + p.skipped + p.failed
	percentage := float64(completed) / float64(p.total) * 100
	elapsed := time.Since(p.startTime)

	// Estimate time remaining
	var eta string
	if completed > 0 {
		rate := float64(completed) / elapsed.Seconds()
		remaining := p.total - completed
		etaSeconds := float64(remaining) / rate
		eta = formatDuration(time.Duration(etaSeconds * float64(time.Second)))
	} else {
		eta = "calculating..."
	}

	fmt.Printf("\r[%3.0f%%] %d/%d images | Processed: %d | Skipped: %d | Failed: %d | Elapsed: %s | ETA: %s",
		percentage,
		completed,
		p.total,
		p.processed,
		p.skipped,
		p.failed,
		formatDuration(elapsed),
		eta,
	)
}

// PrintSummary prints a final summary
func (p *ProgressTracker) PrintSummary() {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Println("\n\n=== Processing Summary ===")
	fmt.Printf("Total images:     %d\n", p.total)
	fmt.Printf("Processed:        %d\n", p.processed)
	fmt.Printf("Skipped (cached): %d\n", p.skipped)
	fmt.Printf("Failed:           %d\n", p.failed)
	fmt.Printf("Total time:       %s\n", formatDuration(time.Since(p.startTime)))

	if p.processed > 0 {
		avgTime := time.Since(p.startTime) / time.Duration(p.processed)
		fmt.Printf("Avg time/image:   %s\n", formatDuration(avgTime))
	}

	if len(p.errors) > 0 {
		fmt.Println("\n=== Errors ===")
		for i, err := range p.errors {
			fmt.Printf("%d. %s (%s): %v\n", i+1, err.Filename, err.URL, err.Error)
		}
	}

	// Calculate cache hit rate
	if p.total > 0 {
		hitRate := float64(p.skipped) / float64(p.total) * 100
		fmt.Printf("\nCache hit rate: %.1f%%\n", hitRate)
	}
}

// GetErrors returns all processing errors
func (p *ProgressTracker) GetErrors() []ProcessingError {
	p.mu.Lock()
	defer p.mu.Unlock()
	errors := make([]ProcessingError, len(p.errors))
	copy(errors, p.errors)
	return errors
}

// HasErrors returns true if any errors occurred
func (p *ProgressTracker) HasErrors() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.errors) > 0
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
