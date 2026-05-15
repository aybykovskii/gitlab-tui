import type { GitLabClient } from '../../../core/gitlab/index.js'
import type { CommentPosition } from '../diff/position.js'
import type { DraftComment, DraftNotesAPI } from './session.js'
import type { InstantCommentsAPI } from './instant.js'
import type { ThreadActionsAPI } from './threadActions.js'

function clientCreds(client: GitLabClient) {
  const base = `${(client as unknown as { host: string }).host}/api/v4`
  const token = (client as unknown as { token: string }).token
  const headers = { 'PRIVATE-TOKEN': token, 'Content-Type': 'application/json' }
  return { base, headers }
}

export function createDraftNotesAPI(
  client: GitLabClient,
  projectPath: string,
  mrIid: number,
): DraftNotesAPI {
  const { base, headers } = clientCreds(client)
  const projectId = encodeURIComponent(projectPath)

  async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const res = await fetch(`${base}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
    if (!res.ok) throw new Error(`GitLab API error ${res.status}: ${await res.text()}`)
    if (res.status === 204) return undefined as T
    return res.json() as Promise<T>
  }

  const endpoint = `/projects/${projectId}/merge_requests/${mrIid}/draft_notes`

  async function create(body: string, position?: CommentPosition | null): Promise<DraftComment> {
    const payload: Record<string, unknown> = { note: body }
    if (position) {
      payload.position = {
        base_sha: position.baseSha,
        head_sha: position.headSha,
        start_sha: position.startSha,
        old_path: position.oldPath,
        new_path: position.newPath,
        old_line: position.oldLine,
        new_line: position.newLine,
        position_type: position.positionType,
      }
    }
    const raw = await request<Record<string, unknown>>('POST', endpoint, payload)
    return rawToDraft(raw)
  }

  async function list(): Promise<DraftComment[]> {
    const raws = await request<Record<string, unknown>[]>('GET', endpoint)
    return raws.map(rawToDraft)
  }

  async function publishAll(summary?: string): Promise<void> {
    await request('POST', `${endpoint}/bulk_publish`, summary ? { note: summary } : {})
  }

  async function remove(id: number): Promise<void> {
    await request('DELETE', `${endpoint}/${id}`)
  }

  function rawToDraft(raw: Record<string, unknown>): DraftComment {
    return {
      id: raw.id as number,
      body: String(raw.note ?? raw.body ?? ''),
      position: null,
    }
  }

  return { create, list, publishAll, remove }
}

export function createInstantCommentsAPI(
  client: GitLabClient,
  projectPath: string,
  mrIid: number,
): InstantCommentsAPI {
  async function postInlineComment(note: string, position: CommentPosition): Promise<void> {
    const pos: Record<string, unknown> = {
      baseSha: position.baseSha,
      headSha: position.headSha,
      startSha: position.startSha,
      oldPath: position.oldPath,
      newPath: position.newPath,
      positionType: 'text',
    }
    if (position.oldLine != null) pos['oldLine'] = String(position.oldLine)
    if (position.newLine != null) pos['newLine'] = String(position.newLine)

    await client.MergeRequestDiscussions.create(projectPath, mrIid, note, {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      position: pos as any,
    })
  }

  async function postMRComment(note: string): Promise<void> {
    await client.MergeRequestNotes.create(projectPath, mrIid, note)
  }

  return { postInlineComment, postMRComment }
}

export function createThreadActionsAPIImpl(
  client: GitLabClient,
  projectPath: string,
  mrIid: number,
): ThreadActionsAPI {
  const { base, headers } = clientCreds(client)
  const projectId = encodeURIComponent(projectPath)
  const mrBase = `/projects/${projectId}/merge_requests/${mrIid}`

  async function replyToThread(discussionId: string, body: string): Promise<void> {
    await client.MergeRequestDiscussions.addNote(projectPath, mrIid, discussionId, body)
  }

  async function resolveThread(discussionId: string, resolved: boolean): Promise<void> {
    // SDK only supports note-level edits; discussion-level resolve requires the raw endpoint
    const res = await fetch(`${base}${mrBase}/discussions/${discussionId}`, {
      method: 'PUT',
      headers,
      body: JSON.stringify({ resolved }),
    })
    if (!res.ok) throw new Error(`GitLab API error ${res.status}: ${await res.text()}`)
  }

  return { replyToThread, resolveThread }
}
