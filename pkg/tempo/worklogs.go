package tempo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	goTime "time"

	log "github.com/sirupsen/logrus"

	"toggl-tempo/pkg/config"
	"toggl-tempo/pkg/time"
)

type WorklogsResult struct {
	Metadata Metadata  `json:"metadata"`
	Results  []Worklog `json:"results"`
	Self     string    `json:"self"`
}

type Worklog struct {
	Self           string `json:"self"`
	TempoWorklogID int    `json:"tempoWorklogId"`
	CreatedAt      string `json:"createdAt"`
	Description    string `json:"description"`
	Issue          struct {
		Self string `json:"self"`
		ID   int    `json:"id"`
	} `json:"issue"`
	Author struct {
		Self      string `json:"self"`
		AccountID string `json:"accountId"`
	} `json:"author"`
	SecondsSpent int    `json:"timeSpentSeconds"`
	StartDate    string `json:"startDate"`
	StartTime    string `json:"startTime"`
}

type WorklogAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type WorklogCreateInput struct {
	AuthorAccountID          string             `json:"authorAccountId"`
	IssueID                  int                `json:"issueId"`
	Description              string             `json:"description"`
	StartDate                string             `json:"startDate"`
	StartTime                string             `json:"startTime"`
	TimeSpentSeconds         int                `json:"timeSpentSeconds"`
	RemainingEstimateSeconds int                `json:"remainingEstimateSeconds"`
	Attributes               []WorklogAttribute `json:"attributes"`
}

// GetWorklogs retrieves all worklogs from Tempo.
func (c *Client) GetWorklogs(params url.Values) ([]Worklog, error) {
	req, _ := http.NewRequest("GET", c.Url+"/worklogs", nil)

	if params == nil {
		params = url.Values{}
	}

	if params.Get("orderBy") == "" {
		params.Add("orderBy", "UPDATED")
	}
	req.URL.RawQuery = params.Encode()

	bodyBytes, resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get worklogs: status code %d", resp.StatusCode)
	}

	log.WithField("body", string(bodyBytes)).
		WithField("status", resp.Status).
		Trace("get worklogs response")

	var result WorklogsResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}
	log.WithField("result", result).Trace("get worklogs result")

	return result.Results, nil
}

func MustGetCurrentWeekEntries(params url.Values) []Worklog {
	GetClient()

	if params == nil {
		params = url.Values{}
	}

	if params.Get("from") == "" {
		params.Add("from", time.WeekStartDate(goTime.Now()).Format("2006-01-02"))
		params.Add("to", goTime.Now().Format("2006-01-02"))
	}
	log.WithFields(log.Fields{
		"from": params.Get("from"),
		"to":   params.Get("to"),
	}).Debug("get current week's worklogs")
	worklogs, err := C.GetWorklogs(params)
	if err != nil {
		panic(err)
	}
	return worklogs
}

// CreateWorklog creates a new worklog in Tempo.
func (c *Client) CreateWorklog(input WorklogCreateInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.Url+"/worklogs", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"header": req.Header,
		"body":   string(body),
	}).Trace("create worklogs request")

	bodyBytes, resp, err := c.Do(req)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"status": resp.Status,
		"body":   string(bodyBytes),
	}).Debug("create worklogs response")

	return nil
}

func MustCreateWorklog(input WorklogCreateInput, workAttrVal string) {
	GetClient()

	if config.C.TempoWorkAttributeKey != "" {
		input.Attributes = []WorklogAttribute{
			{Key: config.C.TempoWorkAttributeKey, Value: workAttrVal},
		}
	}

	log.WithField("input", fmt.Sprintf("%+v", input)).Info("creating worklog")
	err := C.CreateWorklog(input)
	if err != nil {
		panic(err)
	}
}
