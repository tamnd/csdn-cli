// Package cli assembles the csdn command tree on top of the csdn library and the
// any-cli/kit framework. The record-stream commands are kit operations the csdn
// package declares once and exposes as CLI, HTTP, and MCP. The two commands that
// do not fit that shape, the raw byte dump and the version banner, are
// escape-hatch kit.Command commands that share the run state through the context.
package cli

import (
	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/csdn-cli/csdn"
)

// Build metadata, set via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// NewApp builds the kit application: the shared identity, the csdn defaults,
// every record operation and the csdn:// driver (both installed by
// csdn.Register), and the escape-hatch commands.
func NewApp() *kit.App {
	id := csdn.BaseIdentity()
	id.Version = Version

	app := kit.New(id, kit.WithDefaults(csdn.Defaults))
	csdn.Register(app)

	// kit gives every binary an honest default User-Agent. Restore a
	// --user-agent override here: bind it to a local and fold it onto the
	// resolved config the client factory reads.
	var userAgent string
	app.GlobalFlags(func(f *kit.FlagSet) {
		f.StringVar(&userAgent, "user-agent", "", "override the User-Agent sent with each request")
	})
	app.Finalize(func(c *kit.Config) {
		if userAgent != "" {
			c.UserAgent = userAgent
		}
	})

	app.AddCommand(newRawCmd())
	app.AddCommand(newVersionCmd())
	return app
}
