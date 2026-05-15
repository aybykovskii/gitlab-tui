import React, { useState } from 'react'
import { Box, Text, useInput } from 'ink'
import TextInput from 'ink-text-input'
import SelectInput from 'ink-select-input'
import type { Account, RecentProject } from '../config/types.js'
import type { DetectedProject } from './detector.js'

interface Props {
  recentProjects: RecentProject[]
  accounts: Account[]
  onSelect: (project: DetectedProject) => void
}

export function ProjectSelect({ recentProjects, accounts, onSelect }: Props) {
  const [query, setQuery] = useState('')

  const filtered = recentProjects.filter((p) =>
    p.projectPath.toLowerCase().includes(query.toLowerCase()),
  )

  const items = filtered.map((p) => ({
    label: `${p.accountName}: ${p.projectPath}`,
    value: p,
  }))

  function handleSelect(item: { value: RecentProject }) {
    const account = accounts.find((a) => a.name === item.value.accountName)
    if (!account) return
    onSelect({ account, projectPath: item.value.projectPath })
  }

  return (
    <Box flexDirection="column" gap={1}>
      <Text bold>Select project</Text>
      <TextInput
        placeholder="Filter projects..."
        value={query}
        onChange={setQuery}
        onSubmit={() => {}}
      />
      {items.length > 0 ? (
        <SelectInput items={items} onSelect={handleSelect} />
      ) : (
        <Text dimColor>{query ? 'No matches' : 'No recent projects'}</Text>
      )}
    </Box>
  )
}
