---
title: "Output formats"
description: "The output contract every command shares: formats, fields, and templates."
weight: 30
---

Every read command renders through one formatter, so the same flags work
everywhere. Pick a format with `-o`, or let csdn choose: a table when writing to
a terminal, JSONL when piped.

## Formats

```bash
csdn <command> -o table      # aligned columns for reading
csdn <command> -o jsonl      # one JSON object per line, for piping
csdn <command> -o json       # a single JSON array
csdn <command> -o markdown   # a Markdown table
csdn <command> -o csv        # spreadsheet friendly
csdn <command> -o tsv        # tab-separated
csdn <command> -o url        # just the url field
csdn <command> -o raw        # the record's body field, one per line
```

| Format | Best for |
|---|---|
| `table` | Reading on a terminal |
| `jsonl` | Piping into another tool, one object at a time |
| `json` | Loading a whole result as an array |
| `markdown` | Pasting a table into a document |
| `csv` / `tsv` | Spreadsheets and quick column math |
| `url` | Feeding URLs into other commands |
| `raw` | The record's text body (the `content` of an article, the `text` of a comment) |

Long text fields are truncated in `table` output to fit; `json`, `jsonl`, `csv`,
and `tsv` carry the full untruncated value.

## Narrowing columns

Keep only the fields you want, by their lowercase JSON key:

```bash
csdn hot --fields rank,title,score
```

`--no-header` drops the header row in `table`, `csv`, and `tsv` output, which
helps when a downstream tool expects bare rows.

## Templating rows

For full control over each line, apply a Go text/template. Fields are the JSON
keys, capitalised:

```bash
csdn hot --template '{{.Rank}} {{.Title}} {{.Score}}'
```

## Why auto-detection helps

Because the default adapts to the destination, the same command reads well by
hand and parses cleanly in a pipe:

```bash
csdn hot            # a table, because this is a terminal
csdn hot | wc -l    # JSONL, because this is a pipe
```

You only reach for `-o` when you want something other than that default.
