import { describe, expect, it, vi } from 'vitest'

import type { MRActionsAPI } from './mrActions.js'
import { createMRActionsService } from './mrActions.js'

function makeAPI (overrides: Partial<MRActionsAPI> = {}): MRActionsAPI {
  return {
    approveMR: vi.fn().mockResolvedValue(undefined),
    mergeMR: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  }
}

describe('approveMR', () => {
  it('delegates to api with iid', async () => {
    const api = makeAPI()
    const service = createMRActionsService(api)

    await service.approveMR(42)

    expect(api.approveMR).toHaveBeenCalledWith(42)
  })
})

describe('mergeMR', () => {
  it('delegates to api with iid', async () => {
    const api = makeAPI()
    const service = createMRActionsService(api)

    await service.mergeMR(7)

    expect(api.mergeMR).toHaveBeenCalledWith(7)
  })
})
