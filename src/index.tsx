#!/usr/bin/env node
import React, { useState } from 'react'
import { render } from 'ink'
import { createConfigManager, SetupWizard } from './core/config/index.js'
import { createGitRemoteDetector } from './core/git/index.js'
import { createGitLabClient, listUserProjects } from './core/gitlab/index.js'
import { ThemeProvider } from './core/theme/index.js'
import { Navigator } from './core/navigation/index.js'
import { HomeScreen } from './features/home/HomeScreen.js'
import { ProjectScreen } from './features/home/ProjectScreen.js'
import type { Config } from './core/config/types.js'
import type { Screen } from './core/navigation/types.js'

import './features/mrs/index.js'
import './features/pipelines/index.js'

const configManager = createConfigManager()

function buildInitialScreen(config: Config): Screen {
  const detected = createGitRemoteDetector(config).detect()
  const defaultAccount = config.accounts.find((a) => a.name === config.defaultAccount)
    ?? config.accounts[0]

  const loadProjects = defaultAccount
    ? () => listUserProjects(createGitLabClient(defaultAccount), defaultAccount.name)
    : undefined

  if (detected) {
    return {
      id: 'project',
      component: ProjectScreen,
      props: { project: detected, config },
    }
  }

  return {
    id: 'home',
    component: HomeScreen,
    props: { config, configManager: { saveConfig: (c: Config) => configManager.saveConfig(c) }, loadProjects },
  }
}

function App() {
  const [config, setConfig] = useState<Config | null>(
    configManager.configExists() ? configManager.getConfig() : null,
  )

  if (!config) {
    return (
      <ThemeProvider>
        <SetupWizard onComplete={(c) => { configManager.saveConfig(c); setConfig(c) }} />
      </ThemeProvider>
    )
  }

  return (
    <ThemeProvider theme={config.theme}>
      <Navigator
        initialScreen={buildInitialScreen(config)}
        leftColumnWidth={config.ui?.leftColumnWidth}
      />
    </ThemeProvider>
  )
}

render(<App />)
