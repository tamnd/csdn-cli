---
title: "CLI reference"
description: "Every flag and subcommand for csdn."
weight: 10
---

`csdn` reads public CSDN data through its public web surfaces: the hot-rank
board, blog search, article and profile pages, and the comment list under an
article. No API key is required, and nothing is signed. Read operations also work
over HTTP (`csdn serve`) and MCP (`csdn mcp`).

Run `csdn <command> --help` for the exact flag list on any command.

## Global flags

These apply to every command. Output is a table at a terminal and JSONL when
piped.

```
  -o, --output string      output: auto|table|markdown|json|jsonl|csv|tsv|url|raw (default "auto")
      --fields strings     comma-separated columns to show
      --no-header          omit the header row
      --template string    Go text/template applied per record
  -n, --limit int          stop after N records (0 = no limit)
  -q, --quiet              suppress progress output
  -v, --verbose            increase verbosity (repeatable)
      --color string       color: auto|always|never (default "auto")
      --rate duration      minimum delay between requests
      --timeout duration   per-request timeout
      --retries int        retry attempts on rate limit or 5xx (default -1)
      --user-agent string  override the User-Agent sent with each request
      --no-cache           bypass on-disk caches
      --data-dir string    override the data directory
      --profile string     named profile to load
      --db string          tee every record into a store (e.g. out.db, postgres://...)
      --dry-run            print actions, do not perform them
```

## Read commands

### csdn hot

The CSDN hot-rank board (CSDN热榜), in rank order, as ranked article rows.

```
Usage:
  csdn hot [flags]

Flags:
      --type string   rank board type
```

Fields: `rank`, `title`, `author`, `username`, `score`, `views`, `comments`,
`favors`, `url`.

The hot-rank board answers a plain anonymous GET and returns real data. `-n`
caps the rows (default 25).

### csdn search

Blog search hits for a query. The query is variadic, so
`csdn search go modules` is one query of two words. Emits `SearchHit` records.

```
Usage:
  csdn search [query...] [flags]

Flags:
      --type string   search type: blog|all|ask|download|bbs
```

Fields: `type`, `id`, `title`, `author`, `username`, `summary`, `views`, `likes`,
`comments`, `url`. The `--type` flag chooses the search index and defaults to
`blog`.

### csdn article

One article with its body. Accepts a full article URL or a `username/id`
shorthand. A bare numeric id is rejected, because CSDN article URLs carry the
author username in the path, so the page cannot be addressed by id alone. Parsed
from the HTML page plus its embedded JSON-LD. Returns a single `Article`.

```
Usage:
  csdn article <url-or-ref> [flags]
```

Fields: `id`, `title`, `author`, `username`, `summary`, `content`, `tags`,
`published`, `updated`, `views`, `likes`, `collects`, `comments`, `url`.

### csdn user

A public profile, addressed by username. Accepts a username or a profile URL.
Read from the page's `window.__INITIAL_STATE__` blob plus a tab-total JSON for
the content counts. Returns a single `User`.

```
Usage:
  csdn user <username-or-url> [flags]
```

Fields: `username`, `nickname`, `intro`, `level`, `code_age`, `region`,
`school`, `company`, `registered`, `gender`, `vip`, `original_count`, `rank`,
`fans`, `follows`, `loyal_fans`, `total_views`, `blog_count`, `column_count`,
`download_count`, `ask_count`, `avatar`, `url`.

### csdn posts

A user's published articles. Accepts the same input as `user`. Emits `SearchHit`
records.

```
Usage:
  csdn posts <username-or-url> [flags]
```

`-n` caps the rows (default 40).

### csdn comments

Comments under an article. Accepts a full article URL, a `username/id` shorthand,
or a bare numeric id (the comment list is keyed only by the article id). Emits
`Comment` records.

```
Usage:
  csdn comments <url-or-ref> [flags]
```

Fields: `id`, `article_id`, `text`, `author`, `nickname`, `parent_id`,
`post_time`, `likes`, `region`, `url`. `-n` caps the rows (default 50).

This is the surface most likely to be walled from a datacenter IP. When the
edge serves a challenge it exits 4 (needs auth) rather than printing nothing as
if it were a result.

## Escape-hatch and server commands

### csdn raw

Fetch a url and print its body untouched. An inspection escape hatch: fetches one
endpoint and prints its raw body to stdout, including a challenge body when
walled. Not a record stream, and not exposed over HTTP or MCP.

```
Usage:
  csdn raw <url> [flags]
```

### csdn serve

Serve the read operations over HTTP (NDJSON). Each operation mounts at
`/v1/<command>`; flags become query parameters and a positional argument can
trail the path. The server also mounts `GET /healthz` and
`GET /v1/openapi.json`.

```
Usage:
  csdn serve [flags]

Flags:
      --addr string    listen address (default ":8080")
      --allow-writes   expose write operations
```

### csdn mcp

Run as an MCP server over stdio. Each read operation becomes an MCP tool whose
argument schema is the operation's input struct.

```
Usage:
  csdn mcp [flags]
```

### csdn completion

Generate the autocompletion script for bash, zsh, fish, or powershell.

```
Usage:
  csdn completion <shell>
```

### csdn version

Print version information.

```
Usage:
  csdn version [flags]

Flags:
      --short   print just the version number
```

## Exit codes

| Code | Meaning |
|---|---|
| 0 | ok |
| 1 | generic error |
| 2 | usage |
| 3 | no results |
| 4 | needs auth / anti-bot wall |
| 5 | rate limited |
| 6 | not found |
| 7 | unsupported |
| 8 | network |

A surface that the anti-bot edge challenges maps to exit 4 (needs auth). That is
the honest "this needs a residential session" outcome, distinct from a genuine
empty result (exit 3) and from a missing record (exit 6).
