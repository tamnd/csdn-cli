package csdn_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/csdn-cli/csdn"
)

// ─── mock payloads ────────────────────────────────────────────────────────────

const mockSearchResponse = `{
  "total": 42,
  "result_vos": [
    {
      "id": "149862946",
      "articleid": "149862946",
      "title": "全面掌握<em>Golang</em>全栈开发：从入门到实战",
      "description": "本文详细介绍了<em>Golang</em>语言的发展。",
      "nickname": "罗博深",
      "author": "weixin_35950531",
      "create_time_str": "2026-02-15",
      "created_at": "2026-02-15 00:00:00",
      "view": "2649",
      "view_num": "2649",
      "comment": "1",
      "collections": "40",
      "digg": "13",
      "url_location": "https://blog.csdn.net/weixin_35950531/article/details/149862946",
      "url": "https://blog.csdn.net/weixin_35950531/article/details/149862946?utm_term=golang",
      "type": "blog"
    },
    {
      "id": "123456789",
      "articleid": "123456789",
      "title": "Docker入门教程",
      "description": "学习Docker基础知识。",
      "nickname": "",
      "author": "techwriter99",
      "create_time_str": "",
      "created_at": "2025-11-20 10:00:00",
      "view": "500",
      "view_num": "",
      "comment": "5",
      "collections": "12",
      "digg": "8",
      "url_location": "",
      "url": "https://blog.csdn.net/techwriter99/article/details/123456789",
      "type": "blog"
    }
  ]
}`

const mockHotResponse = `{
  "code": 200,
  "message": "success",
  "traceId": "abc-123",
  "data": [
    {
      "period": "2026-06-14-08",
      "hotRankScore": "24495",
      "pcHotRankScore": "2.4w",
      "nickName": "倔强的石头_",
      "userName": "2302_78391795",
      "articleTitle": "从纯文本到具身智能：魔珐星云让国产大模型 Agent 拥有 3D 具身躯壳",
      "articleDetailUrl": "https://blog.csdn.net/2302_78391795/article/details/161890576",
      "commentCount": "35",
      "favorCount": "36",
      "viewCount": "9653",
      "productId": "161890576",
      "productType": "blog"
    },
    {
      "period": "2026-06-14-08",
      "hotRankScore": "17134",
      "pcHotRankScore": "1.7w",
      "nickName": "",
      "userName": "LOVEmy134611",
      "articleTitle": "OpenCV-Python实战（28）——OpenCV计算摄影从HDR图像融合到全景拼接",
      "articleDetailUrl": "https://blog.csdn.net/LOVEmy134611/article/details/161490560",
      "commentCount": "21",
      "favorCount": "30",
      "viewCount": "300",
      "productId": "161490560",
      "productType": "blog"
    }
  ]
}`

// ─── helpers ──────────────────────────────────────────────────────────────────

func newTestClient(searchTS, hotTS *httptest.Server) *csdn.Client {
	cfg := csdn.DefaultConfig()
	cfg.BaseURL = searchTS.URL
	cfg.HotBaseURL = hotTS.URL
	cfg.Rate = 0
	return csdn.NewClient(cfg)
}

func dummyHotServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
}

// ─── Search tests ─────────────────────────────────────────────────────────────

func TestSearchSendsUserAgent(t *testing.T) {
	hotSrv := dummyHotServer(mockHotResponse)
	defer hotSrv.Close()

	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer searchSrv.Close()

	c := newTestClient(searchSrv, hotSrv)
	_, _, err := c.Search(context.Background(), csdn.SearchOpts{Query: "golang"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchParsesResults(t *testing.T) {
	hotSrv := dummyHotServer(mockHotResponse)
	defer hotSrv.Close()

	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer searchSrv.Close()

	c := newTestClient(searchSrv, hotSrv)
	articles, total, err := c.Search(context.Background(), csdn.SearchOpts{Query: "golang"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 42 {
		t.Errorf("total = %d, want 42", total)
	}
	if len(articles) != 2 {
		t.Fatalf("got %d articles, want 2", len(articles))
	}

	a := articles[0]
	if a.Rank != 1 {
		t.Errorf("rank = %d, want 1", a.Rank)
	}
	if a.Author != "罗博深" {
		t.Errorf("author = %q, want nickname", a.Author)
	}
	if a.Views != "2649" {
		t.Errorf("views = %q, want 2649", a.Views)
	}
	if a.Date != "2026-02-15" {
		t.Errorf("date = %q, want 2026-02-15", a.Date)
	}
	if a.URL != "https://blog.csdn.net/weixin_35950531/article/details/149862946" {
		t.Errorf("url = %q, want clean url_location", a.URL)
	}

	// Second article: empty nickname falls back to author; empty view_num uses view; date from created_at.
	a2 := articles[1]
	if a2.Author != "techwriter99" {
		t.Errorf("author = %q, want username fallback", a2.Author)
	}
	if a2.Views != "500" {
		t.Errorf("views = %q, want view fallback 500", a2.Views)
	}
	if a2.Date != "2025-11-20" {
		t.Errorf("date = %q, want 2025-11-20 from created_at", a2.Date)
	}
	if a2.URL != "https://blog.csdn.net/techwriter99/article/details/123456789" {
		t.Errorf("url = %q, want raw url when url_location empty", a2.URL)
	}
}

func TestSearchStripsEmTags(t *testing.T) {
	hotSrv := dummyHotServer(mockHotResponse)
	defer hotSrv.Close()

	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer searchSrv.Close()

	c := newTestClient(searchSrv, hotSrv)
	articles, _, err := c.Search(context.Background(), csdn.SearchOpts{Query: "golang"})
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) == 0 {
		t.Fatal("no articles")
	}
	if strings.Contains(articles[0].Title, "<em>") || strings.Contains(articles[0].Title, "</em>") {
		t.Errorf("title still contains em tags: %q", articles[0].Title)
	}
	want := "全面掌握Golang全栈开发：从入门到实战"
	if articles[0].Title != want {
		t.Errorf("title = %q, want %q", articles[0].Title, want)
	}
}

func TestSearchRetriesOn503(t *testing.T) {
	hotSrv := dummyHotServer(mockHotResponse)
	defer hotSrv.Close()

	var hits int
	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer searchSrv.Close()

	cfg := csdn.DefaultConfig()
	cfg.BaseURL = searchSrv.URL
	cfg.HotBaseURL = hotSrv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := csdn.NewClient(cfg)

	_, _, err := c.Search(context.Background(), csdn.SearchOpts{Query: "golang"})
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

// ─── Hot tests ────────────────────────────────────────────────────────────────

func TestHotParsesResults(t *testing.T) {
	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer searchSrv.Close()

	hotSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer hotSrv.Close()

	c := newTestClient(searchSrv, hotSrv)
	articles, err := c.Hot(context.Background(), 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 2 {
		t.Fatalf("got %d hot articles, want 2", len(articles))
	}

	a := articles[0]
	if a.Rank != 1 {
		t.Errorf("rank = %d, want 1", a.Rank)
	}
	if a.Score != "2.4w" {
		t.Errorf("score = %q, want 2.4w", a.Score)
	}
	if a.Author != "倔强的石头_" {
		t.Errorf("author = %q, want nickName", a.Author)
	}
	if a.Views != "9653" {
		t.Errorf("views = %q, want 9653", a.Views)
	}
	if a.URL != "https://blog.csdn.net/2302_78391795/article/details/161890576" {
		t.Errorf("url = %q", a.URL)
	}

	// Second: empty nickName falls back to userName.
	a2 := articles[1]
	if a2.Author != "LOVEmy134611" {
		t.Errorf("author = %q, want userName fallback", a2.Author)
	}
	if a2.Rank != 2 {
		t.Errorf("rank = %d, want 2", a2.Rank)
	}
}

func TestHotSendsUserAgent(t *testing.T) {
	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer searchSrv.Close()

	hotSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("hot request carried no User-Agent")
		}
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer hotSrv.Close()

	c := newTestClient(searchSrv, hotSrv)
	_, err := c.Hot(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
}
