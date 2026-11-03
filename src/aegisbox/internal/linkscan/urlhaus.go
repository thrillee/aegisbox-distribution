package linkscan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const urlhausURLEndpoint = "https://urlhaus-api.abuse.ch/v1/url/"

// URLhaus looks URLs up in abuse.ch's URLhaus list of malware URLs. It is used only when
// LINKSCAN_URLHAUS_AUTH_KEY is set, after the owner has confirmed the terms for commercial use.
type URLhaus struct {
	key      string
	endpoint string
	client   *http.Client
}

func NewURLhaus(authKey string, client *http.Client) *URLhaus {
	if authKey == "" {
		return nil
	}
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &URLhaus{key: authKey, endpoint: urlhausURLEndpoint, client: client}
}

// Lookup returns a short description ("URLhaus: malware_download") when the URL is listed, "" when
// it is not. Errors mean no signal; they never fail the scan.
func (u *URLhaus) Lookup(ctx context.Context, rawURL string) (string, error) {
	if u == nil || rawURL == "" {
		return "", nil
	}
	form := url.Values{"url": {rawURL}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Auth-Key", u.key)
	resp, err := u.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("URLhaus answered %d", resp.StatusCode)
	}
	var body struct {
		QueryStatus string `json:"query_status"`
		URLStatus   string `json:"url_status"`
		Threat      string `json:"threat"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return "", err
	}
	switch body.QueryStatus {
	case "ok":
		threat := body.Threat
		if threat == "" {
			threat = "listed"
		}
		return fmt.Sprintf("URLhaus: %s (%s)", threat, body.URLStatus), nil
	case "no_results":
		return "", nil
	default:
		return "", fmt.Errorf("URLhaus query status %q", body.QueryStatus)
	}
}
