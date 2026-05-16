import { describe, it, expect } from 'vitest'
import { buildInitialStack } from './stackBuilder.js'
import type { Config } from '../config/types.js'
import type { DetectedProject } from '../git/index.js'

const mockAccount = { name: 'work', url: 'https://gitlab.example.com', token: 'tok' }

const mockDetected: DetectedProject = {
  account: mockAccount,
  projectPath: 'myorg/myrepo',
  localPath: '/home/user/myrepo',
}

const configWithProject: Config = {
  accounts: [mockAccount],
  defaultAccount: 'work',
  recentProjects: [{ accountName: 'work', projectPath: 'myorg/myrepo' }],
  editor: 'nvim',
}

const emptyConfig: Config = {
  accounts: [],
  defaultAccount: '',
  recentProjects: [],
  editor: 'nvim',
}

describe('buildInitialStack', () => {
  it('returns home screen when no args and no detected project', () => {
    const stack = buildInitialStack({ args: [], config: emptyConfig })
    expect(stack).toHaveLength(1)
    expect(stack[0].id).toBe('home')
  })

  it('returns project screen when detected project and no args', () => {
    const stack = buildInitialStack({ args: [], config: emptyConfig, detected: mockDetected })
    expect(stack).toHaveLength(1)
    expect(stack[0].id).toBe('project')
    expect(stack[0].props?.project).toBe(mockDetected)
  })

  it('returns [project, mr-list] when args=["mr"] and detected project', () => {
    const stack = buildInitialStack({ args: ['mr'], config: emptyConfig, detected: mockDetected })
    expect(stack).toHaveLength(2)
    expect(stack[0].id).toBe('project')
    expect(stack[1].id).toBe('mr-list')
  })

  it('returns [project, mr-list] when args=["mrs"] and detected project', () => {
    const stack = buildInitialStack({ args: ['mrs'], config: emptyConfig, detected: mockDetected })
    expect(stack).toHaveLength(2)
    expect(stack[1].id).toBe('mr-list')
  })

  it('returns [project, mr-list with deepLinkIid] when args=["mr","42"] and detected project', () => {
    const stack = buildInitialStack({ args: ['mr', '42'], config: emptyConfig, detected: mockDetected })
    expect(stack).toHaveLength(2)
    expect(stack[1].id).toBe('mr-list')
    expect(stack[1].props?.deepLinkIid).toBe(42)
  })

  it('falls back to recent project from config when no detected project', () => {
    const stack = buildInitialStack({ args: [], config: configWithProject })
    expect(stack).toHaveLength(1)
    expect(stack[0].id).toBe('project')
    expect((stack[0].props?.project as DetectedProject).projectPath).toBe('myorg/myrepo')
  })

  it('returns [project, mr-list] from recent project when args=["mr"]', () => {
    const stack = buildInitialStack({ args: ['mr'], config: configWithProject })
    expect(stack).toHaveLength(2)
    expect(stack[0].id).toBe('project')
    expect(stack[1].id).toBe('mr-list')
  })

  it('skips project path arg and resolves section', () => {
    const stack = buildInitialStack({ args: ['myorg/myrepo', 'mr'], config: configWithProject })
    expect(stack).toHaveLength(2)
    expect(stack[0].id).toBe('project')
    expect(stack[1].id).toBe('mr-list')
  })
})
