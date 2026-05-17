import { describe, expect, it } from 'vitest'

import { createMRTemplateLoader } from './mrTemplates.js'

describe('listTemplates', () => {
  it('returns template names without .md extension', async () => {
    const loader = createMRTemplateLoader('/repo', {
      readDir: async () => ['bug_fix.md', 'feature.md', 'README.md'],
      readFile: async () => '',
    })

    const names = await loader.listTemplates()

    expect(names).toEqual(['bug_fix', 'feature', 'README'])
  })

  it('returns empty array when directory does not exist', async () => {
    const loader = createMRTemplateLoader('/repo', {
      readDir: async () => {
        throw Object.assign(new Error('ENOENT'), { code: 'ENOENT' })
      },
      readFile: async () => '',
    })

    const names = await loader.listTemplates()

    expect(names).toEqual([])
  })
})

describe('getTemplateContent', () => {
  it('returns the file content for the given template name', async () => {
    const content = '## What to build\n\n...'
    const loader = createMRTemplateLoader('/repo', {
      readDir: async () => [],
      readFile: async (path) => {
        if (path.endsWith('feature.md')) return content
        throw new Error(`unexpected path: ${path}`)
      },
    })

    const result = await loader.getTemplateContent('feature')

    expect(result).toBe(content)
  })
})
