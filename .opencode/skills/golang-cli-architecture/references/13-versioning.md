# 13 — Versioning and Distribution

---

## Build Metadata Embedding

Inject version, commit, and build date at compile time via `-ldflags`:

```go
// internal/version/version.go — importable by the Factory and the version command
package version

var (
    Version = "dev"
    Commit  = "none"
    Date    = "unknown"
)
```

```bash
go build -ldflags "-X myapp/internal/version.Version=1.2.3 \
  -X myapp/internal/version.Commit=$(git rev-parse HEAD) \
  -X myapp/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

goreleaser handles this automatically.

---

## Version Command Output

**For humans (default):**

```
myapp version 1.2.3 (commit: abc1234, built: 2024-01-15T10:30:00Z)
```

**For machines (`--output json`):**

```json
{"version": "1.2.3", "commit": "abc1234", "date": "2024-01-15T10:30:00Z", "go": "go1.22.0"}
```

Include Go version — it helps with bug reports.

---

## Update Checking

**Pattern: opt-in update notification.** On first run, ask if the user wants update notifications. If yes, check once per day (cache the last check timestamp). Display a non-blocking notice:

```
A new version of myapp is available: 1.3.0 (current: 1.2.3)
Run `myapp upgrade` or visit https://github.com/you/myapp/releases
```

Rules:
- Never auto-update without consent
- Never block command execution for update checks
- Use a background goroutine with a short timeout
- Write to stderr (it's diagnostic, not primary output)

Libraries: `creativeprojects/go-selfupdate` for GitHub-based self-update.

---

## Distribution

### goreleaser

The standard for Go CLI distribution. Handles: multi-platform builds, Homebrew taps, Docker images, Snapcraft, APT/RPM repos, checksums, signing.

Shell completion scripts can be generated as build artifacts and included in packages.

### Other Options

| Tool | Purpose |
|------|---------|
| `ko` | Build and publish Go container images without Docker |
| `cosign` | Sign and verify binaries/containers |

---

## Flag and Command Deprecation

See [02-command-ux.md](./02-command-ux.md) for the deprecation pattern. The lifecycle:

1. **Announce** deprecation in changelog and `--help` output
2. **Warn** on use (Cobra prints warnings to stderr automatically)
3. **Remove** after 2–3 minor versions
4. **Document** migration path in each step

---

**See also:** `samber/cc-skills-golang@golang-continuous-integration` — goreleaser in GitHub Actions, release pipelines, automated publishing, and security scanning
