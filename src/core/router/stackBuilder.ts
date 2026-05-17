import { HomeScreen } from '../../features/home/HomeScreen.js'
import { ProjectScreen } from '../../features/home/ProjectScreen.js'
import { MRListScreen } from '../../features/mrs/screens/MRListScreen.js'
import type { Config } from '../config/types.js'
import type { DetectedProject } from '../git/index.js'
import type { Screen } from '../navigation/types.js'

import { parseDeepLink } from './deepLink.js'

export interface BuildStackOptions {
  args: string[]
  config: Config
  detected?: DetectedProject
  configManager?: { saveConfig(config: Config): void }
}

function resolveProject (config: Config, detected?: DetectedProject): DetectedProject | null {
  if (detected) return detected
  if (config.recentProjects.length > 0 && config.accounts.length > 0) {
    const recent = config.recentProjects[0]
    const account = config.accounts.find((a) => a.name === recent.accountName)
    if (account) {
      return { account, projectPath: recent.projectPath, localPath: recent.localPath ?? '' }
    }
  }
  return null
}

function makeProjectScreen (
  project: DetectedProject,
  config: Config,
  configManager?: { saveConfig(config: Config): void },
): Screen {
  return {
    id: 'project',
    component: ProjectScreen,
    props: { project, config, configManager: configManager ?? { saveConfig: () => {} } },
  }
}

function makeMRListScreen (project: DetectedProject): Screen {
  return {
    id: 'mr-list',
    component: MRListScreen,
    props: { account: project.account, projectPath: project.projectPath, localPath: project.localPath },
  }
}

export function buildInitialStack ({ args, config, detected, configManager }: BuildStackOptions): Screen[] {
  const project = resolveProject(config, detected)

  if (!project) {
    return [{ id: 'home', component: HomeScreen,
      props: { config, configManager: configManager ?? { saveConfig: () => {} } } }]
  }

  const sectionArgs = args[0] === project.projectPath ? args.slice(1) : args
  const firstArg = sectionArgs[0]

  if (firstArg === 'mr' || firstArg === 'mrs') {
    const deepLink = parseDeepLink(sectionArgs)
    const projectScreen = makeProjectScreen(project, config, configManager)
    const mrListScreen = makeMRListScreen(project)
    if (deepLink.type === 'mr-detail') {
      return [projectScreen, { ...mrListScreen, props: { ...mrListScreen.props, deepLinkIid: deepLink.iid } }]
    }
    return [projectScreen, mrListScreen]
  }

  return [makeProjectScreen(project, config, configManager)]
}
