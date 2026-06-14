// Package csdn is the library behind the csdn command line: the HTTP client, the
// open hot-rank board, the search and HTML surfaces, and the typed data models
// for CSDN (CSDN博客).
//
// CSDN is open: every surface serves anonymously over plain HTTPS GET, with no
// request signing. The catch is an anti-bot edge that returns an HTTP 521 or an
// HTML challenge when a request lacks full browser headers, so the client sends
// a real desktop Chrome User-Agent, an Accept-Language, and a Referer on every
// request. When the edge still wins it serves a challenge shell; the client
// returns ErrWalled and the command layer reports it honestly instead of a
// silent empty result.
package csdn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Host is the CSDN front-page origin, the default Referer for HTML reads.
const Host = "https://www.csdn.net"

// blogHost serves the hot-rank board, article pages, profile pages, and the
// community home-api JSON.
const blogHost = "https://blog.csdn.net"

// soHost serves the search API.
const soHost = "https://so.csdn.net"

// DefaultUserAgent is a current desktop Chrome string. An honest, real
// User-Agent is both polite and the thing most likely to keep a request from
// tripping the anti-bot edge, which scores non-browser callers.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// ErrWalled means the anti-bot edge served a challenge instead of the surface.
// The request needs a residential session (or a retry).
var ErrWalled = errors.New("CSDN served an anti-bot challenge (try again, or this surface needs a residential session)")

// ErrNotFound means the call parsed but carried no record for the request, for
// example an article id or username that does not exist.
var ErrNotFound = errors.New("not found")

// Config holds the tunable client settings.
type Config struct {
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns polite defaults: a 600ms gap between requests, a 30s
// timeout, and five retries on transient errors.
func DefaultConfig() Config {
	return Config{
		UserAgent: DefaultUserAgent,
		Rate:      600 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   5,
	}
}

// Client talks to CSDN over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient builds a Client from cfg.
func NewClient(cfg Config) *Client {
	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// UserAgent returns the configured User-Agent.
func (c *Client) UserAgent() string { return c.cfg.UserAgent }

// GetJSON fetches a JSON endpoint with the JSON Accept header and the given
// referer, and returns its body. A body that is not JSON is the anti-bot
// challenge, so it maps to ErrWalled.
func (c *Client) GetJSON(ctx context.Context, fullURL, referer string) ([]byte, error) {
	body, err := c.get(ctx, fullURL, map[string]string{
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Referer":         referer,
	})
	if err != nil {
		return nil, err
	}
	if walled(body) {
		return nil, ErrWalled
	}
	return body, nil
}

// GetHTML fetches an HTML page with the HTML Accept header and the given
// referer, and returns its body untouched. It does not classify the body: the
// caller inspects it with isNotFoundHTML and isChallengeHTML, because a CSDN
// 404 page is a normal 200 HTML document, not a transport error.
func (c *Client) GetHTML(ctx context.Context, fullURL, referer string) ([]byte, error) {
	return c.get(ctx, fullURL, map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Referer":         referer,
	})
}

// Raw fetches an arbitrary url with the browser HTML headers and returns the
// body untouched. It is the escape hatch behind `csdn raw`: it does no parsing,
// so it shows exactly what a surface returns, challenge included. An empty body
// maps to ErrWalled.
func (c *Client) Raw(ctx context.Context, url string) ([]byte, error) {
	body, err := c.get(ctx, url, map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Referer":         Host + "/",
	})
	if err != nil {
		return nil, err
	}
	if isChallengeHTML(body) {
		return nil, ErrWalled
	}
	return body, nil
}

// walled reports whether a body returned from a JSON endpoint is the anti-bot
// challenge: an empty body, or a body whose first non-space byte is not { or [
// (a JSON endpoint that answers with an HTML challenge shell).
func walled(body []byte) bool {
	s := strings.TrimSpace(string(body))
	if s == "" {
		return true
	}
	return s[0] != '{' && s[0] != '['
}

// isNotFoundHTML reports whether an HTML page is the CSDN 404 page rather than a
// real article or profile.
func isNotFoundHTML(body []byte) bool {
	s := string(body)
	return strings.Contains(s, `"pid":"404"`) || strings.Contains(s, "<title>404")
}

// isChallengeHTML reports whether an HTML body is empty or the anti-bot edge
// challenge shell rather than a real CSDN page.
func isChallengeHTML(body []byte) bool {
	s := strings.TrimSpace(string(body))
	if s == "" {
		return true
	}
	low := strings.ToLower(s)
	if strings.Contains(low, "security check") || strings.Contains(low, "访问验证") {
		return true
	}
	// A real CSDN page is an HTML document. A bare non-HTML body on an HTML read
	// is the edge returning a 521 stub.
	if !strings.Contains(low, "<html") && !strings.Contains(low, "<!doctype") {
		return true
	}
	return false
}

func (c *Client) get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url, headers)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string, headers map[string]string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	// 521 (Cloudflare "web server is down") is how the CSDN edge fronts an
	// anti-bot challenge; it is >= 500 so it retries with the rest of the 5xx.
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	// 403 from the CSDN edge is its bot wall (a cdn_cgi_bs_bot challenge page),
	// not an authorization error. It engages per surface after repeated calls
	// from one address, so report it honestly as walled rather than retrying.
	if resp.StatusCode == http.StatusForbidden {
		return nil, false, ErrWalled
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}
