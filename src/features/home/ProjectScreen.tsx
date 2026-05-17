import React from 'react'
import { Box, Text } from 'ink'
import SelectInput from 'ink-select-input'

import type { Config } from '../../core/config/types.js'
import type { DetectedProject } from '../../core/git/index.js'
import { useNavigation } from '../../core/navigation/index.js'
import type { ScreenProps } from '../../core/navigation/types.js'
import { useTheme } from '../../core/theme/index.js'
import { MRListScreen } from '../mrs/screens/MRListScreen.js'

interface ProjectScreenProps extends ScreenProps {
  project: DetectedProject
  config: Config
}

function ComingSoonScreen (_: ScreenProps) {
  return <Text>Coming soon</Text>
}

const SECTIONS = [
  { label: 'Merge Requests', value: 'mrs' },
  { label: 'Issues', value: 'issues' },
  { label: 'Pipelines', value: 'pipelines' },
]

export function ProjectScreen ({ leftWidth, rightWidth, project, config }: ProjectScreenProps) {
  const { push } = useNavigation()
  const theme = useTheme()

  function handleSelect (item: { value: string }) {
    if (item.value === 'mrs') {
      push({
        id: 'mr-list',
        component: MRListScreen,
        props: { account: project.account, projectPath: project.projectPath, localPath: project.localPath },
      })
    } else {
      push({ id: item.value, component: ComingSoonScreen })
    }
  }

  return (
    <Box>
      <Box width={leftWidth} flexDirection="column" paddingX={1}>
        <Text bold color={theme.secondary}>Projects</Text>
        {config.recentProjects.map((p) => {
          const isSelected = p.projectPath === project.projectPath
          return (
            <Text key={p.projectPath} color={isSelected ? theme.primary : theme.muted} bold={isSelected}>
              {isSelected ? '▶ ' : '  '}
              {p.projectPath}
            </Text>
          )
        })}
      </Box>
      <Box width={rightWidth} flexDirection="column" paddingX={1}>
        <Text bold color={theme.secondary}>Section</Text>
        <SelectInput items={SECTIONS} onSelect={handleSelect} />
      </Box>
    </Box>
  )
}
