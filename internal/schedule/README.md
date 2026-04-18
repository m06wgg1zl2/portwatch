# schedule

Provides a configurable polling schedule with optional jitter for portwatch monitors.

## Config

| Field              | Type | Default | Description                          |
|--------------------|------|---------|--------------------------------------|
| `interval_seconds` | int  | 30      | Base polling interval in seconds     |
| `jitter_seconds`   | int  | 0       | Max random jitter added to interval  |

## Usage

```go
cfg := schedule.Config{IntervalSeconds: 15, JitterSeconds: 3}
s := schedule.New(cfg)

tk := s.Ticker()
defer tk.Stop()

for range tk.C {
    // perform port check
}
```

Jitter helps avoid thundering-herd problems when many monitors start simultaneously.
The jitter value is derived from the current wall-clock nanoseconds modulo the jitter
duration, so it is deterministic within a given nanosecond but varies across ticks.
