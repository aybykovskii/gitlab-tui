# gitlab-tui

Terminal-first GitLab client for browsing projects and reviewing GitLab entities through a persistent two-pane navigation model.

## Language

**Navigation Level**:
A product area or object state that can become the active right pane in the TUI.
_Avoid_: screen, page, route

**Navigation Level Stack**:
An ordered chain of **Navigation Levels** where the right pane shows the active level and the left pane shows its parent level.
_Avoid_: navigation stack, screen stack

**Project List**:
The Navigation Level that lists selectable GitLab projects and highlights the resolved project when one is already known.
_Avoid_: project picker, project selector, project screen

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

**Discussion**:
A GitLab merge request thread with one or more notes, optional diff position, and resolved state.
_Avoid_: comment, note thread

**Changed Files**:
The merge-request context level that lists files changed by the selected merge request.
_Avoid_: diff file picker, file sidebar

**Entity Tab**:
A subordinate view of an **Entity**, such as a merge request diff or a pipeline job.
_Avoid_: subpage, nested screen

**Context Pane**:
The read-only left pane that shows parent-level context for the active right pane.
_Avoid_: sidebar, left navigation

## Relationships

- A **Navigation Level Stack** contains one or more **Navigation Levels**.
- A **Navigation Level** can have at most one active child **Navigation Level**.
- A **Project List** resolves exactly one GitLab project before section navigation begins and remains the parent level of **Sections**.
- A project entered from **Project List** or opened through **Project Override** becomes the first recent project immediately.
- A **Project** contains multiple **Sections**: merge requests, issues, and pipelines.
- A **Section** contains zero or more **Entities**.
- A **Section Alias** resolves to exactly one **Section**.
- A **Project Override** bypasses git remote and Project Selector resolution.
- An **Entity** can expose zero or more **Entity Tabs**.
- A merge request has zero or more **Discussions**.
- A **Discussion** can appear in the merge request Discussions tab and inline in Diff View when it has a diff position.
- The merge request diff **Entity Tab** uses **Changed Files** as its parent context level.
- The **Context Pane** is read-only; `Esc`/back moves to the previous **Navigation Level** instead of focusing the left pane.

## Example dialogue

> **Dev:** "When a merge request diff opens, is the diff a new screen?"
> **Domain expert:** "No — the diff is an **Entity Tab**. The merge request stays visible as the parent level in the left pane while the diff is active in the right pane."

## Flagged ambiguities

- "level" is resolved as **Navigation Level**: a domain/navigation concept, not a code package or router route.
- "вкладка сущности" is resolved as **Entity Tab**: a child level of a selected **Entity**.
- "раздел" is resolved as **Section**: a project area such as merge requests, issues, or pipelines.
- Positional project paths are not part of the canonical CLI grammar; use **Project Override** instead.
