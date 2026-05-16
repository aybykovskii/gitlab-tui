import React, { useState, useEffect, useCallback } from 'react'
import { Box, Text, useInput } from 'ink'
import TextInput from 'ink-text-input'
import { StatusBar } from '../../../ui/StatusBar.js'
import { filterMRs } from '../services/filter.js'
import type { MR, MRState } from '../services/types.js'

const STATE_CYCLE: Array<MRState | 'all'> = ['opened', 'merged', 'closed', 'all']

const PIPELINE_ICON: Record<string, string> = {
  success: '✓',
  failed: '✗',
  running: '●',
  pending: '○',
}

const VISIBLE_MRS = 12

interface Props {
  projectPath: string
  initialState?: MRState | 'all'
  onSelect: (mr: MR) => void
  loadMRs: (state: MRState | 'all') => Promise<MR[]>
  onHighlight?: (mr: MR) => void
  focused?: boolean
}

export function MRList({ projectPath, initialState = 'opened', onSelect, loadMRs, onHighlight, focused = true }: Props) {
  const [mrs, setMrs] = useState<MR[]>([])
  const [query, setQuery] = useState('')
  const [state, setState] = useState<MRState | 'all'>(initialState)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [cursor, setCursor] = useState(0)
  const [offset, setOffset] = useState(0)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await loadMRs(state)
      setMrs(result)
      setCursor(0)
      setOffset(0)
      if (result.length > 0) onHighlight?.(result[0])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load MRs')
    } finally {
      setLoading(false)
    }
  }, [state, loadMRs])

  useEffect(() => { load() }, [load])

  const filtered = filterMRs(mrs, query)

  useInput((input, key) => {
    if (key.return) {
      const mr = filtered[cursor]
      if (mr) onSelect(mr)
      return
    }
    if (input === 'j' || key.downArrow) {
      const next = Math.min(cursor + 1, filtered.length - 1)
      setCursor(next)
      if (filtered[next]) onHighlight?.(filtered[next])
      if (next >= offset + VISIBLE_MRS) setOffset(next - VISIBLE_MRS + 1)
    }
    if (input === 'k' || key.upArrow) {
      const next = Math.max(cursor - 1, 0)
      setCursor(next)
      if (filtered[next]) onHighlight?.(filtered[next])
      if (next < offset) setOffset(next)
    }
    if (input === 'r') load()
    if (input === 's') setState((prev) => {
      const idx = STATE_CYCLE.indexOf(prev)
      return STATE_CYCLE[(idx + 1) % STATE_CYCLE.length]
    })
  }, { isActive: focused })

  if (loading) return <Text dimColor>Loading MRs…</Text>
  if (error) return <Text color="red">{error}</Text>

  const visible = filtered.slice(offset, offset + VISIBLE_MRS)

  return (
    <Box flexDirection="column" gap={1}>
      <Box gap={2}>
        <Text bold>{projectPath}</Text>
        <Text dimColor>[{state}]</Text>
      </Box>
      <TextInput
        placeholder="Search MRs…"
        value={query}
        onChange={setQuery}
        onSubmit={() => {}}
      />
      {visible.length > 0 ? (
        <Box flexDirection="column">
          {visible.map((mr, i) => {
            const absIdx = offset + i
            const isCursor = absIdx === cursor
            const icon = PIPELINE_ICON[mr.pipeline?.status ?? ''] ?? '–'
            const iid = `!${mr.iid}`.padEnd(5)
            return (
              <Box key={mr.iid} flexDirection="column" marginTop={i > 0 ? 1 : 0}>
                <Text inverse={isCursor} bold={isCursor}>
                  {icon} {iid} {mr.title}
                </Text>
                <Text dimColor>        {mr.author.name}  {mr.sourceBranch} → {mr.targetBranch}</Text>
              </Box>
            )
          })}
        </Box>
      ) : (
        <Text dimColor>{query ? 'No matches' : 'No MRs'}</Text>
      )}
      <StatusBar hints={[
        { key: 'j/k', label: 'navigate' },
        { key: 'Enter', label: 'open' },
        { key: 's', label: 'cycle state' },
        { key: 'r', label: 'refresh' },
      ]} />
    </Box>
  )
}
