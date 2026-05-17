import React from 'react'
import { render } from 'ink-testing-library'
import { describe, expect, it } from 'vitest'

import type { Config } from '../../core/config/types.js'
import type { DetectedProject } from '../../core/git/index.js'
import { Navigator } from '../../core/navigation/index.js'
import type { Screen } from '../../core/navigation/types.js'
import { ThemeProvider } from '../../core/theme/index.js'

import { ProjectScreen } from './ProjectScreen.js'

const mockAccount = { name: 'work', url: 'https://gitlab.com', token: 'tok' }

const mockProject: DetectedProject = {
  account: mockAccount,
  projectPath: 'myorg/myrepo',
  localPath: '/home/user/myrepo',
}

const mockConfig: Config = {
  accounts: [mockAccount],
  defaultAccount: 'work',
  recentProjects: [
    { accountName: 'work', projectPath: 'myorg/myrepo' },
    { accountName: 'work', projectPath: 'myorg/other' },
  ],
  editor: 'nvim',
}

const projectScreen: Screen = {
  id: 'project',
  component: ProjectScreen,
  props: { project: mockProject, config: mockConfig },
}

function renderProject () {
  return render(
    (
      <ThemeProvider>
        <Navigator initialScreen={projectScreen} />
      </ThemeProvider>
    ),
  )
}

describe('ProjectScreen', () => {
  it('shows the selected project path in the left panel', () => {
    const { lastFrame } = renderProject()
    expect(lastFrame()).toContain('myorg/myrepo')
  })

  it('shows all three sections in the right panel', () => {
    const { lastFrame } = renderProject()
    const frame = lastFrame() ?? ''
    expect(frame).toContain('Merge Requests')
    expect(frame).toContain('Issues')
    expect(frame).toContain('Pipelines')
  })

  it('visually highlights the selected project', () => {
    const { lastFrame } = renderProject()
    expect(lastFrame()).toContain('▶')
  })
})
