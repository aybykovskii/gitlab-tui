import { describe, it, expect } from 'vitest'
import { render } from 'ink-testing-library'
import React from 'react'
import { ThemeProvider } from '../../../core/theme/index.js'
import { Navigator } from '../../../core/navigation/index.js'
import { MRDetailScreen } from './MRDetailScreen.js'
import type { Screen } from '../../../core/navigation/types.js'
import type { MR } from '../services/types.js'

const mockAccount = { name: 'work', url: 'https://gitlab.example.com', token: 'tok' }

const mockMR: MR = {
  iid: 42,
  title: 'Fix critical bug',
  state: 'opened',
  author: { name: 'alice', username: 'alice' },
  sourceBranch: 'fix/bug',
  targetBranch: 'main',
  pipeline: null,
  webUrl: 'https://gitlab.example.com/myorg/myrepo/-/merge_requests/42',
}

const screen: Screen = {
  id: 'mr-detail',
  component: MRDetailScreen,
  props: { mr: mockMR, account: mockAccount, projectPath: 'myorg/myrepo' },
}

function renderScreen() {
  return render(
    <ThemeProvider>
      <Navigator initialScreen={screen} />
    </ThemeProvider>,
  )
}

describe('MRDetailScreen', () => {
  it('shows MR title in left panel', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('Fix critical bug')
  })

  it('shows MR number in left panel', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('!42')
  })

  it('shows loading state in right panel while MR detail is fetched', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('Loading')
  })

  it('shows MR state in left panel', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('opened')
  })
})
