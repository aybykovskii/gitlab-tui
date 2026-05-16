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

## UI layout

**Navigator** — корневой компонент, управляющий Navigation Stack. Рендерит Left Panel, Right Panel и Status Bar. Предоставляет NavigationContext.

**Navigation Stack** — массив Screen'ов. Текущий экран = последний элемент. `push()` — переход вглубь, `pop()` — возврат (q/Esc). Переход = полная перерисовка без анимации.

**Screen** — единица навигации: `{ id: string, component: ComponentType<ScreenProps>, props? }`. Компонент рендерит обе панели сам, координируя состояние между ними.

**Left Panel** — левая колонка: контекст уровня N-1 (замороженный, без фокуса). Обновляется реактивно при смене выбора в Right Panel (например, подсветка текущего файла). Ширина задаётся через `ui.leftColumnWidth` в Config (default: 30%).

**Right Panel** — правая колонка: активный уровень N с фокусом. Занимает оставшиеся 70% ширины терминала (или `100 - ui.leftColumnWidth`%).

**NavigationContext** — React-контекст, доступный любому компоненту внутри Screen. API: `push(screen)`, `pop()`, `setHints(hints)`.

**Status Bar** — двухстрочная полоса снизу, вне Screen'ов. Первая строка: контекстные хоткеи текущего Screen (задаются через `setHints()`). Вторая строка: глобальные хоткеи (q = назад, всегда видны).

**Theme** — набор смысловых цветовых токенов: `primary`, `secondary`, `success`, `warning`, `error`, `muted`, `border`. Загружается из `Config.theme` (имя пресета или объект с переопределениями). Доступен через `ThemeContext` + `useTheme()`. Встроенные пресеты: `default`, `dracula`, `nord`.

**Navigation Tree** — дерево экранов приложения:
- HomeScreen: Left = info о приложении, Right = выбор Project
- ProjectScreen: Left = список Project (selected), Right = выбор раздела (MRs / Issues / Pipelines)
- MRListScreen: Left = список разделов (selected), Right = список MR
- MRDetailScreen: Left = список MR (selected), Right = детали MR
- DiffScreen: Left = список файлов MR (active file highlighted), Right = Diff View
