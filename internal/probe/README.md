# probe

The `probe` package wraps `portcheck`, `backoff`, and `circuit` into a single
high-level component that performs resilient, multi-attempt port probes.

## Behaviour

- **Retries** — a failed probe is retried up to `max_attempts` times.
- **Backoff** — wait between attempts grows linearly or exponentially, capped at
  `max_wait`.
- **Circuit breaker** — when a `circuit.Breaker` is supplied, an open circuit
  causes the probe to return immediately without attempting a connection,
  protecting downstream resources during sustained outages.

## Configuration

```json
{
  "max_attempts": 3,
  "initial_wait": "200ms",
  "max_wait": "5s",
  "exponential": true
}
```

| Field         | Type     | Default | Description                              |
|---------------|----------|---------|------------------------------------------|
| `max_attempts`| int      | 1       | Maximum probe attempts before giving up  |
| `initial_wait`| duration | 0       | Wait before the second attempt           |
| `max_wait`    | duration | 0       | Upper bound on inter-attempt delay       |
| `exponential` | bool     | false   | Double the wait on each subsequent retry |

## Usage

```go
cb := circuit.New(circuit.Config{Threshold: 5, Timeout: 30 * time.Second})
p := probe.New(probe.Config{
    MaxAttempts: 3,
    InitialWait: 200 * time.Millisecond,
    MaxWait:     5 * time.Second,
    Exponential: true,
}, cb)

result := p.Run("db.internal", 5432)
if result.Open {
    log.Println("port reachable after", result.Attempts, "attempt(s)")
}
```
