import React, { useState, useEffect, useCallback } from 'react'
import { Box, Text, useInput } from 'ink'
import TextInput from 'ink-text-input'
import SelectInput from 'ink-select-input'
import { filterMRs } from '../services/filter.js'
import type { MR, MRState } from '../services/types.js'

const STATE_CYCLE: Array<MRState | 'all'> = ['opened', 'merged', 'closed', 'all']

const PIPELINE_ICON: Record<string, string> = {
  success: '✓',
  failed: '✗',
  running: '●',
  pending: '○',
}

interface Props {
  projectPath: string
  initialState?: MRState | 'all'
  onSelect: (mr: MR) => void
  loadMRs: (state: MRState | 'all') => Promise<MR[]>
}

export function MRList({ projectPath, initialState = 'opened', onSelect, loadMRs }: Props) {
  const [mrs, setMrs] = useState<MR[]>([])
  const [query, setQuery] = useState('')
  const [state, setState] = useState<MRState | 'all'>(initialState)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      setMrs(await loadMRs(state))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load MRs')
    } finally {
      setLoading(false)
    }
  }, [state, loadMRs])

  useEffect(() => { load() }, [load])

  useInput((input) => {
    if (input === 'r') load()
    if (input === 's') setState((prev) => {
      const idx = STATE_CYCLE.indexOf(prev)
      return STATE_CYCLE[(idx + 1) % STATE_CYCLE.length]
    })
  })

  const filtered = filterMRs(mrs, query)
  const items = filtered.map((mr) => ({
    label: formatMR(mr),
    value: mr,
  }))

  if (loading) return <Text dimColor>Loading MRs…</Text>
  if (error) return <Text color="red">{error}</Text>

  return (
    <Box flexDirection="column" gap={1}>
      <Box gap={2}>
        <Text bold>{projectPath}</Text>
        <Text dimColor>[{state}]</Text>
        <Text dimColor>s: cycle state  r: refresh</Text>
      </Box>
      <TextInput
        placeholder="Search MRs…"
        value={query}
        onChange={setQuery}
        onSubmit={() => {}}
      />
      {items.length > 0 ? (
        <SelectInput items={items} onSelect={(item) => onSelect(item.value)} />
      ) : (
        <Text dimColor>{query ? 'No matches' : 'No MRs'}</Text>
      )}
    </Box>
  )
}

function formatMR(mr: MR): string {
  const pipeline = mr.pipeline ? (PIPELINE_ICON[mr.pipeline.status] ?? '?') : ' '
  return `${pipeline} !${mr.iid} ${mr.title} (${mr.sourceBranch} → ${mr.targetBranch})`
}
