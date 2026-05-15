import type { GitLabClient } from '../../../core/gitlab/index.js'
import type { MR, MRDetail, MRState, DiffFile, Thread } from './types.js'
import { mapDiffFile, mapThread } from './mappers.js'

export interface ListMRsOptions {
  state?: MRState | 'all'
}

export function createMRService(client: GitLabClient, projectPath: string) {
  async function listMRs(options: ListMRsOptions = {}): Promise<MR[]> {
    const { state = 'opened' } = options
    const results = await client.MergeRequests.all({
      projectId: projectPath,
      state: state === 'all' ? undefined : state,
      perPage: 100,
    })

    return results.map((mr) => ({
      iid: mr.iid,
      title: mr.title,
      state: mr.state as MRState,
      author: { name: mr.author?.name ?? '', username: mr.author?.username ?? '' },
      sourceBranch: String(mr.source_branch),
      targetBranch: String(mr.target_branch),
      webUrl: String(mr.web_url),
      pipeline: mr.head_pipeline
        ? { status: String((mr.head_pipeline as Record<string, unknown>).status ?? '') }
        : null,
    }))
  }

  async function getMR(iid: number): Promise<MRDetail> {
    const mr = await client.MergeRequests.show(projectPath, iid)
    const raw = mr as Record<string, unknown>
    return {
      iid: mr.iid,
      title: mr.title,
      state: mr.state as MRState,
      author: { name: mr.author?.name ?? '', username: mr.author?.username ?? '' },
      sourceBranch: String(mr.source_branch),
      targetBranch: String(mr.target_branch),
      webUrl: String(mr.web_url),
      pipeline: mr.head_pipeline
        ? { status: String((mr.head_pipeline as Record<string, unknown>).status ?? '') }
        : null,
      description: String(raw.description ?? ''),
      approvalsRequired: Number(raw.approvals_required ?? 0),
      approvalsLeft: Number(raw.approvals_left ?? 0),
    }
  }

  async function getDiffFiles(iid: number): Promise<DiffFile[]> {
    const diffs = await client.MergeRequests.allDiffs(projectPath, iid)
    return (diffs as unknown[]).map((d) => mapDiffFile(d as Parameters<typeof mapDiffFile>[0]))
  }

  async function getThreads(iid: number): Promise<Thread[]> {
    const discussions = await client.MergeRequestDiscussions.all(projectPath, iid)
    return (discussions as unknown[]).map((d) =>
      mapThread(d as Parameters<typeof mapThread>[0]),
    )
  }

  return { listMRs, getMR, getDiffFiles, getThreads }
}
