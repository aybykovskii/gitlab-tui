#!/usr/bin/env node
import { render } from 'ink'
import React from 'react'
import { getRegisteredFeatures } from './core/router/index.js'

import './features/mrs/index.js'
import './features/pipelines/index.js'

function App() {
  const features = getRegisteredFeatures()
  return React.createElement(
    React.Fragment,
    null,
    React.createElement('text', null, `gitlab-tui — ${features.length} feature(s) loaded`),
  )
}

render(React.createElement(App))
