import type { GitLabClient } from '../../../core/gitlab/index.js'
import type { MR, MRState } from './types.js'

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

  return { listMRs }
}
