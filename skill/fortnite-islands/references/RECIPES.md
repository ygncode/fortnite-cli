# Recipes

Copy-paste command snippets for common tasks. Each recipe is one command that returns JSON the agent can parse directly.

## Top 10 most played islands in the last 24 hours

```bash
fortnite top --by plays --interval hour --last 24h --limit 10
```

Returns: `[{"code":"...","title":"...","value":<plays>}, ...]` sorted descending.

## Top 10 by peak CCU over the last 7 days

```bash
fortnite top --by peakCCU --interval day --last 7d --limit 10
```

## Full picture of one island

```bash
fortnite summary 7022-1566-6527
```

Returns `{metadata, metrics}` with all 8 metric series plus retention.

## How has island X trended over 7 days?

```bash
fortnite islands plays 7022-1566-6527 --interval day --last 7d
```

Returns `{intervals: [{value, timestamp}, ...]}`. Walk the array from oldest to newest to describe the trend.

## Does island X retain players (d1/d7)?

```bash
fortnite islands retention 7022-1566-6527 --last 7d
```

## Compare peak CCU for three islands over the last day

```bash
fortnite compare 7022-1566-6527 8495-3177-6945 1234-1234-1234 --metric peakCCU --interval hour --last 24h
```

Returns aligned `{timestamps, series: [{code, title, values}]}`.

## Find race-genre islands

```bash
fortnite search --tag race --max-scan 500
```

`meta.truncated` tells you if there are more beyond the scan limit.

## Find islands by a specific creator

```bash
fortnite search --creator maphits --max-scan 500
```

## Find islands whose title contains a keyword

```bash
fortnite search --title-contains zombie --max-scan 500
```

## Latest 20 published islands

```bash
fortnite islands list --size 20
```

The API sorts newest-released-first, so this is the freshest 20.

## Pull raw API data for one island (debugging)

```bash
fortnite islands metrics 7022-1566-6527 --raw
```

## Handle the "<5 players" null case

When any metric value is `null`, it means the interval had fewer than 5 unique players. Surface it to the user as "insufficient data" rather than "error". Example:

```bash
fortnite islands peak-ccu 7022-1566-6527 --interval minute --last 60m
```

If the entire `intervals` array has `null` values, say: "Island 7022-1566-6527 had fewer than 5 players in every 10-minute bucket over the last hour."
