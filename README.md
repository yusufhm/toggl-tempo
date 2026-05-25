# toggl-tempo

## Clone the repo

```sh
git clone git@github.com:yusufhm/toggl-tempo.git
cd toggl-tempo
```

## Configure

Configuration is loaded from a YAML file in your OS user config directory,
with environment variables taking precedence over file values when set.

### Config file

Create `config.yaml` in the appropriate directory for your OS:

| OS      | Path                                                       |
| ------- | ---------------------------------------------------------- |
| macOS   | `~/Library/Application Support/toggl-tempo/config.yaml`    |
| Linux   | `~/.config/toggl-tempo/config.yaml` (or `$XDG_CONFIG_HOME`)|
| Windows | `%AppData%\toggl-tempo\config.yaml`                        |

Example `config.yaml`:

```yaml
jira-url: https://example.atlassian.net
jira-user: you@example.com
jira-token: ""

tempo-url: https://api.tempo.io/4
tempo-token: ""
# Optional: stores the Toggl entry ID against this Tempo work attribute
tempo-work-attribute-key: ""

toggl-token: ""
# When true, similar entries on the same day are merged into a single
# Tempo worklog.
toggl-group-similar-entries: true
```

### Environment variables

Any of these env vars override the corresponding config file value when set:

> [!TIP]
> [direnv](https://direnv.net/docs/installation.html) is a very handy tool for
> setting up variables per directory.

```sh
export JIRA_URL=
export JIRA_USER=
export JIRA_TOKEN=""
export TEMPO_URL=https://api.tempo.io/4
export TEMPO_TOKEN=""
export TEMPO_WORK_ATTRIBUTE_KEY=""
export TOGGL_TOKEN=""
export TOGGL_GROUP_SIMILAR_ENTRIES=true
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

### Re-sync already-synced entries

By default, Toggl entries tagged `synced` are skipped. Pass `--force` to
re-sync them anyway (useful if a previous sync went to the wrong issue or
the Tempo worklog was deleted).

```sh
# Re-sync the current week, including entries already tagged `synced`
go run . --force

# Re-sync last week
go run . --force --last-week
```

Note: `--force` does not delete or update existing Tempo worklogs; it just
creates worklogs for entries that would otherwise have been skipped. Clean
up any duplicates in Tempo manually if needed.

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
