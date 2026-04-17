# filter

The `filter` package provides time-window and state-based suppression of port-watch alerts.

## Usage

```go
rules := []filter.Rule{
    {FromHour: 8, ToHour: 20, States: []string{"closed"}},
}
f := filter.New(rules)

if f.Allow(state, time.Now()) {
    // forward the alert
}
```

## Rule fields

| Field | Type | Description |
|-------|------|-------------|
| `from_hour` | int | Start of allowed window (0-23, UTC) |
| `to_hour` | int | End of allowed window (exclusive). Wraps midnight when `from > to`. |
| `states` | []string | Allowed port states (`"open"`, `"closed"`). Empty means all states. |

All rules are ANDed together. If no rules are defined every event is allowed.
