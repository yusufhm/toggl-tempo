package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"
	goTime "time"

	origToggl "github.com/jason0x43/go-toggl"
	log "github.com/sirupsen/logrus"

	"toggl-tempo/pkg/config"
	"toggl-tempo/pkg/jira"
	"toggl-tempo/pkg/tempo"
	"toggl-tempo/pkg/time"
	"toggl-tempo/pkg/toggl"
)

var finalLogLevel = log.WarnLevel
var lastWeek bool
var listWorklogs bool
var fromDate string
var toDate string
var force bool

func init() {
	logLevel := flag.String("log-level", "warn", "log level")
	verbose := flag.Bool("verbose", false, "verbose mode - alias for '-log-level info'")
	debug := flag.Bool("debug", false, "debug mode - alias for '-log-level debug'")
	flag.BoolVar(&lastWeek, "last-week", false, "sync last week's entries (or list last week with --list-worklogs)")
	flag.BoolVar(&listWorklogs, "list-worklogs", false, "list Tempo worklogs instead of syncing from Toggl")
	flag.StringVar(&fromDate, "from", "", "start date (YYYY-MM-DD) for --list-worklogs; overrides --last-week")
	flag.StringVar(&toDate, "to", "", "end date (YYYY-MM-DD) for --list-worklogs; overrides --last-week")
	flag.BoolVar(&force, "force", false, "re-sync entries even if already tagged as synced")
	flag.Parse()

	if *verbose {
		*logLevel = "info"
	}
	if *debug {
		*logLevel = "debug"
	}

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.WithError(err).Fatal("failed to parse log level")
	}
	finalLogLevel = level

	log.SetLevel(finalLogLevel)
	config.InitEnvConfig()
}

func main() {
	if listWorklogs {
		runListWorklogs()
		return
	}

	jiraAccountId := jira.MustGetAccountId()

	var entries []origToggl.TimeEntry
	if lastWeek {
		entries = toggl.MustGetLastWeekEntries()
		log.WithField("toggl entries", len(entries)).Info("last week's entries")
	} else {
		entries = toggl.MustGetCurrentWeekEntries()
		log.WithField("toggl entries", len(entries)).Info("current week's entries")
	}
	if log.GetLevel() == log.DebugLevel {
		for _, e := range entries {
			log.WithField("entry", fmt.Sprintf("%+v", e)).Debug()
		}
	}

	filteredEntries, entriesToTag := toggl.FilterEntries(entries, force)
	if log.GetLevel() == log.DebugLevel {
		for _, e := range entries {
			log.WithField("entry", fmt.Sprintf("%+v", e)).Debug("filtered entry")
		}
	}

	for day, entries := range filteredEntries {
		log.WithField("day", day).Info("adding day's entries")
		for _, entry := range entries {
			project := toggl.MustGetProject(*entry.Pid, entry.Wid)
			jiraIssueKey := jira.JiraKeyFromString(project.Name)
			jiraIssueId, jiraIssueEstimate, err := jira.GetIssueIdEstimate(jiraIssueKey)
			if err != nil {
				log.Warn(err)
				continue
			}

			entryId := strconv.Itoa(entry.ID)
			start := entry.Start.In(time.Location())
			input := tempo.WorklogCreateInput{
				IssueID:                  jiraIssueId,
				AuthorAccountID:          jiraAccountId,
				Description:              entry.Description,
				StartDate:                start.Format("2006-01-02"),
				StartTime:                start.Format("15:04:05"),
				TimeSpentSeconds:         int(entry.Duration),
				RemainingEstimateSeconds: jiraIssueEstimate,
			}
			tempo.MustCreateWorklog(input, entryId)
		}
	}

	if len(entriesToTag) > 0 {
		log.WithField("count", len(entriesToTag)).Info("tagging entries as synced")
		if err := toggl.BulkAddTimeEntryTag(entriesToTag, "synced"); err != nil {
			log.WithError(err).WithField("count", len(entriesToTag)).Error("failed to bulk tag entries")
		}
	}
}

func runListWorklogs() {
	params := url.Values{}

	switch {
	case fromDate != "" || toDate != "":
		if fromDate != "" {
			params.Set("from", fromDate)
		}
		if toDate != "" {
			params.Set("to", toDate)
		}
	case lastWeek:
		currentWeekStart := time.WeekStartDate(goTime.Now())
		lastWeekStart := currentWeekStart.AddDate(0, 0, -7)
		lastWeekEnd := currentWeekStart.AddDate(0, 0, -1)
		params.Set("from", lastWeekStart.Format("2006-01-02"))
		params.Set("to", lastWeekEnd.Format("2006-01-02"))
	default:
		params.Set("from", time.WeekStartDate(goTime.Now()).Format("2006-01-02"))
		params.Set("to", goTime.Now().Format("2006-01-02"))
	}

	log.WithFields(log.Fields{
		"from": params.Get("from"),
		"to":   params.Get("to"),
	}).Info("listing tempo worklogs")

	worklogs := tempo.MustGetWorklogs(params)
	printWorklogs(worklogs)
}

func printWorklogs(worklogs []tempo.Worklog) {
	if len(worklogs) == 0 {
		fmt.Println("No worklogs found.")
		return
	}

	sort.Slice(worklogs, func(i, j int) bool {
		if worklogs[i].StartDate != worklogs[j].StartDate {
			return worklogs[i].StartDate < worklogs[j].StartDate
		}
		return worklogs[i].StartTime < worklogs[j].StartTime
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DATE\tSTART\tDURATION\tWORKLOG ID\tISSUE ID\tAUTHOR\tDESCRIPTION")
	for _, wl := range worklogs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\t%s\n",
			wl.StartDate,
			wl.StartTime,
			formatDuration(wl.SecondsSpent),
			wl.TempoWorklogID,
			wl.Issue.ID,
			wl.Author.AccountID,
			wl.Description,
		)
	}
	if err := w.Flush(); err != nil {
		log.WithError(err).Error("failed to flush worklogs output")
	}

	fmt.Fprintf(os.Stderr, "\nTotal: %d worklog(s)\n", len(worklogs))
}

func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh%dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	case m > 0 && s > 0:
		return fmt.Sprintf("%dm%ds", m, s)
	case m > 0:
		return fmt.Sprintf("%dm", m)
	default:
		return fmt.Sprintf("%ds", s)
	}
}
