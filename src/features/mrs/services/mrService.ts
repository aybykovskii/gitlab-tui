import type { GitLabClient } from '../../../core/gitlab/index.js'

import { mapDiffFile, mapThread } from './mappers.js'
import { formatMRTitle } from './mrTitle.js'
import type { DiffFile, MR, MRDetail, MRState, Thread } from './types.js'

export interface ListMRsOptions {
  state?: MRState | 'all'
}

export interface UpdateMRInput {
  title?: string
  description?: string
  labels?: string[]
  draft?: boolean
}

export interface CreateMRInput {
  title: string
  sourceBranch: string
  targetBranch: string
  description?: string
  labels?: string[]
  assigneeIds?: number[]
  reviewerIds?: number[]
  draft?: boolean
}

export function createMRService (client: GitLabClient, projectPath: string) {
  async function listMRs (options: ListMRsOptions = {}): Promise<MR[]> {
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

  async function getMR (iid: number): Promise<MRDetail> {
    const mr = await client.MergeRequests.show(projectPath, iid)
    const raw = mr as Record<string, unknown>
    const rawRefs = raw.diff_refs as Record<string, string> | null
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
      diffRefs: rawRefs
        ? { baseSha: rawRefs.base_sha, headSha: rawRefs.head_sha, startSha: rawRefs.start_sha }
        : null,
    }
  }

  async function getDiffFiles (iid: number): Promise<DiffFile[]> {
    const diffs = await client.MergeRequests.allDiffs(projectPath, iid)
    return (diffs as unknown[]).map((d) => mapDiffFile(d as Parameters<typeof mapDiffFile>[0]))
  }

  async function getThreads (iid: number): Promise<Thread[]> {
    const discussions = await client.MergeRequestDiscussions.all(projectPath, iid)
    return (discussions as unknown[]).map((d) => mapThread(d as Parameters<typeof mapThread>[0]))
  }

  async function getBranches (): Promise<string[]> {
    const branches = await client.Branches.all(projectPath, { perPage: 100 })
    return branches.map((b) => String(b.name))
  }

  async function createMR (input: CreateMRInput): Promise<MR> {
    const title = formatMRTitle(input.title, input.draft ?? false)
    const mr = await client.MergeRequests.create(
      projectPath,
      input.sourceBranch,
      input.targetBranch,
      title,
      {
        description: input.description,
        labels: input.labels?.join(','),
        assigneeIds: input.assigneeIds,
        reviewerIds: input.reviewerIds,
      },
    )
    return {
      iid: mr.iid,
      title: mr.title,
      state: mr.state as MRState,
      author: { name: mr.author?.name ?? '', username: mr.author?.username ?? '' },
      sourceBranch: String(mr.source_branch),
      targetBranch: String(mr.target_branch),
      webUrl: String(mr.web_url),
      pipeline: null,
    }
  }

  async function updateMR (iid: number, changes: UpdateMRInput): Promise<MRDetail> {
    const options: Record<string, unknown> = {}
    if (changes.title !== undefined || changes.draft !== undefined) {
      const currentTitle = changes.title ?? ''
      options.title = formatMRTitle(currentTitle, changes.draft ?? false)
    }
    if (changes.description !== undefined) options.description = changes.description
    if (changes.labels !== undefined) options.labels = changes.labels.join(',')

    const result = await client.MergeRequests.edit(projectPath, iid, options)
    const mr = result as unknown as Record<string, unknown>
    const raw = mr
    return {
      iid: mr.iid as number,
      title: String(mr.title ?? ''),
      state: String(mr.state ?? '') as MRState,
      author: {
        name: String((mr.author as Record<string, unknown>)?.name ?? ''),
        username: String((mr.author as Record<string, unknown>)?.username ?? ''),
      },
      sourceBranch: String(mr.source_branch ?? ''),
      targetBranch: String(mr.target_branch ?? ''),
      webUrl: String(mr.web_url ?? ''),
      pipeline: mr.head_pipeline
        ? { status: String((mr.head_pipeline as Record<string, unknown>).status ?? '') }
        : null,
      description: String(raw.description ?? ''),
      approvalsRequired: Number(raw.approvals_required ?? 0),
      approvalsLeft: Number(raw.approvals_left ?? 0),
      diffRefs: null,
    }
  }

  return { listMRs, getMR, getDiffFiles, getThreads, getBranches, createMR, updateMR }
}
