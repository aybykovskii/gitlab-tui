import { describe, it, expect } from 'vitest'
import { resolveTheme } from './resolver.js'
import type { Theme } from './types.js'

describe('resolveTheme', () => {
  it('returns default preset when called with undefined', () => {
    const theme = resolveTheme(undefined)
    expect(theme.primary).toBeDefined()
    expect(theme.secondary).toBeDefined()
    expect(theme.success).toBeDefined()
    expect(theme.warning).toBeDefined()
    expect(theme.error).toBeDefined()
    expect(theme.muted).toBeDefined()
    expect(theme.border).toBeDefined()
  })

  it('returns dracula preset when called with "dracula"', () => {
    const theme = resolveTheme('dracula')
    expect(theme.primary).toBe('#bd93f9')
    expect(theme.error).toBe('#ff5555')
  })

  it('returns nord preset when called with "nord"', () => {
    const theme = resolveTheme('nord')
    expect(theme.primary).toBe('#88c0d0')
    expect(theme.success).toBe('#a3be8c')
  })

  it('falls back to default when preset name is unknown', () => {
    const defaultTheme = resolveTheme(undefined)
    const theme = resolveTheme('nonexistent')
    expect(theme).toEqual(defaultTheme)
  })

  it('merges partial overrides onto default, changing only specified tokens', () => {
    const defaultTheme = resolveTheme(undefined)
    const theme = resolveTheme({ primary: '#ff0000' })
    expect(theme.primary).toBe('#ff0000')
    expect(theme.secondary).toBe(defaultTheme.secondary)
    expect(theme.success).toBe(defaultTheme.success)
    expect(theme.border).toBe(defaultTheme.border)
  })
})
