# reaper

The `reaper` package provides a background expiry sweeper for tracked keys.

## Overview

A `Reaper` maintains a set of keys, each associated with a TTL deadline.
At a configurable interval it scans for expired keys, removes them, and
fires a user-supplied callback for each one.

## Usage

```go
r := reaper.New(reaper.Config{
    Interval: 10 * time.Second,
}, func(key string) {
    log.Printf("key expired: %s", key)
})

r.Start()
defer r.Stop()

// Register a key with a 30-second TTL.
r.Track("host:9200", 30*time.Second)

// Remove a key early if the resource is explicitly closed.
r.Remove("host:9200")
```

## Config

| Field      | Type          | Default | Description                        |
|------------|---------------|---------|------------------------------------|
| `Interval` | `time.Duration` | 30s   | How often the reap pass runs.      |

## Notes

- `Track` is idempotent; calling it again on the same key resets its TTL.
- The callback is invoked outside the internal lock — safe for I/O.
- Call `Stop` to cleanly shut down the background goroutine.
