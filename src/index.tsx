#!/usr/bin/env node
import React, { useState } from 'react'
import { render } from 'ink'
import { createConfigManager, SetupWizard } from './core/config/index.js'
import { createGitRemoteDetector, ProjectSelect } from './core/git/index.js'
import { createGitLabClient } from './core/gitlab/index.js'
import { parseDeepLink } from './core/router/index.js'
import { createMRService } from './features/mrs/services/mrService.js'
import { MRList } from './features/mrs/screens/MRList.js'
import type { Config } from './core/config/index.js'
import type { DetectedProject } from './core/git/index.js'
import type { MR } from './features/mrs/services/types.js'

import './features/mrs/index.js'
import './features/pipelines/index.js'

const configManager = createConfigManager()
const deepLink = parseDeepLink(process.argv.slice(2))

function App() {
  const [config, setConfig] = useState<Config | null>(
    configManager.configExists() ? configManager.getConfig() : null,
  )
  const [project, setProject] = useState<DetectedProject | null>(null)
  const [selectedMR, setSelectedMR] = useState<MR | null>(null)

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

  if (selectedMR) {
    return <>{`MR !${selectedMR.iid}: ${selectedMR.title}`}</>
  }

  const client = createGitLabClient(project.account)
  const mrService = createMRService(client, project.projectPath)

  return (
    <MRList
      projectPath={project.projectPath}
      initialState={deepLink.type === 'mr-detail' ? 'opened' : 'opened'}
      loadMRs={(state) => mrService.listMRs({ state })}
      onSelect={setSelectedMR}
    />
  )
}

render(<App />)
