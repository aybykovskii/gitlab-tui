import React, { useState, useEffect } from 'react'
import { Box, Text, useInput } from 'ink'
import TextInput from 'ink-text-input'
import SelectInput from 'ink-select-input'
import type { CreateMRInput } from '../services/mrService.js'

type Step = 'title' | 'target' | 'template' | 'description' | 'labels' | 'draft' | 'confirm'

interface Props {
  sourceBranch: string
  loadBranches: () => Promise<string[]>
  loadTemplates: () => Promise<string[]>
  loadTemplateContent: (name: string) => Promise<string>
  onSubmit: (input: CreateMRInput) => void
  onBack: () => void
}

export function CreateMRForm({ sourceBranch, loadBranches, loadTemplates, loadTemplateContent, onSubmit, onBack }: Props) {
  const [step, setStep] = useState<Step>('title')
  const [title, setTitle] = useState('')
  const [targetBranch, setTargetBranch] = useState('')
  const [description, setDescription] = useState('')
  const [labels, setLabels] = useState('')
  const [draft, setDraft] = useState(false)
  const [branches, setBranches] = useState<string[]>([])
  const [templates, setTemplates] = useState<string[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    Promise.all([loadBranches(), loadTemplates()])
      .then(([b, t]) => { setBranches(b); setTemplates(t) })
      .finally(() => setLoading(false))
  }, [loadBranches, loadTemplates])

  useInput((input, key) => {
    if (key.escape) onBack()
  })

  if (loading) return <Text dimColor>Loading branches…</Text>

  if (step === 'title') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Create MR — Title</Text>
        <Text dimColor>Source: <Text color="cyan">{sourceBranch}</Text>  Esc: cancel</Text>
        <TextInput
          value={title}
          onChange={setTitle}
          onSubmit={(v) => {
            if (v.trim()) { setTitle(v.trim()); setStep('target') }
          }}
        />
      </Box>
    )
  }

  if (step === 'target') {
    const items = branches.map((b) => ({ label: b, value: b }))
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Create MR — Target branch</Text>
        <Text dimColor>Title: <Text color="cyan">{title}</Text></Text>
        {items.length > 0
          ? <SelectInput items={items} onSelect={(item) => { setTargetBranch(item.value); setStep(templates.length > 0 ? 'template' : 'description') }} />
          : <Text dimColor>No branches found</Text>
        }
      </Box>
    )
  }

  if (step === 'template') {
    const items = [
      { label: '(no template)', value: '' },
      ...templates.map((t) => ({ label: t, value: t })),
    ]
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Create MR — Template</Text>
        <SelectInput
          items={items}
          onSelect={async (item) => {
            if (item.value) {
              const content = await loadTemplateContent(item.value)
              setDescription(content)
            }
            setStep('description')
          }}
        />
      </Box>
    )
  }

  if (step === 'description') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Create MR — Description <Text dimColor>(Enter to continue, leave empty to skip)</Text></Text>
        <TextInput
          value={description}
          onChange={setDescription}
          onSubmit={() => setStep('labels')}
        />
      </Box>
    )
  }

  if (step === 'labels') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Create MR — Labels <Text dimColor>(comma-separated, Enter to skip)</Text></Text>
        <TextInput
          value={labels}
          onChange={setLabels}
          onSubmit={() => setStep('draft')}
        />
      </Box>
    )
  }

  if (step === 'draft') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Create as draft? <Text dimColor>(y/n)</Text></Text>
        <Text dimColor>Current: {draft ? 'yes' : 'no'}</Text>
        <Text dimColor>Press y or n, then Enter to confirm</Text>
        <SelectInput
          items={[
            { label: 'No — ready for review', value: 'no' },
            { label: 'Yes — draft / WIP', value: 'yes' },
          ]}
          onSelect={(item) => { setDraft(item.value === 'yes'); setStep('confirm') }}
        />
      </Box>
    )
  }

  const labelList = labels.split(',').map((l) => l.trim()).filter(Boolean)

  return (
    <Box flexDirection="column" gap={1}>
      <Text bold>Create MR — Confirm</Text>
      <Text><Text dimColor>Title:  </Text>{draft ? 'Draft: ' : ''}{title}</Text>
      <Text><Text dimColor>Source: </Text>{sourceBranch}</Text>
      <Text><Text dimColor>Target: </Text>{targetBranch}</Text>
      {description ? <Text><Text dimColor>Desc:   </Text>{description.slice(0, 60)}…</Text> : null}
      {labelList.length > 0 && <Text><Text dimColor>Labels: </Text>{labelList.join(', ')}</Text>}
      <Text dimColor>Enter: create  Esc: cancel</Text>
      <SelectInput
        items={[
          { label: 'Create MR', value: 'create' },
          { label: 'Cancel', value: 'cancel' },
        ]}
        onSelect={(item) => {
          if (item.value === 'create') {
            onSubmit({
              title,
              sourceBranch,
              targetBranch,
              description: description || undefined,
              labels: labelList.length > 0 ? labelList : undefined,
              draft,
            })
          } else {
            onBack()
          }
        }}
      />
    </Box>
  )
}
