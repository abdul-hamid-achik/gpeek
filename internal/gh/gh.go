package gh

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type Client struct {
	available bool
}

func New() *Client {
	c := &Client{}
	c.available = c.checkAvailability()
	return c
}

func (c *Client) checkAvailability() bool {
	cmd := exec.Command("gh", "--version")
	return cmd.Run() == nil
}

func (c *Client) IsAvailable() bool {
	return c.available
}

type PullRequest struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	Author    string `json:"author"`
	Branch    string `json:"headRefName"`
	BaseBranch string `json:"baseRefName"`
	URL       string `json:"url"`
	IsDraft   bool   `json:"isDraft"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type prListItem struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	State       string `json:"state"`
	HeadRefName string `json:"headRefName"`
	BaseRefName string `json:"baseRefName"`
	URL         string `json:"url"`
	IsDraft     bool   `json:"isDraft"`
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Author      struct {
		Login string `json:"login"`
	} `json:"author"`
}

func (c *Client) ListPullRequests(state string) ([]PullRequest, error) {
	if !c.available {
		return nil, nil
	}

	args := []string{"pr", "list", "--json",
		"number,title,state,headRefName,baseRefName,url,isDraft,additions,deletions,createdAt,updatedAt,author"}

	if state != "" {
		args = append(args, "--state", state)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var items []prListItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, err
	}

	var prs []PullRequest
	for _, item := range items {
		prs = append(prs, PullRequest{
			Number:     item.Number,
			Title:      item.Title,
			State:      item.State,
			Author:     item.Author.Login,
			Branch:     item.HeadRefName,
			BaseBranch: item.BaseRefName,
			URL:        item.URL,
			IsDraft:    item.IsDraft,
			Additions:  item.Additions,
			Deletions:  item.Deletions,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}

	return prs, nil
}

func (c *Client) CheckoutPR(number int) error {
	if !c.available {
		return nil
	}

	cmd := exec.Command("gh", "pr", "checkout", string(rune(number+'0')))
	return cmd.Run()
}

func (c *Client) ViewPR(number int) (string, error) {
	if !c.available {
		return "", nil
	}

	cmd := exec.Command("gh", "pr", "view", string(rune(number+'0')))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (c *Client) GetCurrentPR() (*PullRequest, error) {
	if !c.available {
		return nil, nil
	}

	cmd := exec.Command("gh", "pr", "view", "--json",
		"number,title,state,headRefName,baseRefName,url,isDraft,additions,deletions,createdAt,updatedAt,author")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	var item prListItem
	if err := json.Unmarshal(output, &item); err != nil {
		return nil, err
	}

	return &PullRequest{
		Number:     item.Number,
		Title:      item.Title,
		State:      item.State,
		Author:     item.Author.Login,
		Branch:     item.HeadRefName,
		BaseBranch: item.BaseRefName,
		URL:        item.URL,
		IsDraft:    item.IsDraft,
		Additions:  item.Additions,
		Deletions:  item.Deletions,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}, nil
}

type Issue struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	State     string   `json:"state"`
	Author    string   `json:"author"`
	Labels    []string `json:"labels"`
	URL       string   `json:"url"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

type issueListItem struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	URL       string `json:"url"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

func (c *Client) ListIssues(state string) ([]Issue, error) {
	if !c.available {
		return nil, nil
	}

	args := []string{"issue", "list", "--json",
		"number,title,state,url,createdAt,updatedAt,author,labels"}

	if state != "" {
		args = append(args, "--state", state)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var items []issueListItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, err
	}

	var issues []Issue
	for _, item := range items {
		var labels []string
		for _, l := range item.Labels {
			labels = append(labels, l.Name)
		}
		issues = append(issues, Issue{
			Number:    item.Number,
			Title:     item.Title,
			State:     item.State,
			Author:    item.Author.Login,
			Labels:    labels,
			URL:       item.URL,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	return issues, nil
}

func (c *Client) GetRepoInfo() (owner, name string, err error) {
	if !c.available {
		return "", "", nil
	}

	cmd := exec.Command("gh", "repo", "view", "--json", "owner,name")
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	var info struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return "", "", err
	}

	return info.Owner.Login, info.Name, nil
}

func (c *Client) Run(args ...string) (string, error) {
	if !c.available {
		return "", nil
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
