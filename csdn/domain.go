package csdn

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes CSDN as a kit Domain: a driver a multi-domain host (ant)
// enables with a single blank import,
//
//	import _ "github.com/tamnd/csdn-cli/csdn"
//
// the way a database/sql program enables a driver with `import _
// "github.com/lib/pq"`. The init below registers it; the host then dereferences
// csdn:// URIs by routing to the operations Register installs. The standalone
// csdn binary calls Register through cli.NewApp and shares the same registry, so
// the CLI and the host expose one set of operations.
func init() { kit.Register(Domain{}) }

// Domain is the CSDN driver. It carries no state; the per-run client is built by
// the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against, and
// the identity a host reuses for help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme:   "csdn",
		Aliases:  []string{},
		Hosts:    []string{"csdn.net", "blog.csdn.net", "so.csdn.net", "www.csdn.net", "download.csdn.net"},
		Identity: BaseIdentity(),
	}
}

// BaseIdentity is the help and version identity shared by the standalone binary
// and any host that links the package.
func BaseIdentity() kit.Identity {
	return kit.Identity{
		Binary: "csdn",
		Short:  "A command line for CSDN (CSDN博客).",
		Long: `csdn reads public CSDN (CSDN博客) developer-blog data and prints clean,
pipeable records.

It needs no API key and no login. It reads the same public web surface a
logged-out browser sees: the hot-rank board, blog search, article pages, user
profiles, and the comment threads under an article.

Records come out as table, JSON, JSONL, CSV, TSV, url, or raw.

csdn is an independent tool and is not affiliated with CSDN.`,
		Site: Host,
		Repo: "https://github.com/tamnd/csdn-cli",
	}
}

// Defaults seeds the framework baseline from the csdn defaults, so an unset
// --rate/--retries/--timeout keeps the library's own pacing.
func Defaults(c *kit.Config) {
	d := DefaultConfig()
	c.Rate = d.Rate
	c.Timeout = d.Timeout
	c.Retries = d.Retries
	c.UserAgent = d.UserAgent
}

// Register installs the client factory and every CSDN operation onto app. It is
// the single point both surfaces go through: cli.NewApp calls it for the
// standalone binary, and a host calls Domain.Register for csdn:// URIs.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)
	registerOps(app)
}

// Register is the convenience a host or the binary calls without naming the
// zero-value Domain.
func Register(app *kit.App) { Domain{}.Register(app) }

// Session is the per-run client kit injects into every operation. It pairs the
// HTTP client with the resolved quiet flag, so an operation can pace its own
// stderr progress without reaching for a global.
type Session struct {
	Client *Client
	Quiet  bool
}

// Progressf prints a one-line progress note to stderr unless the run is quiet.
func (s *Session) Progressf(format string, args ...any) {
	if s == nil || s.Quiet {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// newClient is the factory kit calls once per run. It overlays the resolved
// framework globals on the library defaults so --rate, --timeout, --retries, and
// a custom User-Agent reach the HTTP client.
func newClient(_ context.Context, c kit.Config) (any, error) {
	cfg := DefaultConfig()
	if c.UserAgent != "" {
		cfg.UserAgent = c.UserAgent
	}
	if c.Rate > 0 {
		cfg.Rate = c.Rate
	}
	if c.Timeout > 0 {
		cfg.Timeout = c.Timeout
	}
	if c.Retries > 0 {
		cfg.Retries = c.Retries
	}
	return &Session{Client: NewClient(cfg), Quiet: c.Quiet}, nil
}

// Classify turns any accepted input into the canonical (type, id), so `ant
// resolve` and `ant url` need no network. An article url or username/id maps to
// an article; a profile url or a bare username maps to a user.
func (Domain) Classify(input string) (uriType, id string, err error) {
	if username, aid, e := ParseArticleRef(input); e == nil {
		return "article", username + "/" + aid, nil
	}
	if username, e := ParseUsername(input); e == nil {
		return "user", username, nil
	}
	return "", "", errs.Usage("unrecognized CSDN reference: %q", input)
}

// Locate is the inverse: the live https URL for a (type, id), built without a
// fetch. An article id is the username/id pair Classify produced.
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "article":
		username, aid, err := ParseArticleRef(id)
		if err != nil {
			return "", errs.Usage("%s", err.Error())
		}
		return articleURL(username, aid), nil
	case "user":
		return userURL(id), nil
	default:
		return "", errs.Usage("csdn has no resource type %q", uriType)
	}
}

// MapErr converts a library error into the kit error kind that carries the right
// exit code, so a host renders the same walled and not-found outcomes the
// standalone binary does. An anti-bot challenge is exit 4, a missing record is
// exit 6.
func MapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrWalled):
		return errs.NeedAuth("%s", err.Error())
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("%s", err.Error())
	default:
		return err
	}
}
