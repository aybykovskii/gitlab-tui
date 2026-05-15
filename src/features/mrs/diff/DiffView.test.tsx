import { describe, it, expect, vi } from 'vitest'
import { render } from 'ink-testing-library'
import React from 'react'
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
      <DiffView
        filePath="src/foo.ts"
        rawDiff={RAW_DIFF}
        refs={refs}
        onAddComment={vi.fn()}
        onBack={vi.fn()}
      />,
    )
    const frame = lastFrame() ?? ''
    expect(frame).toContain('const b = 2')
    expect(frame).toContain('const b = 99')
  })

  it('calls onAddComment with position when c is pressed', () => {
    const onAddComment = vi.fn()
    const { stdin } = render(
      <DiffView
        filePath="src/foo.ts"
        rawDiff={RAW_DIFF}
        refs={refs}
        onAddComment={onAddComment}
        onBack={vi.fn()}
      />,
    )

    stdin.write('c')

    expect(onAddComment).toHaveBeenCalledOnce()
    const pos = onAddComment.mock.calls[0][0]
    expect(pos.baseSha).toBe('abc')
    expect(pos.positionType).toBe('text')
  })

  it('calls onBack when q is pressed', () => {
    const onBack = vi.fn()
    const { stdin } = render(
      <DiffView
        filePath="src/foo.ts"
        rawDiff={RAW_DIFF}
        refs={refs}
        onAddComment={vi.fn()}
        onBack={onBack}
      />,
    )

    stdin.write('q')

    expect(onBack).toHaveBeenCalledOnce()
  })
})
