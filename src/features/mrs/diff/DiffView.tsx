import React, { useEffect, useState } from 'react'
import chalk from 'chalk'
import { highlight } from 'cli-highlight'
import { Box, Text, useInput } from 'ink'
import TextInput from 'ink-text-input'

import { StatusBar } from '../../../ui/StatusBar.js'
import type { Thread } from '../services/types.js'

import type { SideBySideRow } from './parser.js'
import { parseDiff } from './parser.js'
import type { CommentPosition, DiffRefs, LineRange } from './position.js'
import { buildDiffPosition } from './position.js'

const VISIBLE_LINES = 25
// Layout per row:
// ●(1) ○(1) gutter(2) | lineNo(4)+marker(1)+content(colW) | " │ "(3) | lineNo(4)+marker(1)+content(colW)
const FIXED_COLS = 17

const LANG_MAP: Record<string, string> = {
  ts: 'typescript',
  tsx: 'typescript',
  js: 'javascript',
  jsx: 'javascript',
  py: 'python',
  go: 'go',
  rs: 'rust',
  yaml: 'yaml',
  yml: 'yaml',
  json: 'json',
  css: 'css',
  scss: 'scss',
  html: 'html',
  md: 'markdown',
  sh: 'bash',
  bash: 'bash',
  sql: 'sql',
  rb: 'ruby',
  java: 'java',
  kt: 'kotlin',
  swift: 'swift',
  c: 'c',
  cpp: 'cpp',
  h: 'c',
  php: 'php',
}

function detectLang (filePath: string): string | undefined {
  const ext = filePath.split('.').pop()?.toLowerCase()
  return ext ? LANG_MAP[ext] : undefined
}

function pad (s: string, width: number): string {
  if (s.length >= width) return s.slice(0, width)
  return s + ' '.repeat(width - s.length)
}

function lineNoStr (n: number | null): string {
  return n === null ? '    ' : String(n).padStart(4)
}

function diffMarker (type: NonNullable<SideBySideRow['left']>['type'] | null): string {
  if (type === 'added') return chalk.green('+')
  if (type === 'removed') return chalk.red('-')
  return ' '
}

function syntaxColor (content: string, colWidth: number, lang: string | undefined): string {
  const padded = pad(content, colWidth)
  if (!lang) return padded
  try {
    return highlight(padded, { language: lang, ignoreIllegals: true })
  } catch {
    return padded
  }
}

interface CommentingState {
  position: CommentPosition
  range: LineRange | null
  displayStart: number
  displayEnd: number
}

type CommentMode = 'draft' | 'instant' | 'reply' | 'draft-reply'

interface Props {
  filePath: string
  rawDiff: string
  refs: DiffRefs
  draftComments?: Map<number, string[]>
  draftRangeLines?: Set<number>
  threadComments?: Map<number, Thread[]>
  onAddComment: (position: CommentPosition, body: string) => void
  onAddInstantComment?: (position: CommentPosition, body: string) => void
  onReplyToThread?: (threadId: string, body: string) => Promise<void>
  onDraftReplyToThread?: (threadId: string, body: string) => Promise<void>
  onResolveThread?: (threadId: string, noteId: number, resolved: boolean) => Promise<void>
  onOpenInEditor?: (filePath: string, line: number) => void
  onPrevFile?: () => void
  onNextFile?: () => void
  onBack: () => void
}

export function DiffView ({
  filePath,
  rawDiff,
  refs,
  draftComments,
  draftRangeLines,
  threadComments,
  onAddComment,
  onAddInstantComment,
  onReplyToThread,
  onDraftReplyToThread,
  onResolveThread,
  onOpenInEditor,
  onPrevFile,
  onNextFile,
  onBack,
}: Props) {
  const rows = parseDiff(rawDiff)
  const lang = detectLang(filePath)

  const [cursor, setCursor] = useState(0)
  const [offset, setOffset] = useState(0)
  const [selectionAnchor, setSelectionAnchor] = useState<number | null>(null)
  const [commenting, setCommenting] = useState<CommentingState | null>(null)
  const [commentMode, setCommentMode] = useState<CommentMode>('draft')
  const [commentBody, setCommentBody] = useState('')
  const [threadPanelCursor, setThreadPanelCursor] = useState(0)
  const [replyTarget, setReplyTarget] = useState<Thread | null>(null)

  const cols = process.stdout.columns ?? 120
  const colWidth = Math.floor((cols - FIXED_COLS) / 2)

  const cursorRow = rows[cursor]
  const cursorLineNo = cursorRow?.right?.newLineNo ?? cursorRow?.left?.oldLineNo ?? null
  const currentDrafts = cursorLineNo != null ? (draftComments?.get(cursorLineNo) ?? []) : []
  const currentThreads = cursorLineNo != null ? (threadComments?.get(cursorLineNo) ?? []) : []
  const activeThread = currentThreads[threadPanelCursor] ?? null

  useEffect(() => {
    setThreadPanelCursor(0)
  }, [cursorLineNo])

  const inVisualMode = selectionAnchor !== null
  const selStart = inVisualMode ? Math.min(selectionAnchor, cursor) : cursor
  const selEnd = inVisualMode ? Math.max(selectionAnchor, cursor) : cursor

  const inputActive = !commenting && !replyTarget

  // Main navigation handler — disabled while typing
  useInput((input, key) => {
    if (input === 'q' || key.escape) {
      if (inVisualMode) {
        setSelectionAnchor(null)
        return
      }
      onBack()
      return
    }

    if (key.leftArrow && onPrevFile) {
      onPrevFile()
      return
    }
    if (key.rightArrow && onNextFile) {
      onNextFile()
      return
    }

    if (input === 'j' || key.downArrow) {
      setCursor((c) => {
        const next = Math.min(c + 1, rows.length - 1)
        if (next >= offset + VISIBLE_LINES) setOffset(next - VISIBLE_LINES + 1)
        return next
      })
    }
    if (input === 'k' || key.upArrow) {
      setCursor((c) => {
        const next = Math.max(c - 1, 0)
        if (next < offset) setOffset(next)
        return next
      })
    }

    if (input === 'v') {
      setSelectionAnchor((a) => (a === null ? cursor : null))
      return
    }

    if (currentThreads.length > 1) {
      if (input === '[') setThreadPanelCursor((c) => Math.max(c - 1, 0))
      if (input === ']') setThreadPanelCursor((c) => Math.min(c + 1, currentThreads.length - 1))
    }
    if (activeThread && input === 'r' && onReplyToThread) {
      setReplyTarget(activeThread)
      setCommentMode('reply')
      setCommentBody('')
      return
    }
    if (activeThread && input === 'u' && onDraftReplyToThread) {
      setReplyTarget(activeThread)
      setCommentMode('draft-reply')
      setCommentBody('')
      return
    }
    if (activeThread && input === 'R' && onResolveThread) {
      onResolveThread(activeThread.id, activeThread.notes[0].id, !activeThread.resolved)
      return
    }

    if ((input === 'c' || input === 'C') && !commenting) {
      const startRow = rows[selStart]
      const endRow = rows[selEnd]
      const startLine = startRow?.right ?? startRow?.left
      const endLine = endRow?.right ?? endRow?.left
      if (!endLine) return

      const displayStart = startLine?.newLineNo ?? startLine?.oldLineNo ?? selStart + 1
      const displayEnd = endLine.newLineNo ?? endLine.oldLineNo ?? selEnd + 1

      const range: LineRange | null = inVisualMode && selStart !== selEnd
        ? {
          startOldLine: startLine?.oldLineNo ?? null,
          startNewLine: startLine?.newLineNo ?? null,
          endOldLine: endLine.oldLineNo,
          endNewLine: endLine.newLineNo,
        }
        : null

      setCommenting({
        position: buildDiffPosition(
          refs,
          { oldPath: filePath, newPath: filePath, oldLineNo: endLine.oldLineNo, newLineNo: endLine.newLineNo },
          range ?? undefined,
        ),
        range,
        displayStart,
        displayEnd,
      })
      setCommentMode(input === 'C' ? 'instant' : 'draft')
      setCommentBody('')
      setSelectionAnchor(null)
    }

    if (input === 'o' && onOpenInEditor) {
      const row = rows[cursor]
      const line = row?.right ?? row?.left
      if (line?.newLineNo) onOpenInEditor(filePath, line.newLineNo)
    }
  }, { isActive: inputActive })

  // Esc-only handler while TextInput is active
  useInput((_, key) => {
    if (key.escape) {
      setCommenting(null)
      setReplyTarget(null)
      setCommentBody('')
    }
  }, { isActive: !inputActive })

  const visible = rows.slice(offset, offset + VISIBLE_LINES)

  // ── Comment / reply input ─────────────────────────────────────────────
  if (commenting || replyTarget) {
    const isReply = !!replyTarget
    let label: string
    if (isReply) {
      label = commentMode === 'draft-reply'
        ? `Draft reply — ${replyTarget.firstNote.slice(0, 50)} (Enter to save, Esc to cancel):`
        : `Reply — ${replyTarget.firstNote.slice(0, 60)} (Enter to send, Esc to cancel):`
    } else if (commenting?.range) {
      label =
        `Draft comment on lines ${commenting.displayStart}–${commenting.displayEnd} (Enter to save, Esc to cancel):`
    } else {
      label = commentMode === 'instant'
        ? `Post comment at line ${commenting?.displayEnd} (Enter to send, Esc to cancel):`
        : `Draft comment at line ${commenting?.displayEnd} (Enter to save, Esc to cancel):`
    }

    return (
      <Box flexDirection="column" gap={1}>
        <Text>{label}</Text>
        <TextInput
          value={commentBody}
          onChange={setCommentBody}
          onSubmit={(body) => {
            const trimmed = body.trim()
            if (trimmed) {
              if (isReply) {
                const t = replyTarget
                if (commentMode === 'draft-reply' && onDraftReplyToThread) onDraftReplyToThread(t.id, trimmed)
                else if (onReplyToThread) onReplyToThread(t.id, trimmed)
              } else if (commenting) {
                if (commentMode === 'instant' && onAddInstantComment) onAddInstantComment(commenting.position, trimmed)
                else onAddComment(commenting.position, trimmed)
              }
            }
            setCommenting(null)
            setReplyTarget(null)
            setCommentBody('')
          }}
        />
      </Box>
    )
  }

  const selStartLine = rows[selStart]?.right?.newLineNo ?? rows[selStart]?.left?.oldLineNo ?? selStart + 1
  const selEndLine = rows[selEnd]?.right?.newLineNo ?? rows[selEnd]?.left?.oldLineNo ?? selEnd + 1

  const diffHints = [
    { key: 'j/k', label: 'navigate' },
    { key: 'v', label: inVisualMode ? 'cancel selection' : 'visual select' },
    { key: 'c', label: inVisualMode ? `draft (${selStartLine}–${selEndLine})` : 'draft comment' },
    { key: 'C', label: 'instant comment' },
    ...(onPrevFile ? [{ key: '←', label: 'prev file' }] : []),
    ...(onNextFile ? [{ key: '→', label: 'next file' }] : []),
    ...(activeThread && onReplyToThread ? [{ key: 'r', label: 'reply' }] : []),
    ...(activeThread && onDraftReplyToThread ? [{ key: 'u', label: 'draft reply' }] : []),
    ...(activeThread && onResolveThread ? [{ key: 'R', label: activeThread.resolved ? 'unresolve' : 'resolve' }] : []),
    ...(onOpenInEditor ? [{ key: 'o', label: 'editor' }] : []),
    { key: 'q', label: 'back' },
  ]

  return (
    <Box flexDirection="column">
      <Box gap={2}>
        <Text bold>{filePath}</Text>
        {inVisualMode && <Text color="yellow">VISUAL {selStartLine}–{selEndLine}</Text>}
      </Box>

      {visible.map((row, i) => {
        const absIdx = offset + i
        const isCursor = absIdx === cursor
        const isSelected = inVisualMode && absIdx >= selStart && absIdx <= selEnd

        const leftLine = row.left
        const rightLine = row.right

        const leftContent = leftLine ? syntaxColor(leftLine.content, colWidth, lang) : ' '.repeat(colWidth)
        const rightContent = rightLine ? syntaxColor(rightLine.content, colWidth, lang) : ' '.repeat(colWidth)

        const lineNum = rightLine?.newLineNo ?? leftLine?.oldLineNo ?? null
        const hasDraft = lineNum != null && (draftComments?.has(lineNum) ?? false)
        const isDraftRange = !hasDraft && lineNum != null && (draftRangeLines?.has(lineNum) ?? false)
        const hasThread = lineNum != null && (threadComments?.has(lineNum) ?? false)

        const gutterColor = isCursor ? 'green' : isSelected ? 'yellow' : undefined
        const gutterChar = isCursor ? '>' : isSelected ? '·' : ' '

        return (
          <Box key={absIdx}>
            <Text color="yellow">{hasDraft ? '●' : isDraftRange ? '·' : ' '}</Text>
            <Text color="cyan">{hasThread ? '○' : ' '}</Text>
            <Text color={gutterColor}>{gutterChar}</Text>
            <Text>
              {lineNoStr(leftLine?.oldLineNo ?? null)}
              {diffMarker(leftLine?.type ?? null)}
              {leftContent}
            </Text>
            <Text dimColor>│</Text>
            <Text>
              {lineNoStr(rightLine?.newLineNo ?? null)}
              {diffMarker(rightLine?.type ?? null)}
              {rightContent}
            </Text>
          </Box>
        )
      })}

      {(currentDrafts.length > 0 || currentThreads.length > 0) && (
        <Box flexDirection="column" borderStyle="single" borderColor="yellow" marginTop={1} paddingX={1}>
          {activeThread && (
            <Box flexDirection="column" gap={1}>
              <Text bold color={activeThread.resolved ? 'green' : 'yellow'}>
                {activeThread.resolved ? '✓' : '○'} {activeThread.author.name}
                {currentThreads.length > 1
                  ? <Text dimColor>[{threadPanelCursor + 1}/{currentThreads.length} [/]: switch]</Text>
                  : null}
              </Text>
              {activeThread.notes.map((note, ni) => (
                <Box key={ni} flexDirection="column" marginLeft={2}>
                  <Text dimColor>{note.author.name}:</Text>
                  <Text>{note.body}</Text>
                </Box>
              ))}
            </Box>
          )}
          {currentDrafts.map((c, i) => <Text key={`d${i}`} color="yellow">● [draft] {c}</Text>)}
        </Box>
      )}

      <StatusBar hints={diffHints} />
    </Box>
  )
}
