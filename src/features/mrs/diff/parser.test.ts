import { describe, expect, it } from 'vitest'

import { parseDiff } from './parser.js'

const SIMPLE_DIFF = `@@ -1,4 +1,5 @@
 const a = 1
-const b = 2
+const b = 3
+const c = 4
 const d = 5
`

describe('parseDiff', () => {
  it('context line appears on both left and right', () => {
    const rows = parseDiff(SIMPLE_DIFF)
    const contextRows = rows.filter((r) => r.left?.type === 'context')

    expect(contextRows.length).toBeGreaterThan(0)
    contextRows.forEach((r) => {
      expect(r.left?.type).toBe('context')
      expect(r.right?.type).toBe('context')
      expect(r.left?.content).toBe(r.right?.content)
    })
  })

  it('removed line appears only on the left', () => {
    const rows = parseDiff(SIMPLE_DIFF)
    const removedRows = rows.filter((r) => r.left?.type === 'removed')

    expect(removedRows.length).toBeGreaterThan(0)
    removedRows.forEach((r) => {
      expect(r.left?.type).toBe('removed')
      // right is either null or an added line (paired), never context/removed
      if (r.right !== null) expect(r.right.type).toBe('added')
    })
  })

  it('added line appears only on the right', () => {
    const rows = parseDiff(SIMPLE_DIFF)
    const addedRows = rows.filter((r) => r.right?.type === 'added' && r.left?.type !== 'removed')

    expect(addedRows.length).toBeGreaterThan(0)
    addedRows.forEach((r) => {
      expect(r.left).toBeNull()
    })
  })

  it('tracks old and new line numbers correctly', () => {
    const rows = parseDiff(SIMPLE_DIFF)

    // First context line: old=1, new=1
    const firstContext = rows.find((r) => r.left?.type === 'context')
    expect(firstContext?.left?.oldLineNo).toBe(1)
    expect(firstContext?.right?.newLineNo).toBe(1)

    // Removed line: has oldLineNo, no newLineNo
    const removedRow = rows.find((r) => r.left?.type === 'removed')
    expect(removedRow?.left?.oldLineNo).toBe(2)
    expect(removedRow?.left?.newLineNo).toBeNull()

    // Paired added line: has newLineNo, no oldLineNo
    expect(removedRow?.right?.newLineNo).toBe(2)
    expect(removedRow?.right?.oldLineNo).toBeNull()
  })

  it('adjacent removed and added lines are paired in the same row', () => {
    const rows = parseDiff(SIMPLE_DIFF)
    const pairedRows = rows.filter(
      (r) => r.left?.type === 'removed' && r.right?.type === 'added',
    )

    expect(pairedRows.length).toBe(1)
    expect(pairedRows[0].left?.content).toBe('const b = 2')
    expect(pairedRows[0].right?.content).toBe('const b = 3')
  })
})
