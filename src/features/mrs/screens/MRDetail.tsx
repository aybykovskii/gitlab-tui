import React, { useState, useEffect, useCallback } from 'react'
import { Box, Text, useInput } from 'ink'
import TextInput from 'ink-text-input'
import SelectInput from 'ink-select-input'
import type { MRDetail, DiffFile, Thread } from '../services/types.js'

type Tab = 'files' | 'threads'
type ThreadModal = { type: 'reply'; thread: Thread } | { type: 'mr-comment' }

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
  onReplyToThread?: (discussionId: string, body: string) => Promise<void>
  onResolveThread?: (discussionId: string, resolved: boolean) => Promise<void>
  onOpenFileLine?: (filePath: string, line: number) => void
  draftCount?: number
  onBack: () => void
}

export function MRDetail({
  mr, loadFiles, loadThreads, onOpenFile, onOpenInBrowser, onSubmitReview,
  onAddMRComment, onReplyToThread, onResolveThread, onOpenFileLine,
  draftCount = 0, onBack,
}: Props) {
  const [tab, setTab] = useState<Tab>('files')
  const [files, setFiles] = useState<DiffFile[]>([])
  const [threads, setThreads] = useState<Thread[]>([])
  const [loading, setLoading] = useState(true)
  const [modal, setModal] = useState<ThreadModal | null>(null)
  const [inputBody, setInputBody] = useState('')
  const [threadCursor, setThreadCursor] = useState(0)

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
    if (modal) return

    if (key.tab) setTab((t) => (t === 'files' ? 'threads' : 'files'))
    if (input === 'r' && tab !== 'threads') load()
    if (input === 'b' && onOpenInBrowser) onOpenInBrowser()
    if (input === 'S' && onSubmitReview) onSubmitReview()
    if (input === 'm' && onAddMRComment) { setModal({ type: 'mr-comment' }); setInputBody('') }
    if (input === 'q' || key.escape) onBack()

    if (tab === 'threads' && threads.length > 0) {
      if (input === 'j' || key.downArrow) setThreadCursor((c) => Math.min(c + 1, threads.length - 1))
      if (input === 'k' || key.upArrow) setThreadCursor((c) => Math.max(c - 1, 0))

      if (input === 'r' && onReplyToThread) {
        const t = threads[threadCursor]
        if (t) { setModal({ type: 'reply', thread: t }); setInputBody('') }
      }

      if (input === 'R' && onResolveThread) {
        const t = threads[threadCursor]
        if (t) {
          onResolveThread(t.id, !t.resolved).then(() => load()).catch(() => undefined)
        }
      }

      if (key.return && onOpenFileLine) {
        const t = threads[threadCursor]
        const line = t?.position?.newLine ?? t?.position?.oldLine
        if (t?.position?.filePath && line) onOpenFileLine(t.position.filePath, line)
      }
    }
  })

  const pipeline = mr.pipeline ? (PIPELINE_ICON[mr.pipeline.status] ?? '?') : '–'

  if (modal?.type === 'reply') {
    const thread = modal.thread
    return (
      <Box flexDirection="column" gap={1}>
        <Text>Reply to thread (Enter to send, Esc to cancel):</Text>
        <Text dimColor>{thread.firstNote.slice(0, 80)}</Text>
        <TextInput
          value={inputBody}
          onChange={setInputBody}
          onSubmit={(body) => {
            if (body.trim() && onReplyToThread) {
              onReplyToThread(thread.id, body.trim()).then(() => load()).catch(() => undefined)
            }
            setModal(null)
            setInputBody('')
          }}
        />
      </Box>
    )
  }

  if (modal?.type === 'mr-comment') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text>Add MR comment (Enter to send, Esc to cancel):</Text>
        <TextInput
          value={inputBody}
          onChange={setInputBody}
          onSubmit={(body) => {
            if (body.trim() && onAddMRComment) onAddMRComment(body.trim())
            setModal(null)
            setInputBody('')
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
          <Text dimColor>Tab: switch  b: browser  m: comment  q: back</Text>
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
            <Text dimColor>j/k: navigate  r: reply  R: resolve  Enter: open in diff</Text>
            {threads.map((t, i) => (
              <Box key={t.id} gap={1}>
                <Text inverse={i === threadCursor} color={t.resolved ? 'green' : 'yellow'}>
                  {t.resolved ? '✓' : '○'}
                </Text>
                <Text inverse={i === threadCursor} dimColor>{t.author.name}:</Text>
                <Text inverse={i === threadCursor}>{t.firstNote.slice(0, 80)}</Text>
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
