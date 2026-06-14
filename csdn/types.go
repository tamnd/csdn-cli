package csdn

// Article is the record emitted for search results.
type Article struct {
	Rank        int    `json:"rank"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Views       string `json:"views"`
	Comments    string `json:"comments"`
	Collections string `json:"collections"`
	Date        string `json:"date"`
	URL         string `json:"url"`
}

// HotArticle is the record emitted for hot-rank results.
type HotArticle struct {
	Rank      int    `json:"rank"`
	Score     string `json:"score"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Views     string `json:"views"`
	Comments  string `json:"comments"`
	Favorites string `json:"favorites"`
	URL       string `json:"url"`
}
