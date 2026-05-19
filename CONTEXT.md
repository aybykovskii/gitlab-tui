# gitlab-tui

Terminal-first GitLab client for browsing projects and reviewing GitLab entities through a persistent two-pane navigation model.

## Language

**Navigation Level**:
A product area or object state that can become the active right pane in the TUI.
_Avoid_: screen, page, route

**Navigation Level Stack**:
An ordered chain of **Navigation Levels** where the right pane shows the active level and the left pane shows its parent level.
_Avoid_: navigation stack, screen stack

**Account**:
A configured GitLab identity with an ID, host URL, and a reference to a token environment variable. The Account ID serves as its human-readable display name.
_Avoid_: user, profile, workspace

**Project List**:
The Navigation Level that lists selectable GitLab projects grouped by **Account**, with a **Recent Projects** section at the top. Highlights the resolved project when one is already known.
_Avoid_: project picker, project selector, project screen

**Recent Projects**:
The cross-account section at the top of **Project List** that shows the most recently opened projects up to the configured limit. Each entry displays its **Account** ID in parentheses.
_Avoid_: history, last used, favorites

**Section**:
A project area that groups GitLab entities, such as merge requests, issues, or pipelines.
_Avoid_: category, module

**Section Alias**:
A short CLI token (`mr`, `issue`, or `pipeline`) that opens a **Section** directly for the resolved project.
_Avoid_: command, route

**Project Override**:
The `--project` CLI option that explicitly selects the GitLab project for a command.
_Avoid_: positional project, project argument

**Entity**:
A GitLab object inside a project section, such as a merge request, issue, or pipeline.
_Avoid_: item, record

**Label**:
A project-scoped tag with a name and a hex color applied to an **Entity**. Labels belong to the project and are loaded once per session when the project opens.
_Avoid_: tag, category, badge

**Draft**:
The merge request state indicating it is not ready for merge. A **Draft** merge request can be toggled to ready-for-review. Displayed alongside the MR title, not as a separate field.
_Avoid_: WIP, work in progress

**Label Selector**:
The Navigation Level within MR Detail where the user interactively toggles **Labels** for the selected merge request. Replaces the right pane content while open.
_Avoid_: label picker, label overlay, label editor

**Discussion**:
A GitLab merge request thread with one or more notes, optional diff position, and resolved state.
_Avoid_: comment, note thread

**Changed Files**:
The merge-request context level that lists files changed by the selected merge request.
_Avoid_: diff file picker, file sidebar

**Thread Panel**:
The dynamic area at the bottom of the Diff View right pane that shows the full **Discussion** thread for the diff line under the cursor. Appears only when the cursor line has a **Discussion** or a **Draft Comment**. Height grows with thread content. Hidden or shown with `t`.
_Avoid_: comment panel, inline thread, thread overlay

**Draft Comment**:
An unsaved inline comment written in Diff View, accumulated locally before being submitted as part of a review. Shown in the **Thread Panel** and listed in the **Review Tab**.
_Avoid_: pending comment, local comment, draft note

**Review Tab**:
The **Entity Tab** in MR Detail that lists all accumulated **Draft Comments** for the selected merge request, provides an optional summary input, and offers submit and discard actions. Navigating to a draft from this tab opens Diff View at the corresponding file and line.
_Avoid_: review screen, draft summary, submission tab

**Entity Tab**:
A subordinate view of an **Entity**, such as a merge request diff or a pipeline job.
_Avoid_: subpage, nested screen

**Context Pane**:
The read-only left pane that shows parent-level context for the active right pane.
_Avoid_: sidebar, left navigation

**Key Bar**:
A bordered strip fixed at the bottom of the screen, always visible, showing two lines: local keys for the active **Navigation Level** and global keys available everywhere. Pressing `h` expands the Key Bar upward to show the full key list for the current level; the main panes shrink by the same amount. Pressing `h` again collapses it.
_Avoid_: hint bar, status bar, help line, action bar

**View Component**:
A sub-struct of the top-level BubbleTea `Model` that owns a logical slice of UI state and renders it via a `View(layout LayoutState) string` method. When a View Component embeds a `bubbles` component (e.g. `viewport.Model`), it also implements `Update(msg tea.Msg) tea.Cmd` to delegate message handling — but never owns business logic.
_Avoid_: sub-model, child model, widget

**Layout State**:
The shared rendering context passed into every **View Component**: terminal width, height, active focus, and current mode. Allows each component to size itself without reaching into the parent `Model`.
_Avoid_: render context, screen state, window state

**Local Keys**:
The key bindings that are active only in the current **Navigation Level**. Shown on the first line of the **Key Bar**. Input modes (comment, reply, filter) disable global keys and replace the local line with input-specific hints.
_Avoid_: context keys, mode keys

**Global Keys**:
Key bindings available in every **Navigation Level**: `q` (quit) and `Esc` (back). Shown on the second line of the **Key Bar**. Input modes suppress global keys for the duration of text entry.
_Avoid_: universal keys, shared keys

## Relationships

- A **Navigation Level Stack** contains one or more **Navigation Levels**.
- A **Navigation Level** can have at most one active child **Navigation Level**.
- A **Project List** resolves exactly one GitLab project before section navigation begins and remains the parent level of **Sections**.
- A **Project List** contains one **Recent Projects** section followed by one section per **Account**.
- Each **Account** section in **Project List** loads its projects from the GitLab API in parallel with other accounts when the level opens.
- **Recent Projects** shows projects across all **Accounts**, up to the limit configured in `recent_projects_limit` (default: 10).
- A project entered from **Project List** or opened through **Project Override** becomes the first recent project immediately.
- A git-remote match bypasses **Project List** and opens **Sections** directly; the resolved project is highlighted in the **Context Pane**.
- A **Project** contains multiple **Sections**: merge requests, issues, and pipelines.
- A **Section** contains zero or more **Entities**.
- A **Section Alias** resolves to exactly one **Section**.
- A **Project Override** bypasses git remote and Project Selector resolution.
- An **Entity** can expose zero or more **Entity Tabs**.
- A merge request has zero or more **Discussions**.
- A **Discussion** can appear in the merge request Discussions tab and inline in Diff View when it has a diff position.
- The merge request diff **Entity Tab** uses **Changed Files** as its parent context level.
- The **Context Pane** is read-only; `Esc`/back moves to the previous **Navigation Level** instead of focusing the left pane.
- A **Label** belongs to a **Project**, not to an individual **Entity**; the full label list is fetched once when the project opens.
- A **Label Selector** is opened from MR Detail Summary with `l`; `Space` toggles a label, `Enter` saves, `Esc` cancels.
- A **Draft** merge request shows its status in the title line; `d` toggles between Draft and ready-for-review.
- The **Thread Panel** shows the **Discussion** or **Draft Comment** for the cursor line; `t` toggles its visibility while keeping gutter markers visible.
- When a line has multiple **Discussions**, `[` and `]` switch between them; a `[1/3]` counter is shown in the **Thread Panel** header.
- The **Review Tab** lists **Draft Comments** with file and line context; `Enter` on a draft navigates to Diff View at that location, `Esc` returns to the **Review Tab**.
- The top-level `Model` is decomposed into **View Components**: `layout`, `projectPicker`, `entityList`, `mrDetail`, `issueDetail`, `diffView`, `labelSelector`, and `input`. Each component owns its own state and renders via `View(LayoutState) string`.
- The application runs in AltScreen mode — it occupies the full terminal height and restores the terminal on exit.
- The **Key Bar** is always visible and spans the full width of the screen below both panes.
- The **Key Bar** has two lines in collapsed state: **Local Keys** (line 1) and **Global Keys** (line 2).
- `h` toggles the **Key Bar** between collapsed and expanded; expanded state shows all **Local Keys** for the active **Navigation Level**.
- Each **Navigation Level** declares its own **Local Keys** and which **Global Keys** it suppresses during input modes.
- The collapsed **Key Bar** local line is truncated to terminal width with `…`; `h` reveals the full list.
- `h` is a **Global Key** that toggles **Key Bar** expansion.

## Example dialogue

> **Dev:** "When a merge request diff opens, is the diff a new screen?"
> **Domain expert:** "No — the diff is an **Entity Tab**. The merge request stays visible as the parent level in the left pane while the diff is active in the right pane."

## Flagged ambiguities

- "level" is resolved as **Navigation Level**: a domain/navigation concept, not a code package or router route.
- "вкладка сущности" is resolved as **Entity Tab**: a child level of a selected **Entity**.
- "раздел" is resolved as **Section**: a project area such as merge requests, issues, or pipelines.
- Positional project paths are not part of the canonical CLI grammar; use **Project Override** instead.
- "аккаунт" is resolved as **Account**: a configured GitLab identity, not a user profile or workspace.
- "недавние проекты" / "последние проекты" is resolved as **Recent Projects**: the cross-account section at the top of Project List.
