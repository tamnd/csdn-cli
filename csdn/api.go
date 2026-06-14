package csdn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	reTitle       = regexp.MustCompile(`(?is)<h1[^>]*id="articleContentId"[^>]*>(.*?)</h1>`)
	reArticleIDJS = regexp.MustCompile(`articleId\s*=\s*(\d+)`)
	reLDJSON      = regexp.MustCompile(`(?is)<script[^>]*type="application/ld\+json"[^>]*>(.*?)</script>`)
	reKeywords    = regexp.MustCompile(`(?is)<meta[^>]*name="keywords"[^>]*content="([^"]*)"`)
	reDescription = regexp.MustCompile(`(?is)<meta[^>]*name="description"[^>]*content="([^"]*)"`)
	reReadSpan    = regexp.MustCompile(`(?is)<span[^>]*class="read-count"[^>]*>(.*?)</span>`)
	reContentDiv  = regexp.MustCompile(`(?is)<div[^>]*id="content_views"[^>]*>(.*)`)
)

// Hot fetches the hot-rank board and returns up to limit ranked entries.
func (c *Client) Hot(ctx context.Context, typ string, limit int) ([]Hot, error) {
	if limit <= 0 {
		limit = 25
	}
	q := url.Values{}
	q.Set("page", "0")
	q.Set("pageSize", strconv.Itoa(limit))
	q.Set("type", typ)
	full := blogHost + "/phoenix/web/blog/hot-rank?" + q.Encode()
	// The hot-rank endpoint answers 403 when the Referer is the bare blog host
	// but 200 from the rank page it is actually called from, so point there.
	body, err := c.GetJSON(ctx, full, blogHost+"/rank/list")
	if err != nil {
		return nil, err
	}
	var res rawHotResp
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	if res.Code != 200 {
		return nil, fmt.Errorf("hot-rank: code %d %s", res.Code, res.Message)
	}
	out := make([]Hot, 0, len(res.Data))
	for i, it := range res.Data {
		out = append(out, hotFrom(i+1, it))
		if len(out) >= limit {
			break
		}
	}
	if len(out) == 0 {
		return nil, ErrNotFound
	}
	return out, nil
}

// Search runs a blog search and returns thin, normalized hits. The typ maps to
// the `t` parameter (blog|all|ask|download|bbs).
func (c *Client) Search(ctx context.Context, query, typ string, limit int) ([]SearchHit, error) {
	if limit <= 0 {
		limit = 20
	}
	if typ == "" {
		typ = "blog"
	}
	q := url.Values{}
	q.Set("q", query)
	q.Set("t", typ)
	q.Set("p", "1")
	q.Set("s", "0")
	q.Set("tm", "0")
	q.Set("pageSize", strconv.Itoa(limit))
	full := soHost + "/api/v3/search?" + q.Encode()
	body, err := c.GetJSON(ctx, full, soHost+"/")
	if err != nil {
		return nil, err
	}
	var res rawSearchResp
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	out := make([]SearchHit, 0, len(res.ResultVOs))
	for _, h := range res.ResultVOs {
		out = append(out, hitFrom(h))
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

// ArticleByRef fetches and parses one article page.
func (c *Client) ArticleByRef(ctx context.Context, username, id string) (Article, error) {
	full := articleURL(username, id)
	body, err := c.GetHTML(ctx, full, Host+"/")
	if err != nil {
		return Article{}, err
	}
	if isNotFoundHTML(body) {
		return Article{}, ErrNotFound
	}
	if isChallengeHTML(body) {
		return Article{}, ErrWalled
	}
	p := parseArticle(string(body), username, id, full)
	if p.Title == "" && p.Content == "" {
		return Article{}, ErrNotFound
	}
	return articleFrom(p), nil
}

// parseArticle pulls the title, JSON-LD dates, tags, summary, read count, and
// body text out of an article page's HTML.
func parseArticle(html, username, id, fullURL string) articlePieces {
	p := articlePieces{ID: id, Username: username, URL: fullURL}

	if m := reTitle.FindStringSubmatch(html); m != nil {
		p.Title = stripTags(m[1])
	}
	if m := reArticleIDJS.FindStringSubmatch(html); m != nil {
		p.ID = m[1]
	}
	if m := reLDJSON.FindStringSubmatch(html); m != nil {
		var ld rawLDJSON
		if json.Unmarshal([]byte(strings.TrimSpace(m[1])), &ld) == nil {
			p.Published = ld.DatePublished
			p.Updated = ld.DateModified
			if p.Title == "" {
				p.Title = ld.Headline
			}
			if len(ld.Author) > 0 {
				p.Author = ld.Author[0].Name
			}
		}
	}
	if m := reKeywords.FindStringSubmatch(html); m != nil {
		for _, t := range strings.Split(m[1], ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				p.Tags = append(p.Tags, t)
			}
		}
	}
	if m := reDescription.FindStringSubmatch(html); m != nil {
		p.Summary = strings.TrimSpace(unescapeHTML(m[1]))
	}
	if m := reReadSpan.FindStringSubmatch(html); m != nil {
		if n := firstInt(stripTags(m[1])); n != "" {
			p.Views, _ = strconv.ParseInt(n, 10, 64)
		}
	}
	if m := reContentDiv.FindStringSubmatch(html); m != nil {
		p.Content = stripTags(closeAtDiv(m[1]))
	}
	return p
}

// closeAtDiv trims a content_views capture down to the inner HTML of the div by
// cutting at the balancing close tag. The regex grabs to end-of-document, so we
// walk nested <div> depth and stop at the matching </div>.
func closeAtDiv(s string) string {
	lower := strings.ToLower(s)
	depth := 1
	i := 0
	for i < len(lower) {
		open := strings.Index(lower[i:], "<div")
		close := strings.Index(lower[i:], "</div>")
		if close == -1 {
			return s
		}
		if open != -1 && open < close {
			depth++
			i += open + 4
			continue
		}
		depth--
		if depth == 0 {
			return s[:i+close]
		}
		i += close + 6
	}
	return s
}

// UserByName fetches and parses one profile page, then merges the tab totals.
func (c *Client) UserByName(ctx context.Context, username string) (User, error) {
	full := userURL(username)
	body, err := c.GetHTML(ctx, full, Host+"/")
	if err != nil {
		return User{}, err
	}
	if isNotFoundHTML(body) {
		return User{}, ErrNotFound
	}
	if isChallengeHTML(body) {
		return User{}, ErrWalled
	}
	st, ok := parseInitialState(string(body))
	if !ok {
		return User{}, ErrNotFound
	}
	tab := c.tabTotal(ctx, username)
	u := userFrom(st, tab, username)
	if u.Username == "" {
		return User{}, ErrNotFound
	}
	return u, nil
}

// parseInitialState extracts and decodes window.__INITIAL_STATE__ from a profile
// page.
func parseInitialState(html string) (rawInitialState, bool) {
	const marker = "window.__INITIAL_STATE__="
	idx := strings.Index(html, marker)
	if idx == -1 {
		return rawInitialState{}, false
	}
	obj, ok := extractBalancedJSON(html, idx+len(marker))
	if !ok {
		return rawInitialState{}, false
	}
	var st rawInitialState
	if err := json.Unmarshal([]byte(obj), &st); err != nil {
		return rawInitialState{}, false
	}
	return st, true
}

// tabTotal fetches the per-tab content counts for a user. A failure here is not
// fatal: the profile still reports everything the page carried.
func (c *Client) tabTotal(ctx context.Context, username string) rawTabTotalResp {
	q := url.Values{}
	q.Set("username", username)
	full := blogHost + "/community/home-api/v1/get-tab-total?" + q.Encode()
	body, err := c.GetJSON(ctx, full, userURL(username))
	if err != nil {
		return rawTabTotalResp{}
	}
	var res rawTabTotalResp
	_ = json.Unmarshal(body, &res)
	return res
}

// Posts pages a user's article list and returns up to limit articles.
func (c *Client) Posts(ctx context.Context, username string, limit int) ([]Article, error) {
	if limit <= 0 {
		limit = 40
	}
	const size = 20
	var out []Article
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("size", strconv.Itoa(size))
		q.Set("businessType", "blog")
		q.Set("username", username)
		full := blogHost + "/community/home-api/v1/get-business-list?" + q.Encode()
		body, err := c.GetJSON(ctx, full, userURL(username))
		if err != nil {
			return out, err
		}
		var res rawBusinessResp
		if err := json.Unmarshal(body, &res); err != nil {
			return out, err
		}
		if len(res.Data.List) == 0 {
			break
		}
		for _, it := range res.Data.List {
			out = append(out, postFrom(username, it))
			if len(out) >= limit {
				return out[:limit], nil
			}
		}
		if len(res.Data.List) < size {
			break
		}
	}
	return out, nil
}

// Comments pages the comment list under an article and returns up to limit
// top-level comments.
func (c *Client) Comments(ctx context.Context, articleID string, limit int) ([]Comment, error) {
	if limit <= 0 {
		limit = 50
	}
	const size = 20
	var out []Comment
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("size", strconv.Itoa(size))
		q.Set("commentId", "")
		q.Set("order", "")
		full := blogHost + "/phoenix/web/v1/comment/list/" + articleID + "?" + q.Encode()
		body, err := c.GetJSON(ctx, full, blogHost+"/")
		if err != nil {
			return out, err
		}
		var res rawCommentResp
		if err := json.Unmarshal(body, &res); err != nil {
			return out, err
		}
		if len(res.Data.List) == 0 {
			break
		}
		for _, node := range res.Data.List {
			out = append(out, commentFrom(node.Info))
			if len(out) >= limit {
				return out[:limit], nil
			}
		}
		if len(res.Data.List) < size {
			break
		}
	}
	return out, nil
}
