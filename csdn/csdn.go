// Package csdn is the library behind the csdn command: the HTTP client,
// request shaping, and the typed data models for CSDN (csdn.net).
//
// The Client hits two public CSDN endpoints:
//   - so.csdn.net/api/v3/search  — full-text search across blog posts
//   - blog.csdn.net/phoenix/web/blog/hot-rank — hourly hot article ranking
//
// No authentication is required. It sets a real User-Agent, paces requests,
// and retries transient 429/5xx errors with exponential backoff.
package csdn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// DefaultUserAgent identifies the client to CSDN.
const DefaultUserAgent = "csdn/dev (+https://github.com/tamnd/csdn-cli)"

// emTag strips <em> and </em> HTML emphasis tags from CSDN search titles.
var emTag = regexp.MustCompile(`</?em>`)

// Config holds constructor parameters for Client.
type Config struct {
	BaseURL    string        // search base: https://so.csdn.net
	HotBaseURL string        // hot rank base: https://blog.csdn.net
	UserAgent  string
	Rate       time.Duration
	Retries    int
	Timeout    time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:    "https://so.csdn.net",
		HotBaseURL: "https://blog.csdn.net",
		UserAgent:  DefaultUserAgent,
		Rate:       300 * time.Millisecond,
		Retries:    3,
		Timeout:    30 * time.Second,
	}
}

// Client talks to the CSDN public APIs.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// get performs a GET request with pacing and retry. URL must be fully formed.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

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
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// ─── Search ──────────────────────────────────────────────────────────────────

// SearchOpts controls the search request.
type SearchOpts struct {
	Query    string // required
	Type     string // "blog" (default), "ask", "download"
	Sort     int    // 0=relevance, 1=newest, 2=hottest
	Page     int    // 1-based, default 1
	PageSize int    // default 20
}

// wireSearchResult is the JSON shape returned by so.csdn.net/api/v3/search.
type wireSearchResponse struct {
	Total      int                `json:"total"`
	ResultVos  []wireSearchResult `json:"result_vos"`
}

type wireSearchResult struct {
	ID          string `json:"id"`
	ArticleID   string `json:"articleid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Digest      string `json:"digest"`
	Nickname    string `json:"nickname"`
	Author      string `json:"author"`
	CreateTimeStr string `json:"create_time_str"`
	CreatedAt   string `json:"created_at"`
	View        string `json:"view"`
	ViewNum     string `json:"view_num"`
	Comment     string `json:"comment"`
	Collections string `json:"collections"`
	Digg        string `json:"digg"`
	URLLocation string `json:"url_location"`
	URL         string `json:"url"`
	Type        string `json:"type"`
}

// Search returns blog articles matching query. Page is 1-based.
func (c *Client) Search(ctx context.Context, opts SearchOpts) ([]Article, int, error) {
	t := opts.Type
	if t == "" {
		t = "blog"
	}
	page := opts.Page
	if page < 1 {
		page = 1
	}

	params := url.Values{
		"q":    {opts.Query},
		"t":    {t},
		"sort": {strconv.Itoa(opts.Sort)},
		"p":    {strconv.Itoa(page)},
		"s":    {"0"},
		"tm":   {"0"},
		"lv":   {"-1"},
		"ft":   {"0"},
	}
	rawURL := c.cfg.BaseURL + "/api/v3/search?" + params.Encode()

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, 0, err
	}

	var resp wireSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, fmt.Errorf("decode search: %w", err)
	}

	out := make([]Article, 0, len(resp.ResultVos))
	for i, r := range resp.ResultVos {
		out = append(out, wireToArticle(r, i+1))
	}
	return out, resp.Total, nil
}

func wireToArticle(r wireSearchResult, rank int) Article {
	title := emTag.ReplaceAllString(r.Title, "")
	if title == "" {
		title = emTag.ReplaceAllString(r.Description, "")
	}

	author := r.Nickname
	if author == "" {
		author = r.Author
	}

	date := r.CreateTimeStr
	if date == "" && len(r.CreatedAt) >= 10 {
		date = r.CreatedAt[:10]
	}

	views := r.ViewNum
	if views == "" {
		views = r.View
	}

	// Prefer the clean url_location (no utm params) if available.
	u := r.URLLocation
	if u == "" {
		u = r.URL
	}

	return Article{
		Rank:        rank,
		Title:       title,
		Author:      author,
		Views:       views,
		Comments:    r.Comment,
		Collections: r.Collections,
		Date:        date,
		URL:         u,
	}
}

// ─── Hot rank ─────────────────────────────────────────────────────────────────

// wireHotResponse is the JSON shape from blog.csdn.net/phoenix/web/blog/hot-rank.
type wireHotResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    []wireHotArticle `json:"data"`
}

type wireHotArticle struct {
	Period         string `json:"period"`
	HotRankScore   string `json:"hotRankScore"`
	PcHotRankScore string `json:"pcHotRankScore"`
	NickName       string `json:"nickName"`
	UserName       string `json:"userName"`
	ArticleTitle   string `json:"articleTitle"`
	ArticleDetailURL string `json:"articleDetailUrl"`
	CommentCount   string `json:"commentCount"`
	FavorCount     string `json:"favorCount"`
	ViewCount      string `json:"viewCount"`
	ProductID      string `json:"productId"`
}

// Hot returns the current hot article ranking from CSDN.
// page is 0-based. pageSize defaults to 20 when 0.
func (c *Client) Hot(ctx context.Context, page, pageSize int) ([]HotArticle, error) {
	if pageSize <= 0 {
		pageSize = 20
	}

	params := url.Values{
		"page":     {strconv.Itoa(page)},
		"pageSize": {strconv.Itoa(pageSize)},
		"type":     {""},
	}
	rawURL := c.cfg.HotBaseURL + "/phoenix/web/blog/hot-rank?" + params.Encode()

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	var resp wireHotResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode hot-rank: %w", err)
	}
	if resp.Code != 0 && resp.Code != 200 {
		return nil, fmt.Errorf("hot-rank API error %d: %s", resp.Code, resp.Message)
	}

	out := make([]HotArticle, 0, len(resp.Data))
	offset := page * pageSize
	for i, a := range resp.Data {
		out = append(out, wireToHotArticle(a, offset+i+1))
	}
	return out, nil
}

func wireToHotArticle(a wireHotArticle, rank int) HotArticle {
	score := a.PcHotRankScore
	if score == "" {
		score = a.HotRankScore
	}
	author := a.NickName
	if author == "" {
		author = a.UserName
	}
	return HotArticle{
		Rank:      rank,
		Score:     score,
		Title:     a.ArticleTitle,
		Author:    author,
		Views:     a.ViewCount,
		Comments:  a.CommentCount,
		Favorites: a.FavorCount,
		URL:       a.ArticleDetailURL,
	}
}
