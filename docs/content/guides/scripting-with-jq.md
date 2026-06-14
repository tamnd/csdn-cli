---
title: "Scripting with jq"
description: "Combine csdn with jq for flexible data extraction."
weight: 10
---

`csdn` outputs JSONL when piped, making it simple to process with `jq`.

```bash
# Print just the article titles from the hot-rank board
csdn hot | jq -r '.title'

# Title and score as a tab-separated pair
csdn hot | jq -r '[.title, .score] | @tsv'

# Top hot-rank rows over a score threshold
csdn hot -n 25 | jq 'select(.score > 100000) | .title'

# Hot-rank board as CSV
csdn hot -n 10 -o csv

# Search hits, just the fields you want
csdn search golang | jq '{title, username, views}'

# Search result links only
csdn search rust | jq -r '.url'

# Pull a few fields from one article
csdn article LOVEmy134611/161870903 | jq '{title, views, likes, comments}'

# A user's published articles, sorted by views
csdn posts LOVEmy134611 -n 40 | jq -s 'sort_by(-.views) | .[] | .title'
```

CSDN has no key gate, so most surfaces return real records from a plain GET. The
one thing that can stop a command is the anti-bot edge: when it walls a request
the command exits 4 and emits no records, so a downstream `jq` simply sees an
empty stream. Check the exit code in a script if you need to tell a wall (exit 4)
apart from a genuinely empty result (exit 3). See
[troubleshooting](/reference/troubleshooting/) for the exit-code meanings.
