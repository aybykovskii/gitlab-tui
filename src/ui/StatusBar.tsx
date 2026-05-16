import React from 'react'
import { Box, Text } from 'ink'

export interface Hint {
  key: string
  label: string
}

interface Props {
  hints: Hint[]
}

export function StatusBar({ hints }: Props) {
  return (
    <Box borderStyle="single" borderColor="gray" paddingX={1} gap={2} flexWrap="wrap">
      {hints.map((h) => (
        <Text key={h.key}>
          <Text bold color="cyan">{h.key}</Text>
          <Text dimColor>: {h.label}</Text>
        </Text>
      ))}
    </Box>
  )
}
