---
title: "Introduction"
description: "How csdn-cli is put together."
weight: 10
---

`csdn` is a single Go binary built on the any-cli/kit framework. It reads CSDN
(CSDN博客), China's largest developer-blog community, through its public web
surfaces and prints clean structured records you can pipe into your tools. No API
key, no login.

![csdn reading the hot-rank board into a table and piping it through jq](/demo.gif)

## Read operations, three ways

The read operations are defined once and exposed three ways:

- on the CLI, as `csdn hot`, `csdn search`, `csdn article`, `csdn user`,
  `csdn posts`, and `csdn comments`;
- over HTTP, with `csdn serve` (NDJSON responses; endpoints include `/healthz`
  and `/v1/openapi.json`);
- over MCP, with `csdn mcp` (an MCP server over stdio).

The same code answers all three, so a record you get on the CLI is the same
record you get over HTTP or MCP.

## Surfaces

- `hot` reads the CSDN hot-rank board (CSDN热榜), in rank order, as ranked
  article rows.
- `search` reads blog search hits for a query, with `--type blog|all|ask|download|bbs`.
- `article` reads one article with its body, parsed from the HTML page and the
  embedded JSON-LD.
- `user` reads a public profile, from the page's `window.__INITIAL_STATE__` blob
  plus a tab-total JSON for the content counts.
- `posts` reads a user's published articles.
- `comments` reads the comment thread under an article.

## No signing

This is the one big contrast with the rest of the fleet (douyin, bilibili,
xiaohongshu, and the like), which all carry a reimplemented request signer. CSDN
is open: there is no signature to compute anywhere. Every read surface answers a
plain anonymous GET, as long as the request carries a full browser header set
(a real Chrome User-Agent, `Accept-Language: zh-CN`, and the `Referer` each
surface expects). The client builds exactly that and picks the right Referer per
surface, so the request looks like a logged-out browser.

## Output

Every read command returns records. At a terminal the default format is a table;
piped, it is JSONL. Pass `-o json`, `-o csv`, `-o tsv`, `-o url`, `-o markdown`,
or `-o raw` to change format explicitly, and `--fields` or `--template` to shape
what each record shows.

## The one wall

CSDN has no key gate, so the only thing that can stop a request is the anti-bot
edge. It scores each caller and, when it decides a request is not a real browser,
it fronts one of three blocks:

- a 521 status,
- a 403 `cdn_cgi_bs_bot` challenge page,
- or an HTML challenge shell in place of the page.

The client sends real browser headers and the Referer each surface expects, then
on a wall returns a clean exit 4 (needs a residential session). It never crashes
and never fakes data. A 521 is retried; a 403 maps straight to walled because the
edge engages it per surface after repeated calls from one address.

Live from a datacenter IP, `hot`, `search`, `article`, `user`, and `posts` all
returned rich real records. `comments` was the one surface the bot wall held shut
there (a sticky 403), reported honestly as exit 4. From a normal residential IP,
all six return real data.

## Not affiliated

csdn is an independent tool and is not affiliated with, authorized, or endorsed
by CSDN (北京创新乐知网络技术有限公司).
