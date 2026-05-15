import { describe, it, expect, vi } from 'vitest'
import { createInstantCommentService } from './instant.js'
import type { InstantCommentsAPI } from './instant.js'
import type { CommentPosition } from '../diff/position.js'

const position: CommentPosition = {
  baseSha: 'abc',
  headSha: 'def',
  startSha: 'ghi',
  oldPath: 'src/foo.ts',
  newPath: 'src/foo.ts',
  oldLine: null,
  newLine: 10,
  positionType: 'text',
}

function makeAPI(overrides: Partial<InstantCommentsAPI> = {}): InstantCommentsAPI {
  return {
    postInlineComment: vi.fn().mockResolvedValue(undefined),
    postMRComment: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  }
}

describe('postInlineComment', () => {
  it('delegates to api with position and body', async () => {
    const api = makeAPI()
    const service = createInstantCommentService(api)

    await service.postInlineComment(position, 'Looks wrong')

    expect(api.postInlineComment).toHaveBeenCalledWith('Looks wrong', position)
  })
})

describe('postMRComment', () => {
  it('delegates to api with body', async () => {
    const api = makeAPI()
    const service = createInstantCommentService(api)

    await service.postMRComment('Great work!')

    expect(api.postMRComment).toHaveBeenCalledWith('Great work!')
  })
})
