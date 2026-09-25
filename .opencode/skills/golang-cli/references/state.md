# State and Persistence

XDG locations, atomic writes, embedded stores, and locking between concurrent invocations. Schema design, migrations, and queries once you reach SQLite → See `samber/cc-skills-golang@golang-database`.

## Contents

- [What CLIs persist](#what-clis-persist)
- [Storage options](#storage-options)
- [Concurrent invocation](#concurrent-invocation)
- [Auth token lifecycle](#auth-token-lifecycle)

## What CLIs Persist

| Data | XDG Location | Format |
| --- | --- | --- |
| Config (preferences, defaults) | `$XDG_CONFIG_HOME/<app>/` | YAML, TOML, JSON |
| Auth tokens | OS keyring, else a 0600 file in `$XDG_CONFIG_HOME/<app>/` | Keyring entry, JSON |
| Cache (API responses, computed data) | `$XDG_CACHE_HOME/<app>/` | Any |
| Application data (local DB, user-created content) | `$XDG_DATA_HOME/<app>/` | SQLite, bbolt, JSON |
| State (history, recent items, last update check) | `$XDG_STATE_HOME/<app>/` | JSON, bbolt |
| Lock files shared by cron and login sessions | `$XDG_STATE_HOME/<app>/` | Files |
| Session-scoped runtime files (sockets, PID for the login session) | `$XDG_RUNTIME_DIR/<app>/` | Files |

Use `adrg/xdg` for cross-platform XDG path resolution.

## Storage Options

Library picks (`bbolt`, `modernc.org/sqlite`, `go-keyring`, `gofrs/flock`) are in [libraries.md](./libraries.md#storage); this section covers when to reach for each.

### Plain Files (JSON/YAML)

Simplest. Good for config and small state. No concurrent access guarantees.

For atomic writes, write a uniquely named temp file in the same directory, sync it, then rename — a fixed `path + ".tmp"` name collides when two invocations run at once, and a temp file on another filesystem breaks the atomic rename:

```go
func atomicWrite(path string, data []byte, perm os.FileMode) error {
    tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
    if err != nil {
        return err
    }
    defer os.Remove(tmp.Name()) // no-op after a successful rename
    if _, err := tmp.Write(data); err != nil {
        tmp.Close()
        return err
    }
    if err := tmp.Sync(); err != nil {
        tmp.Close()
        return err
    }
    if err := tmp.Close(); err != nil {
        return err
    }
    if err := os.Chmod(tmp.Name(), perm); err != nil {
        return err
    }
    return os.Rename(tmp.Name(), path)
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

For sensitive credentials (API tokens, passwords). Use `zalando/go-keyring` for cross-platform access (macOS Keychain, Linux Secret Service, Windows Credential Manager). When no keyring is available (headless Linux, containers), fall back to a file written with mode 0600 and print a one-time stderr warning that the token is stored in plain text — the `gh` model. An encrypted file whose key sits on the same disk adds no protection.

## Concurrent Invocation

When multiple instances might run simultaneously (parallel CI, multiple terminals):

**File locking** with `gofrs/flock`:

```go
// cron often lacks XDG_RUNTIME_DIR, so the state dir keeps one path for every caller.
// flock releases when the holder dies, so a crash leaves no stale lock.
path, err := xdg.StateFile("myapp/sync.lock")
if err != nil {
    return err
}
lock := flock.New(path)
locked, err := lock.TryLock()
if !locked {
    return fmt.Errorf("another instance is running")
}
defer lock.Unlock()
```

The atomic rename prevents truncation; the lock prevents lost updates. Hold it across the whole read-modify-write, or two instances each load the old file and the second save drops the first one's change:

```go
lock := flock.New(statePath + ".lock")
if err := lock.Lock(); err != nil {
    return err
}
defer lock.Unlock()
state, err := load(statePath) // read inside the lock
if err != nil {
    return err
}
state.Add(project)
return atomicWrite(statePath, state.Marshal(), 0o644)
```

**SQLite WAL mode** handles concurrent reads naturally. For writes, `_busy_timeout` retries rather than failing immediately.

**bbolt** allows multiple readers but one writer. Use a lock file around bbolt write operations if contention is expected.

## Auth Token Lifecycle

Ship `login` together with `status` and `logout`: a token the user cannot inspect or remove is a support ticket. Every command that needs the token fails through the [auth hook](./commands.md#middleware-via-cobra-hooks) with the hint `run 'myapp auth login'`. The pattern (used by `gh`, `gcloud`, `aws`):

1. `myapp auth login` — starts OAuth flow (opens browser, listens on localhost callback)
2. Token stored in config/keyring
3. Subsequent commands read token from storage
4. `myapp auth status` — show current auth state
5. `myapp auth refresh` — refresh expired token
6. `myapp auth logout` — clear stored credentials

Browser-based OAuth for CLIs: spawn a temporary localhost HTTP server, open the browser to the authorization URL with redirect to `http://localhost:<port>/callback`, receive the code, exchange for token.
