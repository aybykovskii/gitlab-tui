import { describe, it, expect } from 'vitest'
import { filterMRs } from './filter.js'
import type { MR } from './types.js'

const mrs: MR[] = [
  {
    iid: 1,
    title: 'Add authentication flow',
    state: 'opened',
    author: { name: 'Alice', username: 'alice' },
    sourceBranch: 'feat/auth',
    targetBranch: 'main',
    webUrl: 'https://gitlab.com/ns/p/-/merge_requests/1',
    pipeline: null,
  },
  {
    iid: 2,
    title: 'Fix login bug',
    state: 'opened',
    author: { name: 'Bob', username: 'bob' },
    sourceBranch: 'fix/login',
    targetBranch: 'main',
    webUrl: 'https://gitlab.com/ns/p/-/merge_requests/2',
    pipeline: { status: 'success' },
  },
  {
    iid: 3,
    title: 'Refactor database layer',
    state: 'merged',
    author: { name: 'Alice', username: 'alice' },
    sourceBranch: 'refactor/db',
    targetBranch: 'main',
    webUrl: 'https://gitlab.com/ns/p/-/merge_requests/3',
    pipeline: { status: 'failed' },
  },
]

describe('filterMRs', () => {
  it('returns all MRs when query is empty', () => {
    expect(filterMRs(mrs, '')).toEqual(mrs)
  })

  it('filters by case-insensitive title substring', () => {
    expect(filterMRs(mrs, 'LOGIN')).toEqual([mrs[1]])
  })

  it('returns empty array when no MRs match the query', () => {
    expect(filterMRs(mrs, 'nonexistent')).toEqual([])
  })
})
