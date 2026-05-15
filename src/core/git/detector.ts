import { existsSync, readFileSync } from 'node:fs'
import { join, dirname } from 'node:path'
import type { Account, Config } from '../config/types.js'

export interface DetectedProject {
  account: Account
  projectPath: string
}

function findGitConfig(cwd: string): string | null {
  let dir = cwd
  while (true) {
    const candidate = join(dir, '.git', 'config')
    if (existsSync(candidate)) return candidate
    const parent = dirname(dir)
    if (parent === dir) return null
    dir = parent
  }
}

function parseRemoteOriginUrl(gitConfigContent: string): string | null {
  const match = gitConfigContent.match(/\[remote "origin"\][\s\S]*?url\s*=\s*(.+)/)
  return match ? match[1].trim() : null
}

function extractProjectPath(remoteUrl: string): string | null {
  // SSH: git@gitlab.com:namespace/project.git
  const sshMatch = remoteUrl.match(/^git@[^:]+:(.+?)(?:\.git)?$/)
  if (sshMatch) return sshMatch[1]

  // HTTPS: https://gitlab.com/namespace/project.git
  try {
    const path = new URL(remoteUrl).pathname.replace(/^\//, '').replace(/\.git$/, '')
    return path || null
  } catch {
    return null
  }
}

function extractHostname(remoteUrl: string): string | null {
  const sshMatch = remoteUrl.match(/^git@([^:]+):/)
  if (sshMatch) return sshMatch[1]
  try {
    return new URL(remoteUrl).hostname
  } catch {
    return null
  }
}

export function createGitRemoteDetector(config: Config) {
  function detect(cwd = process.cwd()): DetectedProject | null {
    const gitConfigPath = findGitConfig(cwd)
    if (!gitConfigPath) return null

    const content = readFileSync(gitConfigPath, 'utf-8')
    const remoteUrl = parseRemoteOriginUrl(content)
    if (!remoteUrl) return null

    const hostname = extractHostname(remoteUrl)
    if (!hostname) return null

    const account = config.accounts.find((a) => new URL(a.url).hostname === hostname)
    if (!account) return null

    const projectPath = extractProjectPath(remoteUrl)
    if (!projectPath) return null

    return { account, projectPath }
  }

  return { detect }
}
