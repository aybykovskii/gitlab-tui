import React, { useEffect, useState } from 'react'
import { Box, Text } from 'ink'
import SelectInput from 'ink-select-input'
import TextInput from 'ink-text-input'

import type { Account, RecentProject } from '../config/types.js'

import type { DetectedProject } from './detector.js'

interface Props {
  recentProjects: RecentProject[]
  accounts: Account[]
  onSelect: (project: DetectedProject) => void
  loadProjects?: () => Promise<{ accountName: string; projectPath: string }[]>
}

export function ProjectSelect ({ recentProjects, accounts, onSelect, loadProjects }: Props) {
  const [query, setQuery] = useState('')
  const [apiProjects, setApiProjects] = useState<{ accountName: string; projectPath: string }[]>([])
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)

  useEffect(() => {
    if (!loadProjects) return
    setLoading(true)
    loadProjects()
      .then(setApiProjects)
      .catch((e: unknown) => setLoadError(e instanceof Error ? e.message : String(e)))
      .finally(() => setLoading(false))
  }, [])

  const recentPaths = new Set(recentProjects.map((p) => p.projectPath))
  const merged = [
    ...recentProjects,
    ...apiProjects.filter((p) => !recentPaths.has(p.projectPath)),
  ]

  const filtered = merged.filter((p) => p.projectPath.toLowerCase().includes(query.toLowerCase()))

  const items = filtered.map((p, i) => ({
    key: String(i),
    label: `${p.accountName}: ${p.projectPath}`,
    value: `${p.accountName}\0${p.projectPath}`,
  }))

  function handleSelect (item: { value: string }) {
    const [accountName, projectPath] = item.value.split('\0')
    const account = accounts.find((a) => a.name === accountName)
    if (!account) return
    const recent = recentProjects.find(
      (p) => p.accountName === accountName && p.projectPath === projectPath,
    )
    onSelect({ account, projectPath, localPath: recent?.localPath ?? '' })
  }

  return (
    <Box flexDirection="column" gap={1}>
      <Text bold>Select project</Text>
      <TextInput placeholder="Filter projects..." value={query} onChange={setQuery} onSubmit={() => {}} />
      {loading
        ? <Text dimColor>Loading projects…</Text>
        : loadError
        ? <Text color="red">Failed to load projects: {loadError}</Text>
        : items.length > 0
        ? <SelectInput items={items} onSelect={handleSelect} />
        : <Text dimColor>{query ? 'No matches' : 'No projects found'}</Text>}
    </Box>
  )
}
