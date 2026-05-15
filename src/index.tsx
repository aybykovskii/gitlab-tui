#!/usr/bin/env node
import React, { useState } from 'react'
import { render } from 'ink'
import { createConfigManager, SetupWizard } from './core/config/index.js'
import { getRegisteredFeatures } from './core/router/index.js'
import type { Config } from './core/config/index.js'

import './features/mrs/index.js'
import './features/pipelines/index.js'

const configManager = createConfigManager()

function App() {
  const [config, setConfig] = useState<Config | null>(
    configManager.configExists() ? configManager.getConfig() : null,
  )

  if (!config) {
    return <SetupWizard onComplete={setConfig} />
  }

  const features = getRegisteredFeatures()
  return <>{features.map((f) => f.name).join(', ')}</>
}

render(<App />)
