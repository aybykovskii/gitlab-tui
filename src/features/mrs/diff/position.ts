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

export interface LineRange {
  startOldLine: number | null
  startNewLine: number | null
  endOldLine: number | null
  endNewLine: number | null
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
  lineRange?: LineRange
}

export function buildDiffPosition(
  refs: DiffRefs,
  line: LineInfo,
  range?: LineRange,
): CommentPosition {
  return {
    baseSha: refs.baseSha,
    headSha: refs.headSha,
    startSha: refs.startSha,
    oldPath: line.oldPath,
    newPath: line.newPath,
    oldLine: line.oldLineNo,
    newLine: line.newLineNo,
    positionType: 'text',
    ...(range ? { lineRange: range } : {}),
  }
}
