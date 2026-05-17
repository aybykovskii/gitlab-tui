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

type MergeRequestApprovalsClient interface {
	GetConfiguration(pid any, mr int64, options ...glab.RequestOptionFunc) (*glab.MergeRequestApprovals, *glab.Response, error)
}

type Client struct {
	mergeRequests MergeRequestClient
	approvals     MergeRequestApprovalsClient
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

	return Client{mergeRequests: client.MergeRequests, approvals: client.MergeRequestApprovals}, nil
}

func NewClientWithMergeRequests(mergeRequests MergeRequestClient) Client {
	return Client{mergeRequests: mergeRequests}
}

func NewClientWithServices(mergeRequests MergeRequestClient, approvals MergeRequestApprovalsClient) Client {
	return Client{mergeRequests: mergeRequests, approvals: approvals}
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
			mapped := MapMergeRequest(item)
			if c.approvals != nil && item != nil {
				approval, _, approvalErr := c.approvals.GetConfiguration(projectPath, item.IID, glab.WithContext(ctx))
				if approvalErr != nil {
					return nil, approvalErr
				}
				mapped.Approvals = formatApprovals(approval)
			}
			result = append(result, mapped)
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
	authorUsername := ""
	if item.Author != nil {
		author = item.Author.Name
		authorUsername = item.Author.Username
		if author == "" {
			author = authorUsername
		}
	}

	pipeline := strings.TrimSpace(item.DetailedMergeStatus)
	if pipeline == "" {
		pipeline = "unknown"
	}

	return mr.MergeRequest{
		IID:            int(item.IID),
		Title:          item.Title,
		Author:         author,
		AuthorUsername: authorUsername,
		SourceBranch:   item.SourceBranch,
		TargetBranch:   item.TargetBranch,
		State:          item.State,
		Pipeline:       pipeline,
		Approvals:      "—",
		Description:    item.Description,
		WebURL:         item.WebURL,
	}
}

func formatApprovals(approval *glab.MergeRequestApprovals) string {
	if approval == nil {
		return "—"
	}
	approved := approval.ApprovalsRequired - approval.ApprovalsLeft
	if approved < 0 {
		approved = 0
	}
	return fmt.Sprintf("%d/%d", approved, approval.ApprovalsRequired)
}
