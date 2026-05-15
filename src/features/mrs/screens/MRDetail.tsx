import React, { useState, useEffect, useCallback } from 'react'
import { Box, Text, useInput } from 'ink'
import TextInput from 'ink-text-input'
import SelectInput from 'ink-select-input'
import type { MRDetail, DiffFile, Thread } from '../services/types.js'

type Tab = 'files' | 'threads'

const PIPELINE_ICON: Record<string, string> = {
  success: '✓', failed: '✗', running: '●', pending: '○',
}

interface Props {
  mr: MRDetail
  loadFiles: () => Promise<DiffFile[]>
  loadThreads: () => Promise<Thread[]>
  onOpenFile: (file: DiffFile) => void
  onOpenInBrowser?: () => void
  onSubmitReview?: () => void
  onAddMRComment?: (body: string) => void
  draftCount?: number
  onBack: () => void
}

export function MRDetail({ mr, loadFiles, loadThreads, onOpenFile, onOpenInBrowser, onSubmitReview, onAddMRComment, draftCount = 0, onBack }: Props) {
  const [tab, setTab] = useState<Tab>('files')
  const [files, setFiles] = useState<DiffFile[]>([])
  const [threads, setThreads] = useState<Thread[]>([])
  const [loading, setLoading] = useState(true)
  const [commenting, setCommenting] = useState(false)
  const [commentBody, setCommentBody] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [f, t] = await Promise.all([loadFiles(), loadThreads()])
      setFiles(f)
      setThreads(t)
    } finally {
      setLoading(false)
    }
  }, [loadFiles, loadThreads])

  useEffect(() => { load() }, [load])

  useInput((input, key) => {
    if (commenting) return
    if (key.tab) setTab((t) => (t === 'files' ? 'threads' : 'files'))
    if (input === 'r') load()
    if (input === 'b' && onOpenInBrowser) onOpenInBrowser()
    if (input === 'S' && onSubmitReview) onSubmitReview()
    if (input === 'm' && onAddMRComment) { setCommenting(true); setCommentBody('') }
    if (input === 'q' || key.escape) onBack()
  })

  const pipeline = mr.pipeline ? (PIPELINE_ICON[mr.pipeline.status] ?? '?') : '–'

  if (commenting) {
    return (
      <Box flexDirection="column" gap={1}>
        <Text>Add MR comment (Enter to send, Esc to cancel):</Text>
        <TextInput
          value={commentBody}
          onChange={setCommentBody}
          onSubmit={(body) => {
            if (body.trim() && onAddMRComment) onAddMRComment(body.trim())
            setCommenting(false)
            setCommentBody('')
          }}
        />
      </Box>
    )
  }

  return (
    <Box flexDirection="column" gap={1}>
      <Box flexDirection="column">
        <Text bold>{mr.title}</Text>
        <Text>
          <Text dimColor>Author: </Text>{mr.author.name}
          {'  '}
          <Text dimColor>{mr.sourceBranch}</Text>
          {' → '}
          <Text dimColor>{mr.targetBranch}</Text>
        </Text>
        <Text>
          <Text dimColor>State: </Text>{mr.state}
          {'  '}
          <Text dimColor>Pipeline: </Text>{pipeline}
          {'  '}
          <Text dimColor>Approvals: </Text>
          {mr.approvalsRequired - mr.approvalsLeft}/{mr.approvalsRequired}
        </Text>
        {mr.description ? <Text dimColor>{mr.description}</Text> : null}
      </Box>

      <Box gap={2}>
        <Text bold={tab === 'files'} underline={tab === 'files'}>Files ({files.length})</Text>
        <Text bold={tab === 'threads'} underline={tab === 'threads'}>
          Threads ({threads.filter((t) => !t.resolved).length} open)
        </Text>
        <Box gap={2}>
          <Text dimColor>Tab: switch  r: refresh  b: browser  m: comment  q: back</Text>
          {draftCount > 0 && (
            <Text color="yellow">● {draftCount} draft{draftCount > 1 ? 's' : ''}  S: submit review</Text>
          )}
        </Box>
      </Box>

      {loading ? <Text dimColor>Loading…</Text> : null}

      {!loading && tab === 'files' && (
        files.length > 0 ? (
          <SelectInput
            items={files.map((f) => ({ label: formatFile(f), value: f }))}
            onSelect={(item) => onOpenFile(item.value)}
          />
        ) : <Text dimColor>No changed files</Text>
      )}

      {!loading && tab === 'threads' && (
        threads.length > 0 ? (
          <Box flexDirection="column">
            {threads.map((t) => (
              <Box key={t.id} gap={1}>
                <Text color={t.resolved ? 'green' : 'yellow'}>{t.resolved ? '✓' : '○'}</Text>
                <Text dimColor>{t.author.name}:</Text>
                <Text>{t.firstNote.slice(0, 80)}</Text>
              </Box>
            ))}
          </Box>
        ) : <Text dimColor>No threads</Text>
      )}
    </Box>
  )
}

function formatFile(f: DiffFile): string {
  const tag = f.isNew ? '[new]' : f.isDeleted ? '[del]' : f.isRenamed ? '[ren]' : ''
  const path = f.isRenamed ? `${f.oldPath} → ${f.newPath}` : f.newPath
  return `${tag ? tag + ' ' : ''}${path} +${f.addedLines} -${f.removedLines}`
}
