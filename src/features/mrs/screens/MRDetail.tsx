import React, { useState } from 'react'
import { Box, Text, useInput } from 'ink'
import TextInput from 'ink-text-input'
import SelectInput from 'ink-select-input'
import { StatusBar } from '../../../ui/StatusBar.js'
import type { Hint } from '../../../ui/StatusBar.js'
import type { MRDetail, DiffFile, Thread } from '../services/types.js'

type Tab = 'files' | 'threads'
type Modal =
  | { type: 'reply'; thread: Thread }
  | { type: 'draft-reply'; thread: Thread }
  | { type: 'mr-comment' }
  | { type: 'merge-confirm' }
  | { type: 'discard-confirm' }
  | { type: 'error'; message: string }

const PIPELINE_ICON: Record<string, string> = {
  success: '✓', failed: '✗', running: '●', pending: '○',
}

interface Props {
  mr: MRDetail
  files: DiffFile[]
  threads: Thread[]
  loading: boolean
  onReload: () => void
  onOpenFile: (file: DiffFile, index: number) => void
  onOpenInBrowser?: () => void
  onSubmitReview?: () => void
  onAddMRComment?: (body: string) => void
  onReplyToThread?: (discussionId: string, body: string) => Promise<void>
  onDraftReplyToThread?: (discussionId: string, body: string) => Promise<void>
  onResolveThread?: (discussionId: string, resolved: boolean) => Promise<void>
  onOpenFileLine?: (filePath: string, line: number) => void
  onApprove?: () => Promise<void>
  onMerge?: () => Promise<void>
  onEdit?: () => void
  onDiscardDrafts?: () => Promise<void>
  draftCount?: number
  onBack: () => void
  focused?: boolean
}

export function MRDetail({
  mr, files, threads, loading, onReload,
  onOpenFile, onOpenInBrowser, onSubmitReview,
  onAddMRComment, onReplyToThread, onDraftReplyToThread, onResolveThread, onOpenFileLine,
  onApprove, onMerge, onEdit, onDiscardDrafts,
  draftCount = 0, onBack, focused = true,
}: Props) {
  const [tab, setTab] = useState<Tab>('files')
  const [modal, setModal] = useState<Modal | null>(null)
  const [inputBody, setInputBody] = useState('')
  const [threadCursor, setThreadCursor] = useState(0)
  const [threadDetail, setThreadDetail] = useState<Thread | null>(null)

  useInput((input, key) => {
    if (threadDetail) {
      if (key.escape || input === 'q') { setThreadDetail(null); return }
      if (input === 'r' && onReplyToThread) {
        setModal({ type: 'reply', thread: threadDetail }); setInputBody(''); setThreadDetail(null); return
      }
      if (input === 'd' && onDraftReplyToThread) {
        setModal({ type: 'draft-reply', thread: threadDetail }); setInputBody(''); setThreadDetail(null); return
      }
      if (input === 'R' && onResolveThread) {
        onResolveThread(threadDetail.id, !threadDetail.resolved)
          .then(onReload)
          .catch((e: unknown) => setModal({ type: 'error', message: String(e) }))
        setThreadDetail(null)
        return
      }
      return
    }

    if (modal) {
      if (modal.type === 'merge-confirm') {
        if (input === 'y' && onMerge) {
          onMerge().then(onReload).catch((e: unknown) => setModal({ type: 'error', message: String(e) }))
          setModal(null)
        }
        if (input === 'n' || key.escape) setModal(null)
      }
      if (modal.type === 'discard-confirm') {
        if (input === 'y' && onDiscardDrafts) {
          onDiscardDrafts().catch((e: unknown) => setModal({ type: 'error', message: String(e) }))
          setModal(null)
        }
        if (input === 'n' || key.escape) setModal(null)
      }
      if (modal.type === 'error') {
        if (input === 'q' || key.escape) setModal(null)
      }
      return
    }

    if (key.tab) setTab((t) => (t === 'files' ? 'threads' : 'files'))
    if (input === 'b' && onOpenInBrowser) onOpenInBrowser()
    if (input === 'S' && onSubmitReview) onSubmitReview()
    if (input === 'X' && onDiscardDrafts && draftCount > 0) setModal({ type: 'discard-confirm' })
    if (input === 'm' && onAddMRComment) { setModal({ type: 'mr-comment' }); setInputBody('') }
    if (input === 'q' || key.escape) onBack()
    if (input === 'a' && onApprove) {
      onApprove().then(onReload).catch((e: unknown) => setModal({ type: 'error', message: String(e) }))
    }
    if (input === 'M' && onMerge) setModal({ type: 'merge-confirm' })
    if (input === 'e' && onEdit) onEdit()

    if (tab === 'threads' && threads.length > 0) {
      if (input === 'j' || key.downArrow) setThreadCursor((c) => Math.min(c + 1, threads.length - 1))
      if (input === 'k' || key.upArrow) setThreadCursor((c) => Math.max(c - 1, 0))
      if (key.return) { const t = threads[threadCursor]; if (t) setThreadDetail(t) }
      if (input === 'o' && onOpenFileLine) {
        const t = threads[threadCursor]
        const line = t?.position?.newLine ?? t?.position?.oldLine
        if (t?.position?.filePath && line) onOpenFileLine(t.position.filePath, line)
      }
      if (input === 'r' && onReplyToThread) {
        const t = threads[threadCursor]
        if (t) { setModal({ type: 'reply', thread: t }); setInputBody('') }
      }
      if (input === 'd' && onDraftReplyToThread) {
        const t = threads[threadCursor]
        if (t) { setModal({ type: 'draft-reply', thread: t }); setInputBody('') }
      }
      if (input === 'R' && onResolveThread) {
        const t = threads[threadCursor]
        if (t) {
          onResolveThread(t.id, !t.resolved)
            .then(onReload)
            .catch((e: unknown) => setModal({ type: 'error', message: String(e) }))
        }
      }
    }
  }, { isActive: focused })

  const pipeline = mr.pipeline ? (PIPELINE_ICON[mr.pipeline.status] ?? '?') : '–'

  if (threadDetail) {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>
          <Text color={threadDetail.resolved ? 'green' : 'yellow'}>{threadDetail.resolved ? '✓' : '○'}</Text>
          {' '}Thread  <Text dimColor>{threadDetail.author.name}</Text>
        </Text>
        <Box flexDirection="column" gap={1}>
          {threadDetail.notes.map((note, i) => (
            <Box key={i} flexDirection="column" borderStyle="single" borderColor="gray" paddingX={1}>
              <Text bold color="cyan">{note.author.name}</Text>
              <Text>{note.body}</Text>
            </Box>
          ))}
        </Box>
        <StatusBar hints={[
          ...(onReplyToThread ? [{ key: 'r', label: 'reply' }] : []),
          ...(onDraftReplyToThread ? [{ key: 'd', label: 'draft reply' }] : []),
          ...(onResolveThread ? [{ key: 'R', label: threadDetail.resolved ? 'unresolve' : 'resolve' }] : []),
          { key: 'q / Esc', label: 'back' },
        ]} />
      </Box>
    )
  }

  if (modal?.type === 'merge-confirm') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold>Merge MR: {mr.title}</Text>
        <Text>Are you sure? <Text bold>(y/n)</Text></Text>
      </Box>
    )
  }
  if (modal?.type === 'discard-confirm') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text bold color="yellow">Discard draft review</Text>
        <Text>Delete all {draftCount} draft{draftCount !== 1 ? 's' : ''}? <Text bold>(y/n)</Text></Text>
      </Box>
    )
  }
  if (modal?.type === 'error') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text color="red">Error: {modal.message}</Text>
        <Text dimColor>q / Esc: dismiss</Text>
      </Box>
    )
  }
  if (modal?.type === 'draft-reply') {
    const thread = modal.thread
    return (
      <Box flexDirection="column" gap={1}>
        <Text color="yellow">Draft reply (Enter to save, Esc to cancel):</Text>
        <Text dimColor>{thread.firstNote.slice(0, 80)}</Text>
        <TextInput value={inputBody} onChange={setInputBody} onSubmit={(body) => {
          if (body.trim() && onDraftReplyToThread) {
            onDraftReplyToThread(thread.id, body.trim()).then(onReload)
              .catch((e: unknown) => setModal({ type: 'error', message: String(e) }))
          }
          setModal(null); setInputBody('')
        }} />
      </Box>
    )
  }
  if (modal?.type === 'reply') {
    const thread = modal.thread
    return (
      <Box flexDirection="column" gap={1}>
        <Text>Reply (Enter to send, Esc to cancel):</Text>
        <Text dimColor>{thread.firstNote.slice(0, 80)}</Text>
        <TextInput value={inputBody} onChange={setInputBody} onSubmit={(body) => {
          if (body.trim() && onReplyToThread) {
            onReplyToThread(thread.id, body.trim()).then(onReload)
              .catch((e: unknown) => setModal({ type: 'error', message: String(e) }))
          }
          setModal(null); setInputBody('')
        }} />
      </Box>
    )
  }
  if (modal?.type === 'mr-comment') {
    return (
      <Box flexDirection="column" gap={1}>
        <Text>MR comment (Enter to send, Esc to cancel):</Text>
        <TextInput value={inputBody} onChange={setInputBody} onSubmit={(body) => {
          if (body.trim() && onAddMRComment) onAddMRComment(body.trim())
          setModal(null); setInputBody('')
        }} />
      </Box>
    )
  }

  const hints = buildHints({ tab, draftCount, onApprove, onMerge, onEdit, onOpenInBrowser, onAddMRComment, onSubmitReview, onDiscardDrafts, onReplyToThread, onDraftReplyToThread, onResolveThread, onOpenFileLine })

  return (
    <Box flexDirection="column" gap={1}>
      <Box flexDirection="column">
        <Text bold>{mr.title}</Text>
        <Text>
          <Text dimColor>Author: </Text>{mr.author.name}
          {'  '}<Text dimColor>{mr.sourceBranch}</Text>{' → '}<Text dimColor>{mr.targetBranch}</Text>
        </Text>
        <Text>
          <Text dimColor>State: </Text>{mr.state}
          {'  '}<Text dimColor>Pipeline: </Text>{pipeline}
          {'  '}<Text dimColor>Approvals: </Text>{mr.approvalsRequired - mr.approvalsLeft}/{mr.approvalsRequired}
        </Text>
        {mr.description ? <Text dimColor>{mr.description}</Text> : null}
      </Box>

      <Box gap={2}>
        <Text bold={tab === 'files'} underline={tab === 'files'}>Files ({files.length})</Text>
        <Text bold={tab === 'threads'} underline={tab === 'threads'}>
          Threads ({threads.filter((t) => !t.resolved).length} open)
        </Text>
        {draftCount > 0 && <Text color="yellow">● {draftCount} draft{draftCount > 1 ? 's' : ''}</Text>}
      </Box>

      {loading ? <Text dimColor>Loading…</Text> : null}

      {!loading && tab === 'files' && (
        files.length > 0 ? (
          <SelectInput
            items={files.map((f, i) => ({ label: formatFile(f), value: String(i) }))}
            onSelect={(item) => {
              const i = Number(item.value)
              const f = files[i]
              if (f !== undefined) onOpenFile(f, i)
            }}
          />
        ) : <Text dimColor>No changed files</Text>
      )}

      {!loading && tab === 'threads' && (
        threads.length > 0 ? (
          <Box flexDirection="column">
            {threads.map((t, i) => (
              <Box key={t.id} gap={1}>
                <Text inverse={i === threadCursor} color={t.resolved ? 'green' : 'yellow'}>
                  {t.resolved ? '✓' : '○'}
                </Text>
                <Text inverse={i === threadCursor} dimColor>{t.author.name}:</Text>
                <Text inverse={i === threadCursor}>
                  {t.firstNote.slice(0, 70)}
                  {t.notes.length > 1 ? <Text dimColor> +{t.notes.length - 1}</Text> : null}
                </Text>
              </Box>
            ))}
          </Box>
        ) : <Text dimColor>No threads</Text>
      )}

      <StatusBar hints={hints} />
    </Box>
  )
}

function formatFile(f: DiffFile): string {
  const tag = f.isNew ? '[new]' : f.isDeleted ? '[del]' : f.isRenamed ? '[ren]' : ''
  const path = f.isRenamed ? `${f.oldPath} → ${f.newPath}` : f.newPath
  return `${tag ? tag + ' ' : ''}${path} +${f.addedLines} -${f.removedLines}`
}

interface BuildHintsOpts {
  tab: Tab; draftCount: number
  onApprove?: Props['onApprove']; onMerge?: Props['onMerge']; onEdit?: Props['onEdit']
  onOpenInBrowser?: Props['onOpenInBrowser']; onAddMRComment?: Props['onAddMRComment']
  onSubmitReview?: Props['onSubmitReview']; onDiscardDrafts?: Props['onDiscardDrafts']
  onReplyToThread?: Props['onReplyToThread']; onDraftReplyToThread?: Props['onDraftReplyToThread']
  onResolveThread?: Props['onResolveThread']; onOpenFileLine?: Props['onOpenFileLine']
}

function buildHints(opts: BuildHintsOpts): Hint[] {
  const h: Hint[] = [{ key: 'Tab', label: 'switch tab' }]
  if (opts.onApprove) h.push({ key: 'a', label: 'approve' })
  if (opts.onMerge) h.push({ key: 'M', label: 'merge' })
  if (opts.onEdit) h.push({ key: 'e', label: 'edit' })
  if (opts.onOpenInBrowser) h.push({ key: 'b', label: 'browser' })
  if (opts.onAddMRComment) h.push({ key: 'm', label: 'comment' })
  if (opts.onSubmitReview && opts.draftCount > 0) h.push({ key: 'S', label: 'submit review' })
  if (opts.onDiscardDrafts && opts.draftCount > 0) h.push({ key: 'X', label: 'discard drafts' })
  h.push({ key: 'q', label: 'back' })
  if (opts.tab === 'threads') {
    h.push({ key: 'j/k', label: 'navigate' }, { key: 'Enter', label: 'view thread' })
    if (opts.onReplyToThread) h.push({ key: 'r', label: 'reply' })
    if (opts.onDraftReplyToThread) h.push({ key: 'd', label: 'draft reply' })
    if (opts.onResolveThread) h.push({ key: 'R', label: 'resolve' })
    if (opts.onOpenFileLine) h.push({ key: 'o', label: 'open in diff' })
  }
  return h
}
