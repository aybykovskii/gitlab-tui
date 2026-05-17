import { describe, expect, it } from 'vitest'

import { parseMRTitle } from './mrTitle.js'

describe('parseMRTitle', () => {
  it('detects Draft prefix and strips it from title', () => {
    expect(parseMRTitle('Draft: Fix the bug')).toEqual({ title: 'Fix the bug', draft: true })
  })

  it('returns draft=false and keeps title unchanged when no prefix', () => {
    expect(parseMRTitle('Fix the bug')).toEqual({ title: 'Fix the bug', draft: false })
  })
})
