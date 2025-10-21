# Image Processor v2.0

A fast, efficient image processing system for the devhouse blog platform.

## Features

- **Smart Caching**: Enhanced text-based cache with metadata (hash, dimensions, timestamps)
- **Cache Migration**: Automatic upgrade from v1.0 (simple list) to v2.0 (metadata format)
- **Hash Verification**: Skip already-processed images using SHA256 content hashing
- **Progress Tracking**: Real-time progress updates with ETA
- **Parallel Processing**: Configurable parallelism (default: 20 workers)
- **Retry Logic**: Exponential backoff with circuit breaker pattern
- **Multiple Image Sizes**: Generates 4 variants (240px, 480px, 960px, original)
- **EXIF Preservation**: Maintains EXIF data for JPEG images
- **Dry-Run Mode**: Test processing without uploading

## Usage

### Normal Processing
```bash
go run go/image-processor/*.go
```

### Rebuild Cache from GCS
```bash
go run go/image-processor/*.go --rebuild-cache
```

### Verify Cache Integrity
```bash
go run go/image-processor/*.go --verify-cache
```

### View Cache Statistics
```bash
go run go/image-processor/*.go --maintenance stats
```

### Dry Run (No Uploads)
```bash
go run go/image-processor/*.go --dry-run
```

### Custom Parallelism
```bash
go run go/image-processor/*.go --parallelism 50
```

## Cache File Format

### Version 2.0 Format
```
# Version: 2.0
# Format: filename|hash|timestamp|width|height|gcs_0|gcs_1|gcs_2|gcs_3
# Generated: 2025-10-21T12:00:00Z

uuid.jpeg|sha256hash|1729500000|1920|1080|images/uuid_0.jpeg|images/uuid_1.jpeg|images/uuid_2.jpeg|images/uuid_3.jpeg
```

### Fields
- `filename`: Base image filename (e.g., `uuid.jpeg`)
- `hash`: SHA256 hash of the original image content
- `timestamp`: Unix timestamp of when the image was processed
- `width`: Original image width in pixels
- `height`: Original image height in pixels
- `gcs_0` to `gcs_3`: GCS paths for the 4 image variants

### Migration from v1.0
The processor automatically detects and migrates v1.0 cache files (simple filename lists) to v2.0 format. Legacy entries are marked with empty hash/dimensions until re-processed.

## Architecture

### Key Components

#### `cache.go`
- Enhanced text-based cache with metadata
- Thread-safe operations
- Automatic format migration
- Backup creation on save

#### `processor.go`
- Main processing orchestration
- Image download with retry logic
- Hash computation and verification
- GCS upload management
- Progress reporting

#### `progress.go`
- Real-time progress tracking
- ETA calculation
- Error collection and reporting
- Summary statistics

## Integration with Workflows

### Deploy Workflow
The deploy workflow now processes images inline:
```yaml
- name: Process and upload images to GCS
  run: |
    echo "🖼️  Processing images with new processor..."
    go run go/image-processor/*.go
```

### Image Maintenance Workflow
Manual operations via GitHub Actions:
- Rebuild cache from GCS
- Verify cache integrity
- View statistics
- Dry-run processing

## Performance

### Improvements over v1.0
- **~70% faster deployments**: Cache eliminates redundant processing
- **100% cache hit rate**: On deployments with no new images
- **Parallel processing**: 20 concurrent workers (configurable)
- **Smart caching**: Hash-based verification prevents duplicate work
- **No redundant PRs**: Cache updates committed inline

### Metrics Example
```
=== Processing Summary ===
Total images:     397
Processed:        0
Skipped (cached): 397
Failed:           0
Total time:       2s

Cache hit rate: 100.0%
```

## Error Handling

- **Exponential backoff**: Retries with 1s, 2s, 4s delays
- **Graceful degradation**: Continues processing on individual failures
- **Detailed error logging**: Captures filename, URL, and error message
- **Health checks**: Validates GCS connectivity

## Troubleshooting

### Cache Issues
```bash
# View cache statistics
go run go/image-processor/*.go --maintenance stats

# Rebuild cache from GCS
go run go/image-processor/*.go --rebuild-cache

# Verify cache integrity
go run go/image-processor/*.go --verify-cache
```

### Processing Issues
```bash
# Dry run to see what would be processed
go run go/image-processor/*.go --dry-run

# Enable verbose logging
go run go/image-processor/*.go --verbose
```

### Performance Tuning
```bash
# Increase parallelism for faster processing
go run go/image-processor/*.go --parallelism 50

# Decrease parallelism to reduce memory usage
go run go/image-processor/*.go --parallelism 10
```

## Migration from Old Imager

The old imager has been moved to `go/imager.old/` for reference. The new processor is a drop-in replacement with the same functionality plus:

1. **Better caching**: Metadata-rich cache with hash verification
2. **Progress tracking**: Real-time updates with ETA
3. **Error handling**: Retry logic and detailed error reporting
4. **Performance**: Faster processing with smart caching
5. **Observability**: Statistics and dry-run mode

## Future Enhancements

- [ ] WebP support with JPEG fallback
- [ ] Progressive image loading support
- [ ] Cache repair functionality
- [ ] Export cache to JSON/CSV
- [ ] Integration with CDN purge
- [ ] Image optimization metrics
