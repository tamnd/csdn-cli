---
title: "Configuration"
description: "Global flags and the data directory."
weight: 20
---

`csdn` needs almost no configuration. There is no account, no API key, and no
required environment variable. The settings that exist are the kit global flags,
which apply to every command.

## Request behavior

| Flag | Meaning |
|---|---|
| `--rate` | Minimum delay between requests (e.g. `--rate 1s`) |
| `--timeout` | Per-request timeout |
| `--retries` | Retry attempts on rate limit or 5xx |
| `--user-agent` | Override the User-Agent sent with each request |

By default `csdn` sends a full desktop Chrome header set (a real User-Agent,
`Accept-Language: zh-CN`, and the `Referer` each surface expects), paces
requests, and retries transient 429 and 5xx failures. Those headers are what
keep a plain GET out of the anti-bot challenge, so override `--user-agent` only
if you know what you are doing.

## Output

| Flag | Meaning |
|---|---|
| `-o`, `--output` | Output format (see [output formats](/reference/output/)) |
| `--fields` | Comma-separated columns to show |
| `--no-header` | Omit the header row |
| `--template` | Go text/template applied per record |
| `-n`, `--limit` | Stop after N records |
| `-q`, `--quiet` | Suppress progress output |
| `--color` | Color: `auto`, `always`, or `never` |

## State and caching

| Flag | Meaning |
|---|---|
| `--data-dir` | Override the data directory |
| `--no-cache` | Bypass on-disk caches |
| `--profile` | Named profile to load |
| `--db` | Tee every record into a store (e.g. `out.db`, `postgres://...`) |
| `--dry-run` | Print actions, do not perform them |

## Environment variables

None are required. `csdn` reads no credential or token from the environment.
There is nothing to sign and no key to carry, so there is nothing to configure to
open a surface. The only block is the anti-bot edge at the IP and session layer;
see [troubleshooting](/reference/troubleshooting/) for that picture.
