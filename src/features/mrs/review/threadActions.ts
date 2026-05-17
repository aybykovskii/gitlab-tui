export interface ThreadActionsAPI {
  replyToThread(discussionId: string, body: string): Promise<void>
  resolveThread(discussionId: string, noteId: number, resolved: boolean): Promise<void>
}

export function createThreadActionsService (api: ThreadActionsAPI) {
  async function replyToThread (discussionId: string, body: string): Promise<void> {
    return api.replyToThread(discussionId, body)
  }

  async function resolveThread (discussionId: string, noteId: number, resolved: boolean): Promise<void> {
    return api.resolveThread(discussionId, noteId, resolved)
  }

  return { replyToThread, resolveThread }
}
