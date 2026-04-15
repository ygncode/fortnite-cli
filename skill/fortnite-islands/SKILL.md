---
name: fortnite-islands
description: Query Fortnite island metadata and engagement metrics (peak CCU, plays, unique players, retention, favorites, minutes played) via the official Fortnite Ecosystem API. Use when the user asks about Fortnite Creative/UEFN islands, island popularity, player counts, play time, retention, or wants to compare or rank islands by any engagement metric. Covers only island data — not player stats, matches, cosmetics, or shop.
license: MIT
compatibility: Requires the `fortnite` CLI binary (install instructions in the repo README). Network access to api.fortnite.com.
metadata:
  version: "0.1.0"
  repository: https://github.com/ygncode/fortnite-cli
  minimum-cli-version: "0.1.0"
---

# Fortnite Islands

Query the official Fortnite Ecosystem v1 Data API: island metadata and engagement metrics (peak CCU, plays, unique players, retention, favorites, minutes played). The API's historical window is capped at **7 days**. The `null` metric value means "fewer than 5 unique players in that bucket" — it is not an error. Treat it as data.

## Preflight

Before the first call in a session, verify the CLI is installed:

```bash
fortnite --version
```

If the command is not found, the user needs to install it. Point them at the repository README's install instructions and stop.

## Prefer composite commands

When you can answer the user in one command, use these. Each one is internally doing multiple API calls, so picking the right composite saves many turns and tokens.

| User intent | Command |
| --- | --- |
| "Top N islands by <metric>" | `fortnite top --by <metric> --last <duration> --limit <N>` |
| Full picture of one island | `fortnite summary <code>` |
| Compare specific islands on one metric | `fortnite compare <code1> <code2> [...] --metric <m> --last <duration>` |
| Find islands by tag / creator / title | `fortnite search --tag <t> --creator <c> --title-contains <s>` |
| Otherwise (you want a specific endpoint) | Drop to the 1:1 layer below. |

## 1:1 endpoint cheat sheet

```
fortnite islands list [--size N] [--after CURSOR] [--before CURSOR]
fortnite islands get <code>
fortnite islands metrics <code> [--interval day|hour|minute] [--metric M]... [--last DUR | --from T --to T]
fortnite islands peak-ccu <code> --interval day|hour|minute [--last DUR]
fortnite islands favorites <code> --interval day|hour|minute [--last DUR]
fortnite islands minutes-played <code> --interval day|hour|minute [--last DUR]
fortnite islands avg-minutes <code> --interval day [--last DUR]         # day-only
fortnite islands recommendations <code> --interval day|hour|minute [--last DUR]
fortnite islands unique-players <code> --interval day|hour|minute [--last DUR]
fortnite islands plays <code> --interval day|hour|minute [--last DUR]
fortnite islands retention <code> [--last DUR]                          # day-only
```

Full flags, defaults, and response examples: `references/COMMANDS.md`.

## Output shape and parsing

- Every command returns JSON on stdout (exit 0) or a structured JSON error on stderr (exit 2 for API errors, 3 for network errors, 1 for flag errors).
- Default output is *slim JSON*: pagination envelopes are flattened, timestamps are rounded to the interval granularity (`"2026-04-14"` for day, `"2026-04-14T07:00"` for hour/minute), and metric arrays containing only `null` values collapse to a literal `null`.
- Use `--raw` if you need the exact API shape. You rarely do.
- Error shape: `{"error": "rate_limited|not_found|bad_request|upstream_error|api_error|generic", "message": "...", "status": 429}`.

Example `fortnite top --by peakCCU --last 1h --limit 3` output:

```json
[
  {"code":"1234-1234-1234","title":"Battle Royale","value":65432},
  {"code":"5555-5555-5555","title":"Fall Guys","value":45000},
  {"code":"9999-9999-9999","title":"OG","value":33000}
]
```

Example `fortnite islands peak-ccu <code> --interval hour --last 3h` output:

```json
{
  "intervals": [
    {"value": 120, "timestamp": "2026-04-14T07:00"},
    {"value": 95, "timestamp": "2026-04-14T08:00"},
    {"value": null, "timestamp": "2026-04-14T09:00"}
  ]
}
```

## Token-efficiency rules

1. **Don't call `top` when you already have island codes.** `top` scans up to 500 islands by default and fires one metric call per island. It is the most expensive command. If the user names specific islands, use `summary` or `compare` instead.
2. **Don't iterate `islands list` in a loop.** Use `search` with `--max-scan` to filter client-side in a single call. The API has no server-side search.
3. **Parse the JSON once.** Don't ask the CLI to re-run with different flags just to reformat output; adjust your parsing.
4. **Use `--last` over `--from`/`--to`** when the user is asking about "the last N hours/days" — it's shorter and less error-prone.
5. **Check exit code.** Exit 2 means the API responded with an error (rate-limit, not-found, etc.). The JSON on stderr tells you which.

## Caveats

- `null` in a metric value means "fewer than 5 unique players in that interval." It is normal, not an error. Surface it to the user as "insufficient data," not as a failure.
- The API's historical window is 7 days. Anything older returns an error.
- `avg-minutes` (average minutes per player) is only available at `day` interval.
- `retention` (d1/d7) is only available at `day` interval.
- `top` scans the newest-released-first island list, not a popularity-ranked list. The first 100 islands will skew toward recent releases. Default scan depth is 500 to mitigate this. If you truly need a global top-N, pass `--scan 1000` and expect it to take tens of seconds.

## References

- `references/COMMANDS.md` — complete per-command reference with all flags, defaults, and example output.
- `references/RECIPES.md` — copy-paste command snippets for common user tasks.
