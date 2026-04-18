# portwatch

A lightweight CLI daemon that monitors port availability and triggers configurable webhook or shell callbacks on state changes.

---

## Installation

```bash
go install github.com/youruser/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/youruser/portwatch.git && cd portwatch && go build -o portwatch .
```

---

## Usage

```bash
portwatch --port 8080 --interval 10s \
  --on-open "curl -X POST https://hooks.example.com/up" \
  --on-close "curl -X POST https://hooks.example.com/down"
```

**Flags:**

| Flag | Description | Default |
|------|-------------|---------|
| `--port` | Port number to monitor | `80` |
| `--host` | Host to monitor | `localhost` |
| `--interval` | Poll interval | `5s` |
| `--on-open` | Shell command or webhook URL to trigger when port opens | — |
| `--on-close` | Shell command or webhook URL to trigger when port closes | — |
| `--config` | Path to a YAML config file | — |
| `--timeout` | TCP dial timeout per attempt | `3s` |

**Example config file (`portwatch.yaml`):**

```yaml
host: localhost
port: 5432
interval: 10s
timeout: 3s
on_open: "echo 'DB is up'"
on_close: "alertmanager notify --service postgres"
```

Run with a config file:

```bash
portwatch --config portwatch.yaml
```

---

## How It Works

`portwatch` continuously polls the specified host/port using TCP dial attempts. When the port state changes (open → closed or closed → open), it executes the configured shell command or fires an HTTP POST to the provided webhook URL.

---

## License

MIT © 2024 youruser
