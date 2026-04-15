# fortnite-cli

Agent-friendly Go CLI for the official Fortnite Ecosystem v1 Data API, plus a companion [agent skill](https://agentskills.io) (`fortnite-islands`).

## What it does

Queries Fortnite island (Creative / UEFN) metadata and engagement metrics: peak CCU, plays, unique players, retention, favorites, minutes played, recommendations, average minutes per player. It does **not** cover player stats, matches, cosmetics, shop, or BR stats — those are not part of the Ecosystem API.

## Install

### One-liner (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/ygncode/fortnite-cli/main/install.sh | sh
```

This drops the `fortnite` binary into `$HOME/.local/bin`. Re-run to upgrade.

### From source

```bash
go install github.com/ygncode/fortnite-cli/cmd/fortnite@latest
```

### Verify

```bash
fortnite --version
```

## Quickstart

```bash
fortnite islands list --size 5
fortnite top --by peakCCU --last 24h --limit 10
fortnite summary 7022-1566-6527
```

Output is compact JSON by default. Pass `--raw` for the raw API response.

## Using the agent skill

Copy `skill/fortnite-islands/` into your agent's skills directory. Installation paths vary by agent — see your agent's docs for where skills live. After that, ask your agent things like "What are the top 10 Fortnite islands by peak CCU right now?" and it will use the CLI.

## Command reference

| Command | Description |
| --- | --- |
| `fortnite islands list` | List islands (paginated). |
| `fortnite islands get <code>` | Island metadata. |
| `fortnite islands metrics <code>` | Bundled metrics (day) or filterable with `--interval`. |
| `fortnite islands peak-ccu <code>` | Peak concurrent users. |
| `fortnite islands favorites <code>` | Favorites count. |
| `fortnite islands minutes-played <code>` | Total minutes played. |
| `fortnite islands avg-minutes <code>` | Average minutes per player (day only). |
| `fortnite islands recommendations <code>` | Recommendation count. |
| `fortnite islands unique-players <code>` | Unique players. |
| `fortnite islands plays <code>` | Play count. |
| `fortnite islands retention <code>` | D1/D7 retention (day only). |
| `fortnite top --by <metric>` | Top islands by metric. |
| `fortnite summary <code>` | One-shot metadata + metrics. |
| `fortnite compare <code>...` | Compare islands on one metric. |
| `fortnite search` | Client-side filter by tag/creator/title. |

Run any command with `--help` for flag details.

## Design

See `docs/superpowers/specs/2026-04-15-fortnite-cli-and-skill-design.md` for the full design spec.

## License

MIT.
