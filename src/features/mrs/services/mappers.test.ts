import { describe, it, expect } from 'vitest'
import { mapDiffFile, mapThread } from './mappers.js'

const makeRawDiff = (overrides = {}) => ({
  old_path: 'src/foo.ts',
  new_path: 'src/foo.ts',
  added_lines: 5,
  removed_lines: 2,
  new_file: false,
  deleted_file: false,
  renamed_file: false,
  diff: '',
  ...overrides,
})

const makeRawThread = (overrides = {}) => ({
  id: 'thread-1',
  notes: [
    {
      author: { name: 'Alice', username: 'alice' },
      body: 'Looks good',
      position: null,
    },
  ],
  ...overrides,
})

describe('mapThread', () => {
  it('maps resolved thread with author and first note', () => {
    const result = mapThread({ ...makeRawThread(), resolved: true } as any)
    expect(result.resolved).toBe(true)
    expect(result.author).toEqual({ name: 'Alice', username: 'alice' })
    expect(result.firstNote).toBe('Looks good')
  })

  it('maps unresolved thread', () => {
    const result = mapThread({ ...makeRawThread(), resolved: false } as any)
    expect(result.resolved).toBe(false)
  })

  it('maps inline thread with diff position', () => {
    const raw = makeRawThread({
      notes: [
        {
          author: { name: 'Alice', username: 'alice' },
          body: 'Why this?',
          position: { new_path: 'src/foo.ts', old_line: null, new_line: 42 },
        },
      ],
    })
    const result = mapThread(raw as any)
    expect(result.position).toEqual({ filePath: 'src/foo.ts', oldLine: null, newLine: 42 })
  })

  it('maps general thread without position', () => {
    const result = mapThread(makeRawThread() as any)
    expect(result.position).toBeNull()
  })
})

describe('mapDiffFile', () => {
  it('maps added and removed line counts', () => {
    const result = mapDiffFile(makeRawDiff())
    expect(result.addedLines).toBe(5)
    expect(result.removedLines).toBe(2)
  })

  it('marks file as new when new_file is true', () => {
    const result = mapDiffFile(makeRawDiff({ new_file: true }))
    expect(result.isNew).toBe(true)
    expect(result.isDeleted).toBe(false)
    expect(result.isRenamed).toBe(false)
  })

  it('marks file as deleted when deleted_file is true', () => {
    const result = mapDiffFile(makeRawDiff({ deleted_file: true }))
    expect(result.isDeleted).toBe(true)
  })

  it('marks file as renamed when renamed_file is true', () => {
    const result = mapDiffFile(
      makeRawDiff({ renamed_file: true, old_path: 'src/old.ts', new_path: 'src/new.ts' }),
    )
    expect(result.isRenamed).toBe(true)
    expect(result.oldPath).toBe('src/old.ts')
    expect(result.newPath).toBe('src/new.ts')
  })
})
