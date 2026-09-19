# 10 — State and Persistence

This reference owns CLI-specific persistence: XDG locations, atomic file writes, single-binary embedded stores, and cross-invocation locking. For relational schema design, migrations, query patterns, and transaction handling once you reach SQLite, defer to `samber/cc-skills-golang@golang-database`.

---

## What CLIs Persist

| Data | XDG Location | Format |
|------|-------------|--------|
| Config (preferences, defaults) | `$XDG_CONFIG_HOME/<app>/` | YAML, TOML, JSON |
| Auth tokens, session data | `$XDG_CONFIG_HOME/<app>/` or OS keyring | JSON, encrypted |
| Cache (API responses, computed data) | `$XDG_CACHE_HOME/<app>/` | Any |
| Application data (local DB, history) | `$XDG_DATA_HOME/<app>/` | SQLite, bbolt, JSON |
| Runtime/transient (lock files, PID) | `$XDG_RUNTIME_DIR/<app>/` | Files |

Use `adrg/xdg` for cross-platform XDG path resolution.

---

## Storage Options

Library picks (`bbolt`, `modernc.org/sqlite`, `go-keyring`, `gofrs/flock`) are in
[14-libraries.md](./14-libraries.md#storage); this section covers when to reach for each.

### Plain Files (JSON/YAML)

Simplest. Good for config and small state. No concurrent access guarantees.

For atomic writes, use the write-temp-then-rename pattern:

```go
func atomicWrite(path string, data []byte, perm os.FileMode) error {
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, data, perm); err != nil {
        return err
    }
    return os.Rename(tmp, path)
}
```

### bbolt (etcd-io/bbolt)

Embedded key/value store. Pure Go, single file, ACID transactions, crash-safe.

**Good for:** Token storage, session caches, local indexes. Single writer, multiple readers.

### SQLite

Full relational database. Two options:
- `mattn/go-sqlite3` — requires CGO
- `modernc.org/sqlite` — pure Go port, better for cross-compilation

**Good for:** Complex queries, large datasets, migration-friendly schemas.

Use `_journal_mode=WAL` and `_busy_timeout=5000` for better concurrent access.

### OS Keyring

For sensitive credentials (API tokens, passwords). Use `zalando/go-keyring` for cross-platform access (macOS Keychain, Linux Secret Service, Windows Credential Manager). Fall back to encrypted file if keyring unavailable.

---

## Concurrent Invocation

When multiple instances might run simultaneously (parallel CI, multiple terminals):

**File locking** with `gofrs/flock`:

```go
lock := flock.New("/path/to/app.lock")
locked, err := lock.TryLock()
if !locked {
    return fmt.Errorf("another instance is running")
}
defer lock.Unlock()
```

**SQLite WAL mode** handles concurrent reads naturally. For writes, `_busy_timeout` retries rather than failing immediately.

**bbolt** allows multiple readers but one writer. Use a lock file around bbolt write operations if contention is expected.

---

## Auth Token Lifecycle

Common pattern (used by `gh`, `gcloud`, `aws`):

1. `myapp auth login` — starts OAuth flow (opens browser, listens on localhost callback)
2. Token stored in config/keyring
3. Subsequent commands read token from storage
4. `myapp auth status` — show current auth state
5. `myapp auth refresh` — refresh expired token
6. `myapp auth logout` — clear stored credentials

Browser-based OAuth for CLIs: spawn a temporary localhost HTTP server, open the browser to the authorization URL with redirect to `http://localhost:<port>/callback`, receive the code, exchange for token.
