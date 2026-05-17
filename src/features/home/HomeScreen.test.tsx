import React from 'react'
import { render } from 'ink-testing-library'
import { describe, expect, it, vi } from 'vitest'

import type { Config } from '../../core/config/types.js'
import { Navigator } from '../../core/navigation/index.js'
import type { Screen } from '../../core/navigation/types.js'
import { ThemeProvider } from '../../core/theme/index.js'

import { HomeScreen } from './HomeScreen.js'

const mockConfig: Config = {
  accounts: [],
  defaultAccount: '',
  recentProjects: [],
  editor: 'nvim',
}

const mockConfigManager = { saveConfig: vi.fn() }

const homeScreen: Screen = {
  id: 'home',
  component: HomeScreen,
  props: { config: mockConfig, configManager: mockConfigManager },
}

function renderHome () {
  return render(
    (
      <ThemeProvider>
        <Navigator initialScreen={homeScreen} />
      </ThemeProvider>
    ),
  )
}

describe('HomeScreen', () => {
  it('shows app name in left panel', () => {
    const { lastFrame } = renderHome()
    expect(lastFrame()).toContain('gitlab-tui')
  })

  it('shows project picker in right panel', () => {
    const { lastFrame } = renderHome()
    expect(lastFrame()).toContain('Select project')
  })
})
