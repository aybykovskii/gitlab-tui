import React, { useState } from 'react'
import { Box, Text, useInput } from 'ink'
import TextInput from 'ink-text-input'
import SelectInput from 'ink-select-input'
import type { MRDetail } from '../services/types.js'
import type { UpdateMRInput } from '../services/mrService.js'
import { parseMRTitle } from '../services/mrTitle.js'

type Step = 'title' | 'description' | 'labels' | 'draft' | 'confirm'

interface Props {
  mr: MRDetail
  onSubmit: (changes: UpdateMRInput) => void
  onBack: () => void
}

export function EditMRForm({ mr, onSubmit, onBack }: Props) {
  const parsed = parseMRTitle(mr.title)
  const [step, setStep] = useState<Step>('title')
  const [title, setTitle] = useState(parsed.title)
  const [description, setDescription] = useState(mr.description)
  const [labels, setLabels] = useState('')
  const [draft, setDraft] = useState(parsed.draft)

  useInput((_input, key) => {
    if (key.escape) onBack()
  })

  if (step === 'title') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Edit MR — Title</Text>
        <Text dimColor>Esc: cancel</Text>
        <TextInput
          value={title}
          onChange={setTitle}
          onSubmit={(v) => { if (v.trim()) { setTitle(v.trim()); setStep('description') } }}
        />
      </Box>
    )
  }

  if (step === 'description') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Edit MR — Description</Text>
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
        <Text bold>Edit MR — Labels <Text dimColor>(comma-separated, Enter to skip)</Text></Text>
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
        <Text bold>Edit MR — Draft status</Text>
        <SelectInput
          items={[
            { label: 'No — ready for review', value: 'no' },
            { label: 'Yes — draft / WIP', value: 'yes' },
          ]}
          initialIndex={draft ? 1 : 0}
          onSelect={(item) => { setDraft(item.value === 'yes'); setStep('confirm') }}
        />
      </Box>
    )
  }

  const labelList = labels.split(',').map((l) => l.trim()).filter(Boolean)

  return (
    <Box flexDirection="column" gap={1}>
      <Text bold>Edit MR — Confirm</Text>
      <Text><Text dimColor>Title: </Text>{draft ? 'Draft: ' : ''}{title}</Text>
      {description ? <Text><Text dimColor>Desc:  </Text>{description.slice(0, 60)}</Text> : null}
      {labelList.length > 0 && <Text><Text dimColor>Labels:</Text> {labelList.join(', ')}</Text>}
      <SelectInput
        items={[
          { label: 'Save changes', value: 'save' },
          { label: 'Cancel', value: 'cancel' },
        ]}
        onSelect={(item) => {
          if (item.value === 'save') {
            onSubmit({
              title,
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
