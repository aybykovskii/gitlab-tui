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

type DiscussionsClient interface {
	ListMergeRequestDiscussions(pid any, mergeRequest int64, opt *glab.ListMergeRequestDiscussionsOptions, options ...glab.RequestOptionFunc) ([]*glab.Discussion, *glab.Response, error)
}

type ProjectsClient interface {
	ListProjects(opt *glab.ListProjectsOptions, options ...glab.RequestOptionFunc) ([]*glab.Project, *glab.Response, error)
}

type Client struct {
	mergeRequests MergeRequestClient
	approvals     MergeRequestApprovalsClient
	discussions   DiscussionsClient
	projects      ProjectsClient
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

	return Client{mergeRequests: client.MergeRequests, approvals: client.MergeRequestApprovals, discussions: client.Discussions, projects: client.Projects}, nil
}

func NewClientWithMergeRequests(mergeRequests MergeRequestClient) Client {
	return Client{mergeRequests: mergeRequests}
}

func NewClientWithProjects(projects ProjectsClient) Client {
	return Client{projects: projects}
}

func NewClientWithServices(mergeRequests MergeRequestClient, approvals MergeRequestApprovalsClient) Client {
	return Client{mergeRequests: mergeRequests, approvals: approvals}
}

func (c Client) ListProjects(ctx context.Context, limit int) ([]string, error) {
	if c.projects == nil {
		return nil, fmt.Errorf("projects client is not configured")
	}

	membership := true
	orderBy := "last_activity_at"
	items, _, err := c.projects.ListProjects(&glab.ListProjectsOptions{
		Membership: &membership,
		OrderBy:    &orderBy,
		ListOptions: glab.ListOptions{
			PerPage: int64(limit),
		},
	}, glab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil {
			paths = append(paths, item.PathWithNamespace)
		}
	}

	return paths, nil
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
	if approval == nil || approval.ApprovalsRequired == 0 {
		return "—"
	}
	approved := approval.ApprovalsRequired - approval.ApprovalsLeft
	if approved < 0 {
		approved = 0
	}
	return fmt.Sprintf("%d/%d", approved, approval.ApprovalsRequired)
}

func (c Client) MergeRequestDiscussions(ctx context.Context, projectPath string, iid int) ([]mr.Discussion, error) {
	if c.discussions == nil {
		return nil, fmt.Errorf("discussions client is not configured")
	}
	opt := &glab.ListMergeRequestDiscussionsOptions{
		ListOptions: glab.ListOptions{PerPage: 100, Page: 1},
	}
	var result []mr.Discussion
	for {
		items, response, err := c.discussions.ListMergeRequestDiscussions(projectPath, int64(iid), opt, glab.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			d := MapDiscussion(item)
			if len(d.Notes) > 0 {
				result = append(result, d)
			}
		}
		if response == nil || response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}
	return result, nil
}

func (c Client) MergeRequestChangedFiles(ctx context.Context, projectPath string, iid int) ([]mr.ChangedFile, error) {
	if c.mergeRequests == nil {
		return nil, fmt.Errorf("merge requests client is not configured")
	}
	opt := &glab.ListMergeRequestDiffsOptions{
		ListOptions: glab.ListOptions{PerPage: 50, Page: 1},
	}
	var result []mr.ChangedFile
	for {
		items, response, err := c.mergeRequests.ListMergeRequestDiffs(projectPath, int64(iid), opt, glab.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item != nil {
				result = append(result, MapChangedFile(item))
			}
		}
		if response == nil || response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}
	return result, nil
}

func MapDiscussion(item *glab.Discussion) mr.Discussion {
	if item == nil {
		return mr.Discussion{}
	}
	d := mr.Discussion{ID: item.ID, Resolved: true}
	for _, note := range item.Notes {
		if note == nil || note.System {
			continue
		}
		author := note.Author.Name
		if author == "" {
			author = note.Author.Username
		}
		if !note.Resolved {
			d.Resolved = false
		}
		d.Notes = append(d.Notes, mr.Note{
			Author:   author,
			Body:     note.Body,
			Resolved: note.Resolved,
		})
	}
	if len(d.Notes) == 0 {
		d.Resolved = true
	}
	return d
}

func MapChangedFile(item *glab.MergeRequestDiff) mr.ChangedFile {
	if item == nil {
		return mr.ChangedFile{}
	}
	path := item.NewPath
	if path == "" {
		path = item.OldPath
	}
	rows := diff.Parse(item.Diff)
	added, removed := 0, 0
	for _, row := range rows {
		if row.OldLine == 0 && row.NewLine != 0 {
			added++
		} else if row.OldLine != 0 && row.NewLine == 0 {
			removed++
		}
	}
	return mr.ChangedFile{
		Path:         path,
		OldPath:      item.OldPath,
		IsNew:        item.NewFile,
		IsDeleted:    item.DeletedFile,
		IsRenamed:    item.RenamedFile,
		AddedLines:   added,
		RemovedLines: removed,
	}
}
