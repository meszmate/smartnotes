package captcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TurnstileConfig struct {
	Secret        string
	SiteVerifyURL string
	HTTPClient    *http.Client
	ExpectedHost  string // optional: verify hostname
}

type Response struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	Action      string   `json:"action,omitempty"`
	CData       string   `json:"cdata,omitempty"`
}

type Turnstile struct {
	cfg TurnstileConfig
}

func NewTurnstile(turnstileSecret string) *Turnstile {
	return newTurnstile(TurnstileConfig{
		Secret:        turnstileSecret,
		SiteVerifyURL: "https://challenges.cloudflare.com/turnstile/v0/siteverify",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	})
}

func newTurnstile(cfg TurnstileConfig) *Turnstile {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	if cfg.SiteVerifyURL == "" {
		cfg.SiteVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	}
	return &Turnstile{cfg: cfg}
}

func (t *Turnstile) Verify(ctx context.Context, token, remoteIP string) (bool, error) {
	if token == "" {
		return false, errors.New("empty token")
	}

	data := url.Values{
		"secret":   {t.cfg.Secret},
		"response": {token},
	}
	if remoteIP != "" {
		if net.ParseIP(remoteIP) == nil {
			return false, fmt.Errorf("invalid remote IP: %s", remoteIP)
		}
		data.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.cfg.SiteVerifyURL, strings.NewReader(data.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.cfg.HTTPClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("turnstile request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("turnstile bad status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return false, fmt.Errorf("invalid JSON from turnstile: %w", err)
	}

	if !r.Success {
		return false, fmt.Errorf("turnstile verification failed: %v", r.ErrorCodes)
	}

	// Optional: hostname check
	if t.cfg.ExpectedHost != "" && r.Hostname != t.cfg.ExpectedHost {
		return false, fmt.Errorf("hostname mismatch: got %s, expected %s", r.Hostname, t.cfg.ExpectedHost)
	}

	// Optional: timestamp check
	if r.ChallengeTs != "" {
		ts, err := time.Parse(time.RFC3339, r.ChallengeTs)
		if err != nil {
			return false, fmt.Errorf("invalid challenge_ts: %w", err)
		}
		if time.Since(ts) > 5*time.Minute {
			return false, errors.New("challenge timestamp too old")
		}
	}

	return true, nil
}
