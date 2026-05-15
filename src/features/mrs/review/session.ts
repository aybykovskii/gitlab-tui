import type { CommentPosition } from '../diff/position.js'

export interface DraftComment {
  id: number
  body: string
  position: CommentPosition | null
}

export interface DraftNotesAPI {
  create(body: string, position?: CommentPosition | null): Promise<DraftComment>
  list(): Promise<DraftComment[]>
  publishAll(summary?: string): Promise<void>
  remove(id: number): Promise<void>
}

export function createReviewSession(api: DraftNotesAPI) {
  async function addDraftComment(
    position: CommentPosition | null,
    body: string,
  ): Promise<DraftComment> {
    return api.create(body, position)
  }

  async function getDraftComments(): Promise<DraftComment[]> {
    return api.list()
  }

  async function publishReview(summary?: string): Promise<void> {
    return api.publishAll(summary)
  }

  async function discardDraft(id: number): Promise<void> {
    return api.remove(id)
  }

  return { addDraftComment, getDraftComments, publishReview, discardDraft }
}
