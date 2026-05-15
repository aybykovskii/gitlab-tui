#!/usr/bin/env node
import React, { useState } from 'react'
import { render } from 'ink'
import { createConfigManager, SetupWizard } from './core/config/index.js'
import { createGitRemoteDetector, ProjectSelect } from './core/git/index.js'
import type { Config } from './core/config/index.js'
import type { DetectedProject } from './core/git/index.js'

import './features/mrs/index.js'
import './features/pipelines/index.js'

const configManager = createConfigManager()

function App() {
  const [config, setConfig] = useState<Config | null>(
    configManager.configExists() ? configManager.getConfig() : null,
  )
  const [project, setProject] = useState<DetectedProject | null>(null)

  if (!config) {
    return <SetupWizard onComplete={setConfig} />
  }

  if (!project) {
    const detected = createGitRemoteDetector(config).detect()
    if (detected) {
      setProject(detected)
      return null
    }
    return (
      <ProjectSelect
        recentProjects={config.recentProjects}
        accounts={config.accounts}
        onSelect={(p) => {
          const updated = {
            ...config,
            recentProjects: [
              { accountName: p.account.name, projectPath: p.projectPath },
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

  return <>{`${project.account.name}: ${project.projectPath}`}</>
}

render(<App />)
