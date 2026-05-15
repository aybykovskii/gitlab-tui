export type DeepLink = { type: 'mr-list' } | { type: 'mr-detail'; iid: number }

export function parseDeepLink(args: string[]): DeepLink {
  if (args[0] === 'mr' && args[1] !== undefined) {
    const iid = Number(args[1])
    if (Number.isInteger(iid) && iid > 0) {
      return { type: 'mr-detail', iid }
    }
  }
  return { type: 'mr-list' }
}
