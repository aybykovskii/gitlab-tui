import React, { useState } from 'react'
import { Box, Text, useInput } from 'ink'
import chalk from 'chalk'
import { parseDiff } from './parser.js'
import { buildDiffPosition } from './position.js'
import type { DiffRefs, CommentPosition } from './position.js'
import type { SideBySideRow } from './parser.js'

const VISIBLE_LINES = 30

interface Props {
  filePath: string
  rawDiff: string
  refs: DiffRefs
  onAddComment: (position: CommentPosition) => void
  onOpenInEditor?: (filePath: string, line: number) => void
  onBack: () => void
}

function colorLine(content: string, type: SideBySideRow['left'] extends null ? never : NonNullable<SideBySideRow['left']>['type']): string {
  if (type === 'added') return chalk.green(content)
  if (type === 'removed') return chalk.red(content)
  return content
}

function pad(s: string, width: number): string {
  if (s.length >= width) return s.slice(0, width)
  return s + ' '.repeat(width - s.length)
}

function lineNo(n: number | null): string {
  return n === null ? '    ' : String(n).padStart(4)
}

export function DiffView({ filePath, rawDiff, refs, onAddComment, onOpenInEditor, onBack }: Props) {
  const rows = parseDiff(rawDiff)
  const [cursor, setCursor] = useState(0)
  const [offset, setOffset] = useState(0)

  const colWidth = Math.floor((process.stdout.columns ?? 120) / 2) - 6

  useInput((input, key) => {
    if (input === 'q' || key.escape) { onBack(); return }

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

    if (input === 'c') {
      const row = rows[cursor]
      const line = row?.right ?? row?.left
      if (!line) return
      onAddComment(buildDiffPosition(refs, {
        oldPath: filePath,
        newPath: filePath,
        oldLineNo: line.oldLineNo,
        newLineNo: line.newLineNo,
      }))
    }

    if (input === 'o' && onOpenInEditor) {
      const row = rows[cursor]
      const line = row?.right ?? row?.left
      if (line?.newLineNo) onOpenInEditor(filePath, line.newLineNo)
    }
  })

  const visible = rows.slice(offset, offset + VISIBLE_LINES)

  return (
    <Box flexDirection="column">
      <Text bold>{filePath}</Text>
      <Text dimColor>j/k: navigate  c: comment  q: back</Text>
      {visible.map((row, i) => {
        const absIdx = offset + i
        const isCursor = absIdx === cursor

        const leftLine = row.left
        const rightLine = row.right

        const leftContent = leftLine
          ? colorLine(pad(leftLine.content, colWidth), leftLine.type)
          : ' '.repeat(colWidth)
        const rightContent = rightLine
          ? colorLine(pad(rightLine.content, colWidth), rightLine.type)
          : ' '.repeat(colWidth)

        return (
          <Box key={absIdx}>
            <Text inverse={isCursor}>
              {lineNo(leftLine?.oldLineNo ?? null)} {leftContent}
            </Text>
            <Text dimColor> │ </Text>
            <Text inverse={isCursor}>
              {lineNo(rightLine?.newLineNo ?? null)} {rightContent}
            </Text>
          </Box>
        )
      })}
    </Box>
  )
}
