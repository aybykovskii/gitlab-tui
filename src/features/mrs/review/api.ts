import { createHash } from 'node:crypto'
import { appendFileSync } from 'node:fs'
import type { GitLabClient } from '../../../core/gitlab/index.js'
import type { CommentPosition, LineRange } from '../diff/position.js'
import type { DraftComment, DraftNotesAPI } from './session.js'
import type { InstantCommentsAPI } from './instant.js'
import type { ThreadActionsAPI } from './threadActions.js'

// GitLab line_code format: SHA1(filePath)_oldLine_newLine
function fileLineCode(filePath: string, oldLine: number | null, newLine: number | null): string {
  const hash = createHash('sha1').update(filePath).digest('hex')
  return `${hash}_${oldLine ?? 0}_${newLine ?? 0}`
}

function parseLineCode(lc: string): { oldLine: number | null; newLine: number | null } {
  // strip the 40-char hex hash + underscore, then split "oldLine_newLine"
  const rest = lc.slice(41)
  const [oldStr, newStr] = rest.split('_')
  return {
    oldLine: oldStr ? (Number(oldStr) || null) : null,
    newLine: newStr ? (Number(newStr) || null) : null,
  }
}

function makeHeaders(token: string) {
  return { 'PRIVATE-TOKEN': token, 'Content-Type': 'application/json' }
}

async function request<T>(
  baseUrl: string,
  token: string,
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  if (process.env['GITLAB_TUI_DEBUG']) {
    appendFileSync('/tmp/gitlab-tui.log', `${method} ${path}\n${JSON.stringify(body, null, 2)}\n\n`)
  }
  const res = await fetch(`${baseUrl}/api/v4${path}`, {
    method,
    headers: makeHeaders(token),
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const text = await res.text()
    if (process.env['GITLAB_TUI_DEBUG']) appendFileSync('/tmp/gitlab-tui.log', `ERROR ${res.status}: ${text}\n\n`)
    throw new Error(`GitLab API error ${res.status}: ${text}`)
  }
  if (res.status === 204) return undefined as T
  const data = await res.json() as T
  if (process.env['GITLAB_TUI_DEBUG'] && method !== 'GET') {
    appendFileSync('/tmp/gitlab-tui.log', `RESPONSE: ${JSON.stringify(data, null, 2)}\n\n`)
  }
  return data
}

function rawToDraft(raw: Record<string, unknown>): DraftComment {
  const rawPos = raw.position as Record<string, unknown> | null | undefined

  let lineRange: LineRange | undefined
  if (rawPos?.line_range) {
    const lr = rawPos.line_range as Record<string, unknown>
    const start = lr.start as Record<string, unknown> | undefined
    const end = lr.end as Record<string, unknown> | undefined
    const s = parseLineCode((start?.line_code as string | undefined) ?? '')
    const e = parseLineCode((end?.line_code as string | undefined) ?? '')
    lineRange = {
      startOldLine: s.oldLine,
      startNewLine: s.newLine,
      endOldLine: e.oldLine,
      endNewLine: e.newLine,
    }
  }

  const position: CommentPosition | null = rawPos
    ? {
        baseSha: String(rawPos.base_sha ?? ''),
        headSha: String(rawPos.head_sha ?? ''),
        startSha: String(rawPos.start_sha ?? ''),
        oldPath: String(rawPos.old_path ?? ''),
        newPath: String(rawPos.new_path ?? ''),
        oldLine: rawPos.old_line != null ? Number(rawPos.old_line) : null,
        newLine: rawPos.new_line != null ? Number(rawPos.new_line) : null,
        positionType: 'text',
        ...(lineRange ? { lineRange } : {}),
      }
    : null
  return {
    id: raw.id as number,
    body: String(raw.note ?? raw.body ?? ''),
    position,
  }
}

export function createDraftNotesAPI(
  baseUrl: string,
  token: string,
  projectPath: string,
  mrIid: number,
): DraftNotesAPI {
  const projectId = encodeURIComponent(projectPath)
  const endpoint = `/projects/${projectId}/merge_requests/${mrIid}/draft_notes`

  async function create(body: string, position?: CommentPosition | null): Promise<DraftComment> {
    const payload: Record<string, unknown> = { note: body }
    if (position) {
      const pos: Record<string, unknown> = {
        base_sha: position.baseSha,
        head_sha: position.headSha,
        start_sha: position.startSha,
        old_path: position.oldPath,
        new_path: position.newPath,
        old_line: position.oldLine,
        new_line: position.newLine,
        position_type: position.positionType,
      }
      if (position.lineRange) {
        const lr = position.lineRange
        pos['line_range'] = {
          start: {
            type: lr.startNewLine != null ? 'new' : 'old',
            line_code: fileLineCode(position.newPath, lr.startOldLine, lr.startNewLine),
          },
          end: {
            type: lr.endNewLine != null ? 'new' : 'old',
            line_code: fileLineCode(position.newPath, lr.endOldLine, lr.endNewLine),
          },
        }
      }
      payload.position = pos
    }
    const raw = await request<Record<string, unknown>>(baseUrl, token, 'POST', endpoint, payload)
    return rawToDraft(raw)
  }

  async function createReply(discussionId: string, body: string): Promise<DraftComment> {
    const raw = await request<Record<string, unknown>>(baseUrl, token, 'POST', endpoint, {
      note: body,
      discussion_id: discussionId,
    })
    return rawToDraft(raw)
  }

  async function list(): Promise<DraftComment[]> {
    const raws = await request<Record<string, unknown>[]>(baseUrl, token, 'GET', endpoint)
    return raws.map(rawToDraft)
  }

  async function publishAll(summary?: string): Promise<void> {
    await request(baseUrl, token, 'POST', `${endpoint}/bulk_publish`, summary ? { note: summary } : {})
  }

  async function remove(id: number): Promise<void> {
    await request(baseUrl, token, 'DELETE', `${endpoint}/${id}`)
  }

  async function removeAll(): Promise<void> {
    const drafts = await list()
    await Promise.all(drafts.map((d) => remove(d.id)))
  }

  return { create, createReply, list, publishAll, remove, removeAll }
}

export function createInstantCommentsAPI(
  client: GitLabClient,
  baseUrl: string,
  token: string,
  projectPath: string,
  mrIid: number,
): InstantCommentsAPI {
  const projectId = encodeURIComponent(projectPath)

  async function postInlineComment(body: string, position: CommentPosition): Promise<void> {
    const pos: Record<string, unknown> = {
      base_sha: position.baseSha,
      head_sha: position.headSha,
      start_sha: position.startSha,
      old_path: position.oldPath,
      new_path: position.newPath,
      position_type: 'text',
    }
    if (position.oldLine != null) pos['old_line'] = position.oldLine
    if (position.newLine != null) pos['new_line'] = position.newLine
    if (position.lineRange) {
      const lr = position.lineRange
      pos['line_range'] = {
        start: {
          type: lr.startNewLine != null ? 'new' : 'old',
          line_code: fileLineCode(position.newPath, lr.startOldLine, lr.startNewLine),
        },
        end: {
          type: lr.endNewLine != null ? 'new' : 'old',
          line_code: fileLineCode(position.newPath, lr.endOldLine, lr.endNewLine),
        },
      }
    }
    await request(baseUrl, token, 'POST',
      `/projects/${projectId}/merge_requests/${mrIid}/discussions`,
      { body, position: pos },
    )
  }

  async function postMRComment(body: string): Promise<void> {
    await client.MergeRequestNotes.create(projectPath, mrIid, body)
  }

  return { postInlineComment, postMRComment }
}

export function createThreadActionsAPIImpl(
  client: GitLabClient,
  baseUrl: string,
  token: string,
  projectPath: string,
  mrIid: number,
): ThreadActionsAPI {
  const projectId = encodeURIComponent(projectPath)
  const mrBase = `/projects/${projectId}/merge_requests/${mrIid}`

  async function replyToThread(discussionId: string, body: string): Promise<void> {
    await client.MergeRequestDiscussions.addNote(projectPath, mrIid, discussionId, body)
  }

  async function resolveThread(discussionId: string, resolved: boolean): Promise<void> {
    await request(baseUrl, token, 'PUT', `${mrBase}/discussions/${discussionId}`, { resolved })
  }

  return { replyToThread, resolveThread }
}
