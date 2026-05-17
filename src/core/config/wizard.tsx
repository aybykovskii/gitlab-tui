import React, { useState } from 'react'
import { Box, Text, useApp } from 'ink'
import SelectInput from 'ink-select-input'
import TextInput from 'ink-text-input'

import { createConfigManager } from './manager.js'
import type { Config } from './types.js'

const EDITOR_PRESETS = [
  { label: 'Windsurf', value: 'windsurf' },
  { label: 'VS Code', value: 'code' },
  { label: 'IntelliJ IDEA', value: 'idea' },
  { label: 'Neovim', value: 'nvim' },
]

type Step = 'url' | 'token' | 'editor' | 'done'

interface Props {
  onComplete: (config: Config) => void
}

export function SetupWizard ({ onComplete }: Props) {
  const [step, setStep] = useState<Step>('url')
  const [url, setUrl] = useState('https://gitlab.com')
  const [token, setToken] = useState('')

  function handleUrlSubmit (value: string) {
    setUrl(value)
    setStep('token')
  }

  function handleTokenSubmit (value: string) {
    setToken(value)
    setStep('editor')
  }

  function handleEditorSelect (item: { value: string }) {
    const config: Config = {
      accounts: [{ name: 'default', url, token }],
      defaultAccount: 'default',
      recentProjects: [],
      editor: item.value,
    }
    const manager = createConfigManager()
    manager.saveConfig(config)
    onComplete(config)
  }

  if (step === 'url') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Welcome to gitlab-tui!</Text>
        <Text>GitLab instance URL:</Text>
        <TextInput value={url} onChange={setUrl} onSubmit={handleUrlSubmit} />
      </Box>
    )
  }

  if (step === 'token') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text>Personal Access Token (scopes: api, read_user):</Text>
        <TextInput value={token} onChange={setToken} onSubmit={handleTokenSubmit} mask="*" />
      </Box>
    )
  }

  if (step === 'editor') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text>Preferred editor:</Text>
        <SelectInput items={EDITOR_PRESETS} onSelect={handleEditorSelect} />
      </Box>
    )
  }

  return null
}
