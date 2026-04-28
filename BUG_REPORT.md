# Bug Report - File Keepalive Service

## Critical Bugs

### 1. ✅ FIXED: Unused Import in main.go
- **File**: `main.go:6`
- **Issue**: `fmt` package imported but not used
- **Status**: Fixed - removed unused import

## High Priority Issues

### 2. Race Condition in AntiBotManager
- **File**: `antibot.go:17`
- **Issue**: `rand.Rand` is not thread-safe. Concurrent access from multiple goroutines will cause data races.
- **Impact**: Potential crashes or undefined behavior in concurrent scenarios
- **Fix**: Use `sync.Mutex` to protect rand access or use `math/rand` package-level functions (which are thread-safe in Go 1.20+)

### 3. Context Not Respected in Retry Logic
- **File**: `keepalive.go:267`
- **Issue**: Recursive retry call doesn't check if context is cancelled before retrying
- **Impact**: Service may not shut down gracefully when interrupted during retry
- **Fix**: Check `ctx.Done()` before recursive call

### 4. Memory Leak in StateManager
- **File**: `state.go:23`
- **Issue**: `ProcessedFiles` map grows indefinitely, never cleaned up
- **Impact**: Memory usage grows unbounded over time, especially with long-running services
- **Fix**: Implement periodic cleanup or use a bounded cache structure

## Medium Priority Issues

### 5. Silent Error Handling
- **File**: `keepalive.go:88`
- **Issue**: `_ = ks.supabase.UpdateLastAccessed(recordingID)` ignores errors
- **Impact**: Failures to update last_accessed timestamp go unnoticed
- **Fix**: Log the error at minimum

### 6. Inefficient String Contains Implementation
- **File**: `antibot.go:149-165`
- **Issue**: Custom `contains()` and `containsMiddle()` functions are complex and inefficient
- **Impact**: Unnecessary CPU usage
- **Fix**: Use `strings.Contains()` from standard library

### 7. Cross-Platform File Permissions
- **File**: `state.go:95`
- **Issue**: Unix-style permissions `0644` may not work correctly on Windows
- **Impact**: Potential file access issues on Windows
- **Fix**: Use `0666` or handle platform-specific permissions

## Low Priority Issues

### 8. HTTP Client Timeout Set to 0
- **File**: `keepalive.go:38`
- **Issue**: `Timeout: 0` means no timeout for large downloads
- **Impact**: Hung connections could block indefinitely
- **Note**: This might be intentional for large file downloads, but should be documented

### 9. Magic Numbers in Delay Logic
- **File**: `keepalive.go:145-146`
- **Issue**: Hard-coded 300 seconds (5 minutes) delay
- **Impact**: Not configurable without code changes
- **Suggestion**: Make this a command-line flag or configuration option

### 10. No Validation of Environment Variables
- **File**: `main.go:24-26`
- **Issue**: No validation that URLs are well-formed or keys are non-empty strings
- **Impact**: Cryptic errors later in execution
- **Fix**: Add basic validation (URL format, minimum key length)

## Recommendations

1. Add unit tests, especially for the anti-bot logic and state management
2. Consider using a proper HTTP client library with built-in retry logic (e.g., `hashicorp/go-retryablehttp`)
3. Add structured logging (e.g., `zerolog` or `zap`) instead of standard `log` package
4. Consider using a proper rate limiter library instead of custom implementation
5. Add metrics/monitoring support (Prometheus, etc.)
6. Document the expected Supabase schema (especially the `last_accessed` column requirement)
