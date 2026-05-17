import React, { useCallback, useRef, useState } from 'react'
import { Box, useInput, useStdout } from 'ink'

import { type Hint, StatusBar } from '../../ui/StatusBar.js'
import { useTheme } from '../theme/index.js'

import { NavigationProvider } from './context.js'
import type { Screen } from './types.js'

const GLOBAL_HINTS: Hint[] = [{ key: 'q', label: 'назад' }]
const DEFAULT_LEFT_PERCENT = 30

interface Props {
  initialScreen?: Screen
  initialStack?: Screen[]
  leftColumnWidth?: number
}

export function Navigator ({ initialScreen, initialStack, leftColumnWidth = DEFAULT_LEFT_PERCENT }: Props) {
  const { stdout } = useStdout()
  const totalWidth = stdout?.columns ?? 80
  const leftWidth = Math.floor((totalWidth * leftColumnWidth) / 100)
  const rightWidth = totalWidth - leftWidth

  const startStack = initialStack ?? (initialScreen ? [initialScreen] : [])
  const [stack, setStack] = useState<Screen[]>(startStack)
  const [hints, setHints] = useState<Hint[]>([])
  const theme = useTheme()

  const push = useCallback((screen: Screen) => {
    setStack((s) => [...s, screen])
  }, [])

  const pop = useCallback(() => {
    setStack((s) => (s.length > 1 ? s.slice(0, -1) : s))
  }, [])

  useInput((input, key) => {
    if (input === 'q' || key.escape) pop()
  })

  const current = stack[stack.length - 1]
  const ScreenComponent = current.component

  return (
    <NavigationProvider value={{ push, pop, setHints }}>
      <Box flexDirection="column" width={totalWidth}>
        <Box flexGrow={1}>
          <ScreenComponent leftWidth={leftWidth} rightWidth={rightWidth} {...current.props} />
        </Box>
        <Box flexDirection="column">
          <StatusBar hints={hints} />
          <StatusBar hints={GLOBAL_HINTS} />
        </Box>
      </Box>
    </NavigationProvider>
  )
}
