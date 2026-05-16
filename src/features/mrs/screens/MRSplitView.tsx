import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { spawn } from 'node:child_process'
import { join } from 'node:path'
import { Box, Text } from 'ink'
import { MRList } from './MRList.js'
import { MRDetail } from './MRDetail.js'
import { EditMRForm } from './EditMRForm.js'
import { DiffView } from '../diff/DiffView.js'
import { createMRService } from '../services/mrService.js'
import { createMRActionsAPIImpl, createMRActionsService } from '../services/mrActions.js'
import { createDraftNotesAPI, createInstantCommentsAPI, createThreadActionsAPIImpl } from '../review/api.js'
import { createReviewSession } from '../review/session.js'
import { createInstantCommentService } from '../review/instant.js'
import { createThreadActionsService } from '../review/threadActions.js'
import { createGitLabClient } from '../../../core/gitlab/index.js'
import type { Account } from '../../../core/config/types.js'
import type { MR, MRDetail as MRDetailType, DiffFile, MRState, Thread } from '../services/types.js'
import type { UpdateMRInput } from '../services/mrService.js'
import type { CommentPosition } from '../diff/position.js'
import type { DraftComment } from '../review/session.js'

type Panel = 'list' | 'detail'
type Mode = 'browse' | 'edit' | 'diff'

interface DiffState {
  file: DiffFile
  fileIndex: number
  draftComments: Map<number, string[]>
  draftRangeLines: Set<number>
  threadComments: Map<number, Thread[]>
  fileThreads: Thread[]
}

interface Props {
  account: Account
  projectPath: string
  localPath?: string
  editor?: string
  initialMRState?: MRState | 'all'
}

export function MRSplitView({
  account, projectPath, localPath, editor = 'code', initialMRState = 'opened',
}: Props) {
  const [panel, setPanel] = useState<Panel>('list')
  const [mode, setMode] = useState<Mode>('browse')
  const [highlighted, setHighlighted] = useState<MR | null>(null)
  const [activeMR, setActiveMR] = useState<MRDetailType | null>(null)
  const [loadingDetail, setLoadingDetail] = useState(false)
  const [files, setFiles] = useState<DiffFile[]>([])
  const [threads, setThreads] = useState<Thread[]>([])
  const [dataLoading, setDataLoading] = useState(false)
  const [diffState, setDiffState] = useState<DiffState | null>(null)
  const [draftCount, setDraftCount] = useState(0)
  const [mrListKey, setMrListKey] = useState(0)

  const [termWidth, setTermWidth] = useState(process.stdout.columns ?? 120)
  useEffect(() => {
    let timer: ReturnType<typeof setTimeout>
    const handler = () => {
      clearTimeout(timer)
      timer = setTimeout(() => setTermWidth(process.stdout.columns ?? 120), 150)
    }
    process.stdout.on('resize', handler)
    return () => { clearTimeout(timer); process.stdout.off('resize', handler) }
  }, [])
  const leftWidth = Math.floor(termWidth * 0.33)

  const client = useMemo(() => createGitLabClient(account), [account.url, account.token])
  const mrService = useMemo(() => createMRService(client, projectPath), [client, projectPath])
  const mrActions = useMemo(
    () => createMRActionsService(createMRActionsAPIImpl(client, projectPath)),
    [client, projectPath],
  )

  function makeDraftSession(iid: number) {
    return createReviewSession(createDraftNotesAPI(account.url, account.token, projectPath, iid))
  }
  function makeInstantComments(iid: number) {
    return createInstantCommentService(createInstantCommentsAPI(client, account.url, account.token, projectPath, iid))
  }
  function makeThreadActions(iid: number) {
    return createThreadActionsService(
      createThreadActionsAPIImpl(client, account.url, account.token, projectPath, iid),
    )
  }

  const openInEditor = localPath
    ? (filePath: string, line: number) => {
        const absolute = join(localPath, filePath)
        spawn(editor, editorArgs(editor, absolute, line), { detached: true, stdio: 'ignore' }).unref()
      }
    : undefined

  const loadMRs = useCallback(
    (state: MRState | 'all') => mrService.listMRs({ state }),
    [mrService],
  )

  const loadDraftCount = useCallback(
    async (iid: number) => {
      try {
        const drafts = await makeDraftSession(iid).getDraftComments()
        setDraftCount(drafts.length)
      } catch {
        setDraftCount(0)
      }
    },
    [account.url, account.token, projectPath],
  )

  const reloadMRData = useCallback(async (iid: number) => {
    setDataLoading(true)
    try {
      const [f, t] = await Promise.all([
        mrService.getDiffFiles(iid),
        mrService.getThreads(iid),
      ])
      setFiles(f)
      setThreads(t)
    } finally {
      setDataLoading(false)
    }
  }, [mrService])

  useEffect(() => {
    if (activeMR) {
      loadDraftCount(activeMR.iid)
      reloadMRData(activeMR.iid)
    }
  }, [activeMR?.iid])

  const openDetail = useCallback(async (mr: MR) => {
    setLoadingDetail(true)
    try {
      const detail = await mrService.getMR(mr.iid)
      setActiveMR(detail)
      setPanel('detail')
    } catch {
      // stay on list
    } finally {
      setLoadingDetail(false)
    }
  }, [mrService])

  async function openDiff(file: DiffFile, index: number) {
    if (!activeMR) return
    const drafts = await makeDraftSession(activeMR.iid).getDraftComments().catch(() => [] as DraftComment[])
    const fileThreads = threads.filter((t) =>
      t.position && (t.position.filePath === file.newPath || t.position.filePath === file.oldPath),
    )
    const { draftComments, draftRangeLines } = buildDraftData(drafts, file)
    setDiffState({
      file,
      fileIndex: index,
      draftComments,
      draftRangeLines,
      threadComments: buildThreadMap(fileThreads),
      fileThreads,
    })
    setMode('diff')
  }

  // ── Edit mode ─────────────────────────────────────────────────────────
  if (mode === 'edit' && activeMR) {
    return (
      <EditMRForm
        mr={activeMR}
        onSubmit={async (changes: UpdateMRInput) => {
          const updated = await mrService.updateMR(activeMR.iid, changes)
          setActiveMR(updated)
          setMode('browse')
          setMrListKey((k) => k + 1)
        }}
        onBack={() => setMode('browse')}
      />
    )
  }

  // ── Diff mode ─────────────────────────────────────────────────────────
  if (mode === 'diff' && diffState && activeMR) {
    const refs = activeMR.diffRefs
    if (!refs) { setMode('browse'); return null }

    const draftSession = makeDraftSession(activeMR.iid)
    const instantComments = makeInstantComments(activeMR.iid)

    const threadActions = makeThreadActions(activeMR.iid)

    const prevFile = diffState.fileIndex > 0 ? files[diffState.fileIndex - 1] : undefined
    const nextFile = diffState.fileIndex < files.length - 1 ? files[diffState.fileIndex + 1] : undefined

    return (
      <DiffView
        filePath={diffState.file.newPath}
        rawDiff={diffState.file.rawDiff}
        refs={refs}
        draftComments={diffState.draftComments}
        draftRangeLines={diffState.draftRangeLines}
        threadComments={diffState.threadComments}
        onAddComment={async (position: CommentPosition, body: string) => {
          await draftSession.addDraftComment(position, body)
          const allDrafts = await draftSession.getDraftComments()
          setDraftCount(allDrafts.length)
          setDiffState((prev) => prev ? { ...prev, ...buildDraftData(allDrafts, prev.file) } : prev)
        }}
        onAddInstantComment={(position: CommentPosition, body: string) => {
          instantComments.postInlineComment(position, body)
        }}
        onReplyToThread={(id, body) => threadActions.replyToThread(id, body)}
        onDraftReplyToThread={(id, body) =>
          draftSession.addDraftReply(id, body).then(() => loadDraftCount(activeMR.iid))
        }
        onResolveThread={(id, resolved) => threadActions.resolveThread(id, resolved)}
        onOpenInEditor={openInEditor}
        onPrevFile={prevFile ? () => openDiff(prevFile, diffState.fileIndex - 1) : undefined}
        onNextFile={nextFile ? () => openDiff(nextFile, diffState.fileIndex + 1) : undefined}
        onBack={() => setMode('browse')}
      />
    )
  }

  // ── Browse mode ───────────────────────────────────────────────────────
  return (
    <Box flexDirection="row">
      {panel === 'list' && (
        <Box width={leftWidth} borderStyle="single" borderColor="green" flexDirection="column">
          <Text bold color="green"> Merge Requests</Text>
          <MRList
            key={mrListKey}
            projectPath={projectPath}
            initialState={initialMRState}
            focused={true}
            onHighlight={setHighlighted}
            onSelect={openDetail}
            loadMRs={loadMRs}
          />
        </Box>
      )}

      <Box
        flexGrow={1}
        borderStyle="single"
        borderColor={panel === 'detail' ? 'green' : 'gray'}
        flexDirection="column"
      >
        <Text bold color={panel === 'detail' ? 'green' : 'white'}> Details</Text>

        {panel === 'detail' && activeMR ? (
          <MRDetail
            mr={activeMR}
            files={files}
            threads={threads}
            loading={dataLoading}
            focused={true}
            draftCount={draftCount}
            onReload={() => reloadMRData(activeMR.iid)}
            onOpenFile={(file, index) => openDiff(file, index)}
            onApprove={() => mrActions.approveMR(activeMR.iid)}
            onMerge={() =>
              mrActions.mergeMR(activeMR.iid).then(() => {
                setPanel('list')
                setActiveMR(null)
                setMrListKey((k) => k + 1)
              })
            }
            onEdit={() => setMode('edit')}
            onSubmitReview={() =>
              makeDraftSession(activeMR.iid)
                .publishReview()
                .then(() => loadDraftCount(activeMR.iid))
            }
            onDiscardDrafts={() =>
              makeDraftSession(activeMR.iid)
                .discardAll()
                .then(() => setDraftCount(0))
            }
            onAddMRComment={(body) => makeInstantComments(activeMR.iid).postMRComment(body)}
            onReplyToThread={(id, body) => makeThreadActions(activeMR.iid).replyToThread(id, body)}
            onDraftReplyToThread={(id, body) =>
              makeDraftSession(activeMR.iid)
                .addDraftReply(id, body)
                .then(() => loadDraftCount(activeMR.iid))
            }
            onResolveThread={(id, resolved) =>
              makeThreadActions(activeMR.iid).resolveThread(id, resolved)
            }
            onBack={() => setPanel('list')}
          />
        ) : loadingDetail ? (
          <Box paddingX={1} paddingY={1}><Text dimColor>Loading…</Text></Box>
        ) : highlighted ? (
          <MRPreview mr={highlighted} />
        ) : (
          <Box paddingX={1} paddingY={1}>
            <Text dimColor>↑↓ navigate  Enter: open details</Text>
          </Box>
        )}
      </Box>
    </Box>
  )
}

function buildDraftData(
  drafts: DraftComment[],
  file: DiffFile,
): { draftComments: Map<number, string[]>; draftRangeLines: Set<number> } {
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

function buildThreadMap(threads: Thread[]): Map<number, Thread[]> {
  const map = new Map<number, Thread[]>()
  for (const t of threads) {
    if (!t.position) continue
    const line = t.position.newLine ?? t.position.oldLine
    if (line == null) continue
    map.set(line, [...(map.get(line) ?? []), t])
  }
  return map
}

function MRPreview({ mr }: { mr: MR }) {
  const PIPELINE_ICON: Record<string, string> = {
    success: '✓', failed: '✗', running: '●', pending: '○',
  }
  const pipeline = mr.pipeline ? (PIPELINE_ICON[mr.pipeline.status] ?? '?') : '–'
  return (
    <Box flexDirection="column" paddingX={1} paddingY={1} gap={1}>
      <Text bold>!{mr.iid} {mr.title}</Text>
      <Box flexDirection="column">
        <Text><Text dimColor>Author:   </Text>{mr.author.name}</Text>
        <Text>
          <Text dimColor>Branches: </Text>
          <Text color="cyan">{mr.sourceBranch}</Text>
          <Text dimColor> → </Text>
          <Text color="cyan">{mr.targetBranch}</Text>
        </Text>
        <Text>
          <Text dimColor>State:    </Text>{mr.state}
          {'   '}
          <Text dimColor>Pipeline: </Text>{pipeline}
        </Text>
      </Box>
      <Text dimColor>Enter: open details</Text>
    </Box>
  )
}

function editorArgs(editor: string, filePath: string, line: number): string[] {
  switch (editor) {
    case 'nvim':
    case 'vim':
    case 'vi':
      return [`+${line}`, filePath]
    case 'idea':
      return ['--line', String(line), filePath]
    default:
      return ['--goto', `${filePath}:${line}`]
  }
}
