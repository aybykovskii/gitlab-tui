import type { DiffFile, Thread, ThreadNote, DiffPosition } from './types.js'

interface RawDiff {
  old_path: string
  new_path: string
  added_lines: number
  removed_lines: number
  new_file: boolean
  deleted_file: boolean
  renamed_file: boolean
  diff: string
}

interface RawNote {
  id: number
  author: { name: string; username: string }
  body: string
  position: {
    new_path?: string
    old_line?: number | null
    new_line?: number | null
  } | null
}

interface RawThread {
  id: string
  resolved?: boolean
  notes: RawNote[]
}

function countDiffLines(diff: string): { added: number; removed: number } {
  let added = 0
  let removed = 0
  for (const line of diff.split('\n')) {
    if (line.startsWith('+') && !line.startsWith('+++')) added++
    else if (line.startsWith('-') && !line.startsWith('---')) removed++
  }
  return { added, removed }
}

export function mapDiffFile(raw: RawDiff): DiffFile {
  const counts = countDiffLines(raw.diff)
  return {
    oldPath: raw.old_path,
    newPath: raw.new_path,
    addedLines: raw.added_lines ?? counts.added,
    removedLines: raw.removed_lines ?? counts.removed,
    isNew: raw.new_file,
    isDeleted: raw.deleted_file,
    isRenamed: raw.renamed_file,
    rawDiff: raw.diff,
  }
}

export function mapThread(raw: RawThread): Thread {
  const first = raw.notes[0]
  const pos = first?.position

  const position: DiffPosition | null = pos
    ? {
        filePath: pos.new_path ?? '',
        oldLine: pos.old_line ?? null,
        newLine: pos.new_line ?? null,
      }
    : null

  const notes: ThreadNote[] = raw.notes.map((n) => ({
    id: n.id,
    author: { name: n.author.name, username: n.author.username },
    body: n.body,
  }))

  return {
    id: raw.id,
    resolved: raw.resolved ?? false,
    author: { name: first?.author.name ?? '', username: first?.author.username ?? '' },
    firstNote: first?.body ?? '',
    position,
    notes,
  }
}
