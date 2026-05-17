import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { dirname } from 'node:path'
import { join } from 'node:path'

import type { Account, Config } from './types.js'

const DEFAULT_CONFIG_PATH = join(homedir(), '.config', 'gitlab-tui', 'config.json')

function extractHostname (remoteUrl: string): string | null {
  // SSH: git@gitlab.com:ns/project.git
  const sshMatch = /^git@([^:]+):/.exec(remoteUrl)
  if (sshMatch) return sshMatch[1]

  // HTTPS: https://gitlab.com/ns/project.git
  try {
    return new URL(remoteUrl).hostname
  } catch {
    return null
  }
}

export function createConfigManager (configPath = DEFAULT_CONFIG_PATH) {
  function configExists (): boolean {
    return existsSync(configPath)
  }

  function getConfig (): Config {
    return JSON.parse(readFileSync(configPath, 'utf-8')) as Config
  }

  function saveConfig (config: Config): void {
    mkdirSync(dirname(configPath), { recursive: true })
    writeFileSync(configPath, JSON.stringify(config, null, 2), 'utf-8')
  }

  function getAccountForUrl (remoteUrl: string): Account | null {
    const hostname = extractHostname(remoteUrl)
    if (!hostname) return null
    const config = getConfig()
    return config.accounts.find((a) => new URL(a.url).hostname === hostname) ?? null
  }

  return { configExists, getConfig, saveConfig, getAccountForUrl }
}
