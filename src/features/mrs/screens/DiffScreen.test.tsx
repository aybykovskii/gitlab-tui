import React from 'react'
import { render } from 'ink-testing-library'
import { describe, expect, it } from 'vitest'

import { Navigator } from '../../../core/navigation/index.js'
import type { Screen } from '../../../core/navigation/types.js'
import { ThemeProvider } from '../../../core/theme/index.js'
import type { DiffFile, MRDetail } from '../services/types.js'

import { DiffScreen } from './DiffScreen.js'

const mockAccount = { name: 'work', url: 'https://gitlab.example.com', token: 'tok' }

const RAW_DIFF = `@@ -1,3 +1,3 @@
 const a = 1
-const b = 2
+const b = 99
 const c = 3
`

const mockFile: DiffFile = {
  oldPath: 'src/foo.ts',
  newPath: 'src/foo.ts',
  addedLines: 1,
  removedLines: 1,
  isNew: false,
  isDeleted: false,
  isRenamed: false,
  rawDiff: RAW_DIFF,
}

const mockMRDetail: MRDetail = {
  iid: 42,
  title: 'Fix bug',
  state: 'opened',
  author: { name: 'alice', username: 'alice' },
  sourceBranch: 'fix/bug',
  targetBranch: 'main',
  pipeline: null,
  webUrl: 'https://gitlab.example.com/myorg/myrepo/-/merge_requests/42',
  description: '',
  approvalsRequired: 1,
  approvalsLeft: 1,
  diffRefs: { baseSha: 'abc', headSha: 'def', startSha: 'ghi' },
}

const screen: Screen = {
  id: 'diff',
  component: DiffScreen,
  props: {
    files: [mockFile],
    initialFileIndex: 0,
    activeMR: mockMRDetail,
    account: mockAccount,
    projectPath: 'myorg/myrepo',
    allThreads: [],
    initialDraftComments: new Map(),
    initialDraftRangeLines: new Set(),
    initialThreadComments: new Map(),
  },
}

function renderScreen () {
  return render(
    (
      <ThemeProvider>
        <Navigator initialScreen={screen} />
      </ThemeProvider>
    ),
  )
}

describe('DiffScreen', () => {
  it('shows current file path in left panel', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('src/foo.ts')
  })

  it('highlights the active file with indicator', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('▶')
  })

  it('renders diff content in right panel', () => {
    const { lastFrame } = renderScreen()
    expect(lastFrame()).toContain('const b = 2')
    expect(lastFrame()).toContain('const b = 99')
  })
})
