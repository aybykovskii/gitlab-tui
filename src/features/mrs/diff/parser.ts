export interface DiffLine {
  type: 'context' | 'added' | 'removed'
  content: string
  oldLineNo: number | null
  newLineNo: number | null
}

export interface SideBySideRow {
  left: DiffLine | null
  right: DiffLine | null
}

export function parseDiff (rawDiff: string): SideBySideRow[] {
  const lines = rawDiff.split('\n')
  const rows: SideBySideRow[] = []

  let oldLine = 0
  let newLine = 0

  // Collect pending removed lines to pair with following added lines
  const pendingRemoved: DiffLine[] = []

  function flushRemoved () {
    for (const r of pendingRemoved) {
      rows.push({ left: r, right: null })
    }
    pendingRemoved.length = 0
  }

  for (const raw of lines) {
    if (raw.startsWith('@@')) {
      flushRemoved()
      const match = /@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/.exec(raw)
      if (match) {
        oldLine = parseInt(match[1], 10) - 1
        newLine = parseInt(match[2], 10) - 1
      }
      continue
    }

    if (raw.startsWith('-')) {
      oldLine++
      pendingRemoved.push({
        type: 'removed',
        content: raw.slice(1),
        oldLineNo: oldLine,
        newLineNo: null,
      })
      continue
    }

    if (raw.startsWith('+')) {
      newLine++
      const addedLine: DiffLine = {
        type: 'added',
        content: raw.slice(1),
        oldLineNo: null,
        newLineNo: newLine,
      }
      if (pendingRemoved.length > 0) {
        // Pair with a pending removed line
        rows.push({ left: pendingRemoved.shift()!, right: addedLine })
      } else {
        rows.push({ left: null, right: addedLine })
      }
      continue
    }

    if (raw.startsWith(' ')) {
      flushRemoved()
      oldLine++
      newLine++
      const contextLine: DiffLine = {
        type: 'context',
        content: raw.slice(1),
        oldLineNo: oldLine,
        newLineNo: newLine,
      }
      rows.push({ left: contextLine, right: contextLine })
      continue
    }
  }

  flushRemoved()
  return rows
}
