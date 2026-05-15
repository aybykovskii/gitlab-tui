import { describe, it, expect, vi } from 'vitest'
import { createIDELauncher } from './launcher.js'

function makeSpawn() {
  const fn = vi.fn()
  return { spawn: fn, lastCall: () => fn.mock.calls[0] as [string, string[]] }
}

describe('openFile', () => {
  it('calls windsurf --goto file:line', () => {
    const { spawn, lastCall } = makeSpawn()
    createIDELauncher('windsurf', spawn).openFile('/src/foo.ts', 42)

    expect(lastCall()).toEqual(['windsurf', ['--goto', '/src/foo.ts:42']])
  })

  it('calls code --goto file:line', () => {
    const { spawn, lastCall } = makeSpawn()
    createIDELauncher('code', spawn).openFile('/src/foo.ts', 10)

    expect(lastCall()).toEqual(['code', ['--goto', '/src/foo.ts:10']])
  })

  it('calls idea --line line file', () => {
    const { spawn, lastCall } = makeSpawn()
    createIDELauncher('idea', spawn).openFile('/src/foo.ts', 7)

    expect(lastCall()).toEqual(['idea', ['--line', '7', '/src/foo.ts']])
  })

  it('calls nvim +line file', () => {
    const { spawn, lastCall } = makeSpawn()
    createIDELauncher('nvim', spawn).openFile('/src/foo.ts', 3)

    expect(lastCall()).toEqual(['nvim', ['+3', '/src/foo.ts']])
  })

  it('throws a descriptive error for unknown editor', () => {
    const { spawn } = makeSpawn()
    const launcher = createIDELauncher('emacs', spawn)

    expect(() => launcher.openFile('/src/foo.ts', 1)).toThrow(
      'Unknown editor "emacs"',
    )
  })
})

describe('openUrl', () => {
  it('uses "open" on macOS', () => {
    const { spawn, lastCall } = makeSpawn()
    const launcher = createIDELauncher('code', spawn)
    ;(launcher.openUrl as (url: string, platform: string) => void)(
      'https://gitlab.com/ns/p/-/merge_requests/1',
      'darwin',
    )
    expect(lastCall()).toEqual(['open', ['https://gitlab.com/ns/p/-/merge_requests/1']])
  })

  it('uses "xdg-open" on Linux', () => {
    const { spawn, lastCall } = makeSpawn()
    const launcher = createIDELauncher('code', spawn)
    ;(launcher.openUrl as (url: string, platform: string) => void)(
      'https://gitlab.com/ns/p/-/merge_requests/1',
      'linux',
    )
    expect(lastCall()).toEqual(['xdg-open', ['https://gitlab.com/ns/p/-/merge_requests/1']])
  })
})
