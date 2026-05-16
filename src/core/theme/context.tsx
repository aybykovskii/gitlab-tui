import React, { createContext, useContext } from 'react'
import type { Theme } from './types.js'
import { resolveTheme } from './resolver.js'

const ThemeContext = createContext<Theme>(resolveTheme(undefined))

interface Props {
  theme?: string | Partial<Theme>
  children: React.ReactNode
}

export function ThemeProvider({ theme, children }: Props) {
  return <ThemeContext.Provider value={resolveTheme(theme)}>{children}</ThemeContext.Provider>
}

export function useTheme(): Theme {
  return useContext(ThemeContext)
}
