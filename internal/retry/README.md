# retry

Provides a simple retry policy for portwatch operations such as webhook delivery and port checks.

## Usage

```go
p := retry.New(3, 500*time.Millisecond, 2.0)
err := p.Do(func() error {
    return callWebhook(url, payload)
})
```

## Configuration

| Field | Type | Description |
|-------|------|-------------|
| `max_attempts` | int | Maximum number of attempts (minimum 1) |
| `delay_ms` | int | Initial delay between attempts in milliseconds |
| `backoff` | float64 | Multiplier applied to delay after each attempt (e.g. 2.0 doubles it) |

## Behaviour

- Returns `nil` immediately on first successful attempt.
- Returns the **last** error if all attempts are exhausted.
- A `backoff` of `1.0` gives a constant delay; values greater than `1.0` produce exponential back-off.
