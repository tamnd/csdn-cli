package csdn

// The clean records. Each is a flat struct whose json tags set the wire shape
// and the default column order, and whose url field feeds `-o url`.
//
// The kit tags make a record addressable when a host such as ant drives the
// package: kit:"id" is the key a resource URI and the --db record store use, and
// kit:"body" is the long text `ant cat` prints. A table:",truncate" tag only
// shortens the on-screen table on a terminal; json, csv, and tsv carry the full
// value.

// Hot is one hot-rank board entry, ranked 1-based.
type Hot struct {
	Rank     int    `json:"rank"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title" kit:"id" table:"title,truncate"`
	Author   string `json:"author"`
	Username string `json:"username"`
	Score    int64  `json:"score"`
	Views    int64  `json:"views"`
	Comments int64  `json:"comments"`
	Favors   int64  `json:"favors"`
	URL      string `json:"url"`
	Cover    string `json:"cover"`
	Avatar   string `json:"avatar"`
}

// Article is a single blog post with its body and counters.
type Article struct {
	ID        string   `json:"id" kit:"id"`
	Title     string   `json:"title" table:"title,truncate"`
	Author    string   `json:"author"`
	Username  string   `json:"username"`
	Summary   string   `json:"summary" table:"summary,truncate"`
	Content   string   `json:"content" kit:"body"`
	Tags      []string `json:"tags"`
	Published string   `json:"published"`
	Updated   string   `json:"updated"`
	Views     int64    `json:"views"`
	Likes     int64    `json:"likes"`
	Collects  int64    `json:"collects"`
	Comments  int64    `json:"comments"`
	Pinned    bool     `json:"pinned"`
	Cover     string   `json:"cover"`
	URL       string   `json:"url"`
}

// User is a public CSDN profile, addressed by its username.
type User struct {
	Username      string `json:"username" kit:"id"`
	Nickname      string `json:"nickname"`
	Intro         string `json:"intro" kit:"body" table:"intro,truncate"`
	Level         string `json:"level"`
	CodeAge       string `json:"code_age"`
	Region        string `json:"region"`
	School        string `json:"school"`
	Company       string `json:"company"`
	Registered    string `json:"registered"`
	Gender        string `json:"gender"`
	VIP           bool   `json:"vip"`
	OriginalCount int64  `json:"original_count"`
	Rank          int64  `json:"rank"`
	Fans          int64  `json:"fans"`
	Follows       int64  `json:"follows"`
	LoyalFans     int64  `json:"loyal_fans"`
	TotalViews    int64  `json:"total_views"`
	BlogCount     int64  `json:"blog_count"`
	ColumnCount   int64  `json:"column_count"`
	DownloadCount int64  `json:"download_count"`
	AskCount      int64  `json:"ask_count"`
	Avatar        string `json:"avatar"`
	URL           string `json:"url"`
}

// Comment is one comment under an article, with its parent id for replies.
type Comment struct {
	ID         string `json:"id" kit:"id"`
	ArticleID  string `json:"article_id"`
	Text       string `json:"text" kit:"body" table:"text,truncate"`
	Author     string `json:"author"`
	Nickname   string `json:"nickname"`
	ParentID   string `json:"parent_id"`
	ParentNick string `json:"parent_nick"`
	PostTime   string `json:"post_time"`
	Likes      int64  `json:"likes"`
	Region     string `json:"region"`
	Avatar     string `json:"avatar"`
	URL        string `json:"url"`
}

// SearchHit is a thin, normalized search result row.
type SearchHit struct {
	Type      string   `json:"type"`
	ID        string   `json:"id" kit:"id"`
	Title     string   `json:"title" table:"title,truncate"`
	Author    string   `json:"author"`
	Username  string   `json:"username"`
	Summary   string   `json:"summary" table:"summary,truncate"`
	Published string   `json:"published"`
	Tags      []string `json:"tags"`
	Views     int64    `json:"views"`
	Likes     int64    `json:"likes"`
	Collects  int64    `json:"collects"`
	Comments  int64    `json:"comments"`
	URL       string   `json:"url"`
}
