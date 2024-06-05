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
