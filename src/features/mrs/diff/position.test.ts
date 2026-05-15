import { describe, it, expect } from 'vitest'
import { buildDiffPosition } from './position.js'

const refs = {
  baseSha: 'abc123',
  headSha: 'def456',
  startSha: 'ghi789',
}

describe('buildDiffPosition', () => {
  it('builds position for an added line', () => {
    const pos = buildDiffPosition(refs, {
      oldPath: 'src/foo.ts',
      newPath: 'src/foo.ts',
      oldLineNo: null,
      newLineNo: 42,
    })

    expect(pos).toEqual({
      baseSha: 'abc123',
      headSha: 'def456',
      startSha: 'ghi789',
      oldPath: 'src/foo.ts',
      newPath: 'src/foo.ts',
      oldLine: null,
      newLine: 42,
      positionType: 'text',
    })
  })

  it('builds position for a removed line', () => {
    const pos = buildDiffPosition(refs, {
      oldPath: 'src/foo.ts',
      newPath: 'src/foo.ts',
      oldLineNo: 10,
      newLineNo: null,
    })

    expect(pos.oldLine).toBe(10)
    expect(pos.newLine).toBeNull()
  })
})
