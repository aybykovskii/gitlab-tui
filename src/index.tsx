#!/usr/bin/env node
import React, { useState } from 'react'
import './features/mrs/index.js'
import './features/pipelines/index.js'
import { render } from 'ink'

import { createConfigManager, SetupWizard } from './core/config/index.js'
import type { Config } from './core/config/types.js'
import { createGitRemoteDetector } from './core/git/index.js'
import { Navigator } from './core/navigation/index.js'
import { buildInitialStack } from './core/router/stackBuilder.js'
import { ThemeProvider } from './core/theme/index.js'

const configManager = createConfigManager()
const configManagerAdapter = { saveConfig: (c: Config) => configManager.saveConfig(c) }

function App () {
  const [config, setConfig] = useState<Config | null>(
    configManager.configExists() ? configManager.getConfig() : null,
  )

  if (!config) {
    return (
      <ThemeProvider>
        <SetupWizard
          onComplete={(c) => {
            configManager.saveConfig(c)
            setConfig(c)
          }}
        />
      </ThemeProvider>
    )
  }

  const detected = createGitRemoteDetector(config).detect()
  const args = process.argv.slice(2)
  const initialStack = buildInitialStack({ args, config, detected: detected ?? undefined,
    configManager: configManagerAdapter })

  return (
    <ThemeProvider theme={config.theme}>
      <Navigator initialStack={initialStack} leftColumnWidth={config.ui?.leftColumnWidth} />
    </ThemeProvider>
  )
}

render(<App />)
