import { describe, it, expect, vi } from 'vitest'
import { createThreadActionsService } from './threadActions.js'
import type { ThreadActionsAPI } from './threadActions.js'

function makeAPI(overrides: Partial<ThreadActionsAPI> = {}): ThreadActionsAPI {
  return {
    replyToThread: vi.fn().mockResolvedValue(undefined),
    resolveThread: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  }
}

describe('replyToThread', () => {
  it('delegates to api with discussionId and body', async () => {
    const api = makeAPI()
    const service = createThreadActionsService(api)

    await service.replyToThread('disc-123', 'Agreed, will fix')

    expect(api.replyToThread).toHaveBeenCalledWith('disc-123', 'Agreed, will fix')
  })
})

describe('resolveThread', () => {
  it('delegates to api with resolved=true', async () => {
    const api = makeAPI()
    const service = createThreadActionsService(api)

    await service.resolveThread('disc-123', 456, true)

    expect(api.resolveThread).toHaveBeenCalledWith('disc-123', 456, true)
  })

  it('delegates to api with resolved=false', async () => {
    const api = makeAPI()
    const service = createThreadActionsService(api)

    await service.resolveThread('disc-123', 456, false)

    expect(api.resolveThread).toHaveBeenCalledWith('disc-123', 456, false)
  })
})
