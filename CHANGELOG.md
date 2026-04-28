# Changelog

All notable changes to this project will be documented in this file.

## [1.1.0] - 2026-04-28

### Fixed

#### Critical Bugs
- **Fixed unused import error**: Removed unused `fmt` import from `main.go` that prevented compilation
- **Fixed race condition in AntiBotManager**: Added `sync.Mutex` to protect concurrent access to `rand.Rand` instance
- **Fixed context cancellation in retry logic**: Retry operations now properly respect context cancellation for graceful shutdown
- **Fixed memory leak in StateManager**: Implemented bounded cache with automatic cleanup (max 10,000 entries) to prevent unbounded memory growth

#### High Priority Issues
- **Fixed silent error handling**: `UpdateLastAccessed()` errors are now logged instead of being silently ignored
- **Replaced inefficient string functions**: Removed custom `contains()` and `containsMiddle()` functions, now using `strings.Contains()` from standard library
- **Fixed cross-platform file permissions**: Changed from Unix-style `0644` to cross-platform `0666` for state file writes

#### Medium Priority Issues
- **Added URL validation**: Supabase URL is now validated to ensure it starts with `http://` or `https://`
- **Added API key validation**: Basic validation to ensure API key is not empty and has minimum length
- **Made delay configurable**: Added `--delay-seconds` flag to make inter-file delay configurable (default: 300 seconds)

### Added
- **New command-line flag**: `--delay-seconds` to configure delay between file downloads
- **Memory management**: Automatic cleanup of old processed files to prevent memory leaks
- **Better error logging**: All errors are now properly logged with context
- **Thread-safety**: All concurrent operations are now properly synchronized with mutexes

### Changed
- **Project structure**: Organized files into logical folders:
  - `deploy/` - Docker and deployment configurations
  - `scripts/` - Setup and utility scripts
- **State management**: Added `ProcessedOrder` slice to track insertion order for cleanup
- **State struct**: Added `MaxProcessedFiles` field to control memory usage
- **Anti-bot manager**: All random number generation is now thread-safe

### Improved
- **Code quality**: Passed `go vet` with no warnings
- **Documentation**: Comprehensive README with usage examples and troubleshooting
- **Error messages**: More descriptive error messages for common issues
- **Shutdown handling**: Better graceful shutdown with proper context propagation

## [1.0.0] - Initial Release

### Features
- Automatic file access to prevent deletion
- Supabase integration for file tracking
- Anti-bot detection avoidance
- Rate limiting
- State management with persistence
- Docker support
- Configurable intervals and age filters
