const DRAFT_PREFIX = 'Draft: '

export function parseMRTitle(raw: string): { title: string; draft: boolean } {
  if (raw.startsWith(DRAFT_PREFIX)) {
    return { title: raw.slice(DRAFT_PREFIX.length), draft: true }
  }
  return { title: raw, draft: false }
}

export function formatMRTitle(title: string, draft: boolean): string {
  return draft ? `${DRAFT_PREFIX}${title}` : title
}
