import type { GitLabClient } from './client.js'

export interface ProjectSummary {
  accountName: string
  projectPath: string
}

export async function listUserProjects (
  client: GitLabClient,
  accountName: string,
  limit = 10,
): Promise<ProjectSummary[]> {
  const projects = await client.Projects.all({
    membership: true,
    perPage: limit,
    maxPages: 1,
  })
  return (projects as Record<string, unknown>[]).map((p) => ({
    accountName,
    projectPath: String(p.path_with_namespace ?? ''),
  }))
}
