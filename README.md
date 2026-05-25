# toggl-tempo

## Clone the repo

```sh
git clone git@github.com:yusufhm/toggl-tempo.git
cd toggl-tempo
```

## Setup environment variables

> [!TIP]
> [direnv](https://direnv.net/docs/installation.html) is a very handy tool for
> setting up variables per directory.

```sh
export JIRA_URL=
export JIRA_USER=
export JIRA_TOKEN=""
export TEMPO_URL=https://api.tempo.io/4
export TEMPO_TOKEN=""
export TOGGL_TOKEN=""
```

## Run the sync

```sh
go run .

# More info:
go run . --verbose

# Debug:
go run . --debug

# Even more (trace)
go run . --log-level trace
```

## List Tempo worklogs

Inspect worklogs that already exist in Tempo without syncing anything new.

```sh
# Current week (Monday to today)
go run . --list-worklogs

# Last week (previous Monday to Sunday)
go run . --list-worklogs --last-week

# Explicit range (overrides --last-week)
go run . --list-worklogs --from 2026-05-18 --to 2026-05-24
```
