import React, { useEffect, useMemo } from 'react'
import { Box, Text } from 'ink'

import type { Account } from '../../../core/config/types.js'
import { createGitLabClient } from '../../../core/gitlab/index.js'
import { useNavigation } from '../../../core/navigation/index.js'
import type { ScreenProps } from '../../../core/navigation/types.js'
import { useTheme } from '../../../core/theme/index.js'
import { createMRService } from '../services/mrService.js'
import type { MR } from '../services/types.js'

import { MRDetailScreen } from './MRDetailScreen.js'
import { MRList } from './MRList.js'

const SECTIONS = [
  { label: 'Merge Requests', value: 'mrs' },
  { label: 'Issues', value: 'issues' },
  { label: 'Pipelines', value: 'pipelines' },
]

interface MRListScreenProps extends ScreenProps {
  account: Account
  projectPath: string
  localPath?: string
  editor?: string
}

export function MRListScreen ({ leftWidth, rightWidth, account, projectPath, localPath, editor }: MRListScreenProps) {
  const { push, setHints } = useNavigation()
  const theme = useTheme()

  useEffect(() => {
    setHints([
      { key: 'j/k', label: 'navigate' },
      { key: 'Enter', label: 'open MR' },
      { key: 's', label: 'cycle state' },
      { key: 'r', label: 'refresh' },
    ])
  }, [])

  const client = useMemo(() => createGitLabClient(account), [account.url, account.token])
  const mrService = useMemo(() => createMRService(client, projectPath), [client, projectPath])

  function handleSelectMR (mr: MR) {
    push({
      id: 'mr-detail',
      component: MRDetailScreen,
      props: { mr, account, projectPath, localPath, editor },
    })
  }

  return (
    <Box>
      <Box width={leftWidth} flexDirection="column" paddingX={1}>
        <Text bold color={theme.secondary}>Section</Text>
        {SECTIONS.map((s) => {
          const isActive = s.value === 'mrs'
          return (
            <Text key={s.value} color={isActive ? theme.primary : theme.muted} bold={isActive}>
              {isActive ? '▶ ' : '  '}
              {s.label}
            </Text>
          )
        })}
      </Box>
      <Box width={rightWidth} flexDirection="column">
        <MRList projectPath={projectPath} onSelect={handleSelectMR} loadMRs={(state) => mrService.listMRs({ state })} />
      </Box>
    </Box>
  )
}
