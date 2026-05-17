import type { CommentPosition } from '../diff/position.js'

export interface InstantCommentsAPI {
  postInlineComment(body: string, position: CommentPosition): Promise<void>
  postMRComment(body: string): Promise<void>
}

export function createInstantCommentService (api: InstantCommentsAPI) {
  async function postInlineComment (position: CommentPosition, body: string): Promise<void> {
    return api.postInlineComment(body, position)
  }

  async function postMRComment (body: string): Promise<void> {
    return api.postMRComment(body)
  }

  return { postInlineComment, postMRComment }
}
