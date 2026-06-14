package csdn

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// flexInt decodes a counter CSDN sends sometimes as a JSON number and sometimes
// as a quoted string. Both shapes land here as int64; empty or null is 0.
type flexInt int64

func (f *flexInt) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		// CSDN formats counters for display, e.g. "146,206" — drop the commas
		// before parsing.
		s = strings.ReplaceAll(s, ",", "")
		if s == "" {
			*f = 0
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			*f = 0
			return nil
		}
		*f = flexInt(n)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	v, err := n.Int64()
	if err != nil {
		*f = 0
		return nil
	}
	*f = flexInt(v)
	return nil
}

// --- hot-rank board ---

type rawHotResp struct {
	Code    flexInt      `json:"code"`
	Message string       `json:"message"`
	Data    []rawHotItem `json:"data"`
}

type rawHotItem struct {
	HotRankScore     flexInt `json:"hotRankScore"`
	PcHotRankScore   flexInt `json:"pcHotRankScore"`
	NickName         string  `json:"nickName"`
	UserName         string  `json:"userName"`
	ArticleTitle     string  `json:"articleTitle"`
	ArticleDetailURL string  `json:"articleDetailUrl"`
	CommentCount     flexInt `json:"commentCount"`
	FavorCount       flexInt `json:"favorCount"`
	ViewCount        flexInt `json:"viewCount"`
}

func hotFrom(rank int, it rawHotItem) Hot {
	score := int64(it.PcHotRankScore)
	if score == 0 {
		score = int64(it.HotRankScore)
	}
	return Hot{
		Rank:     rank,
		Title:    it.ArticleTitle,
		Author:   it.NickName,
		Username: it.UserName,
		Score:    score,
		Views:    int64(it.ViewCount),
		Comments: int64(it.CommentCount),
		Favors:   int64(it.FavorCount),
		URL:      it.ArticleDetailURL,
	}
}

// --- search (so.csdn.net/api/v3/search) ---

type rawSearchResp struct {
	ResultVOs []rawSearchHit `json:"result_vos"`
	Total     flexInt        `json:"total"`
}

type rawSearchHit struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Nickname    string  `json:"nickname"`
	Username    string  `json:"username"`
	ViewNum     flexInt `json:"view_num"`
	View        flexInt `json:"view"`
	Digg        flexInt `json:"digg"`
	Comment     flexInt `json:"comment"`
	Type        string  `json:"type"`
	ArticleID   string  `json:"articleid"`
	Description string  `json:"description"`
}

func hitFrom(h rawSearchHit) SearchHit {
	views := int64(h.ViewNum)
	if views == 0 {
		views = int64(h.View)
	}
	typ := h.Type
	if typ == "" {
		typ = "blog"
	}
	return SearchHit{
		Type:     typ,
		ID:       h.ArticleID,
		Title:    stripEm(h.Title),
		Author:   h.Nickname,
		Username: h.Username,
		Summary:  stripEm(h.Description),
		Views:    views,
		Likes:    int64(h.Digg),
		Comments: int64(h.Comment),
		URL:      h.URL,
	}
}

// --- article (parsed from HTML + JSON-LD) ---

// articlePieces is the bundle the HTML parser extracts before it is folded into
// an Article. It keeps the parse and the assembly separate so each is testable.
type articlePieces struct {
	ID        string
	Username  string
	Title     string
	Summary   string
	Content   string
	Tags      []string
	Published string
	Updated   string
	Author    string
	Views     int64
	URL       string
}

func articleFrom(p articlePieces) Article {
	return Article{
		ID:        p.ID,
		Title:     p.Title,
		Author:    p.Author,
		Username:  p.Username,
		Summary:   p.Summary,
		Content:   p.Content,
		Tags:      p.Tags,
		Published: p.Published,
		Updated:   p.Updated,
		Views:     p.Views,
		URL:       p.URL,
	}
}

// rawLDJSON is the application/ld+json block CSDN embeds in an article page.
type rawLDJSON struct {
	Headline      string `json:"headline"`
	DatePublished string `json:"datePublished"`
	DateModified  string `json:"dateModified"`
	Author        []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"author"`
}

// --- user profile (window.__INITIAL_STATE__ + get-tab-total) ---

type rawInitialState struct {
	PageData struct {
		Data struct {
			BaseInfo struct {
				UserModule struct {
					Avatar           string     `json:"avatar"`
					Username         string     `json:"username"`
					Nickname         string     `json:"nickname"`
					VIP              bool       `json:"vip"`
					BlogURL          string     `json:"blogUrl"`
					Level            flexInt    `json:"level"`
					Introduction     string     `json:"introduction"`
					School           string     `json:"school"`
					Company          string     `json:"company"`
					RegistrationTime string     `json:"registrationTime"`
					CodeAge          rawCodeAge `json:"codeAge"`
					Gender           flexInt    `json:"gender"`
					Region           rawRegion  `json:"region"`
				} `json:"userModule"`
				AchievementModule struct {
					OriginalCount      flexInt      `json:"originalCount"`
					Rank               flexInt      `json:"rank"`
					FansCount          flexInt      `json:"fansCount"`
					FollowCount        flexInt      `json:"followCount"`
					LoyalFansCount     flexInt      `json:"loyalFansCount"`
					WholeSiteViewCount rawViewCount `json:"wholeSiteViewCount"`
				} `json:"achievementModule"`
			} `json:"baseInfo"`
		} `json:"data"`
	} `json:"pageData"`
}

// rawCodeAge is the codeAge block, which CSDN sends as an object carrying a
// display string such as "码龄15年".
type rawCodeAge struct {
	Desc string `json:"desc"`
}

// rawRegion is the IP-region block, whose region field reads like
// "IP 属地：河南省".
type rawRegion struct {
	Region string `json:"region"`
}

// rawViewCount is the whole-site view counter, an object whose total holds the
// number (formatted with commas, decoded by flexInt).
type rawViewCount struct {
	Total flexInt `json:"total"`
}

// genderText maps the CSDN gender code to a label. 1 is male, 2 is female;
// anything else is unknown and reported as empty.
func genderText(code int64) string {
	switch code {
	case 1:
		return "male"
	case 2:
		return "female"
	default:
		return ""
	}
}

type rawTabTotalResp struct {
	Code flexInt `json:"code"`
	Data struct {
		Blog       flexInt `json:"blog"`
		BlogColumn flexInt `json:"blogColumn"`
		Download   flexInt `json:"download"`
		Bbs        flexInt `json:"bbs"`
		Ask        flexInt `json:"ask"`
	} `json:"data"`
}

func userFrom(st rawInitialState, tab rawTabTotalResp, username string) User {
	um := st.PageData.Data.BaseInfo.UserModule
	am := st.PageData.Data.BaseInfo.AchievementModule
	name := um.Username
	if name == "" {
		name = username
	}
	level := ""
	if um.Level != 0 {
		level = strconv.FormatInt(int64(um.Level), 10)
	}
	region := strings.TrimSpace(um.Region.Region)
	region = strings.TrimPrefix(region, "IP 属地：")
	return User{
		Username:      name,
		Nickname:      um.Nickname,
		Intro:         um.Introduction,
		Level:         level,
		CodeAge:       um.CodeAge.Desc,
		Region:        region,
		School:        um.School,
		Company:       um.Company,
		Registered:    um.RegistrationTime,
		Gender:        genderText(int64(um.Gender)),
		VIP:           um.VIP,
		OriginalCount: int64(am.OriginalCount),
		Rank:          int64(am.Rank),
		Fans:          int64(am.FansCount),
		Follows:       int64(am.FollowCount),
		LoyalFans:     int64(am.LoyalFansCount),
		TotalViews:    int64(am.WholeSiteViewCount.Total),
		BlogCount:     int64(tab.Data.Blog),
		ColumnCount:   int64(tab.Data.BlogColumn),
		DownloadCount: int64(tab.Data.Download),
		AskCount:      int64(tab.Data.Ask),
		Avatar:        um.Avatar,
		URL:           userURL(name),
	}
}

// --- posts (get-business-list) ---

type rawBusinessResp struct {
	Code flexInt `json:"code"`
	Data struct {
		List []rawBusinessItem `json:"list"`
	} `json:"data"`
}

type rawBusinessItem struct {
	ArticleID    flexInt `json:"articleId"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	URL          string  `json:"url"`
	Type         flexInt `json:"type"`
	Top          bool    `json:"top"`
	ViewCount    flexInt `json:"viewCount"`
	CommentCount flexInt `json:"commentCount"`
	DiggCount    flexInt `json:"diggCount"`
	CollectCount flexInt `json:"collectCount"`
	PostTime     string  `json:"postTime"`
	FormatTime   string  `json:"formatTime"`
	Tags         rawTags `json:"tags"`
}

// rawTags captures the article tag list, which CSDN sends as an array of objects
// ({"name":"go"}) on some surfaces and an array of strings ("go") on others.
type rawTags []string

func (t *rawTags) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*t = nil
		return nil
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		*t = nil
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		r = bytes.TrimSpace(r)
		if len(r) == 0 {
			continue
		}
		if r[0] == '"' {
			var s string
			if json.Unmarshal(r, &s) == nil && s != "" {
				out = append(out, s)
			}
			continue
		}
		var obj struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(r, &obj) == nil && obj.Name != "" {
			out = append(out, obj.Name)
		}
	}
	*t = out
	return nil
}

func postFrom(username string, it rawBusinessItem) Article {
	id := strconv.FormatInt(int64(it.ArticleID), 10)
	url := it.URL
	if url == "" {
		url = articleURL(username, id)
	}
	published := it.FormatTime
	if published == "" {
		published = it.PostTime
	}
	return Article{
		ID:        id,
		Title:     it.Title,
		Username:  username,
		Summary:   strings.TrimSpace(it.Description),
		Tags:      []string(it.Tags),
		Published: published,
		Views:     int64(it.ViewCount),
		Likes:     int64(it.DiggCount),
		Collects:  int64(it.CollectCount),
		Comments:  int64(it.CommentCount),
		URL:       url,
	}
}

// --- comments ---

type rawCommentResp struct {
	Code flexInt `json:"code"`
	Data struct {
		Count     flexInt          `json:"count"`
		PageCount flexInt          `json:"pageCount"`
		List      []rawCommentNode `json:"list"`
	} `json:"data"`
}

type rawCommentNode struct {
	Info  rawCommentInfo   `json:"info"`
	Reply []rawCommentNode `json:"reply"`
}

type rawCommentInfo struct {
	CommentID  flexInt `json:"commentId"`
	ArticleID  flexInt `json:"articleId"`
	ParentID   flexInt `json:"parentId"`
	PostTime   string  `json:"postTime"`
	Content    string  `json:"content"`
	UserName   string  `json:"userName"`
	NickName   string  `json:"nickName"`
	Digg       flexInt `json:"digg"`
	Region     string  `json:"region"`
	DateFormat string  `json:"dateFormat"`
}

func commentFrom(in rawCommentInfo) Comment {
	post := in.DateFormat
	if post == "" {
		post = in.PostTime
	}
	parent := ""
	if in.ParentID != 0 {
		parent = strconv.FormatInt(int64(in.ParentID), 10)
	}
	username := in.UserName
	return Comment{
		ID:        strconv.FormatInt(int64(in.CommentID), 10),
		ArticleID: strconv.FormatInt(int64(in.ArticleID), 10),
		Text:      strings.TrimSpace(in.Content),
		Author:    username,
		Nickname:  in.NickName,
		ParentID:  parent,
		PostTime:  post,
		Likes:     int64(in.Digg),
		Region:    in.Region,
		URL:       userURL(username),
	}
}
