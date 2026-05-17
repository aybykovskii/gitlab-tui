import React from 'react'
import { render } from 'ink-testing-library'
import { describe, expect, it, vi } from 'vitest'

import { DiffView } from './DiffView.js'
import type { DiffRefs } from './position.js'

const refs: DiffRefs = { baseSha: 'abc', headSha: 'def', startSha: 'ghi' }

const RAW_DIFF = `@@ -1,3 +1,3 @@
 const a = 1
-const b = 2
+const b = 99
 const c = 3
`

describe('DiffView', () => {
  it('renders content from both old and new sides', () => {
    const { lastFrame } = render(
      <DiffView filePath="src/foo.ts" rawDiff={RAW_DIFF} refs={refs} onAddComment={vi.fn()} onBack={vi.fn()} />,
    )
    const frame = lastFrame() ?? ''
    expect(frame).toContain('const b = 2')
    expect(frame).toContain('const b = 99')
  })

  it('does not immediately call onAddComment when c is pressed — opens input first', () => {
    const onAddComment = vi.fn()
    const { stdin } = render(
      <DiffView filePath="src/foo.ts" rawDiff={RAW_DIFF} refs={refs} onAddComment={onAddComment} onBack={vi.fn()} />,
    )

    stdin.write('c')

    // Comment is only sent after body is entered and submitted
    expect(onAddComment).not.toHaveBeenCalled()
  })

  it('calls onBack when q is pressed', () => {
    const onBack = vi.fn()
    const { stdin } = render(
      <DiffView filePath="src/foo.ts" rawDiff={RAW_DIFF} refs={refs} onAddComment={vi.fn()} onBack={onBack} />,
    )

    stdin.write('q')

    expect(onBack).toHaveBeenCalledOnce()
  })
})
