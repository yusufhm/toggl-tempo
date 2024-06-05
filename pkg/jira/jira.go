package jira

import (
	"context"
	"strconv"
	"strings"

	jira "github.com/andygrunwald/go-jira/v2/cloud"

	"toggl-tempo/pkg/config"
)

var Client *jira.Client

func GetClient() (*jira.Client, error) {
	if Client != nil {
		return Client, nil
	}

	tp := jira.BasicAuthTransport{
		Username: config.C.JiraUser,
		APIToken: config.C.JiraToken,
	}

	var err error
	Client, err = jira.NewClient(config.C.JiraURL, tp.Client())
	if err != nil {
		return nil, err
	}

	return Client, nil
}

func GetAccountId() (string, error) {
	_, err := GetClient()
	if err != nil {
		return "", err
	}

	user, _, err := Client.User.GetCurrentUser(context.Background())
	if err != nil {
		return "", err
	}
	return user.AccountID, nil
}

func MustGetAccountId() string {
	accountId, err := GetAccountId()
	if err != nil {
		panic(err)
	}
	return accountId
}

func GetIssueIdEstimate(issueKey string) (int, int, error) {
	_, err := GetClient()
	if err != nil {
		return 0, 0, err
	}

	issue, _, err := Client.Issue.Get(context.Background(),
		issueKey, &jira.GetQueryOptions{})
	if err != nil {
		return 0, 0, err
	}
	issueID, _ := strconv.Atoi(issue.ID)
	return issueID, issue.Fields.TimeEstimate, nil
}

func MustGetIssueIdEstimate(issueKey string) (int, int) {
	issueID, issueEstimate, err := GetIssueIdEstimate(issueKey)
	if err != nil {
		panic(err)
	}
	return issueID, issueEstimate
}

// JiraKeyFromString extracts the Jira issue key from the
// beginning of a string after splitting it by a space.
func JiraKeyFromString(s string) string {
	return strings.Split(s, " ")[0]
}
