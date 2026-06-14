---
title: "Quick start"
description: "Run your first csdn command."
weight: 30
---

## Hot-rank board

```bash
# The CSDN hot-rank board in rank order (table at TTY, JSONL piped)
csdn hot

# A few columns, top 8
csdn hot --fields rank,title,score -n 8

# Output as JSON
csdn hot -o json

# Just the article links
csdn hot -o url
```

The hot-rank board (CSDN热榜) returns real data from a plain anonymous GET. Use
`--type` to pick a board variant and `-n` to cap the rows (default 25).

## Search

```bash
# The query is variadic, so this is one query of two words
csdn search go modules -n 20

# A few columns
csdn search rust --fields title,username,views

# Just the result links
csdn search golang -o url
```

`--type` chooses the search index: `blog` (the default), `all`, `ask`,
`download`, or `bbs`.

## An article, a user, posts, comments

```bash
# One article. A bare numeric id is rejected, because CSDN article URLs carry the
# author username in the path. Pass a full URL or the username/id shorthand.
csdn article LOVEmy134611/161870903

# A profile by username or profile URL
csdn user LOVEmy134611

# A user's published articles
csdn posts LOVEmy134611 -n 10

# Comments under an article (accepts a bare numeric id too)
csdn comments 161490560
```

`article`, `user`, and `posts` parse the public HTML pages (article body from the
page plus its embedded JSON-LD, the profile from the page's
`window.__INITIAL_STATE__` blob). `comments` reads the comment list keyed by the
article id, so it takes a bare id, a `username/id`, or a full URL.

If the anti-bot edge scores your request as a non-browser it fronts a challenge,
and the command exits 4 (needs a residential session) with a hint, rather than
faking data. From a residential IP all six read commands return real records.
`comments` is the surface most likely to wall on a datacenter IP.

## Serve over HTTP or MCP

```bash
csdn serve                # HTTP, NDJSON; /healthz and /v1/openapi.json
csdn mcp                  # MCP server over stdio
```

## Pipe to jq

```bash
csdn hot -n 5 -o jsonl | jq -r '.title'
csdn search golang | jq '{title, username, views}'
```
