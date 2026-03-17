# Version Package

Package: `internal/version`

## Purpose

Holds build-time injected version metadata following the germinator pattern. Version, Commit, and Date variables are injected by GoReleaser or local build tasks.

## Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `Version` | Version string | `"dev"` |
| `Commit` | Full commit hash | `""` (empty) |
| `Date` | Build date | `""` (empty) |

## Build-Time Injection

Variables are injected via Go ldflags:

| Build System | ldflags Path |
|--------------|--------------|
| GoReleaser | `-X twiggit/internal/version.Version` |
| Local builds | `-X twiggit/internal/version.Version` |

**Note:** Case difference - GoReleaser uses `Version` (capitalized), local build configs use `version` (lowercase) due to Go syntax in shell scripts.

## String() Function

Returns formatted version string with variations:

| State | Format | Example |
|-------|--------|---------|
| Full info | `<version> (<commit>) <date>` | `1.0.0 (abc123def456) 2025-03-17` |
| No date | `<version> (<commit>) ` | `1.0.0 (abc123def456) ` |
| Dev build | `<version> () ` | `dev () ` |

**Note:** String() does NOT include the "twiggit " prefix - the command layer prepends it.

## Usage Pattern

Command layer imports and calls `version.String()`:

```go
import "twiggit/internal/version"

// In command output
fmt.Printf("twiggit %s", version.String())
```
