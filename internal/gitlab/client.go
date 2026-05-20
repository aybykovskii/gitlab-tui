//nolint:mnd // GitLab pagination sizes are stable API tuning values.
package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	glab "gitlab.com/gitlab-org/api/client-go"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/debuglog"
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
	"github.com/aybykovskii/gitlab-tui/pkg/diff"
)

type MergeRequestClient interface {
	ListProjectMergeRequests(pid any, opt *glab.ListProjectMergeRequestsOptions, options ...glab.RequestOptionFunc) ([]*glab.BasicMergeRequest, *glab.Response, error)
	ListMergeRequestDiffs(pid any, mergeRequest int64, opt *glab.ListMergeRequestDiffsOptions, options ...glab.RequestOptionFunc) ([]*glab.MergeRequestDiff, *glab.Response, error)
	AcceptMergeRequest(pid any, mergeRequest int64, opt *glab.AcceptMergeRequestOptions, options ...glab.RequestOptionFunc) (*glab.MergeRequest, *glab.Response, error)
}

type MergeRequestApprovalsClient interface {
	GetConfiguration(pid any, mr int64, options ...glab.RequestOptionFunc) (*glab.MergeRequestApprovals, *glab.Response, error)
	ApproveMergeRequest(pid any, mr int64, opt *glab.ApproveMergeRequestOptions, options ...glab.RequestOptionFunc) (*glab.MergeRequestApprovals, *glab.Response, error)
}

type IssuesClient interface {
	ListProjectIssues(pid any, opt *glab.ListProjectIssuesOptions, options ...glab.RequestOptionFunc) ([]*glab.Issue, *glab.Response, error)
	UpdateIssue(pid any, issue int64, opt *glab.UpdateIssueOptions, options ...glab.RequestOptionFunc) (*glab.Issue, *glab.Response, error)
}

type DiscussionsClient interface {
	ListMergeRequestDiscussions(pid any, mergeRequest int64, opt *glab.ListMergeRequestDiscussionsOptions, options ...glab.RequestOptionFunc) ([]*glab.Discussion, *glab.Response, error)
	ListIssueDiscussions(pid any, issue int64, opt *glab.ListIssueDiscussionsOptions, options ...glab.RequestOptionFunc) ([]*glab.Discussion, *glab.Response, error)
	CreateIssueDiscussion(pid any, issue int64, opt *glab.CreateIssueDiscussionOptions, options ...glab.RequestOptionFunc) (*glab.Discussion, *glab.Response, error)
	CreateMergeRequestDiscussion(pid any, mergeRequest int64, opt *glab.CreateMergeRequestDiscussionOptions, options ...glab.RequestOptionFunc) (*glab.Discussion, *glab.Response, error)
	AddMergeRequestDiscussionNote(pid any, mergeRequest int64, discussion string, opt *glab.AddMergeRequestDiscussionNoteOptions, options ...glab.RequestOptionFunc) (*glab.Note, *glab.Response, error)
	ResolveMergeRequestDiscussion(pid any, mergeRequest int64, discussion string, opt *glab.ResolveMergeRequestDiscussionOptions, options ...glab.RequestOptionFunc) (*glab.Discussion, *glab.Response, error)
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

type DraftNotesClient interface {
	CreateDraftNote(pid any, mergeRequest int64, opt *glab.CreateDraftNoteOptions, options ...glab.RequestOptionFunc) (*glab.DraftNote, *glab.Response, error)
	PublishAllDraftNotes(pid any, mergeRequest int64, options ...glab.RequestOptionFunc) (*glab.Response, error)
	PublishDraftNote(pid any, mergeRequest int64, note int64, options ...glab.RequestOptionFunc) (*glab.Response, error)
	ListDraftNotes(pid any, mergeRequest int64, opt *glab.ListDraftNotesOptions, options ...glab.RequestOptionFunc) ([]*glab.DraftNote, *glab.Response, error)
	DeleteDraftNote(pid any, mergeRequest int64, note int64, options ...glab.RequestOptionFunc) (*glab.Response, error)
}

type Client struct {
	mergeRequests MergeRequestClient
	approvals     MergeRequestApprovalsClient
	discussions   DiscussionsClient
	projects      ProjectsClient
	labels        LabelsClient
	mrEdit        MergeRequestEditClient
	draftNotes    DraftNotesClient
	issues        IssuesClient
}

func NewClient(account config.Account, env []string) (Client, error) {
	debuglog.Log("NewClient: account=%s host=%s", account.ID, account.Host)

	token, err := account.Token(env)
	if err != nil {
		debuglog.Log("NewClient: token error: %v", err)
		return Client{}, err
	}

	opts := []glab.ClientOptionFunc{glab.WithBaseURL(account.Host)}

	if debuglog.Enabled() {
		opts = append(opts, glab.WithRequestLogHook(func(_ retryablehttp.Logger, req *http.Request, _ int) {
			debuglog.Log("HTTP %s %s", req.Method, req.URL)
		}))
	}

	client, err := glab.NewClient(token, opts...)
	if err != nil {
		debuglog.Log("NewClient: glab.NewClient error: %v", err)
		return Client{}, err
	}

	return Client{
		mergeRequests: client.MergeRequests,
		approvals:     client.MergeRequestApprovals,
		discussions:   client.Discussions,
		projects:      client.Projects,
		labels:        client.Labels,
		mrEdit:        client.MergeRequests,
		draftNotes:    client.DraftNotes,
		issues:        client.Issues,
	}, nil
}

func NewClientWithMergeRequests(mergeRequests MergeRequestClient) Client {
	return Client{mergeRequests: mergeRequests}
}

func NewClientWithProjects(projects ProjectsClient) Client {
	return Client{projects: projects}
}

func NewClientWithIssues(issues IssuesClient) Client {
	return Client{issues: issues}
}

func NewClientWithDiscussions(discussions DiscussionsClient) Client {
	return Client{discussions: discussions}
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

func NewClientWithDraftNotes(draftNotes DraftNotesClient) Client {
	return Client{draftNotes: draftNotes}
}

func (c Client) ListProjects(ctx context.Context, limit int) ([]string, error) {
	if c.projects == nil {
		return nil, ErrClientNotConfigured
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
		return nil, normalizeError(err)
	}

	paths := make([]string, 0, len(items))

	for _, item := range items {
		if item != nil {
			paths = append(paths, item.PathWithNamespace)
		}
	}

	return paths, nil
}

func (c Client) SearchProjects(ctx context.Context, query string, limit int) ([]string, error) {
	if c.projects == nil {
		return nil, ErrClientNotConfigured
	}

	membership := true

	items, _, err := c.projects.ListProjects(&glab.ListProjectsOptions{
		Membership: &membership,
		Search:     &query,
		ListOptions: glab.ListOptions{
			PerPage: int64(limit),
			Page:    1,
		},
	}, glab.WithContext(ctx))
	if err != nil {
		return nil, normalizeError(err)
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
		return nil, ErrClientNotConfigured
	}

	debuglog.Log("OpenMergeRequests: project=%s", projectPath)

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
			debuglog.Log("OpenMergeRequests: error listing MRs for %s: %T %v", projectPath, err, err)

			if errors.Is(err, glab.ErrNotFound) {
				c.debugProjectMRAccess(ctx, projectPath)
				break
			}

			return nil, normalizeError(err)
		}

		debuglog.Log("OpenMergeRequests: page %d — got %d items", options.Page, len(items))

		for _, item := range items {
			mapped := MapMergeRequest(item)

			if c.approvals != nil && item != nil {
				if approval, _, approvalErr := c.approvals.GetConfiguration(projectPath, item.IID, glab.WithContext(ctx)); approvalErr == nil {
					mapped.Approvals = formatApprovals(approval)
				}
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

func (c Client) EditIssue(ctx context.Context, projectPath string, iid int, title, description string) error {
	if c.issues == nil {
		return ErrClientNotConfigured
	}

	_, _, err := c.issues.UpdateIssue(projectPath, int64(iid), &glab.UpdateIssueOptions{Title: &title, Description: &description}, glab.WithContext(ctx))

	return normalizeError(err)
}

func (c Client) UpdateIssueLabels(ctx context.Context, projectPath string, iid int, labels []string) error {
	if c.issues == nil {
		return ErrClientNotConfigured
	}

	labelOptions := glab.LabelOptions(labels)
	_, _, err := c.issues.UpdateIssue(projectPath, int64(iid), &glab.UpdateIssueOptions{Labels: &labelOptions}, glab.WithContext(ctx))

	return normalizeError(err)
}

func (c Client) AssignSelfIssue(ctx context.Context, projectPath string, iid int) error {
	if c.issues == nil {
		return ErrClientNotConfigured
	}

	self := int64(0)
	_, _, err := c.issues.UpdateIssue(projectPath, int64(iid), &glab.UpdateIssueOptions{AssigneeID: &self}, glab.WithContext(ctx))

	return normalizeError(err)
}

func (c Client) UnassignSelfIssue(ctx context.Context, projectPath string, iid int) error {
	if c.issues == nil {
		return ErrClientNotConfigured
	}

	unassigned := int64(0)
	_, _, err := c.issues.UpdateIssue(projectPath, int64(iid), &glab.UpdateIssueOptions{AssigneeID: &unassigned}, glab.WithContext(ctx))

	return normalizeError(err)
}

func (c Client) CloseIssue(ctx context.Context, projectPath string, iid int) error {
	return c.updateIssueState(ctx, projectPath, iid, "close")
}

func (c Client) ReopenIssue(ctx context.Context, projectPath string, iid int) error {
	return c.updateIssueState(ctx, projectPath, iid, "reopen")
}

func (c Client) updateIssueState(ctx context.Context, projectPath string, iid int, stateEvent string) error {
	if c.issues == nil {
		return ErrClientNotConfigured
	}

	_, _, err := c.issues.UpdateIssue(projectPath, int64(iid), &glab.UpdateIssueOptions{StateEvent: &stateEvent}, glab.WithContext(ctx))

	return normalizeError(err)
}

func (c Client) ListProjectIssues(ctx context.Context, projectPath string, state string, search string) ([]issue.Issue, error) {
	if c.issues == nil {
		return nil, ErrClientNotConfigured
	}

	options := &glab.ListProjectIssuesOptions{
		State:  &state,
		Search: &search,
		ListOptions: glab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
	}

	items, _, err := c.issues.ListProjectIssues(projectPath, options, glab.WithContext(ctx))
	if err != nil {
		return nil, normalizeError(err)
	}

	result := make([]issue.Issue, 0, len(items))
	for _, item := range items {
		result = append(result, MapIssue(item))
	}

	return result, nil
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
		return nil, ErrClientNotConfigured
	}

	items, _, err := c.labels.ListLabels(projectPath, &glab.ListLabelsOptions{
		ListOptions: glab.ListOptions{PerPage: 100, Page: 1},
	}, glab.WithContext(ctx))
	if err != nil {
		return nil, normalizeError(err)
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
		return ErrClientNotConfigured
	}

	opts := glab.LabelOptions(labels)
	_, _, err := c.mrEdit.UpdateMergeRequest(projectPath, int64(iid), &glab.UpdateMergeRequestOptions{
		Labels: &opts,
	}, glab.WithContext(ctx))

	return normalizeError(err)
}

func (c Client) ToggleDraftMR(ctx context.Context, projectPath string, iid int, title string, draft bool) error {
	if c.mrEdit == nil {
		return ErrClientNotConfigured
	}

	newTitle := strings.TrimPrefix(title, "Draft: ")
	if draft {
		newTitle = "Draft: " + newTitle
	}

	_, _, err := c.mrEdit.UpdateMergeRequest(projectPath, int64(iid), &glab.UpdateMergeRequestOptions{
		Title: &newTitle,
	}, glab.WithContext(ctx))

	return normalizeError(err)
}

func MapIssue(item *glab.Issue) issue.Issue {
	if item == nil {
		return issue.Issue{}
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

	assignees := make([]string, 0, len(item.Assignees))

	for _, assignee := range item.Assignees {
		if assignee == nil {
			continue
		}

		name := assignee.Name
		if name == "" {
			name = assignee.Username
		}

		assignees = append(assignees, name)
	}

	milestone := ""
	if item.Milestone != nil {
		milestone = item.Milestone.Title
	}

	dueDate := ""
	if item.DueDate != nil {
		dueDate = item.DueDate.String()
	}

	return issue.Issue{
		IID:            int(item.IID),
		Title:          item.Title,
		Author:         author,
		AuthorUsername: authorUsername,
		State:          item.State,
		Labels:         append([]string(nil), item.Labels...),
		Assignees:      assignees,
		Description:    item.Description,
		WebURL:         item.WebURL,
		CommentCount:   int(item.UserNotesCount),
		Milestone:      milestone,
		DueDate:        dueDate,
		Weight:         int(item.Weight),
		Confidential:   item.Confidential,
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
		return nil, ErrClientNotConfigured
	}

	opt := &glab.ListMergeRequestDiscussionsOptions{
		ListOptions: glab.ListOptions{PerPage: 100, Page: 1},
	}

	var result []mr.Discussion

	for {
		items, response, err := c.discussions.ListMergeRequestDiscussions(projectPath, int64(iid), opt, glab.WithContext(ctx))
		if err != nil {
			return nil, normalizeError(err)
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

func (c Client) AddIssueComment(ctx context.Context, projectPath string, iid int, body string) error {
	if c.discussions == nil {
		return ErrClientNotConfigured
	}

	_, _, err := c.discussions.CreateIssueDiscussion(projectPath, int64(iid), &glab.CreateIssueDiscussionOptions{Body: &body}, glab.WithContext(ctx))

	return normalizeError(err)
}

func (c Client) ListIssueDiscussions(ctx context.Context, projectPath string, iid int) ([]issue.Discussion, error) {
	if c.discussions == nil {
		return nil, ErrClientNotConfigured
	}

	opt := &glab.ListIssueDiscussionsOptions{
		ListOptions: glab.ListOptions{PerPage: 100, Page: 1},
	}

	var result []issue.Discussion

	for {
		items, response, err := c.discussions.ListIssueDiscussions(projectPath, int64(iid), opt, glab.WithContext(ctx))
		if err != nil {
			return nil, normalizeError(err)
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
		return nil, ErrClientNotConfigured
	}

	opt := &glab.ListMergeRequestDiffsOptions{
		ListOptions: glab.ListOptions{PerPage: 50, Page: 1},
	}

	var result []mr.ChangedFile

	for {
		items, response, err := c.mergeRequests.ListMergeRequestDiffs(projectPath, int64(iid), opt, glab.WithContext(ctx))
		if err != nil {
			return nil, normalizeError(err)
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
		if d.Position == nil && note.Position != nil && note.Position.NewPath != "" {
			d.Position = &mr.DiffPosition{
				NewPath: note.Position.NewPath,
				NewLine: int(note.Position.NewLine),
				OldPath: note.Position.OldPath,
				OldLine: int(note.Position.OldLine),
			}
		}
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
		Diff:         rows,
	}
}

type projectDetailClient interface {
	GetProject(pid any, opt *glab.GetProjectOptions, options ...glab.RequestOptionFunc) (*glab.Project, *glab.Response, error)
}

// debugProjectMRAccess logs project MR access details when the MR list returns 404.
func (c Client) debugProjectMRAccess(ctx context.Context, projectPath string) {
	getter, ok := c.projects.(projectDetailClient)
	if !ok {
		debuglog.Log("OpenMergeRequests: 404 — projects client does not support GetProject")
		return
	}

	project, _, err := getter.GetProject(projectPath, nil, glab.WithContext(ctx))
	if err != nil {
		debuglog.Log("OpenMergeRequests: 404 — could not fetch project details: %v", err)
		return
	}

	debuglog.Log("OpenMergeRequests: 404 — project %s: merge_requests_access_level=%s",
		projectPath, project.MergeRequestsAccessLevel)
}
