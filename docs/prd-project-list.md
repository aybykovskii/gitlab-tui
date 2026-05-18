# PRD: Project List — Multi-Account Project Selection

## Problem Statement

The current **Project List** is a stub: it only shows projects that were pre-loaded by the caller (recent projects from config or a git-remote match). There is no API-backed project discovery, no grouping by account, and no cross-account recent projects view. Users with multiple GitLab accounts cannot see or switch between their accounts from the TUI. Finding a project requires knowing its exact path and typing it manually.

## Solution

Make **Project List** a fully functional first-class Navigation Level. When it opens, it immediately fires parallel API requests to load the first 15 projects for every configured **Account**. Projects are grouped under per-account sections. A **Recent Projects** section at the top gives instant access to the most recently opened projects across all accounts. A `/` filter narrows the visible list in real time; `i` opens manual path input for paths not in the list.

## User Stories

1. As a developer, I want to see my most recently opened projects at the top of Project List, regardless of which account they belong to, so I can reopen them with a single keystroke.
2. As a developer, I want each account's projects to appear in its own labeled section, so I can browse projects by account without ambiguity.
3. As a developer, I want project loading to start automatically when Project List opens, so I never stare at an empty screen wondering if I need to trigger something.
4. As a developer, I want per-account loading indicators, so I know which accounts are still fetching and which are done.
5. As a developer, I want to filter the project list by typing `/`, so I can quickly narrow down projects without scrolling.
6. As a developer, I want to type `i` for manual path input, so I can open any project whose path I already know, even if it is not in the loaded list.
7. As a developer, I want the number of recent projects shown to be configurable, so I can tune it to my workflow without changing code.
8. As a developer, I want opening a project to move it immediately to the top of Recent Projects, so the list stays sorted by actual usage.

## Screen Layout

### Right pane (active Project List)

```
Recent
  company/backend (work)
  group/frontend (personal)
  group/old-api (personal)

[personal]  gitlab.com
> group/frontend
  group/backend
  group/old-api
  …

[work]  gitlab.company.com
  company/backend
  Loading…
```

- Section headers (`Recent`, `[account-id]  host`) are non-selectable; the cursor jumps over them.
- `>` marks the currently highlighted project.
- If a project is already open (git-remote match), it is highlighted in its account section.
- Per-account sections show up to 15 projects from the API.
- **Recent Projects** shows up to `recent_projects_limit` entries (default 10), each with the account ID in parentheses.
- While an account's API call is in flight, its section shows `Loading…`.
- If an API call fails, the section shows `Error: <message>  r: retry`.

### Left pane (Context Pane)

- When navigating Project List: shows `gitlab-tui` application label (unchanged).
- When in Sections or deeper with a git-remote-resolved project: shows Project List with the current project highlighted (`>`), so the user always knows which project is active.

### Key bindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up (skips headers) |
| `↓` / `j` | Move cursor down (skips headers) |
| `Enter` | Open selected project → navigate to Sections |
| `/` | Open filter input (real-time, across all sections) |
| `Esc` | Close filter (if open) |
| `i` | Switch to manual path input mode |
| `r` | Retry failed account loads |
| `q` / `Ctrl+C` | Quit |

## Config Changes

Add `recent_projects_limit` to the config schema:

```yaml
default_account: personal
recent_projects_limit: 10   # optional, default 10

accounts:
  - id: personal
    host: https://gitlab.com
    token_env: GITLAB_TOKEN
  - id: work
    host: https://gitlab.company.com
    token_env: GITLAB_WORK_TOKEN

recent_projects:
  - account: personal
    path: group/frontend
    last_used_at: 2026-05-18T10:00:00Z
```

- `recent_projects_limit`: non-negative integer; 0 hides the Recent section entirely. Defaults to 10 when absent.
- No other config schema changes. Account ID already serves as the display name.

## Code Changes

### `internal/config`

- Add `RecentProjectsLimit int` field to `Config` (yaml: `recent_projects_limit`).
- `RecentProjects() []RecentProject` — new method returning the top-N recent projects across all accounts, sorted by `LastUsedAt` descending, where N = `RecentProjectsLimit` (default 10 when zero).

### `internal/tui`

- Add `LoadProjectsFunc func(accountID string) ([]string, error)` type.
- Add `LoadProjects []AccountProjectLoader` to `ProjectOptions`, where `AccountProjectLoader` pairs an account ID and its load function.
- Extend `Model` with per-account loading state: `projectsLoading map[string]bool`, `projectsError map[string]string`, `accountProjects map[string][]string`.
- On `Init`, fire one `tea.Cmd` per account in parallel (using `tea.Batch`).
- Add messages: `accountProjectsStartedMsg`, `accountProjectsFinishedMsg`.
- `renderProjectPicker` — rewrite to render Recent section + per-account sections with headers; cursor logic skips header rows.
- Filter (`/`) narrows project paths across all sections simultaneously.
- `ModeProjectSelect` cursor navigation: maintain a flat list of selectable indices, skipping header positions.

### `internal/gitlab`

- Add `ListProjects(accountID string, limit int) ([]string, error)` to the client, calling `GET /projects?membership=true&order_by=last_activity_at&per_page=<limit>`.

### `internal/app`

- Wire `LoadProjects` into `ProjectOptions` when building the TUI: one loader per account in `Config.Accounts`.
- Pass `Config.RecentProjects()` (cross-account, capped) as `Recents` to `ProjectOptions`.

## Navigation Behavior (unchanged)

- **Git-remote match** (`ProjectSourceGitRemote`): bypasses Project List entirely, opens Sections directly. The resolved project path is included in `projectList` so the Context Pane highlights it with `>`.
- **Project Override** (`--project`): bypasses Project List entirely, same as today.
- **Recent only** (`ProjectSourceRecentProjects`): Project List opens with Recent section populated; API loading still fires for all accounts.
- **Manual input** (`ProjectSourceManualInput`): Project List opens with empty Recent section; `i` is the primary path.

## Acceptance Criteria

1. Opening Project List fires one API call per account in parallel; each account section shows `Loading…` until its call resolves.
2. After loading, each account section shows up to 15 project paths.
3. Recent Projects section shows up to `recent_projects_limit` (default 10) entries, each formatted as `path (account-id)`.
4. Cursor moves through projects only, skipping section headers.
5. `/` filters all sections in real time; matching is case-insensitive substring on the project path.
6. `i` switches to manual path input; `Enter` with a non-empty path opens that project.
7. Opening any project moves it to the top of Recent (updates `LastUsedAt`); the next time Project List opens it appears first.
8. `recent_projects_limit: 0` hides the Recent section entirely.
9. A failed account load shows `Error: <message>  r: retry` in that account's section; `r` retries all failed accounts.
10. Git-remote-resolved project is highlighted (`>`) in the Context Pane Project List while the user is in Sections or deeper.
