import type { GitLabClient } from '../../../core/gitlab/index.js'

export interface MRActionsAPI {
  approveMR(iid: number): Promise<void>
  mergeMR(iid: number): Promise<void>
}

export function createMRActionsService (api: MRActionsAPI) {
  async function approveMR (iid: number): Promise<void> {
    return api.approveMR(iid)
  }

  async function mergeMR (iid: number): Promise<void> {
    return api.mergeMR(iid)
  }

  return { approveMR, mergeMR }
}

export function createMRActionsAPIImpl (
  client: GitLabClient,
  projectPath: string,
): MRActionsAPI {
  async function approveMR (iid: number): Promise<void> {
    await client.MergeRequestApprovals.approve(projectPath, iid)
  }

  async function mergeMR (iid: number): Promise<void> {
    await client.MergeRequests.merge(projectPath, iid)
  }

  return { approveMR, mergeMR }
}
