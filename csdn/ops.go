package csdn

import (
	"context"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// ops.go declares the record-stream commands as kit operations. Each one is
// declared once and exposed as a CLI subcommand, an HTTP route, and an MCP tool;
// kit renders the records in every format, applies --limit, tees them into --db,
// and (for the URI-tagged ops) lets a host dereference them by csdn:// URI.
//
// The single-record reads (article, user) carry URI metadata so a host can
// dereference csdn://article/<user/id> and csdn://user/<username>. The list
// reads (posts, comments) are members of a parent resource, so they answer `ant
// ls`. The raw byte dump and the version banner do not fit the emit-records
// shape and stay as escape-hatch commands in the cli package.
func registerOps(app *kit.App) {
	registerHot(app)
	registerSearch(app)
	registerArticle(app)
	registerUser(app)
	registerPosts(app)
	registerComments(app)
}

// effectiveLimit picks the per-command default when -n is unset.
func effectiveLimit(n, def int) int {
	if n > 0 {
		return n
	}
	return def
}

// --- hot ---

type hotIn struct {
	Type  string   `kit:"flag" help:"rank board type"`
	Limit int      `kit:"flag,inherit" help:"max records"`
	Sess  *Session `kit:"inject"`
}

func registerHot(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "hot", Group: "read",
		Summary: "Hot-rank board (CSDN热榜)",
	}, func(ctx context.Context, in hotIn, emit func(Hot) error) error {
		limit := effectiveLimit(in.Limit, 25)
		in.Sess.Progressf("fetching hot-rank board")
		hot, err := in.Sess.Client.Hot(ctx, in.Type, limit)
		if err != nil {
			return MapErr(err)
		}
		return emitAll(hot, emit)
	})
}

// --- search ---

type searchIn struct {
	Query []string `kit:"arg,variadic" help:"search terms"`
	Type  string   `kit:"flag" help:"search type: blog|all|ask|download|bbs"`
	Limit int      `kit:"flag,inherit" help:"max records"`
	Sess  *Session `kit:"inject"`
}

func registerSearch(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "search", Group: "read",
		Summary: "Search CSDN blog articles",
		Args:    []kit.Arg{{Name: "query", Help: "search terms", Variadic: true}},
	}, func(ctx context.Context, in searchIn, emit func(SearchHit) error) error {
		q := strings.Join(in.Query, " ")
		typ := in.Type
		if typ == "" {
			typ = "blog"
		}
		in.Sess.Progressf("searching for %q", q)
		hits, err := in.Sess.Client.Search(ctx, q, typ, effectiveLimit(in.Limit, 20))
		if err != nil {
			return MapErr(err)
		}
		return emitAll(hits, emit)
	})
}

// --- article ---

type articleIn struct {
	Ref  string   `kit:"arg" help:"article url or username/id"`
	Sess *Session `kit:"inject"`
}

func registerArticle(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "article", Group: "read", Single: true,
		Summary:  "One article with its body",
		URIType:  "article",
		Resolver: true,
		Args:     []kit.Arg{{Name: "url-or-ref", Help: "article url or username/id"}},
	}, func(ctx context.Context, in articleIn, emit func(Article) error) error {
		username, id, err := ParseArticleRef(in.Ref)
		if err != nil {
			return errs.Usage("%s", err.Error())
		}
		in.Sess.Progressf("fetching article %s/%s", username, id)
		a, err := in.Sess.Client.ArticleByRef(ctx, username, id)
		if err != nil {
			return MapErr(err)
		}
		return emit(a)
	})
}

// --- user ---

type userIn struct {
	Ref  string   `kit:"arg" help:"username or profile url"`
	Sess *Session `kit:"inject"`
}

func registerUser(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "user", Group: "read", Single: true,
		Summary:  "Profile record for a CSDN user",
		URIType:  "user",
		Resolver: true,
		Args:     []kit.Arg{{Name: "username-or-url", Help: "username or profile url"}},
	}, func(ctx context.Context, in userIn, emit func(User) error) error {
		username, err := ParseUsername(in.Ref)
		if err != nil {
			return errs.Usage("%s", err.Error())
		}
		in.Sess.Progressf("fetching profile %s", username)
		u, err := in.Sess.Client.UserByName(ctx, username)
		if err != nil {
			return MapErr(err)
		}
		return emit(u)
	})
}

// --- posts (members of a user) ---

type postsIn struct {
	Ref   string   `kit:"arg" help:"username or profile url"`
	Limit int      `kit:"flag,inherit" help:"max records"`
	Sess  *Session `kit:"inject"`
}

func registerPosts(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "posts", Group: "read", List: true,
		Summary: "A user's published articles",
		URIType: "user",
		Args:    []kit.Arg{{Name: "username-or-url", Help: "username or profile url"}},
	}, func(ctx context.Context, in postsIn, emit func(Article) error) error {
		username, err := ParseUsername(in.Ref)
		if err != nil {
			return errs.Usage("%s", err.Error())
		}
		in.Sess.Progressf("fetching posts for %s", username)
		arts, err := in.Sess.Client.Posts(ctx, username, effectiveLimit(in.Limit, 40))
		if err != nil {
			return MapErr(err)
		}
		return emitAll(arts, emit)
	})
}

// --- comments (members of an article) ---

type commentsIn struct {
	Ref   string   `kit:"arg" help:"article url, username/id, or id"`
	Limit int      `kit:"flag,inherit" help:"max records"`
	Sess  *Session `kit:"inject"`
}

func registerComments(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "comments", Group: "read", List: true,
		Summary: "Comments under an article",
		URIType: "article",
		Args:    []kit.Arg{{Name: "url-or-ref", Help: "article url, username/id, or id"}},
	}, func(ctx context.Context, in commentsIn, emit func(Comment) error) error {
		id, err := ParseArticleID(in.Ref)
		if err != nil {
			return errs.Usage("%s", err.Error())
		}
		in.Sess.Progressf("fetching comments for article %s", id)
		comments, err := in.Sess.Client.Comments(ctx, id, effectiveLimit(in.Limit, 50))
		if err != nil {
			return MapErr(err)
		}
		return emitAll(comments, emit)
	})
}

// emitAll streams a slice through emit, stopping on the first error (kit's stop
// sentinel once --limit is reached, or a real one).
func emitAll[T any](items []T, emit func(T) error) error {
	for _, it := range items {
		if err := emit(it); err != nil {
			return err
		}
	}
	return nil
}
