# csdn

[![CI](https://github.com/tamnd/csdn-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/tamnd/csdn-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/tamnd/csdn-cli)](https://github.com/tamnd/csdn-cli/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tamnd/csdn-cli.svg)](https://pkg.go.dev/github.com/tamnd/csdn-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/tamnd/csdn-cli)](https://goreportcard.com/report/github.com/tamnd/csdn-cli)
[![License](https://img.shields.io/github/license/tamnd/csdn-cli)](./LICENSE)

A command line for CSDN (CSDN博客), China's largest developer blog community.
`csdn` reads public CSDN data and prints clean, pipeable records. One pure-Go
binary, no API key, no login.

[Install](#install) • [Commands](#commands) • [Usage](#usage) • [Anti-bot](#anti-bot) • [Serve](#serve-it) • [Docs](https://csdn-cli.tamnd.com)

![csdn reading the hot-rank board into a table and piping it through jq](docs/static/demo.gif)

`csdn` reads the same public web surface a logged-out browser sees: the hot-rank
board, the blog search API, and the article, profile, post-list, and comment
endpoints the pages themselves call. CSDN serves all of it anonymously over plain
HTTPS with no request signing, so the tool carries no signer at all. It returns
records as a table, JSON, JSONL, CSV, TSV, or URLs, and serves the same read
operations over HTTP and MCP. Output is a table at a terminal and JSONL when
piped, so `csdn hot | jq` works with no flags.

`csdn` is an independent tool. It is not affiliated with, authorized, or endorsed
by CSDN (北京创新乐知网络技术有限公司).

## Install

```bash
go install github.com/tamnd/csdn-cli/cmd/csdn@latest
```

Or with Homebrew:

```bash
brew install tamnd/tap/csdn
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/csdn-cli/releases),
or run the container image:

```bash
docker run --rm ghcr.io/tamnd/csdn:latest hot
```

Shell completion is built in: `csdn completion bash|zsh|fish|powershell`.

## Commands

Read operations work from the CLI, over HTTP (`csdn serve`), and over MCP
(`csdn mcp`).

| Command | Description |
|---|---|
| `hot` | Hot-rank board (CSDN热榜), in rank order |
| `search <query>` | Blog search hits (`--type` blog\|all\|ask\|download\|bbs) |
| `article <url\|username/id>` | One article with its body |
| `user <username\|url>` | A public profile |
| `posts <username\|url>` | A user's published articles |
| `comments <url\|username/id\|id>` | Comments under an article |

| Command | Description |
|---|---|
| `serve` | Serve the read operations over HTTP (NDJSON) |
| `mcp` | Run as an MCP server over stdio |
| `raw <url>` | Fetch a URL and print its body untouched |
| `version` | Print version information |

## Usage

```bash
csdn hot                                 # hot-rank board, in rank order
csdn hot --fields rank,title,score -n 10 # pick the columns
csdn search golang -n 20                 # blog search hits
csdn article LOVEmy134611/161870903      # one article with its body
csdn user LOVEmy134611                   # a public profile
csdn posts LOVEmy134611 -n 10            # a user's articles
csdn hot -o json                         # JSON output
csdn hot | jq                            # JSONL piped to jq
```

Output is a table when you are at a terminal and JSONL when you pipe, so the
last line works with no flags. Every command takes the same global flags
(`-o/--output`, `--fields`, `-n/--limit`, `--rate`, `--timeout`, and more); see
the [CLI reference](https://csdn-cli.tamnd.com/reference/cli/) for the full
surface.

## Anti-bot

CSDN is open: there is no request signing, and every read surface answers a plain
anonymous GET as long as the request carries a full browser header set. The one
failure mode is the anti-bot edge, which fronts a 521 status, a 403
`cdn_cgi_bs_bot` challenge, or an HTML challenge shell when it scores a caller as
non-browser. The tool is honest about this: it sends real browser headers, picks
the Referer each surface expects, and on a wall exits cleanly rather than emitting
a challenge page or faking data.

- **Open** from a normal IP, no key: `hot`, `search`, `article`, `user`, `posts`
  all return real records.
- **Bot-walled** under repeated or datacenter access: `comments` guards hardest
  and answers a sticky 403 from a datacenter IP. The command detects this and
  exits with code 4 (needs a residential session or a retry). The parser is
  correct and serves from a residential IP. The wall is at the IP and session
  layer, so no key the client could send opens it.

## Serve it

The read operations are also an HTTP API and an MCP server, backed by the same
code as the CLI.

```bash
csdn serve                   # HTTP on :8080, NDJSON; /healthz and /v1/openapi.json
csdn mcp                     # MCP server over stdio
```

## Docs

Full documentation lives at [csdn-cli.tamnd.com](https://csdn-cli.tamnd.com).

## License

Apache-2.0. See [LICENSE](./LICENSE). `csdn` is an independent tool and is not
affiliated with CSDN.
