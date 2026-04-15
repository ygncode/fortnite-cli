# Fortnite CLI + Agent Skill — Design Spec

**Date:** 2026-04-15
**Repository:** `github.com/ygncode/fortnite-cli`
**Status:** Approved for implementation planning

## 1. Overview

Build two artifacts, shipped together in one repo:

1. **`fortnite`** — a Go CLI that wraps the official Fortnite Ecosystem v1 Data API (`https://api.fortnite.com/ecosystem/v1`). It provides a thin 1:1 command for every endpoint plus a small set of task-oriented composite verbs.
2. **`fortnite-islands`** — an Agent Skill (following the agentskills.io SKILL.md format) that teaches any compatible agent how to use the CLI efficiently.

Both live in `github.com/ygncode/fortnite-cli`. They version in lockstep. One release triggers both.

The design goal is **agent-friendly, token-efficient access to Fortnite island data**. The CLI exists primarily to be driven by agents via the skill, with human CLI use as a secondary (still supported) mode.

## 2. Scope

### In scope

- All 12 endpoints of the Fortnite Ecosystem v1 Data API (island metadata + engagement metrics).
- A thin 1:1 command for each endpoint.
- Four composite verbs: `top`, `summary`, `compare`, `search`.
- Slim-JSON-by-default output with a `--raw` passthrough.
- `--last DUR` time-range sugar on top of raw `--from`/`--to` timestamps.
- 429/5xx retry with exponential backoff.
- Cross-platform prebuilt binaries via GoReleaser.
- An agentskills.io-compatible skill (`SKILL.md` + `references/COMMANDS.md` + `references/RECIPES.md`).

### Out of scope (non-goals)

- Player stats, match history, cosmetics, shop, BR stats, or any unofficial community API.
- Any feature that requires authentication — the Ecosystem v1 API is public.
- Persistent state across invocations: no local DB, no response caching, no watch mode.
- Interactive TUI, dashboards, colors, progress bars, emoji.
- OpenAPI code generation — the surface is small enough that hand-written types are clearer.
- Config files, logging libraries, viper. Env vars only if ever needed, via stdlib.

### Anti-goals (deliberate refusals)

- **No disk-based response cache.** The API is fast; cache invalidation is not worth the complexity.
- **No subcommands that don't correspond to an API endpoint or a documented composite verb.** Composite verbs are the only value-add layer; everything else stays thin.
- **No human-oriented formatting in v1** (tables, colors, emoji). Agents don't care, and it complicates parsing.

## 3. API Surface

The Fortnite Ecosystem v1 Data API exposes 12 GET endpoints, all under `/islands`:

| # | Endpoint | Description |
| --- | --- | --- |
| 1 | `GET /islands` | Cursor-paginated list of islands, sorted by initial release date (newest first). |
| 2 | `GET /islands/{code}` | Metadata for one island. |
| 3 | `GET /islands/{code}/metrics` | Bundled metrics at `day` interval. |
| 4 | `GET /islands/{code}/metrics/{interval}` | Filterable bundled metrics. Interval ∈ {`day`, `hour`, `minute`}. |
| 5 | `GET /islands/{code}/metrics/{interval}/peak-ccu` | Peak concurrent users. |
| 6 | `GET /islands/{code}/metrics/{interval}/favorites` | Times favorited. |
| 7 | `GET /islands/{code}/metrics/{interval}/minutes-played` | Total player minutes. |
| 8 | `GET /islands/{code}/metrics/{interval}/average-minutes-per-player` | Avg minutes per player (day only). |
| 9 | `GET /islands/{code}/metrics/{interval}/recommendations` | Player recommendations. |
| 10 | `GET /islands/{code}/metrics/{interval}/unique-players` | Unique players. |
| 11 | `GET /islands/{code}/metrics/{interval}/plays` | Play counts. |
| 12 | `GET /islands/{code}/metrics/{interval}/retention` | D1 / D7 retention (day only). |

**Authentication:** none. Empirically verified against the live API on 2026-04-15 — all endpoints return 200 without any `Authorization` header. The OpenAPI spec declares an OAuth2 `securitySchemes` block but no endpoint references it under `security:`, so it is not enforced.

**Historical window:** 7 days. The API rejects ranges beyond that.

**Null values:** the spec states that islands with fewer than 5 unique players in an interval return `null` for that bucket. This is not an error — it is the normal representation of "insufficient data." The CLI and skill must both treat `null` as data, not failure.

**Rate limiting:** 429 responses have a `text/plain` body. `Retry-After` is honored when present.

## 4. Command Surface

### 4.1 Thin 1:1 layer

| # | API endpoint | CLI command |
| --- | --- | --- |
| 1 | `GET /islands` | `fortnite islands list [--size N] [--after CURSOR] [--before CURSOR]` |
| 2 | `GET /islands/{code}` | `fortnite islands get <code>` |
| 3 | `GET /islands/{code}/metrics` | `fortnite islands metrics <code> [--from T] [--to T] [--last DUR]` |
| 4 | `GET /islands/{code}/metrics/{interval}` | `fortnite islands metrics <code> --interval {day\|hour\|minute} [--metric M]... [--from/--to/--last]` |
| 5 | `GET .../peak-ccu` | `fortnite islands peak-ccu <code> --interval I [--from/--to/--last]` |
| 6 | `GET .../favorites` | `fortnite islands favorites <code> --interval I [...]` |
| 7 | `GET .../minutes-played` | `fortnite islands minutes-played <code> --interval I [...]` |
| 8 | `GET .../average-minutes-per-player` | `fortnite islands avg-minutes <code> --interval I [...]` |
| 9 | `GET .../recommendations` | `fortnite islands recommendations <code> --interval I [...]` |
| 10 | `GET .../unique-players` | `fortnite islands unique-players <code> --interval I [...]` |
| 11 | `GET .../plays` | `fortnite islands plays <code> --interval I [...]` |
| 12 | `GET .../retention` | `fortnite islands retention <code> --interval day [...]` (only `day` is valid per spec) |

**Notes:**
- **Dispatch between endpoints #3 and #4:** `fortnite islands metrics <code>` routes to endpoint #3 (`/islands/{code}/metrics`, day buckets, all metrics) when `--interval` is *not* provided. When `--interval` *is* provided, it routes to endpoint #4 (`/islands/{code}/metrics/{interval}`) and may additionally accept repeatable `--metric` filters. This preserves a 1:1 mapping to both endpoints under one ergonomic command name.
- Endpoint #4's repeatable `metrics` query param maps to a repeatable `--metric` flag. Repeating `--metric peakCCU --metric plays` is equivalent to the API's `metrics=peakCCU&metrics=plays`.
- `avg-minutes` is the short name for `average-minutes-per-player`. The full name is too long for ergonomic CLI use.
- Per-metric commands (#5–12) are flat children of `islands`, not nested under `metrics`, because flat is fewer keystrokes and composes more naturally in agent-generated command strings.
- `fortnite islands retention` and `fortnite islands avg-minutes`: `--interval` defaults to `day` and only `day` is accepted. Passing `--interval hour` or `--interval minute` fails with exit 1 before any API call (see §7.1).

### 4.2 Composite verbs

| Command | Internally calls | Returns |
| --- | --- | --- |
| `fortnite top --by METRIC [--interval {day\|hour\|minute}] [--last DUR] [--limit N] [--scan N]` | Paginates `/islands` up to `--scan` records (default 500, max 1000), then parallel per-island metric calls with bounded concurrency, then sorts descending. | Compact JSON array of `{code, title, value}`. Islands with entirely-null metric series are dropped. |
| `fortnite summary <code> [--last DUR]` | `/islands/{code}` + `/islands/{code}/metrics` | One merged object: `{metadata: {...}, metrics: {peakCCU, plays, uniquePlayers, minutesPlayed, favorites, recommendations, averageMinutesPerPlayer, retention}}` |
| `fortnite compare <code1> <code2>... --metric M [--interval I] [--last DUR]` | Parallel `/islands/{code}/metrics/{interval}/{metric}` calls | Aligned-timestamp table: `{timestamps: [...], series: [{code, title, values: [...]}]}` |
| `fortnite search [--tag T]... [--creator C] [--title-contains S] [--max-scan N]` | Paginates `/islands`, filters client-side, stops at `--max-scan` (default 500, max 1000). | Slimmed array of matching islands plus `{meta: {scanned, matched, truncated}}` |

**`top` scan depth rationale:** the `/islands` list is sorted by initial release date (newest first), not by popularity. Scanning only 100 islands biases `top` toward recently released content and returns misleading results. 500 is a more honest default sample. Callers can override with `--scan`. This is the single most expensive command in the CLI and the skill will say so explicitly.

### 4.3 Cross-cutting flags (available on every command)

- `--raw` — bypass slimming; emit the API response as-is.
- `--format {json|ndjson}` — default `json`. `ndjson` is useful for streaming multi-result commands (`top`, `search`).
- `--no-retry` — disable 429/5xx retry.
- `--timeout DUR` — HTTP timeout override (default 30s).
- `--concurrency N` — worker-pool size for composite verbs (default 8).
- `-h`/`--help`.

## 5. Output Shape

### 5.1 Slim JSON rules (default)

1. **Unwrap list envelopes.** `{data: [...], links, meta}` becomes `[...]` with pagination cursors moved to a top-level `_pagination: {next, prev}` object (omitted when both are null).
2. **Round timestamps to interval granularity.** `2026-04-14T00:00:00.000Z` → `2026-04-14` for `day`, `2026-04-14T07:00` for `hour`, `2026-04-14T07:20` for `minute`. `--raw` preserves full ISO8601.
3. **Collapse all-null metric arrays to `null`.** A metric array whose every `value` is `null` is replaced with a literal `null`. Mixed arrays are preserved intact. This loses bucket-count information, but `null = no data` is the agent-relevant signal and bucket count is recoverable from interval + range.
4. **Omit absent optional fields.** `displayName`, `creatorCode`, `category` and similar are omitted when absent rather than sent as `null`.
5. **Preserve everything else verbatim.** Numbers stay numbers. Field names stay camelCase to match the API. No reordering beyond what the above rules require.

### 5.2 Slim example (before/after)

Raw `GET /islands/{code}/metrics` response (abridged):

```json
{
  "averageMinutesPerPlayer": [
    {"value": null, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"value": null, "timestamp": "2026-04-15T00:00:00.000Z"}
  ],
  "peakCCU": [
    {"value": 120, "timestamp": "2026-04-14T00:00:00.000Z"},
    {"value": 95,  "timestamp": "2026-04-15T00:00:00.000Z"}
  ],
  "favorites": [{"value": null, "timestamp": "..."}, {"value": null, "timestamp": "..."}],
  "plays": [...],
  "retention": [{"d1": null, "d7": null, "timestamp": "..."}, ...]
}
```

After slimming:

```json
{
  "averageMinutesPerPlayer": null,
  "peakCCU": [
    {"value": 120, "timestamp": "2026-04-14"},
    {"value": 95,  "timestamp": "2026-04-15"}
  ],
  "favorites": null,
  "plays": null,
  "retention": null
}
```

### 5.3 Error shape

All errors go to **stderr** as a single line of JSON, regardless of source:

```json
{"error": "<short_code>", "message": "<human message>", "status": <http_status_or_null>}
```

Exit codes:
- `0` — success
- `1` — usage / flag validation error
- `2` — API error (4xx or 5xx after retries exhausted)
- `3` — network / timeout / IO error

API 4xx responses decode the spec's `ErrorResponse` schema (`errorCode`, `errorMessage`, `uuid`) and surface those values in `error`, `message`, and an additional `uuid` field.

## 6. Architecture

### 6.1 Repository layout

```
fortnite-cli/
├── cmd/
│   └── fortnite/
│       └── main.go              # entrypoint: wires cobra, zero business logic
├── internal/
│   ├── api/
│   │   ├── client.go            # HTTP client: base URL, timeout, retry, error mapping
│   │   ├── islands.go           # typed methods, one per endpoint
│   │   └── types.go             # request/response structs matching openapi.yaml
│   ├── slim/
│   │   ├── slim.go              # pure functions implementing the slimming rules
│   │   └── slim_test.go         # golden-file tests
│   ├── timerange/
│   │   ├── parse.go             # --last / --from / --to parsing and validation
│   │   └── parse_test.go
│   ├── output/
│   │   └── writer.go            # JSON / NDJSON writers, --raw bypass
│   └── commands/
│       ├── root.go              # root cobra command + shared flags
│       ├── islands_list.go      # one file per subcommand (kept small, target <150 lines)
│       ├── islands_get.go
│       ├── islands_metrics.go
│       ├── islands_peak_ccu.go
│       ├── islands_favorites.go
│       ├── islands_minutes_played.go
│       ├── islands_avg_minutes.go
│       ├── islands_recommendations.go
│       ├── islands_unique_players.go
│       ├── islands_plays.go
│       ├── islands_retention.go
│       ├── top.go
│       ├── summary.go
│       ├── compare.go
│       └── search.go
├── skill/
│   └── fortnite-islands/
│       ├── SKILL.md
│       └── references/
│           ├── COMMANDS.md
│           └── RECIPES.md
├── testdata/
│   ├── raw/                     # captured API responses (small, committed)
│   └── slim/                    # expected slim output goldens
├── .goreleaser.yaml
├── install.sh
├── .github/
│   └── workflows/
│       ├── ci.yaml              # test, vet, lint, skills-ref validate
│       └── release.yaml         # goreleaser on tag
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

### 6.2 Layer separation

- **`cmd/fortnite`** wires cobra. Zero business logic. If `main.go` grows logic, the boundary is wrong.
- **`internal/api`** is a pure API client. Knows nothing about CLI flags, cobra, or stdout. Exposes an interface the command layer depends on so tests can fake it.
- **`internal/slim`** is pure functions from `any`/typed structs to slimmed output. No I/O, no HTTP. Fully unit-testable with golden files.
- **`internal/timerange`** is pure: parses and validates time ranges. Every command uses it so `--last`/`--from`/`--to` semantics are identical everywhere.
- **`internal/output`** handles JSON/NDJSON writing and the `--raw` switch. Pure, takes an `io.Writer`.
- **`internal/commands`** is glue. Each file: parse flags → call `timerange` → call `api` → call `slim` unless `--raw` → write via `output`. Adding an endpoint or composite verb should touch one file, not four.

### 6.3 Dependencies

- `github.com/spf13/cobra` + `github.com/spf13/pflag` — CLI framework.
- `github.com/stretchr/testify` — test assertions.
- Standard library for `encoding/json`, `net/http`, `time`, `context`, `sync`.

Deliberate exclusions:
- **No OpenAPI generator.** 12 endpoints are trivially hand-writable and hand-written types are more readable and editable.
- **No logging library.** Errors go to stderr via `fmt.Fprintln`.
- **No `viper`.** Config file is not needed.

### 6.4 Concurrency model

- `top` and `compare` use a bounded worker pool (default 8, overridable via `--concurrency`) for parallel per-island API calls.
- Every other command is a single sequential HTTP request.
- No goroutines, no background work, no watch loops.

## 7. Core Behaviors

### 7.1 Time-range parsing (`internal/timerange`)

Accepted inputs:
- `--last 24h`, `--last 60m`, `--last 7d` (Go duration syntax, capped at 7 days).
- `--from 2026-04-14T00:00:00Z --to 2026-04-15T00:00:00Z` (ISO8601).
- If both `--last` and `--from`/`--to` are given, `--last` wins and a warning goes to stderr.
- If neither is given, endpoint defaults apply: previous day (for `day`), last 24h (for `hour`), last 60m (for `minute`).

Client-side validation (fail with exit 1 before any API call):
- `from > to` → error.
- `now - from > 7d` → error with an explicit "7-day historical limit" message.
- `--last` > 7d → same.
- `to > now` → clamp to `now` with a stderr warning.
- `fortnite islands retention` with `--interval hour` or `--interval minute` → error: "retention is only available at day interval."
- `fortnite islands avg-minutes` with `--interval hour` or `--interval minute` → error: "average-minutes-per-player is only available at day interval." (The OpenAPI spec contains an internal contradiction: endpoint #8's own docstring says day-only and returns 404 for hour/minute, while the filterable endpoint #4's docstring says avg-minutes is available for day and hour. The dedicated endpoint's own description is authoritative for the dedicated command. When called through the filterable `fortnite islands metrics --interval hour --metric averageMinutesPerPlayer` path, the CLI does not pre-validate and lets the API respond — if the API returns data, we pass it through; if it omits or 404s, we surface that.)

### 7.2 Retry & rate limit (`internal/api`)

- **On 429:** read `Retry-After` if present, otherwise 1s initial. Retry up to 3 times with exponential backoff (1s, 2s, 4s) capped at any larger `Retry-After`. On 4th failure, exit 2 with a structured error containing the API's plain-text response body.
- **On 5xx:** same backoff pattern. Distinct `error` code so agents can distinguish rate-limiting from upstream failure.
- **On network errors** (DNS, connect, TLS, read timeout): retry 3× with backoff. `context.DeadlineExceeded` from the outer `--timeout` propagates without extra retries.
- **On 4xx other than 429:** no retry. Retrying won't help.
- **`--no-retry`** disables all of the above; the first failure exits immediately.

No circuit breaker, no jitter, no adaptive rate limiting — a stateless single-process CLI has nowhere to keep the required state.

### 7.3 End-to-end flow (`fortnite top --by peakCCU --last 24h --limit 10`)

1. `cmd/fortnite/main.go` parses flags via cobra → `internal/commands/top.go` runs.
2. `timerange.Parse` translates `--last 24h` into `from = now - 24h`, `to = now`, interval `hour`.
3. `api.Client.ListIslands` paginates `/islands?size=100` up to 500 records (the `--scan` default).
4. Bounded worker pool (8) fires parallel `api.Client.GetIslandMetric(code, "hour", "peak-ccu", from, to)` calls.
5. For each response, take `max(non-null values)` as the island's 24h peak. Drop islands whose series is entirely null.
6. Sort descending by value; take top `--limit`.
7. `slim.TopResult` emits `[{"code", "title", "value"}, ...]`.
8. `output.Write` JSON-encodes to stdout. Exit 0.

## 8. Agent Skill (`skill/fortnite-islands/`)

### 8.1 `SKILL.md` frontmatter

```yaml
---
name: fortnite-islands
description: Query Fortnite island metadata and engagement metrics (peak CCU, plays, unique players, retention, favorites, minutes played) via the official Fortnite Ecosystem API. Use when the user asks about Fortnite Creative/UEFN islands, island popularity, player counts, play time, retention, or wants to compare or rank islands by any engagement metric. Covers only island data — not player stats, matches, cosmetics, or shop.
license: MIT
compatibility: Requires the `fortnite` CLI binary (install instructions in README). Network access to api.fortnite.com.
metadata:
  version: "1.0"
  repository: https://github.com/ygncode/fortnite-cli
  minimum-cli-version: "0.1.0"
---
```

### 8.2 `SKILL.md` body

Target: ≤2000 tokens loaded. Sections in order:

1. **What this skill does** — one paragraph. Scope, 7-day historical limit, `null` ≠ error.
2. **Preflight check** — `fortnite --version`. If it fails, point at install docs and stop.
3. **Prefer composite commands** — decision tree:
   - User wants "top N islands by X" → `fortnite top --by X --last DUR --limit N`.
   - User wants a full picture of one island → `fortnite summary <code>`.
   - User wants to compare specific islands on one metric → `fortnite compare ...`.
   - User wants to find islands matching tag/creator/title → `fortnite search ...`.
   - Otherwise → drop to the 1:1 layer below.
4. **1:1 endpoint cheat sheet** — one-line table of every command, no flag detail. Points to `references/COMMANDS.md` for flags.
5. **Output shape & parsing** — "slim JSON by default, `--raw` for passthrough; errors are JSON on stderr; check exit code." Two example outputs.
6. **Token-efficiency rules** — "Don't call `top` when you have specific island codes. Don't iterate `islands list` in a loop; use `search` with `--max-scan`. Parse JSON once; don't ask the CLI to re-format. `top` scans up to 500 islands by default — it is the expensive command."
7. **Caveats** — `null` means "<5 unique players in that bucket" not error; the 7-day historical limit; `average-minutes-per-player` is day-only; `retention` is day-only.
8. **References** — links to `references/COMMANDS.md` and `references/RECIPES.md`.

### 8.3 `references/COMMANDS.md`

Full per-subcommand reference. One section per command with: purpose, flags (with defaults), one example input, one example slimmed output (truncated to the interesting part), common errors. Target: 3–5 KB of markdown. Loaded on demand by the agent when it needs flag details — not pulled into context at startup.

### 8.4 `references/RECIPES.md`

10–15 task-oriented snippets. Each is exactly one CLI command and the expected output shape. Starter list:

- "Top 10 most played islands in the last 24 hours"
- "Top 10 by peak CCU over the last 7 days"
- "Full picture of island `<code>` including retention"
- "How has island `<code>` trended over 7 days?"
- "Does island `<code>` retain players? (d1/d7)"
- "Compare peak CCU for these three islands over the last 24 hours"
- "Find race-genre islands"
- "Find islands by creator `<name>`"
- "Find islands whose title contains `<keyword>`"
- "Latest 20 published islands"
- "Pull raw API data for island `<code>` for debugging"
- "Handle the <5-players null case"

This file is the highest-leverage token saver — agents copy recipes verbatim instead of deriving flag combinations.

## 9. Install & Distribution

### 9.1 CLI install

User-facing:

```sh
curl -fsSL https://raw.githubusercontent.com/ygncode/fortnite-cli/main/install.sh | sh
```

`install.sh` behavior:
1. Detect OS (`darwin`/`linux`/`windows` via MSYS) and arch (`amd64`/`arm64`).
2. Query the GitHub Releases API for the latest tag.
3. Download `fortnite_<os>_<arch>.tar.gz` (`.zip` on Windows) and `checksums.txt`.
4. Verify checksum.
5. Extract `fortnite` into `$HOME/.local/bin` (create if needed, chmod 755).
6. Print a PATH hint if `$HOME/.local/bin` is not already in `$PATH`.
7. Idempotent — re-running upgrades in place.

### 9.2 Skill install

Users copy `skill/fortnite-islands/` into their agent's skills directory. The README provides one-liner instructions for major agents (Claude Code, Claude Desktop, Cursor, OpenCode, Goose, etc.) plus a generic `git clone` + symlink fallback. v1 does not ship a skill auto-installer — skill directories vary too much per agent and the skill is three files.

### 9.3 GoReleaser

`.goreleaser.yaml` cross-compiles to:
- `darwin/amd64`, `darwin/arm64`
- `linux/amd64`, `linux/arm64`, `linux/armv7`
- `windows/amd64`

Produces `.tar.gz` / `.zip` archives containing the `fortnite` binary + `LICENSE` + `README.md`. Generates `checksums.txt`. Drafts a GitHub Release. Triggered on tag push (`v*`).

## 10. Testing Strategy

Three layers, each minimal:

1. **Unit tests** (the bulk) for `internal/slim/`, `internal/timerange/`, and `internal/api` (using `httptest.Server`). `slim` uses golden files: a small corpus of `testdata/raw/*.json` captured once from the live API, committed, with expected `testdata/slim/*.json` alongside. When the slimmer changes, diffs are reviewed and goldens updated via `go test -update`. Every slimming rule must be exercised by at least one golden.

2. **Command tests** for `internal/commands/` using a fake implementation of the `api.Client` interface. Tests verify flag-parsing → API-method mapping correctness and that `--raw` bypasses the slimmer. Composite verbs (`top`, `summary`, `compare`, `search`) get dedicated tests because they contain real business logic.

3. **One end-to-end smoke test** behind a `-tags=e2e` build tag, hitting the real API: `fortnite islands list --size 5` exits 0 and returns valid JSON. Runs in CI only on `main` (not fork PRs). Catches upstream API shape drift that frozen fixtures will miss.

**Deliberate non-testing:**
- No mocking `net/http` via interfaces inside `api.Client` (the interface boundary is at the command layer, not the HTTP layer).
- No stdout snapshot tests (brittle, low signal).
- No fuzzing in v1.

## 11. Versioning & Release

- **SemVer**, starting at `v0.1.0`. `v1.0.0` when flag stability is committed to.
- **Skill and CLI version together.** `metadata.version` in SKILL.md bumps in lockstep with the binary tag. One `git tag` releases both.
- **Breaking changes** to CLI flags or JSON output shape: minor bump pre-1.0, major bump post-1.0.
- **Skill compatibility:** SKILL.md's `metadata.minimum-cli-version` names the minimum CLI version the skill body assumes. The preflight check in the skill warns clearly if the installed CLI is older.

## 12. Open Questions

None at this time. Ready for implementation planning.
