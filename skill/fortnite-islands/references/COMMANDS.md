# fortnite command reference

Complete per-command reference. This file is loaded on demand — do not load it for every call.

## Global flags (available on every command)

- `--raw` — emit the raw API response instead of the slimmed form.
- `--format json|ndjson` — default `json`.
- `--no-retry` — disable 429/5xx exponential backoff.
- `--timeout DURATION` — HTTP timeout. Default `30s`. Go duration syntax.
- `--concurrency N` — worker-pool size for `top` / `compare`. Default `8`.

## Exit codes

- `0` success
- `1` flag validation error (printed to stderr)
- `2` API error (4xx/5xx after retries; JSON on stderr)
- `3` network or IO error

## `fortnite islands list`

List islands (cursor-paginated, newest released first).

```
fortnite islands list [--size N] [--after CURSOR] [--before CURSOR]
```

Flags:
- `--size N` — default 100, max 1000.
- `--after CURSOR` / `--before CURSOR` — from a previous response's `_pagination.next` / `_pagination.prev`.

Example output (slimmed):
```json
{
  "_pagination": {"next": "/ecosystem/v1/islands?after=...&size=100"},
  "data": [
    {"code":"1234-1234-1234","title":"...","creatorCode":"...","createdIn":"UEFN","tags":["race"],"cursor":"..."}
  ]
}
```

## `fortnite islands get <code>`

Metadata for one island.

Example:
```bash
fortnite islands get 7022-1566-6527
```

## `fortnite islands metrics <code>`

Without `--interval`: bundled day-interval metrics (all metrics at once).
With `--interval`: filterable, supports `--metric M` (repeatable).

```
fortnite islands metrics <code>
fortnite islands metrics <code> --interval day --metric peakCCU --metric plays
fortnite islands metrics <code> --interval hour --last 6h
```

Flags:
- `--interval day|hour|minute`
- `--metric M` — repeatable. Valid: peakCCU, plays, uniquePlayers, minutesPlayed, favorites, recommendations, averageMinutesPerPlayer, retention.
- `--last DURATION` — e.g. `24h`, `7d`, `60m`.
- `--from RFC3339` / `--to RFC3339`.

## `fortnite islands peak-ccu <code>`
`fortnite islands favorites <code>`
`fortnite islands minutes-played <code>`
`fortnite islands recommendations <code>`
`fortnite islands unique-players <code>`
`fortnite islands plays <code>`

Single-metric endpoints. Same flag shape:

```
fortnite islands peak-ccu <code> --interval day|hour|minute [--last DUR | --from T --to T]
```

Default `--interval` is `day`.

Example output:
```json
{
  "intervals": [
    {"value": 120, "timestamp": "2026-04-14T07:00"},
    {"value": null, "timestamp": "2026-04-14T08:00"}
  ]
}
```

## `fortnite islands avg-minutes <code>`

Average minutes per player. **Day-only.** Passing `--interval hour` or `--interval minute` exits 1.

## `fortnite islands retention <code>`

D1 / D7 retention. **Day-only.** Example output:

```json
{
  "intervals": [
    {"d1": 80, "d7": 45, "timestamp": "2026-04-14"},
    {"d1": null, "d7": null, "timestamp": "2026-04-15"}
  ]
}
```

## `fortnite top`

Top islands by a metric. **Expensive.** Scans up to `--scan` islands (default 500, max 1000), making one list pagination call per 100 plus one metric call per island. With default concurrency 8, scanning 500 islands takes ~30–60 seconds.

```
fortnite top --by peakCCU --interval hour --last 24h --limit 10 --scan 500
```

Valid `--by` values: `peakCCU`, `plays`, `uniquePlayers`, `minutesPlayed`, `favorites`, `recommendations`, `averageMinutesPerPlayer`.

Output:
```json
[
  {"code":"1234-1234-1234","title":"...","value":65432}
]
```

Islands whose metric series is entirely null are dropped.

## `fortnite summary <code>`

Metadata + bundled metrics in one call.

```
fortnite summary <code> [--last DUR]
```

Output:
```json
{
  "metadata": {"code":"...","title":"...","tags":[]},
  "metrics": {"peakCCU":[], "plays":[], "uniquePlayers":[]}
}
```

## `fortnite compare`

Compare 2+ islands on one metric.

```
fortnite compare <code1> <code2> [...] --metric peakCCU --interval hour --last 24h
```

Output:
```json
{
  "timestamps": ["2026-04-14T07:00:00.000Z"],
  "metric": "peakCCU",
  "interval": "hour",
  "series": [
    {"code":"A","title":"Alpha","values":[10,20,null]},
    {"code":"B","title":"Bravo","values":[30,null,50]}
  ]
}
```

`timestamps` are the raw ISO8601 strings from the first non-empty series (used for alignment).

## `fortnite search`

Client-side filter over paginated `islands list`. No server-side search on the API.

```
fortnite search [--tag T]... [--creator C] [--title-contains S] [--max-scan N]
```

Flags:
- `--tag T` — repeatable. Island must contain ALL specified tags.
- `--creator C` — exact `creatorCode` match.
- `--title-contains S` — case-insensitive substring of `title`.
- `--max-scan N` — default 500, max 1000. `meta.truncated=true` in the output means you hit the cap.

Output:
```json
{
  "data": [{"code":"...","title":"...","tags":[]}],
  "meta": {"scanned": 500, "matched": 12, "truncated": false}
}
```
