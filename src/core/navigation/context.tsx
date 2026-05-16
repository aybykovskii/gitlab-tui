import React, { createContext, useContext } from 'react'
import type { Screen } from './types.js'
import type { Hint } from '../../ui/StatusBar.js'

export interface NavigationContextValue {
  push: (screen: Screen) => void
  pop: () => void
  setHints: (hints: Hint[]) => void
}

const NavigationContext = createContext<NavigationContextValue>({
  push: () => {},
  pop: () => {},
  setHints: () => {},
})

export function NavigationProvider({
  value,
  children,
}: {
  value: NavigationContextValue
  children: React.ReactNode
}) {
  return <NavigationContext.Provider value={value}>{children}</NavigationContext.Provider>
}

export function useNavigation(): NavigationContextValue {
  return useContext(NavigationContext)
}
