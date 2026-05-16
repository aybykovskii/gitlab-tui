export interface Account {
  name: string
  url: string
  token: string
}

export interface RecentProject {
  accountName: string
  projectPath: string
  localPath?: string
}

export interface Config {
  accounts: Account[]
  defaultAccount: string
  recentProjects: RecentProject[]
  editor: string
}
