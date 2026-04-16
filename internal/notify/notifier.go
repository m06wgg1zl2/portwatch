package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Payload is sent to webhook callbacks.
type Payload struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	State  string `json:"state"`
	At     string `json:"at"`
}

// Notifier dispatches webhook or shell callbacks on port state changes.
type Notifier struct {
	client *http.Client
}

// New returns a Notifier with a sensible HTTP timeout.
func New() *Notifier {
	return &Notifier{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Webhook sends a JSON POST request to the given URL.
func (n *Notifier) Webhook(url, host string, port int, state string) error {
	p := Payload{
		Host:  host,
		Port:  port,
		State: state,
		At:    time.Now().UTC().Format(time.RFC3339),
	}
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}
	resp, err := n.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: webhook POST: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notify: webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// Shell executes a shell command, injecting state info via environment-style
// variable substitution: {host}, {port}, {state}.
func (n *Notifier) Shell(command, host string, port int, state string) error {
	replacer := strings.NewReplacer(
		"{host}", host,
		"{port}", fmt.Sprintf("%d", port),
		"{state}", state,
	)
	cmd := replacer.Replace(command)
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("notify: shell command failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}
