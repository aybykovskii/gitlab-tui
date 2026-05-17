import React from 'react'
import { Box, Text } from 'ink'

import { useTheme } from '../core/theme/index.js'

export interface Hint {
  key: string
  label: string
}

interface Props {
  hints: Hint[]
}

export function StatusBar ({ hints }: Props) {
  const theme = useTheme()
  return (
    <Box borderStyle="single" borderColor={theme.border} paddingX={1} gap={2} flexWrap="wrap">
      {hints.map((h) => (
        <Text key={h.key}>
          <Text bold color={theme.primary}>{h.key}</Text>
          <Text color={theme.muted}>: {h.label}</Text>
        </Text>
      ))}
    </Box>
  )
}
