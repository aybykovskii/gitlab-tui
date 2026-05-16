import { createHash } from 'node:crypto'
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

function toDiscussionPosition(position: CommentPosition): Record<string, unknown> {
  const pos: Record<string, unknown> = {
    baseSha: position.baseSha,
    headSha: position.headSha,
    startSha: position.startSha,
    oldPath: position.oldPath,
    newPath: position.newPath,
    positionType: position.positionType,
  }
  if (position.oldLine != null) pos['oldLine'] = position.oldLine
  if (position.newLine != null) pos['newLine'] = position.newLine
  if (position.lineRange) {
    const lr = position.lineRange
    pos['lineRange'] = {
      start: {
        type: lr.startNewLine != null ? 'new' : 'old',
        lineCode: fileLineCode(position.newPath, lr.startOldLine, lr.startNewLine),
      },
      end: {
        type: lr.endNewLine != null ? 'new' : 'old',
        lineCode: fileLineCode(position.newPath, lr.endOldLine, lr.endNewLine),
      },
    }
  }
  return pos
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
  client: GitLabClient,
  projectPath: string,
  mrIid: number,
): DraftNotesAPI {
  async function create(body: string, position?: CommentPosition | null): Promise<DraftComment> {
    const options = position ? { position: toDiscussionPosition(position) } : undefined
    const raw = await client.MergeRequestDraftNotes.create(projectPath, mrIid, body, options as never)
    return rawToDraft(raw as Record<string, unknown>)
  }

  async function createReply(discussionId: string, body: string): Promise<DraftComment> {
    const raw = await client.MergeRequestDraftNotes.create(projectPath, mrIid, body, {
      inReplyToDiscussionId: discussionId,
    } as never)
    return rawToDraft(raw as Record<string, unknown>)
  }

  async function list(): Promise<DraftComment[]> {
    const raws = await client.MergeRequestDraftNotes.all(projectPath, mrIid)
    return (raws as Record<string, unknown>[]).map(rawToDraft)
  }

  async function publishAll(summary?: string): Promise<void> {
    if (summary) await client.MergeRequestNotes.create(projectPath, mrIid, summary)
    await client.MergeRequestDraftNotes.publishBulk(projectPath, mrIid)
  }

  async function remove(id: number): Promise<void> {
    await client.MergeRequestDraftNotes.remove(projectPath, mrIid, id)
  }

  async function removeAll(): Promise<void> {
    const drafts = await list()
    await Promise.all(drafts.map((d) => remove(d.id)))
  }

  return { create, createReply, list, publishAll, remove, removeAll }
}

export function createInstantCommentsAPI(
  client: GitLabClient,
  projectPath: string,
  mrIid: number,
): InstantCommentsAPI {
  async function postInlineComment(body: string, position: CommentPosition): Promise<void> {
    await client.MergeRequestDiscussions.create(projectPath, mrIid, body, {
      position: toDiscussionPosition(position),
    } as never)
  }

  async function postMRComment(body: string): Promise<void> {
    await client.MergeRequestNotes.create(projectPath, mrIid, body)
  }

  return { postInlineComment, postMRComment }
}

export function createThreadActionsAPIImpl(
  client: GitLabClient,
  projectPath: string,
  mrIid: number,
): ThreadActionsAPI {
  async function replyToThread(discussionId: string, body: string): Promise<void> {
    await client.MergeRequestDiscussions.addNote(projectPath, mrIid, discussionId, body)
  }

  async function resolveThread(discussionId: string, noteId: number, resolved: boolean): Promise<void> {
    await client.MergeRequestDiscussions.editNote(projectPath, mrIid, discussionId, noteId, { resolved })
  }

  return { replyToThread, resolveThread }
}
