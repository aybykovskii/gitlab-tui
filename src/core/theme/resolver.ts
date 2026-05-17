import { presets } from './presets.js'
import type { Theme } from './types.js'

export function resolveTheme (config?: string | Partial<Theme>): Theme {
  if (config === undefined) return presets.default
  if (typeof config === 'string') return presets[config] ?? presets.default
  return { ...presets.default, ...config }
}
