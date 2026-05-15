import { describe, it, expect, beforeEach } from 'vitest'
import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { createConfigManager } from './manager.js'
import type { Config } from './types.js'

const makeConfig = (overrides: Partial<Config> = {}): Config => ({
  accounts: [
    { name: 'work', url: 'https://gitlab.mycompany.com', token: 'work-token' },
    { name: 'personal', url: 'https://gitlab.com', token: 'personal-token' },
  ],
  defaultAccount: 'work',
  recentProjects: [],
  editor: 'windsurf',
  ...overrides,
})

let tmpDir: string
let configPath: string

beforeEach(() => {
  tmpDir = mkdtempSync(join(tmpdir(), 'gitlab-tui-test-'))
  configPath = join(tmpDir, 'config.json')
})

const cleanup = () => rmSync(tmpDir, { recursive: true, force: true })

describe('configExists', () => {
  it('returns false when config file does not exist', () => {
    const manager = createConfigManager(configPath)
    expect(manager.configExists()).toBe(false)
  })

  it('returns true after saveConfig', () => {
    const manager = createConfigManager(configPath)
    manager.saveConfig(makeConfig())
    expect(manager.configExists()).toBe(true)
  })
})

describe('saveConfig / getConfig', () => {
  it('roundtrip preserves all config fields', () => {
    const manager = createConfigManager(configPath)
    const config = makeConfig({
      recentProjects: [{ accountName: 'work', projectPath: 'ns/project' }],
    })
    manager.saveConfig(config)
    expect(manager.getConfig()).toEqual(config)
  })
})

describe('getAccountForUrl', () => {
  it('returns matching account for HTTPS remote URL', () => {
    const manager = createConfigManager(configPath)
    manager.saveConfig(makeConfig())

    const account = manager.getAccountForUrl('https://gitlab.com/ns/project.git')
    expect(account?.name).toBe('personal')
  })

  it('returns matching account for SSH remote URL', () => {
    const manager = createConfigManager(configPath)
    manager.saveConfig(makeConfig())

    const account = manager.getAccountForUrl('git@gitlab.mycompany.com:ns/project.git')
    expect(account?.name).toBe('work')
  })

  it('returns null when hostname does not match any account', () => {
    const manager = createConfigManager(configPath)
    manager.saveConfig(makeConfig())

    expect(manager.getAccountForUrl('https://github.com/ns/project.git')).toBeNull()
  })
})
