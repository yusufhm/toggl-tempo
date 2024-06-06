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
