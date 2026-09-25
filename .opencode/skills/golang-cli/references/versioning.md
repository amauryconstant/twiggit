# Versioning and Distribution

## Contents

- [Build metadata](#build-metadata)
- [Version command output](#version-command-output)
- [Update checking](#update-checking)
- [Distribution](#distribution)
- [Flag and command deprecation](#flag-and-command-deprecation)

## Build metadata

Never hardcode the version string: it drifts from tags. Inject it with `-ldflags` and fall back to `debug.ReadBuildInfo` — [version.go](../assets/examples/version.go).

[version.go](../assets/examples/version.go) holds `Version`, `Commit`, and `Date` in `internal/version`, set at build time:

```bash
go build -ldflags "-X github.com/you/myapp/internal/version.Version=v1.2.3 \
  -X github.com/you/myapp/internal/version.Commit=$(git rev-parse --short HEAD) \
  -X github.com/you/myapp/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

- `-X` needs the **full import path** of the package; a short path silently sets nothing.
- `go install module@version` builds carry no ldflags, so `version.Get()` falls back to `debug.ReadBuildInfo()`: the module version plus the `vcs.revision` and `vcs.time` stamps Go embeds.
- goreleaser sets the same flags from the git tag.

## Version command output

**For humans (default):**

```
myapp v1.2.3 (go1.24.0, commit abc1234, built 2024-01-15T10:30:00Z)
```

**For machines (`--output json`)** — [version_cmd.go](../assets/examples/version_cmd.go):

```json
{
  "version": "v1.2.3",
  "commit": "abc1234",
  "date": "2024-01-15T10:30:00Z",
  "go": "go1.24.0"
}
```

Include the Go version — bug reports need it.

## Update checking

**Pattern: opt-in update notification.** On first run, ask if the user wants update notifications. If yes, check once per day (cache the last check timestamp). Display a non-blocking notice:

```
A new version of myapp is available: 1.3.0 (current: 1.2.3)
Run `myapp upgrade` or visit https://github.com/you/myapp/releases
```

Rules:

- Update only with the user's explicit consent.
- Check in a background goroutine with a short timeout, so commands run at full speed.
- Print the notice to stderr — it is diagnostic, not the result.
- Skip the check in CI and when stdout is not a TTY.

Libraries: `creativeprojects/go-selfupdate` for GitHub-based self-update.

## Distribution

### goreleaser

The standard for Go CLI distribution. Handles: multi-platform builds, Homebrew taps, Docker images, Snapcraft, APT/RPM repos, checksums, signing.

Shell completion scripts can be generated as build artifacts and included in packages.

### Other Options

| Tool     | Purpose                                              |
| -------- | ---------------------------------------------------- |
| `ko`     | Build and publish Go container images without Docker |
| `cosign` | Sign and verify binaries/containers                  |

## Flag and command deprecation

See [command-ux.md](./command-ux.md) for the deprecation pattern. The lifecycle:

1. **Announce** deprecation in changelog and `--help` output
2. **Warn** on use (Cobra prints warnings to stderr automatically)
3. **Remove** after 2–3 minor versions
4. **Document** migration path in each step

→ See `samber/cc-skills-golang@golang-continuous-integration` — goreleaser in GitHub Actions, release pipelines, automated publishing, and security scanning
