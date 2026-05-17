package gitlab

import (
	"context"
	"fmt"
	"strings"

	glab "gitlab.com/gitlab-org/api/client-go"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/diff"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type MergeRequestClient interface {
	ListProjectMergeRequests(pid any, opt *glab.ListProjectMergeRequestsOptions, options ...glab.RequestOptionFunc) ([]*glab.BasicMergeRequest, *glab.Response, error)
	ListMergeRequestDiffs(pid any, mergeRequest int64, opt *glab.ListMergeRequestDiffsOptions, options ...glab.RequestOptionFunc) ([]*glab.MergeRequestDiff, *glab.Response, error)
}

type Client struct {
	mergeRequests MergeRequestClient
}

func NewClient(account config.Account, env []string) (Client, error) {
	token, err := account.Token(env)
	if err != nil {
		return Client{}, err
	}

	client, err := glab.NewClient(token, glab.WithBaseURL(account.Host))
	if err != nil {
		return Client{}, err
	}

	return Client{mergeRequests: client.MergeRequests}, nil
}

func NewClientWithMergeRequests(mergeRequests MergeRequestClient) Client {
	return Client{mergeRequests: mergeRequests}
}

func (c Client) OpenMergeRequests(ctx context.Context, projectPath string) ([]mr.MergeRequest, error) {
	if c.mergeRequests == nil {
		return nil, fmt.Errorf("merge requests client is not configured")
	}

	state := "opened"
	options := &glab.ListProjectMergeRequestsOptions{
		State: &state,
		ListOptions: glab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
	}

	var result []mr.MergeRequest
	for {
		items, response, err := c.mergeRequests.ListProjectMergeRequests(projectPath, options, glab.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			result = append(result, MapMergeRequest(item))
		}
		if response == nil || response.NextPage == 0 {
			break
		}
		options.Page = response.NextPage
	}

	return result, nil
}

func (c Client) MergeRequestDiff(ctx context.Context, projectPath string, iid int) ([]mr.DiffRow, error) {
	if c.mergeRequests == nil {
		return nil, fmt.Errorf("merge requests client is not configured")
	}

	options := &glab.ListMergeRequestDiffsOptions{
		ListOptions: glab.ListOptions{PerPage: 50, Page: 1},
	}
	rows := []mr.DiffRow{}
	for {
		items, response, err := c.mergeRequests.ListMergeRequestDiffs(projectPath, int64(iid), options, glab.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item != nil {
				rows = append(rows, diff.Parse(item.Diff)...)
			}
		}
		if response == nil || response.NextPage == 0 {
			break
		}
		options.Page = response.NextPage
	}

	return rows, nil
}

func MapMergeRequest(item *glab.BasicMergeRequest) mr.MergeRequest {
	if item == nil {
		return mr.MergeRequest{}
	}

	author := ""
	if item.Author != nil {
		author = item.Author.Username
		if author == "" {
			author = item.Author.Name
		}
	}

	pipeline := strings.TrimSpace(item.DetailedMergeStatus)
	if pipeline == "" {
		pipeline = "unknown"
	}

	return mr.MergeRequest{
		IID:          int(item.IID),
		Title:        item.Title,
		Author:       author,
		SourceBranch: item.SourceBranch,
		TargetBranch: item.TargetBranch,
		State:        item.State,
		Pipeline:     pipeline,
		Approvals:    "—",
		Description:  item.Description,
	}
}
