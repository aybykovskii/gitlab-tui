import { Gitlab } from '@gitbeaker/rest'
import type { Account } from '../config/types.js'

export type GitLabClient = InstanceType<typeof Gitlab>

export function createGitLabClient(account: Account): GitLabClient {
  return new Gitlab({ host: account.url, token: account.token })
}
