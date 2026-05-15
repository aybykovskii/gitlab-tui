export interface ThreadActionsAPI {
  replyToThread(discussionId: string, body: string): Promise<void>
  resolveThread(discussionId: string, resolved: boolean): Promise<void>
}

export function createThreadActionsService(api: ThreadActionsAPI) {
  async function replyToThread(discussionId: string, body: string): Promise<void> {
    return api.replyToThread(discussionId, body)
  }

  async function resolveThread(discussionId: string, resolved: boolean): Promise<void> {
    return api.resolveThread(discussionId, resolved)
  }

  return { replyToThread, resolveThread }
}
