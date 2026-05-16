import React from 'react'
import { Box, Text } from 'ink'
import { ProjectSelect } from '../../core/git/index.js'
import { useNavigation } from '../../core/navigation/index.js'
import { useTheme } from '../../core/theme/index.js'
import { ProjectScreen } from './ProjectScreen.js'
import type { Config } from '../../core/config/types.js'
import type { ScreenProps } from '../../core/navigation/types.js'
import type { DetectedProject } from '../../core/git/index.js'

const VERSION = '0.1.0'

const HOTKEYS = [
  { key: 'j/k', label: 'навигация' },
  { key: 'Enter', label: 'выбор' },
  { key: 'q', label: 'назад' },
  { key: 'r', label: 'обновить' },
  { key: '?', label: 'помощь' },
]

interface HomeScreenProps extends ScreenProps {
  config: Config
  configManager: { saveConfig(config: Config): void }
}

export function HomeScreen({ leftWidth, rightWidth, config, configManager }: HomeScreenProps) {
  const { push } = useNavigation()
  const theme = useTheme()

  function handleSelectProject(project: DetectedProject) {
    const updated: Config = {
      ...config,
      recentProjects: [
        { accountName: project.account.name, projectPath: project.projectPath, localPath: project.localPath || undefined },
        ...config.recentProjects.filter((r) => r.projectPath !== project.projectPath),
      ],
    }
    configManager.saveConfig(updated)
    push({ id: 'project', component: ProjectScreen, props: { project, config: updated } })
  }

  return (
    <Box>
      <Box width={leftWidth} flexDirection="column" gap={1} paddingX={1}>
        <Text bold color={theme.primary}>gitlab-tui</Text>
        <Text color={theme.muted}>v{VERSION}</Text>
        <Text color={theme.muted}>GitLab TUI client</Text>
        <Box flexDirection="column" marginTop={1}>
          <Text bold color={theme.secondary}>Hotkeys</Text>
          {HOTKEYS.map((h) => (
            <Text key={h.key} color={theme.muted}>
              <Text color={theme.primary}>{h.key}</Text>
              {' '}{h.label}
            </Text>
          ))}
        </Box>
      </Box>
      <Box width={rightWidth} flexDirection="column">
        <ProjectSelect
          recentProjects={config.recentProjects}
          accounts={config.accounts}
          onSelect={handleSelectProject}
        />
      </Box>
    </Box>
  )
}
