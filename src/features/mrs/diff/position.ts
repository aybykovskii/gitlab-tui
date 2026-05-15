export interface DiffRefs {
  baseSha: string
  headSha: string
  startSha: string
}

export interface LineInfo {
  oldPath: string
  newPath: string
  oldLineNo: number | null
  newLineNo: number | null
}

export interface CommentPosition {
  baseSha: string
  headSha: string
  startSha: string
  oldPath: string
  newPath: string
  oldLine: number | null
  newLine: number | null
  positionType: 'text'
}

export function buildDiffPosition(refs: DiffRefs, line: LineInfo): CommentPosition {
  return {
    baseSha: refs.baseSha,
    headSha: refs.headSha,
    startSha: refs.startSha,
    oldPath: line.oldPath,
    newPath: line.newPath,
    oldLine: line.oldLineNo,
    newLine: line.newLineNo,
    positionType: 'text',
  }
}
