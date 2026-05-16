export type MRState = 'opened' | 'closed' | 'merged' | 'locked'

export interface MR {
  iid: number
  title: string
  state: MRState
  author: { name: string; username: string }
  sourceBranch: string
  targetBranch: string
  webUrl: string
  pipeline: { status: string } | null
}

export interface DiffRefs {
  baseSha: string
  headSha: string
  startSha: string
}

export interface MRDetail extends MR {
  description: string
  approvalsRequired: number
  approvalsLeft: number
  diffRefs: DiffRefs | null
}

export interface DiffFile {
  oldPath: string
  newPath: string
  addedLines: number
  removedLines: number
  isNew: boolean
  isDeleted: boolean
  isRenamed: boolean
  rawDiff: string
}

export interface DiffPosition {
  filePath: string
  oldLine: number | null
  newLine: number | null
}

export interface ThreadNote {
  author: { name: string; username: string }
  body: string
}

export interface Thread {
  id: string
  resolved: boolean
  author: { name: string; username: string }
  firstNote: string
  position: DiffPosition | null
  notes: ThreadNote[]
}
