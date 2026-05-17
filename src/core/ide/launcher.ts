import { spawn as nodeSpawn } from 'node:child_process'

export type SpawnFn = (command: string, args: string[]) => void

interface EditorPreset {
  command: string
  args: (file: string, line: number) => string[]
}

const PRESETS: Record<string, EditorPreset> = {
  windsurf: { command: 'windsurf', args: (f, l) => ['--goto', `${f}:${l}`] },
  code: { command: 'code', args: (f, l) => ['--goto', `${f}:${l}`] },
  idea: { command: 'idea', args: (f, l) => ['--line', String(l), f] },
  webstorm: { command: 'webstorm', args: (f, l) => ['--line', String(l), f] },
  nvim: { command: 'nvim', args: (f, l) => [`+${l}`, f] },
}

function defaultSpawn (command: string, args: string[]) {
  nodeSpawn(command, args, { detached: true, stdio: 'ignore' }).unref()
}

export function createIDELauncher (editor: string, spawn: SpawnFn = defaultSpawn) {
  function openFile (filePath: string, line: number): void {
    const preset = PRESETS[editor]
    if (!preset) {
      throw new Error(
        `Unknown editor "${editor}". Supported: ${Object.keys(PRESETS).join(', ')}`,
      )
    }
    spawn(preset.command, preset.args(filePath, line))
  }

  function openUrl (url: string, platform = process.platform): void {
    const command = platform === 'darwin' ? 'open' : 'xdg-open'
    spawn(command, [url])
  }

  return { openFile, openUrl }
}
