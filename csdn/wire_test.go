package csdn

import (
	"encoding/json"
	"testing"
)

func TestFlexInt(t *testing.T) {
	var f flexInt
	if err := f.UnmarshalJSON([]byte(`"123"`)); err != nil || f != 123 {
		t.Errorf("string number: %v %d", err, f)
	}
	if err := f.UnmarshalJSON([]byte(`456`)); err != nil || f != 456 {
		t.Errorf("bare number: %v %d", err, f)
	}
	if err := f.UnmarshalJSON([]byte(`null`)); err != nil || f != 0 {
		t.Errorf("null: %v %d", err, f)
	}
	if err := f.UnmarshalJSON([]byte(`""`)); err != nil || f != 0 {
		t.Errorf("empty string: %v %d", err, f)
	}
	// CSDN sends display-formatted counters such as "146,206".
	if err := f.UnmarshalJSON([]byte(`"146,206"`)); err != nil || f != 146206 {
		t.Errorf("comma string: %v %d", err, f)
	}
}

func TestUserFromProfileShape(t *testing.T) {
	// The profile JSON CSDN embeds in window.__INITIAL_STATE__ sends counters as
	// comma-formatted strings, the gender as a number, and codeAge/region/
	// wholeSiteViewCount as objects. userFrom must decode all of these.
	const in = `{"pageData":{"data":{"baseInfo":{
		"userModule":{
			"avatar":"https://a.png","username":"LOVEmy134611","nickname":"盼小辉丶",
			"vip":true,"level":10,"introduction":"hello","school":null,"company":null,
			"registrationTime":"2011-08-18","gender":1,
			"codeAge":{"desc":"码龄15年"},
			"region":{"ip":"1.2.3.4","region":"IP 属地：河南省","msg":"x"}
		},
		"achievementModule":{
			"originalCount":"848","rank":"74","fansCount":"146,206","followCount":"748",
			"loyalFansCount":"458",
			"wholeSiteViewCount":{"total":"9,757,241","detail":{"blog":"1"}}
		}
	}}}}`
	var st rawInitialState
	if err := json.Unmarshal([]byte(in), &st); err != nil {
		t.Fatal(err)
	}
	tab := rawTabTotalResp{}
	tab.Data.Blog = 848
	tab.Data.BlogColumn = 36
	u := userFrom(st, tab, "LOVEmy134611")
	if u.Username != "LOVEmy134611" || u.Nickname != "盼小辉丶" || u.Level != "10" {
		t.Errorf("user basics = %+v", u)
	}
	if u.CodeAge != "码龄15年" || u.Region != "河南省" || u.Gender != "male" {
		t.Errorf("user object fields = codeAge=%q region=%q gender=%q", u.CodeAge, u.Region, u.Gender)
	}
	if u.Fans != 146206 || u.TotalViews != 9757241 || u.OriginalCount != 848 || u.Rank != 74 {
		t.Errorf("user counters = %+v", u)
	}
	if u.BlogCount != 848 || u.ColumnCount != 36 {
		t.Errorf("user tab totals = %+v", u)
	}
}

func TestHotFrom(t *testing.T) {
	const in = `{
		"hotRankScore":"500","pcHotRankScore":"900","nickName":"Alice",
		"userName":"alice","articleTitle":"A Go primer",
		"articleDetailUrl":"https://blog.csdn.net/alice/article/details/1",
		"commentCount":"12","favorCount":"34","viewCount":"5600"
	}`
	var raw rawHotItem
	if err := json.Unmarshal([]byte(in), &raw); err != nil {
		t.Fatal(err)
	}
	h := hotFrom(1, raw)
	if h.Rank != 1 || h.Title != "A Go primer" || h.Author != "Alice" || h.Username != "alice" {
		t.Errorf("hotFrom basics = %+v", h)
	}
	if h.Score != 900 || h.Views != 5600 || h.Comments != 12 || h.Favors != 34 {
		t.Errorf("hotFrom counts = %+v", h)
	}
	if h.URL != "https://blog.csdn.net/alice/article/details/1" {
		t.Errorf("hotFrom url = %q", h.URL)
	}
}

func TestHitFrom(t *testing.T) {
	const in = `{
		"title":"a <em>Go</em> guide","url":"https://blog.csdn.net/bob/article/details/2",
		"nickname":"Bob","username":"bob","view_num":1200,"digg":"7","comment":"3",
		"type":"blog","articleid":"2","description":"all about <em>Go</em>"
	}`
	var raw rawSearchHit
	if err := json.Unmarshal([]byte(in), &raw); err != nil {
		t.Fatal(err)
	}
	h := hitFrom(raw)
	if h.Title != "a Go guide" || h.Summary != "all about Go" {
		t.Errorf("hitFrom did not strip <em>: %+v", h)
	}
	if h.Type != "blog" || h.ID != "2" || h.Username != "bob" || h.Author != "Bob" {
		t.Errorf("hitFrom basics = %+v", h)
	}
	if h.Views != 1200 || h.Likes != 7 || h.Comments != 3 {
		t.Errorf("hitFrom counts = %+v", h)
	}
}

func TestCommentFrom(t *testing.T) {
	const in = `{
		"commentId":1001,"articleId":2002,"parentId":0,"postTime":"2024-01-02",
		"content":"  nice post  ","userName":"carol","nickName":"Carol","digg":5,
		"region":"Beijing","dateFormat":"1 day ago"
	}`
	var raw rawCommentInfo
	if err := json.Unmarshal([]byte(in), &raw); err != nil {
		t.Fatal(err)
	}
	c := commentFrom(raw)
	if c.ID != "1001" || c.ArticleID != "2002" || c.ParentID != "" {
		t.Errorf("commentFrom ids = %+v", c)
	}
	if c.Text != "nice post" || c.Author != "carol" || c.Nickname != "Carol" {
		t.Errorf("commentFrom text/author = %+v", c)
	}
	if c.Likes != 5 || c.Region != "Beijing" || c.PostTime != "1 day ago" {
		t.Errorf("commentFrom meta = %+v", c)
	}
	if c.URL != userURL("carol") {
		t.Errorf("commentFrom url = %q", c.URL)
	}
}

func TestRawTags(t *testing.T) {
	var objs rawTags
	if err := json.Unmarshal([]byte(`[{"name":"go"},{"name":"web"}]`), &objs); err != nil {
		t.Fatal(err)
	}
	if len(objs) != 2 || objs[0] != "go" || objs[1] != "web" {
		t.Errorf("object tags = %v", objs)
	}
	var strs rawTags
	if err := json.Unmarshal([]byte(`["go","web"]`), &strs); err != nil {
		t.Fatal(err)
	}
	if len(strs) != 2 || strs[0] != "go" {
		t.Errorf("string tags = %v", strs)
	}
}

func TestParseArticleHTML(t *testing.T) {
	const html = `<html><head>
<meta name="keywords" content="go, web, http">
<meta name="description" content="a short summary">
<script type="application/ld+json">{"headline":"H","datePublished":"2024-01-01","dateModified":"2024-02-01","author":[{"name":"Alice","url":"https://blog.csdn.net/alice"}]}</script>
</head><body>
<h1 class="title-article" id="articleContentId">My Title</h1>
<script>var articleId = 123456;</script>
<span class="read-count">587 阅读</span>
<div id="content_views" class="markdown_views"><p>Hello <b>world</b></p><div>nested</div></div>
<div>after</div>
</body></html>`
	p := parseArticle(html, "alice", "123456", "https://blog.csdn.net/alice/article/details/123456")
	if p.Title != "My Title" {
		t.Errorf("title = %q", p.Title)
	}
	if p.ID != "123456" {
		t.Errorf("id = %q", p.ID)
	}
	if p.Published != "2024-01-01" || p.Updated != "2024-02-01" || p.Author != "Alice" {
		t.Errorf("ld-json = %+v", p)
	}
	if len(p.Tags) != 3 || p.Tags[0] != "go" {
		t.Errorf("tags = %v", p.Tags)
	}
	if p.Summary != "a short summary" {
		t.Errorf("summary = %q", p.Summary)
	}
	if p.Views != 587 {
		t.Errorf("views = %d", p.Views)
	}
	if p.Content != "Hello worldnested" {
		t.Errorf("content = %q", p.Content)
	}
	a := articleFrom(p)
	if a.ID != "123456" || a.Title != "My Title" || a.Views != 587 {
		t.Errorf("articleFrom = %+v", a)
	}
}
