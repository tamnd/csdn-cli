---
title: "Troubleshooting"
description: "The handful of things that trip people up, and how to fix each one."
weight: 40
---

Most of these come down to network reality or how CSDN serves its data, not a
bug.

## A command exits 4 (walled)

CSDN itself is open, with no key and no signing, so the only thing that exits 4 is
the anti-bot edge. It scores each caller and, when it decides a request is not a
real browser, it fronts one of three blocks:

- a 521 status,
- a 403 `cdn_cgi_bs_bot` challenge page,
- or an HTML challenge shell in place of the page.

The client already sends a full Chrome header set and the `Referer` each surface
expects, which is what keeps a normal request out of the challenge. When the edge
walls anyway, `csdn` exits 4 (needs auth) with a hint that the surface needs a
residential session, rather than pretending it found nothing or printing the
challenge page as data. A 521 is retried; a 403 maps straight to walled, because
the edge engages it per surface after repeated calls from one address.

What this looks like in practice: from a datacenter IP, `hot`, `search`,
`article`, `user`, and `posts` all returned rich real records, while `comments`
was the one surface the wall held shut (a sticky 403). From a normal residential
IP, all six return real data. If you are walled, run the command from a
residential IP or a warm browser session, or slow down with `--rate`.

## A genuinely empty result (exit 3) versus a wall (exit 4)

`csdn` distinguishes the two. A successful request that simply found no items
exits 3 (no results). A blocked request, where the edge returned a challenge,
exits 4 (needs auth). A specific id or username that does not exist exits 6 (not
found). Check the exit code in scripts to tell them apart.

## article rejects a bare numeric id

CSDN article URLs carry the author username in the path, so an article cannot be
addressed by its numeric id alone. Pass a full article URL or the `username/id`
shorthand (for example `csdn article LOVEmy134611/161870903`). The `comments`
command is different: its list is keyed only by the article id, so a bare numeric
id is fine there.

## Requests start failing or returning 429

CSDN rate-limits like any public site. `csdn` already paces requests and retries
the transient failures, but a hard limit still means backing off. Raise the delay
between requests with `--rate` (for example `--rate 1s`) and retry later. A burst
of 429 or 5xx responses is the site asking you to slow down, not a defect. A 429
after retries exits 5 (rate limited).

## Seeing exactly what a surface returned

When a command exits 4 and you want to see the body for yourself, use the `raw`
escape hatch:

```bash
csdn raw https://blog.csdn.net/LOVEmy134611/article/details/161870903
```

It prints the page or endpoint body untouched, including a challenge shell when
walled, so you can confirm the wall rather than guess. Add `-v` for per-request
detail on any command.

## The binary is not on your PATH

`go install` puts the binary in `$(go env GOPATH)/bin` (usually `~/go/bin`), and
a release archive leaves it wherever you unpacked it. If your shell cannot find
`csdn`, add that directory to your `PATH`. See
[installation](/getting-started/installation/).
