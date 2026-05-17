import node_path from 'node:path'

const TEMPLATES_DIR = '.gitlab/merge_request_templates'

interface TemplateFS {
  readDir(path: string): Promise<string[]>
  readFile(path: string): Promise<string>
}

export function createMRTemplateLoader (repoRoot: string, fs?: TemplateFS) {
  const resolvedFs: TemplateFS = fs ?? {
    readDir: async (p) => {
      const { readdir } = await import('node:fs/promises')
      return readdir(p)
    },
    readFile: async (p) => {
      const { readFile } = await import('node:fs/promises')
      return readFile(p, 'utf8')
    },
  }

  const templatesDir = node_path.join(repoRoot, TEMPLATES_DIR)

  async function listTemplates (): Promise<string[]> {
    try {
      const files = await resolvedFs.readDir(templatesDir)
      return files
        .filter((f) => f.endsWith('.md'))
        .map((f) => f.slice(0, -3))
    } catch (e) {
      if ((e as NodeJS.ErrnoException).code === 'ENOENT') return []
      throw e
    }
  }

  async function getTemplateContent (name: string): Promise<string> {
    return resolvedFs.readFile(node_path.join(templatesDir, `${name}.md`))
  }

  return { listTemplates, getTemplateContent }
}
