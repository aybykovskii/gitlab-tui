import { describe, it, expect, vi } from 'vitest'
import { createReviewSession } from './session.js'
import type { DraftNotesAPI, DraftComment } from './session.js'
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

const makeDraft = (overrides: Partial<DraftComment> = {}): DraftComment => ({
  id: 1,
  body: 'Looks wrong',
  position,
  ...overrides,
})

function makeAPI(overrides: Partial<DraftNotesAPI> = {}): DraftNotesAPI {
  return {
    create: vi.fn().mockResolvedValue(makeDraft()),
    createReply: vi.fn().mockResolvedValue(makeDraft()),
    list: vi.fn().mockResolvedValue([]),
    publishAll: vi.fn().mockResolvedValue(undefined),
    remove: vi.fn().mockResolvedValue(undefined),
    removeAll: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  }
}

describe('addDraftComment', () => {
  it('calls api.create with body and position and returns the draft', async () => {
    const api = makeAPI()
    const session = createReviewSession(api)

    const draft = await session.addDraftComment(position, 'Looks wrong')

    expect(api.create).toHaveBeenCalledWith('Looks wrong', position)
    expect(draft).toEqual(makeDraft())
  })
})

describe('getDraftComments', () => {
  it('delegates to api.list and returns drafts', async () => {
    const drafts = [makeDraft({ id: 1 }), makeDraft({ id: 2, body: 'Another' })]
    const api = makeAPI({ list: vi.fn().mockResolvedValue(drafts) })
    const session = createReviewSession(api)

    const result = await session.getDraftComments()

    expect(api.list).toHaveBeenCalledOnce()
    expect(result).toEqual(drafts)
  })
})

describe('publishReview', () => {
  it('calls api.publishAll without summary by default', async () => {
    const api = makeAPI()
    await createReviewSession(api).publishReview()

    expect(api.publishAll).toHaveBeenCalledWith(undefined)
  })

  it('passes summary to api.publishAll when provided', async () => {
    const api = makeAPI()
    await createReviewSession(api).publishReview('LGTM')

    expect(api.publishAll).toHaveBeenCalledWith('LGTM')
  })
})

describe('discardDraft', () => {
  it('calls api.remove with the given draft id', async () => {
    const api = makeAPI()
    await createReviewSession(api).discardDraft(42)

    expect(api.remove).toHaveBeenCalledWith(42)
  })
})
