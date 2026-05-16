import React, { useState, useMemo, useCallback } from 'react'
import { spawn } from 'node:child_process'
import { join } from 'node:path'
import { Box, Text } from 'ink'
import { DiffView } from '../diff/DiffView.js'
import { createDraftNotesAPI, createInstantCommentsAPI, createThreadActionsAPIImpl } from '../review/api.js'
import { createReviewSession } from '../review/session.js'
import { createInstantCommentService } from '../review/instant.js'
import { createThreadActionsService } from '../review/threadActions.js'
import { createGitLabClient } from '../../../core/gitlab/index.js'
import { useNavigation } from '../../../core/navigation/index.js'
import { useTheme } from '../../../core/theme/index.js'
import type { Account } from '../../../core/config/types.js'
import type { ScreenProps } from '../../../core/navigation/types.js'
import type { DiffFile, MRDetail, Thread } from '../services/types.js'
import type { CommentPosition } from '../diff/position.js'
import type { DraftComment } from '../review/session.js'

interface DiffScreenProps extends ScreenProps {
  files: DiffFile[]
  initialFileIndex: number
  activeMR: MRDetail
  account: Account
  projectPath: string
  localPath?: string
  editor?: string
  allThreads: Thread[]
  initialDraftComments: Map<number, string[]>
  initialDraftRangeLines: Set<number>
  initialThreadComments: Map<number, Thread[]>
}

function buildDraftData(drafts: DraftComment[], file: DiffFile) {
  const map = new Map<number, string[]>()
  const rangeLines = new Set<number>()
  for (const d of drafts) {
    if (!d.position) continue
    if (d.position.newPath !== file.newPath && d.position.oldPath !== file.oldPath) continue
    const endLine = d.position.newLine ?? d.position.oldLine
    if (endLine == null) continue
    map.set(endLine, [...(map.get(endLine) ?? []), d.body])
    if (d.position.lineRange) {
      const startLine = d.position.lineRange.startNewLine ?? d.position.lineRange.startOldLine
      if (startLine != null && startLine !== endLine) {
        map.set(startLine, [...(map.get(startLine) ?? []), d.body])
        for (let l = startLine + 1; l < endLine; l++) rangeLines.add(l)
      }
    }
  }
  return { draftComments: map, draftRangeLines: rangeLines }
}

function buildThreadMap(threads: Thread[]) {
  const map = new Map<number, Thread[]>()
  for (const t of threads) {
    if (!t.position) continue
    const line = t.position.newLine ?? t.position.oldLine
    if (line == null) continue
    map.set(line, [...(map.get(line) ?? []), t])
  }
  return map
}

function editorArgs(editor: string, filePath: string, line: number): string[] {
  switch (editor) {
    case 'nvim': case 'vim': case 'vi': return [`+${line}`, filePath]
    case 'idea': return ['--line', String(line), filePath]
    default: return ['--goto', `${filePath}:${line}`]
  }
}

export function DiffScreen({
  leftWidth, rightWidth,
  files, initialFileIndex,
  activeMR, account, projectPath, localPath, editor = 'code',
  allThreads,
  initialDraftComments, initialDraftRangeLines, initialThreadComments,
}: DiffScreenProps) {
  const { pop } = useNavigation()
  const theme = useTheme()

  const [currentFileIndex, setCurrentFileIndex] = useState(initialFileIndex)
  const [draftComments, setDraftComments] = useState<Map<number, string[]>>(initialDraftComments)
  const [draftRangeLines, setDraftRangeLines] = useState<Set<number>>(initialDraftRangeLines)
  const [threadComments, setThreadComments] = useState<Map<number, Thread[]>>(initialThreadComments)
  const [draftCount, setDraftCount] = useState(0)

  const client = useMemo(() => createGitLabClient(account), [account.url, account.token])

  const draftSession = useMemo(
    () => createReviewSession(createDraftNotesAPI(client, projectPath, activeMR.iid)),
    [client, projectPath, activeMR.iid],
  )
  const instantComments = useMemo(
    () => createInstantCommentService(createInstantCommentsAPI(client, account.url, account.token, projectPath, activeMR.iid)),
    [client, account.url, account.token, projectPath, activeMR.iid],
  )
  const threadActions = useMemo(
    () => createThreadActionsService(createThreadActionsAPIImpl(client, account.url, account.token, projectPath, activeMR.iid)),
    [client, account.url, account.token, projectPath, activeMR.iid],
  )

  const openInEditor = localPath
    ? (filePath: string, line: number) => {
        const absolute = join(localPath, filePath)
        spawn(editor, editorArgs(editor, absolute, line), { detached: true, stdio: 'ignore' }).unref()
      }
    : undefined

  const goToFile = useCallback(async (index: number) => {
    const file = files[index]
    if (!file) return
    setCurrentFileIndex(index)
    const fileThreads = allThreads.filter((t) =>
      t.position && (t.position.filePath === file.newPath || t.position.filePath === file.oldPath),
    )
    setThreadComments(buildThreadMap(fileThreads))
    try {
      const drafts = await draftSession.getDraftComments()
      const { draftComments: dc, draftRangeLines: dr } = buildDraftData(drafts, file)
      setDraftComments(dc)
      setDraftRangeLines(dr)
    } catch { /* keep previous */ }
  }, [files, allThreads, draftSession])

  const currentFile = files[currentFileIndex]
  const refs = activeMR.diffRefs

  return (
    <Box>
      <Box width={leftWidth} flexDirection="column" paddingX={1}>
        <Text bold color={theme.secondary}>Files</Text>
        {files.map((f, i) => {
          const isActive = i === currentFileIndex
          const label = f.isDeleted ? f.oldPath : f.newPath
          const truncated = label.length > leftWidth - 4
            ? '…' + label.slice(-(leftWidth - 5))
            : label
          return (
            <Text key={f.newPath} color={isActive ? theme.primary : theme.muted} bold={isActive}>
              {isActive ? '▶ ' : '  '}{truncated}
            </Text>
          )
        })}
      </Box>
      <Box width={rightWidth} flexDirection="column">
        {currentFile && refs ? (
          <DiffView
            filePath={currentFile.newPath}
            rawDiff={currentFile.rawDiff}
            refs={refs}
            draftComments={draftComments}
            draftRangeLines={draftRangeLines}
            threadComments={threadComments}
            onAddComment={async (position: CommentPosition, body: string) => {
              await draftSession.addDraftComment(position, body)
              const all = await draftSession.getDraftComments()
              setDraftCount(all.length)
              if (currentFile) {
                const { draftComments: dc, draftRangeLines: dr } = buildDraftData(all, currentFile)
                setDraftComments(dc)
                setDraftRangeLines(dr)
              }
            }}
            onAddInstantComment={(position: CommentPosition, body: string) => {
              instantComments.postInlineComment(position, body)
            }}
            onReplyToThread={(id, body) => threadActions.replyToThread(id, body)}
            onDraftReplyToThread={(id, body) =>
              draftSession.addDraftReply(id, body).then(async () => {
                const all = await draftSession.getDraftComments()
                setDraftCount(all.length)
              })
            }
            onResolveThread={(id, resolved) => threadActions.resolveThread(id, resolved)}
            onOpenInEditor={openInEditor}
            onPrevFile={currentFileIndex > 0 ? () => goToFile(currentFileIndex - 1) : undefined}
            onNextFile={currentFileIndex < files.length - 1 ? () => goToFile(currentFileIndex + 1) : undefined}
            onBack={pop}
          />
        ) : (
          <Text color={theme.muted}>No diff available</Text>
        )}
      </Box>
    </Box>
  )
}
