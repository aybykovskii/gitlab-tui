import type { DiffFile, Thread, DiffPosition } from './types.js'

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

export function mapDiffFile(raw: RawDiff): DiffFile {
  return {
    oldPath: raw.old_path,
    newPath: raw.new_path,
    addedLines: raw.added_lines,
    removedLines: raw.removed_lines,
    isNew: raw.new_file,
    isDeleted: raw.deleted_file,
    isRenamed: raw.renamed_file,
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

  return {
    id: raw.id,
    resolved: raw.resolved ?? false,
    author: { name: first?.author.name ?? '', username: first?.author.username ?? '' },
    firstNote: first?.body ?? '',
    position,
  }
}
