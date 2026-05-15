# gitlab-tui — Domain Glossary

## Core entities

**Project** — GitLab-проект (namespace/repo). Определяется автоматически из `git remote origin` текущей директории. Если не в git-репо — выбирается из списка недавних проектов или через поиск по API.

**MR (Merge Request)** — единица изменений в GitLab. Центральный объект утилиты.

**Diff** — набор изменений файлов в MR. Отображается построчно (unified diff). Каждая строка имеет позицию (путь к файлу, номер строки, SHA коммитов), необходимую для привязки комментариев.

## Review flow

**Review (Batch Review)** — набор Draft Comments, накапливаемых в рамках одной сессии ревью. Публикуется целиком одной командой "Submit Review". Соответствует GitLab Draft Notes API.

**Draft Comment** — Inline Comment, сохранённый как часть незавершённого Review. Не виден другим участникам до публикации Review.

**Inline Comment** — комментарий, привязанный к конкретной строке в Diff. Может быть Draft Comment (часть Review) или Instant Comment.

**MR Comment** — общий комментарий к MR, не привязан к строке или файлу. Публикуется немедленно.

**Instant Comment** — Inline Comment или MR Comment, публикуемый немедленно (не часть Review).

**Thread** — цепочка ответов на Inline Comment или MR Comment. Может быть resolved/unresolved.

## Config & accounts

**Account** — настроенное подключение к GitLab-инстансу: `{name, url, token}`. Поддерживается несколько аккаунтов одновременно (личный + рабочий). Хранится в `~/.config/gitlab-tui/config.json`.

**Default Account** — аккаунт, используемый когда авто-детект по `.git` remote не дал результата.

**Recent Projects** — список последних проектов, с которыми работал пользователь. Хранится в глобальном конфиге, обновляется автоматически. Используется как fallback при запуске вне git-репозитория.

## MR actions

**Approve** — действие пользователя, добавляющее его апрув к MR через GitLab API.

**Merge** — действие слияния MR в target branch, выполняемое из TUI.

**Edit MR** — изменение метаданных MR: заголовок, описание, assignee, reviewer, labels, draft-статус.

**MR Creation** — создание нового MR. Поля MVP: source branch (авто из текущей git-ветки), target branch, заголовок, описание (с поддержкой MR Templates), draft-флаг, assignee, reviewer, labels.

**MR Template** — шаблон описания MR из `.gitlab/merge_request_templates/`. Подставляется в поле описания при создании MR.

## Screen structure

**MR List Screen** — список всех MR проекта. Fuzzy-поиск по заголовку, фильтр по статусу (open/merged/closed).

**MR Detail Screen** — детальный экран MR. Шапка: заголовок, автор, ветки (source → target), статус, pipeline, апрувы, описание. Две вкладки: Files и Threads.

**Files Tab** — список изменённых файлов в MR. Переход в Diff View по выбору файла.

**Threads Tab** — список всех тредов/комментариев MR.

**Diff View** — side-by-side просмотр изменений одного файла с возможностью добавления комментариев.

## Diff view

**Diff View** — side-by-side отображение изменений: старый код слева, новый код справа. Обе панели прокручиваются синхронно. Синтаксическая подсветка через `chalk`. Навигация по строкам: `j/k` и стрелки.

**Diff Position** — уникальная координата строки в Diff: `{base_sha, head_sha, start_sha, old_path, new_path, old_line, new_line}`. Передаётся в GitLab API при создании Inline Comment.

## Navigation & UX

**Setup Wizard** — интерактивный мастер первого запуска. Запускается автоматически при отсутствии конфига. Запрашивает GitLab URL, PAT, предпочитаемый редактор.

**Deep Link** — аргумент командной строки, открывающий TUI сразу на конкретном экране: `gitlab-tui mr 123`.

**Manual Refresh** — ручное обновление данных с GitLab API по клавише `r`. Авто-поллинг не используется.
