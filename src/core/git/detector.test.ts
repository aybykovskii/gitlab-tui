import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { mkdtempSync, rmSync, mkdirSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { createGitRemoteDetector } from './detector.js'
import type { Config } from '../config/types.js'

const config: Config = {
  accounts: [
    { name: 'personal', url: 'https://gitlab.com', token: 'token-a' },
    { name: 'work', url: 'https://gitlab.mycompany.com', token: 'token-b' },
  ],
  defaultAccount: 'personal',
  recentProjects: [],
  editor: 'windsurf',
}

let tmpDir: string

function makeGitRepo(dir: string, remoteUrl: string) {
  mkdirSync(join(dir, '.git'), { recursive: true })
  writeFileSync(
    join(dir, '.git', 'config'),
    `[core]\n\trepositoryformatversion = 0\n[remote "origin"]\n\turl = ${remoteUrl}\n\tfetch = +refs/heads/*:refs/remotes/origin/*\n`,
  )
}

beforeEach(() => {
  tmpDir = mkdtempSync(join(tmpdir(), 'gitlab-tui-git-test-'))
})

afterEach(() => {
  rmSync(tmpDir, { recursive: true, force: true })
})

describe('detect', () => {
  it('returns account and projectPath for HTTPS remote', () => {
    makeGitRepo(tmpDir, 'https://gitlab.com/myorg/myproject.git')
    const detector = createGitRemoteDetector(config)

    expect(detector.detect(tmpDir)).toEqual({
      account: config.accounts[0],
      projectPath: 'myorg/myproject',
    })
  })

  it('returns account and projectPath for SSH remote', () => {
    makeGitRepo(tmpDir, 'git@gitlab.mycompany.com:myorg/myproject.git')
    const detector = createGitRemoteDetector(config)

    expect(detector.detect(tmpDir)).toEqual({
      account: config.accounts[1],
      projectPath: 'myorg/myproject',
    })
  })

  it('returns null when remote hostname does not match any account', () => {
    makeGitRepo(tmpDir, 'https://github.com/myorg/myproject.git')
    const detector = createGitRemoteDetector(config)

    expect(detector.detect(tmpDir)).toBeNull()
  })

  it('returns null when no .git directory exists', () => {
    const detector = createGitRemoteDetector(config)

    expect(detector.detect(tmpDir)).toBeNull()
  })

  it('traverses parent directories to find .git/config', () => {
    makeGitRepo(tmpDir, 'https://gitlab.com/myorg/myproject.git')
    const subDir = join(tmpDir, 'src', 'components')
    mkdirSync(subDir, { recursive: true })
    const detector = createGitRemoteDetector(config)

    expect(detector.detect(subDir)).toEqual({
      account: config.accounts[0],
      projectPath: 'myorg/myproject',
    })
  })
})
