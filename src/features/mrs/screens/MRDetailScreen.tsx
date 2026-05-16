import React, { useState, useEffect, useMemo, useCallback } from 'react'
import { spawn } from 'node:child_process'
import { join } from 'node:path'
import { Box, Text } from 'ink'
import { MRDetail } from './MRDetail.js'
import { EditMRForm } from './EditMRForm.js'
import { createMRService } from '../services/mrService.js'
import { createMRActionsAPIImpl, createMRActionsService } from '../services/mrActions.js'
import { createDraftNotesAPI, createInstantCommentsAPI, createThreadActionsAPIImpl } from '../review/api.js'
import { createReviewSession } from '../review/session.js'
import { createInstantCommentService } from '../review/instant.js'
import { createThreadActionsService } from '../review/threadActions.js'
import { createGitLabClient } from '../../../core/gitlab/index.js'
import { DiffScreen } from './DiffScreen.js'
import { useNavigation } from '../../../core/navigation/index.js'
import { useTheme } from '../../../core/theme/index.js'
import type { Account } from '../../../core/config/types.js'
import type { ScreenProps } from '../../../core/navigation/types.js'
import type { MR, MRDetail as MRDetailType, DiffFile, Thread } from '../services/types.js'
import type { UpdateMRInput } from '../services/mrService.js'
import type { CommentPosition } from '../diff/position.js'
import type { DraftComment } from '../review/session.js'

const PIPELINE_ICON: Record<string, string> = {
  success: '✓', failed: '✗', running: '●', pending: '○',
}

interface MRDetailScreenProps extends ScreenProps {
  mr: MR
  account: Account
  projectPath: string
  localPath?: string
  editor?: string
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

export function MRDetailScreen({ leftWidth, rightWidth, mr, account, projectPath, localPath, editor = 'code' }: MRDetailScreenProps) {
  const { push, pop } = useNavigation()
  const theme = useTheme()

  const [activeMR, setActiveMR] = useState<MRDetailType | null>(null)
  const [files, setFiles] = useState<DiffFile[]>([])
  const [threads, setThreads] = useState<Thread[]>([])
  const [loading, setLoading] = useState(true)
  const [draftCount, setDraftCount] = useState(0)
  const [mode, setMode] = useState<'browse' | 'edit'>('browse')

  const client = useMemo(() => createGitLabClient(account), [account.url, account.token])
  const mrService = useMemo(() => createMRService(client, projectPath), [client, projectPath])
  const mrActions = useMemo(
    () => createMRActionsService(createMRActionsAPIImpl(client, projectPath)),
    [client, projectPath],
  )

  const makeDraftSession = useCallback((iid: number) =>
    createReviewSession(createDraftNotesAPI(client, projectPath, iid)),
    [client, projectPath],
  )
  const makeInstantComments = useCallback((iid: number) =>
    createInstantCommentService(createInstantCommentsAPI(client, projectPath, iid)),
    [client, projectPath],
  )
  const makeThreadActions = useCallback((iid: number) =>
    createThreadActionsService(createThreadActionsAPIImpl(client, projectPath, iid)),
    [client, projectPath],
  )

  const openInEditor = localPath
    ? (filePath: string, line: number) => {
        const absolute = join(localPath, filePath)
        spawn(editor, editorArgs(editor, absolute, line), { detached: true, stdio: 'ignore' }).unref()
      }
    : undefined

  const loadDraftCount = useCallback(async (iid: number) => {
    try {
      const drafts = await makeDraftSession(iid).getDraftComments()
      setDraftCount(drafts.length)
    } catch { setDraftCount(0) }
  }, [makeDraftSession])

  const reloadMRData = useCallback(async (iid: number) => {
    setLoading(true)
    try {
      const [f, t] = await Promise.all([mrService.getDiffFiles(iid), mrService.getThreads(iid)])
      setFiles(f)
      setThreads(t)
    } finally { setLoading(false) }
  }, [mrService])

  useEffect(() => {
    mrService.getMR(mr.iid)
      .then((detail) => {
        setActiveMR(detail)
        return Promise.all([reloadMRData(detail.iid), loadDraftCount(detail.iid)])
      })
      .catch(() => setLoading(false))
  }, [mr.iid])

  async function openDiff(file: DiffFile, index: number) {
    if (!activeMR?.diffRefs) return
    const drafts = await makeDraftSession(activeMR.iid).getDraftComments().catch(() => [] as DraftComment[])
    const fileThreads = threads.filter((t) =>
      t.position && (t.position.filePath === file.newPath || t.position.filePath === file.oldPath),
    )
    const { draftComments, draftRangeLines } = buildDraftData(drafts, file)
    push({
      id: 'diff',
      component: DiffScreen,
      props: {
        files,
        initialFileIndex: index,
        activeMR,
        account,
        projectPath,
        localPath,
        editor,
        allThreads: threads,
        initialDraftComments: draftComments,
        initialDraftRangeLines: draftRangeLines,
        initialThreadComments: buildThreadMap(fileThreads),
      },
    })
  }

  const pipeline = mr.pipeline ? (PIPELINE_ICON[mr.pipeline.status] ?? '?') : '–'

  if (mode === 'edit' && activeMR) {
    return (
      <Box>
        <Box width={leftWidth} flexDirection="column" paddingX={1}>
          {renderLeftPanel()}
        </Box>
        <Box width={rightWidth} flexDirection="column">
          <EditMRForm
            mr={activeMR}
            onSubmit={async (changes: UpdateMRInput) => {
              const updated = await mrService.updateMR(activeMR.iid, changes)
              setActiveMR(updated)
              setMode('browse')
            }}
            onBack={() => setMode('browse')}
          />
        </Box>
      </Box>
    )
  }

  function renderLeftPanel() {
    return (
      <>
        <Text bold color={theme.secondary}>MR</Text>
        <Text bold color={theme.primary}>!{mr.iid} {mr.title}</Text>
        <Text color={theme.muted}>{mr.state}  {pipeline}</Text>
        <Text color={theme.muted}>{mr.sourceBranch} → {mr.targetBranch}</Text>
      </>
    )
  }

  return (
    <Box>
      <Box width={leftWidth} flexDirection="column" paddingX={1}>
        {renderLeftPanel()}
      </Box>
      <Box width={rightWidth} flexDirection="column">
        {!activeMR ? (
          <Text color={theme.muted}>Loading…</Text>
        ) : (
          <MRDetail
            mr={activeMR}
            files={files}
            threads={threads}
            loading={loading}
            draftCount={draftCount}
            onReload={() => reloadMRData(activeMR.iid)}
            onOpenFile={(file, index) => openDiff(file, index)}
            onApprove={() => mrActions.approveMR(activeMR.iid)}
            onMerge={() => mrActions.mergeMR(activeMR.iid).then(pop)}
            onEdit={() => setMode('edit')}
            onSubmitReview={() =>
              makeDraftSession(activeMR.iid).publishReview().then(() => loadDraftCount(activeMR.iid))
            }
            onDiscardDrafts={() =>
              makeDraftSession(activeMR.iid).discardAll().then(() => setDraftCount(0))
            }
            onAddMRComment={(body) => makeInstantComments(activeMR.iid).postMRComment(body)}
            onReplyToThread={(id, body) => makeThreadActions(activeMR.iid).replyToThread(id, body)}
            onDraftReplyToThread={(id, body) =>
              makeDraftSession(activeMR.iid).addDraftReply(id, body).then(() => loadDraftCount(activeMR.iid))
            }
            onResolveThread={(id, noteId, resolved) => makeThreadActions(activeMR.iid).resolveThread(id, noteId, resolved)}
            onOpenFileLine={openInEditor}
            onOpenInBrowser={() => { /* browser open handled via webUrl */ }}
            onBack={pop}
          />
        )}
      </Box>
    </Box>
  )
}
