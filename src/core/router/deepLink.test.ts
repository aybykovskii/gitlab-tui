import { describe, it, expect } from 'vitest'
import { parseDeepLink } from './deepLink.js'

describe('parseDeepLink', () => {
  it('returns mr-list when no args provided', () => {
    expect(parseDeepLink([])).toEqual({ type: 'mr-list' })
  })

  it('returns mr-detail with iid when args are ["mr", "123"]', () => {
    expect(parseDeepLink(['mr', '123'])).toEqual({ type: 'mr-detail', iid: 123 })
  })

  it('returns mr-list when "mr" is given without iid', () => {
    expect(parseDeepLink(['mr'])).toEqual({ type: 'mr-list' })
  })

  it('returns mr-list when iid is not a valid number', () => {
    expect(parseDeepLink(['mr', 'abc'])).toEqual({ type: 'mr-list' })
  })
})
