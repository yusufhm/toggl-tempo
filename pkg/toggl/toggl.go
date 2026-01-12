package toggl

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	goTime "time"

	"github.com/jason0x43/go-toggl"
	log "github.com/sirupsen/logrus"

	"toggl-tempo/pkg/config"
	"toggl-tempo/pkg/time"
)

var Client *toggl.Session
var projectsCache = map[string]toggl.Project{}

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
	cacheId := fmt.Sprintf("%d-%d", id, wid)
	if p, ok := projectsCache[cacheId]; ok {
		return p
	}

	GetClient()
	project, err := Client.GetProject(id, wid)
	if err != nil {
		panic(err)
	}
	log.WithField("project", project).Trace("fetched project")

	projectsCache[cacheId] = project
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

// BulkAddTimeEntryTag adds a tag to multiple time entries using the Toggl API v9 bulk editing endpoint.
// Entries are grouped by workspace and batched in groups of up to 100 IDs per request.
func BulkAddTimeEntryTag(entries []toggl.TimeEntry, tag string) error {
	if len(entries) == 0 {
		return nil
	}

	// Group entries by workspace_id
	entriesByWorkspace := make(map[int][]toggl.TimeEntry)
	for _, entry := range entries {
		entriesByWorkspace[entry.Wid] = append(entriesByWorkspace[entry.Wid], entry)
	}

	// Process each workspace
	for workspaceID, workspaceEntries := range entriesByWorkspace {
		// Batch entries in groups of 100 (API limit)
		const batchSize = 100
		for i := 0; i < len(workspaceEntries); i += batchSize {
			end := i + batchSize
			if end > len(workspaceEntries) {
				end = len(workspaceEntries)
			}
			batch := workspaceEntries[i:end]

			// Collect entry IDs for this batch
			entryIDs := make([]string, len(batch))
			for j, entry := range batch {
				entryIDs[j] = strconv.Itoa(entry.ID)
			}

			// Make bulk update request
			if err := bulkUpdateTimeEntries(workspaceID, entryIDs, tag); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"workspace_id": workspaceID,
					"entry_ids":    entryIDs,
					"tag":          tag,
				}).Error("failed to bulk update time entries")
				// Continue processing other batches even if one fails
			} else {
				log.WithFields(log.Fields{
					"workspace_id": workspaceID,
					"count":        len(batch),
					"tag":          tag,
				}).Debug("successfully bulk updated time entries")
			}
		}
	}

	return nil
}

// bulkUpdateTimeEntries makes a PATCH request to the Toggl API v9 bulk editing endpoint
func bulkUpdateTimeEntries(workspaceID int, entryIDs []string, tag string) error {
	// Build URL: /api/v9/workspaces/{workspace_id}/time_entries/{comma_separated_ids}
	entryIDsStr := strings.Join(entryIDs, ",")
	url := fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%d/time_entries/%s", workspaceID, entryIDsStr)

	// Create JSON Patch request body
	patchOps := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/tags",
			"value": []string{tag},
		},
	}

	requestBody, err := json.Marshal(patchOps)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Set Basic Auth: Toggl API uses token:api_token format
	auth := base64.StdEncoding.EncodeToString([]byte(config.C.TogglToken + ":api_token"))
	req.Header.Set("Authorization", "Basic "+auth)

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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
