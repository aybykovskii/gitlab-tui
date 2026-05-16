import { describe, it, expect, vi } from 'vitest'
import { render } from 'ink-testing-library'
import React from 'react'
import { Text } from 'ink'
import { Navigator } from './Navigator.js'
import { useNavigation } from './context.js'
import { ThemeProvider } from '../theme/index.js'
import type { Screen, ScreenProps } from './types.js'

function wrap(node: React.ReactNode) {
  return <ThemeProvider>{node}</ThemeProvider>
}

function ScreenA({ leftWidth, rightWidth }: ScreenProps) {
  return <Text>Screen A lw={leftWidth} rw={rightWidth}</Text>
}

function ScreenWithCallbacks({ onContextReady }: ScreenProps & { onContextReady?: (nav: ReturnType<typeof useNavigation>) => void }) {
  const nav = useNavigation()
  React.useEffect(() => { onContextReady?.(nav) }, [])
  return <Text>Screen with nav</Text>
}

const screenA: Screen = { id: 'a', component: ScreenA }

describe('Navigator', () => {
  it('renders the initial screen', () => {
    const { lastFrame } = render(wrap(<Navigator initialScreen={screenA} />))
    expect(lastFrame()).toContain('Screen A')
  })

  it('passes calculated leftWidth and rightWidth to the screen', () => {
    const { lastFrame } = render(wrap(<Navigator initialScreen={screenA} leftColumnWidth={25} />))
    const frame = lastFrame() ?? ''
    // ink-testing-library uses terminal width 100: left=25, right=75
    expect(frame).toContain('lw=25')
    expect(frame).toContain('rw=75')
  })

  it('always shows global q hint in status bar', () => {
    const { lastFrame } = render(wrap(<Navigator initialScreen={screenA} />))
    expect(lastFrame()).toContain('q')
    expect(lastFrame()).toContain('назад')
  })

  it('exposes push and pop via useNavigation', () => {
    const navRef = { current: null as ReturnType<typeof useNavigation> | null }
    const screen: Screen = {
      id: 'cb',
      component: (props: ScreenProps) => (
        <ScreenWithCallbacks {...props} onContextReady={(nav) => { navRef.current = nav }} />
      ),
    }
    render(wrap(<Navigator initialScreen={screen} />))
    // useNavigation context is accessible within the tree
    expect(typeof navRef.current?.push).toBe('function')
    expect(typeof navRef.current?.pop).toBe('function')
    expect(typeof navRef.current?.setHints).toBe('function')
  })
})
