#!/usr/bin/env node
import React, { useState, useEffect } from 'react'
import { render } from 'ink'
import { createConfigManager, SetupWizard } from './core/config/index.js'
import { createGitRemoteDetector, ProjectSelect } from './core/git/index.js'
import { createGitLabClient, listUserProjects } from './core/gitlab/index.js'
import type { Account } from './core/config/types.js'
import { parseDeepLink } from './core/router/index.js'
import { MRSplitView } from './features/mrs/screens/MRSplitView.js'
import type { Config } from './core/config/index.js'
import type { DetectedProject } from './core/git/index.js'

import './features/mrs/index.js'
import './features/pipelines/index.js'

const configManager = createConfigManager()
const deepLink = parseDeepLink(process.argv.slice(2))

function App() {
  const [config, setConfig] = useState<Config | null>(
    configManager.configExists() ? configManager.getConfig() : null,
  )
  const [project, setProject] = useState<DetectedProject | null>(null)

  useEffect(() => {
    if (config && !project) {
      const detected = createGitRemoteDetector(config).detect()
      if (detected) setProject(detected)
    }
  }, [config])

  if (!config) {
    return <SetupWizard onComplete={setConfig} />
  }

  if (!project) {
    const detected = createGitRemoteDetector(config).detect()
    if (detected) return null

    const defaultAccount = config.accounts.find((a) => a.name === config.defaultAccount)
      ?? config.accounts[0]

    const loadProjects = defaultAccount
      ? () => listUserProjects(createGitLabClient(defaultAccount), defaultAccount.name)
      : undefined

    return (
      <ProjectSelect
        recentProjects={config.recentProjects}
        accounts={config.accounts}
        loadProjects={loadProjects}
        onSelect={(p) => {
          const updated = {
            ...config,
            recentProjects: [
              { accountName: p.account.name, projectPath: p.projectPath, localPath: p.localPath || undefined },
              ...config.recentProjects.filter((r) => r.projectPath !== p.projectPath),
            ],
          }
          configManager.saveConfig(updated)
          setConfig(updated)
          setProject(p)
        }}
      />
    )
  }

  return (
    <MRSplitView
      account={project.account}
      projectPath={project.projectPath}
      localPath={project.localPath || undefined}
      editor={config.editor}
      initialMRState={deepLink.type === 'mr-detail' ? 'opened' : 'opened'}
    />
  )
}

render(<App />)
