package tempo

import (
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"toggl-tempo/pkg/config"
)

type Client struct {
	*http.Client
	Url   string
	Token string
}

type Metadata struct {
	Count    int    `json:"count"`
	Limit    int    `json:"limit"`
	Next     string `json:"next"`
	Offset   int    `json:"offset"`
	Previous string `json:"previous"`
}

var C *Client

func GetClient() *Client {
	if C != nil {
		return C
	}

	C = &Client{
		Client: &http.Client{},
		Url:    config.C.TempoURL,
		Token:  config.C.TempoToken,
	}

	return C
}

func (c *Client) Do(req *http.Request) ([]byte, *http.Response, error) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.Token)

	log.Tracef("Request: %s %s", req.Method, req.URL)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	log.Tracef("Response body: %s", string(bodyBytes))

	return bodyBytes, resp, nil
}
