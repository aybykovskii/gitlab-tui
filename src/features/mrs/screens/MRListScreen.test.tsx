import { describe, it, expect } from 'vitest'
import { render } from 'ink-testing-library'
import React from 'react'
import { ThemeProvider } from '../../../core/theme/index.js'
import { Navigator } from '../../../core/navigation/index.js'
import { MRListScreen } from './MRListScreen.js'
import type { Screen } from '../../../core/navigation/types.js'

const mockAccount = { name: 'work', url: 'https://gitlab.example.com', token: 'tok' }

const screen: Screen = {
  id: 'mr-list',
  component: MRListScreen,
  props: { account: mockAccount, projectPath: 'myorg/myrepo' },
}

function renderScreen() {
  return render(
    <ThemeProvider>
      <Navigator initialScreen={screen} />
    </ThemeProvider>,
  )
}

describe('MRListScreen', () => {
  it('shows Merge Requests highlighted in left panel', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('Merge Requests')
  })

  it('shows all three sections in left panel', () => {
    const { lastFrame } = renderScreen()
    const frame = lastFrame() ?? ''
    expect(frame).toContain('Issues')
    expect(frame).toContain('Pipelines')
  })

  it('shows MR list loading state in right panel', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('Loading MRs')
  })

  it('highlights active section with indicator', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('▶')
  })
})
