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

type LabelsClient interface {
	ListLabels(pid any, opt *glab.ListLabelsOptions, options ...glab.RequestOptionFunc) ([]*glab.Label, *glab.Response, error)
}

type MergeRequestEditClient interface {
	UpdateMergeRequest(pid any, mergeRequest int64, opt *glab.UpdateMergeRequestOptions, options ...glab.RequestOptionFunc) (*glab.MergeRequest, *glab.Response, error)
}

type Client struct {
	mergeRequests MergeRequestClient
	approvals     MergeRequestApprovalsClient
	discussions   DiscussionsClient
	projects      ProjectsClient
	labels        LabelsClient
	mrEdit        MergeRequestEditClient
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

	return Client{
		mergeRequests: client.MergeRequests,
		approvals:     client.MergeRequestApprovals,
		discussions:   client.Discussions,
		projects:      client.Projects,
		labels:        client.Labels,
		mrEdit:        client.MergeRequests,
	}, nil
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

func NewClientWithLabels(labels LabelsClient) Client {
	return Client{labels: labels}
}

func NewClientWithMergeRequestEdit(mrEdit MergeRequestEditClient) Client {
	return Client{mrEdit: mrEdit}
}

func (c Client) ListProjects(ctx context.Context, limit int) ([]string, error) {
	if c.projects == nil {
		return nil, fmt.Errorf("projects client is not configured")
	}

	membership := true
	items, _, err := c.projects.ListProjects(&glab.ListProjectsOptions{
		Membership: &membership,
		ListOptions: glab.ListOptions{
			PerPage: int64(limit),
			Page:    1,
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
		Labels:         []string(item.Labels),
		Draft:          item.Draft,
		Reviewers:      userNames(item.Reviewers),
		Assignees:      userNames(item.Assignees),
	}
}

func userNames(users []*glab.BasicUser) []string {
	names := make([]string, 0, len(users))
	for _, u := range users {
		if u == nil {
			continue
		}
		name := u.Name
		if name == "" {
			name = u.Username
		}
		names = append(names, name)
	}
	return names
}

func (c Client) ListProjectLabels(ctx context.Context, projectPath string) ([]mr.Label, error) {
	if c.labels == nil {
		return nil, fmt.Errorf("labels client is not configured")
	}
	items, _, err := c.labels.ListLabels(projectPath, &glab.ListLabelsOptions{
		ListOptions: glab.ListOptions{PerPage: 100, Page: 1},
	}, glab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	result := make([]mr.Label, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, mr.Label{Name: item.Name, Color: item.Color})
		}
	}
	return result, nil
}

func (c Client) UpdateMRLabels(ctx context.Context, projectPath string, iid int, labels []string) error {
	if c.mrEdit == nil {
		return fmt.Errorf("merge request edit client is not configured")
	}
	opts := glab.LabelOptions(labels)
	_, _, err := c.mrEdit.UpdateMergeRequest(projectPath, int64(iid), &glab.UpdateMergeRequestOptions{
		Labels: &opts,
	}, glab.WithContext(ctx))
	return err
}

func (c Client) ToggleDraftMR(ctx context.Context, projectPath string, iid int, title string, draft bool) error {
	if c.mrEdit == nil {
		return fmt.Errorf("merge request edit client is not configured")
	}
	newTitle := strings.TrimPrefix(title, "Draft: ")
	if draft {
		newTitle = "Draft: " + newTitle
	}
	_, _, err := c.mrEdit.UpdateMergeRequest(projectPath, int64(iid), &glab.UpdateMergeRequestOptions{
		Title: &newTitle,
	}, glab.WithContext(ctx))
	return err
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
