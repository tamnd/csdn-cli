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
	// hotRankScore is the exact integer; pcHotRankScore is CSDN's display form
	// ("2.4w"), which is lossy and does not parse, so hotFrom must take the exact
	// one. productId/productType carry the article id and kind.
	const in = `{
		"hotRankScore":"24056","pcHotRankScore":"2.4w","nickName":"Alice",
		"userName":"alice","articleTitle":"A Go primer",
		"articleDetailUrl":"https://blog.csdn.net/alice/article/details/1",
		"commentCount":"12","favorCount":"34","viewCount":"5600",
		"avatarUrl":"https://a.png","picList":["https://cover.png","https://b.png"],
		"productId":"1","productType":"blog"
	}`
	var raw rawHotItem
	if err := json.Unmarshal([]byte(in), &raw); err != nil {
		t.Fatal(err)
	}
	h := hotFrom(1, raw)
	if h.Rank != 1 || h.Title != "A Go primer" || h.Author != "Alice" || h.Username != "alice" {
		t.Errorf("hotFrom basics = %+v", h)
	}
	if h.ID != "1" || h.Type != "blog" {
		t.Errorf("hotFrom id/type = id=%q type=%q", h.ID, h.Type)
	}
	if h.Score != 24056 || h.Views != 5600 || h.Comments != 12 || h.Favors != 34 {
		t.Errorf("hotFrom counts = %+v", h)
	}
	if h.URL != "https://blog.csdn.net/alice/article/details/1" {
		t.Errorf("hotFrom url = %q", h.URL)
	}
	if h.Cover != "https://cover.png" || h.Avatar != "https://a.png" {
		t.Errorf("hotFrom media = cover=%q avatar=%q", h.Cover, h.Avatar)
	}
}

func TestHotFromScoreFallback(t *testing.T) {
	// When hotRankScore is absent, fall back to pcHotRankScore.
	const in = `{"pcHotRankScore":"900","articleTitle":"T","productId":"2"}`
	var raw rawHotItem
	if err := json.Unmarshal([]byte(in), &raw); err != nil {
		t.Fatal(err)
	}
	if h := hotFrom(1, raw); h.Score != 900 {
		t.Errorf("score fallback = %d", h.Score)
	}
}

func TestHitFrom(t *testing.T) {
	const in = `{
		"title":"a <em>Go</em> guide",
		"url":"https://blog.csdn.net/bob/article/details/2?ops_request_misc=x&utm_term=go",
		"nickname":"Bob","username":"bob","view_num":1200,"digg":"7","comment":"3",
		"collections":"40","create_time_str":"2024-08-09","tags":["go","web"],
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
	if h.Views != 1200 || h.Likes != 7 || h.Comments != 3 || h.Collects != 40 {
		t.Errorf("hitFrom counts = %+v", h)
	}
	if h.Published != "2024-08-09" || len(h.Tags) != 2 || h.Tags[0] != "go" {
		t.Errorf("hitFrom published/tags = %+v", h)
	}
	// The tracking query tail must be stripped to the canonical url.
	if h.URL != "https://blog.csdn.net/bob/article/details/2" {
		t.Errorf("hitFrom url not cleaned: %q", h.URL)
	}
}

func TestPostFrom(t *testing.T) {
	// postTime is the exact timestamp; formatTime is the fuzzy UI label. postFrom
	// must keep the exact one and capture the cover from picList[0].
	const in = `{
		"articleId":161490560,"title":"OpenCV in practice","description":"  a body  ",
		"url":"https://blog.csdn.net/u/article/details/161490560",
		"top":true,"viewCount":679,"commentCount":35,"diggCount":45,"collectCount":45,
		"postTime":"2026-06-13 16:23:43","formatTime":"前天 16:23",
		"tags":["opencv","python"],"picList":["https://cover.png","https://b.png"]
	}`
	var raw rawBusinessItem
	if err := json.Unmarshal([]byte(in), &raw); err != nil {
		t.Fatal(err)
	}
	a := postFrom("u", raw)
	if a.ID != "161490560" || a.Title != "OpenCV in practice" || a.Summary != "a body" {
		t.Errorf("postFrom basics = %+v", a)
	}
	if a.Views != 679 || a.Comments != 35 || a.Likes != 45 || a.Collects != 45 {
		t.Errorf("postFrom counts = %+v", a)
	}
	if !a.Pinned || a.Published != "2026-06-13 16:23:43" {
		t.Errorf("postFrom pinned/published = pinned=%v published=%q", a.Pinned, a.Published)
	}
	if a.Cover != "https://cover.png" || len(a.Tags) != 2 || a.Tags[0] != "opencv" {
		t.Errorf("postFrom cover/tags = cover=%q tags=%v", a.Cover, a.Tags)
	}
}

func TestPostFromPublishedFallback(t *testing.T) {
	// When postTime is absent, fall back to the formatTime label.
	const in = `{"articleId":5,"formatTime":"前天 16:23"}`
	var raw rawBusinessItem
	if err := json.Unmarshal([]byte(in), &raw); err != nil {
		t.Fatal(err)
	}
	if a := postFrom("u", raw); a.Published != "前天 16:23" {
		t.Errorf("published fallback = %q", a.Published)
	}
}

func TestHitFromPublishedFallback(t *testing.T) {
	// When create_time_str is absent, fall back to created_at.
	const in = `{"articleid":"3","created_at":"2024-01-02 03:04:05","type":"blog"}`
	var raw rawSearchHit
	if err := json.Unmarshal([]byte(in), &raw); err != nil {
		t.Fatal(err)
	}
	if h := hitFrom(raw); h.Published != "2024-01-02 03:04:05" {
		t.Errorf("published fallback = %q", h.Published)
	}
}

func TestCommentFrom(t *testing.T) {
	const in = `{
		"commentId":1001,"articleId":2002,"parentId":0,"postTime":"2024-01-02",
		"content":"  nice post  ","userName":"carol","nickName":"Carol","digg":5,
		"avatar":"https://c.png","region":"IP：Beijing","dateFormat":"1 day ago"
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
	if c.Likes != 5 || c.Region != "Beijing" || c.PostTime != "1 day ago" || c.Avatar != "https://c.png" {
		t.Errorf("commentFrom meta = %+v", c)
	}
	if c.URL != userURL("carol") {
		t.Errorf("commentFrom url = %q", c.URL)
	}
}

func TestCommentNodeRepliesUnderSub(t *testing.T) {
	// CSDN nests replies under "sub" as flat comment objects, and a reply
	// carries its parentId and parentNickName.
	const in = `{
		"info":{"commentId":1,"articleId":9,"content":"top","userName":"a","nickName":"A"},
		"sub":[
			{"commentId":2,"articleId":9,"parentId":1,"parentNickName":"A","content":"reply","userName":"b","nickName":"B","region":"IP：浙江省"}
		]
	}`
	var node rawCommentNode
	if err := json.Unmarshal([]byte(in), &node); err != nil {
		t.Fatal(err)
	}
	if len(node.Sub) != 1 {
		t.Fatalf("expected 1 reply under sub, got %d", len(node.Sub))
	}
	r := commentFrom(node.Sub[0])
	if r.ID != "2" || r.ParentID != "1" || r.ParentNick != "A" || r.Author != "b" {
		t.Errorf("reply = %+v", r)
	}
	if r.Text != "reply" || r.Region != "浙江省" {
		t.Errorf("reply text/region = %+v", r)
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
<span class="read-count" id="blog-digg-num"> 42 </span>
<a class="get-collection " data-num="41" id="get-collection">收藏</a>
<span class="unlogin-comment-tit">30</span>
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
	if p.Likes != 42 || p.Collects != 41 || p.Comments != 30 {
		t.Errorf("counters = likes=%d collects=%d comments=%d", p.Likes, p.Collects, p.Comments)
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
	if a.Likes != 42 || a.Collects != 41 || a.Comments != 30 {
		t.Errorf("articleFrom counters = %+v", a)
	}
}
