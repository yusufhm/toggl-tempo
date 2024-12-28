package toggl

import (
	goTime "time"

	"github.com/jason0x43/go-toggl"
	log "github.com/sirupsen/logrus"

	"toggl-tempo/pkg/config"
	"toggl-tempo/pkg/time"
)

var Client *toggl.Session

func GetClient() *toggl.Session {
	if Client != nil {
		return Client
	}

	c := toggl.OpenSession(config.C.TogglToken)
	if log.GetLevel() < log.DebugLevel {
		toggl.DisableLog()
	}

	Client = &c
	return Client
}

func GetCurrentWeekEntries() ([]toggl.TimeEntry, error) {
	entries, err := Client.GetTimeEntries(time.WeekStartDate(goTime.Now()), goTime.Now())
	if err != nil {
		return nil, err
	}
	log.WithField("entries", entries).Trace("fetched entries")
	return entries, nil
}

func MustGetCurrentWeekEntries() []toggl.TimeEntry {
	GetClient()
	entries, err := GetCurrentWeekEntries()
	if err != nil {
		panic(err)
	}
	return entries
}

func GetLastWeekEntries() ([]toggl.TimeEntry, error) {
	currentWeekStart := time.WeekStartDate(goTime.Now())
	lastWeekStart := currentWeekStart.AddDate(0, 0, -7)
	entries, err := Client.GetTimeEntries(lastWeekStart, goTime.Now())
	if err != nil {
		return nil, err
	}
	log.WithField("entries", entries).Trace("fetched entries")
	return entries, nil
}

func MustGetLastWeekEntries() []toggl.TimeEntry {
	GetClient()
	entries, err := GetLastWeekEntries()
	if err != nil {
		panic(err)
	}
	return entries
}

func MustGetProject(id int, wid int) toggl.Project {
	GetClient()
	project, err := Client.GetProject(id, wid)
	if err != nil {
		panic(err)
	}
	log.WithField("project", project).Trace("fetched project")
	return project
}

func AddTimeEntryTag(entry toggl.TimeEntry, tag string) error {
	entry.Tags = append(entry.Tags, tag)
	_, err := Client.UpdateTimeEntry(entry)
	if err != nil {
		return err
	}
	return nil
}

func FilterEntries(entries []toggl.TimeEntry) (map[string][]toggl.TimeEntry, []toggl.TimeEntry) {
	filtered := map[string][]toggl.TimeEntry{}
	entriesToTag := []toggl.TimeEntry{}
	for _, entry := range entries {
		if entry.Stop == nil {
			continue
		}

		if entry.Pid == nil {
			continue
		}

		if entry.HasTag("synced") {
			log.WithField("entry", entry).Trace("entry already synced")
			continue
		}
		entriesToTag = append(entriesToTag, entry)

		startDate := entry.Start.In(time.Location())
		date := startDate.Format("2006-01-02")
		log.WithFields(log.Fields{
			"entry":      entry,
			"local-date": startDate}).Debug("entry to sync")
		if _, ok := filtered[date]; !ok {
			filtered[date] = []toggl.TimeEntry{}
		}

		if !config.C.TogglGroupSimilarEntries {
			filtered[date] = append(filtered[date], entry)
			continue
		}

		idx, smlrEntry := FindSimilarEntry(filtered[date], entry)
		if smlrEntry == nil {
			filtered[date] = append(filtered[date], entry)
			continue
		}

		smlrEntry.Duration += entry.Duration

		// Use the earliest start time.
		if entry.StartTime().Before(smlrEntry.StartTime()) {
			smlrEntry.Start = entry.Start
		}
		filtered[date][idx] = *smlrEntry
	}

	filteredCount := 0
	for _, entries := range filtered {
		filteredCount += len(entries)
	}
	log.WithField("toggl entries", filteredCount).Info("filtered entries")

	return filtered, entriesToTag
}

func FindSimilarEntry(entries []toggl.TimeEntry, entry toggl.TimeEntry) (int, *toggl.TimeEntry) {
	for i, e := range entries {
		if *e.Pid == *entry.Pid && e.Description == entry.Description {
			return i, &e
		}
	}
	return -1, nil
}
