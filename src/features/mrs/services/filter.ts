import type { MR } from './types.js'

export function filterMRs (mrs: MR[], query: string): MR[] {
  if (!query) return mrs
  const q = query.toLowerCase()
  return mrs.filter((mr) => mr.title.toLowerCase().includes(q))
}
