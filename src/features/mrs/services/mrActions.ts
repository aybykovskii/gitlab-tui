import type { GitLabClient } from '../../../core/gitlab/index.js'

export interface MRActionsAPI {
  approveMR(iid: number): Promise<void>
  mergeMR(iid: number): Promise<void>
}

export function createMRActionsService(api: MRActionsAPI) {
  async function approveMR(iid: number): Promise<void> {
    return api.approveMR(iid)
  }

  async function mergeMR(iid: number): Promise<void> {
    return api.mergeMR(iid)
  }

  return { approveMR, mergeMR }
}

export function createMRActionsAPIImpl(
  client: GitLabClient,
  projectPath: string,
): MRActionsAPI {
  const base = `${(client as unknown as { host: string }).host}/api/v4`
  const token = (client as unknown as { token: string }).token
  const headers = { 'PRIVATE-TOKEN': token, 'Content-Type': 'application/json' }
  const projectId = encodeURIComponent(projectPath)

  async function request(method: string, path: string): Promise<void> {
    const res = await fetch(`${base}${path}`, { method, headers, body: '{}' })
    if (!res.ok) throw new Error(`GitLab API error ${res.status}: ${await res.text()}`)
  }

  async function approveMR(iid: number): Promise<void> {
    await request('POST', `/projects/${projectId}/merge_requests/${iid}/approve`)
  }

  async function mergeMR(iid: number): Promise<void> {
    await request('POST', `/projects/${projectId}/merge_requests/${iid}/merge`)
  }

  return { approveMR, mergeMR }
}
