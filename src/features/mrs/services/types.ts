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
