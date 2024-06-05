package tempo

import (
	"encoding/json"
	"net/http"
)

type AccountsResult struct {
	Metadata Metadata  `json:"metadata"`
	Results  []Account `json:"results"`
	Self     string    `json:"self"`
}

type Account struct {
	Contact struct {
		AccountID string `json:"accountId"`
		Type      string `json:"type"`
		Self      string `json:"self"`
	} `json:"contact"`
	Customer struct {
		ID   string `json:"id"`
		Key  string `json:"key"`
		Name string `json:"name"`
		Self string `json:"self"`
	} `json:"customer"`
	Lead struct {
		AccountID string `json:"accountId"`
		Self      string `json:"self"`
	} `json:"lead"`
	Name string `json:"name"`
	Self string `json:"self"`
}

func (c *Client) GetAccounts() ([]Account, error) {
	req, _ := http.NewRequest("GET", c.Url+"/accounts", nil)
	newQry := req.URL.Query()
	newQry.Add("orderBy", "UPDATED")
	req.URL.RawQuery = newQry.Encode()

	bodyBytes, _, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	var result AccountsResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}
