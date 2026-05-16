import { describe, it, expect, vi } from 'vitest'
import { createDraftNotesAPI } from './api.js'
import type { GitLabClient } from '../../../core/gitlab/index.js'

function makeClient(): GitLabClient {
  return {
    MergeRequestDraftNotes: {
      all: vi.fn().mockResolvedValue([]),
      create: vi.fn().mockResolvedValue({ id: 1, note: 'draft', position: null }),
      publishBulk: vi.fn().mockResolvedValue([]),
      remove: vi.fn().mockResolvedValue(undefined),
    },
    MergeRequestNotes: {
      create: vi.fn().mockResolvedValue({ id: 2, body: 'summary' }),
    },
  } as unknown as GitLabClient
}

describe('createDraftNotesAPI', () => {
  it('creates draft notes through the SDK', async () => {
    const client = makeClient()
    const api = createDraftNotesAPI(client, 'group/project', 7)

    await api.create('Please fix this')

    expect(client.MergeRequestDraftNotes.create).toHaveBeenCalledWith(
      'group/project',
      7,
      'Please fix this',
      undefined,
    )
  })

  it('creates draft replies with inReplyToDiscussionId', async () => {
    const client = makeClient()
    const api = createDraftNotesAPI(client, 'group/project', 7)

    await api.createReply('discussion-123', 'Reply draft')

    expect(client.MergeRequestDraftNotes.create).toHaveBeenCalledWith(
      'group/project',
      7,
      'Reply draft',
      { inReplyToDiscussionId: 'discussion-123' },
    )
  })

  it('lists and removes draft notes through the SDK', async () => {
    const client = makeClient()
    const api = createDraftNotesAPI(client, 'group/project', 7)

    await api.list()
    await api.remove(123)

    expect(client.MergeRequestDraftNotes.all).toHaveBeenCalledWith('group/project', 7)
    expect(client.MergeRequestDraftNotes.remove).toHaveBeenCalledWith('group/project', 7, 123)
  })

  it('creates a merge request summary note before publishing all drafts', async () => {
    const client = makeClient()
    const api = createDraftNotesAPI(client, 'group/project', 7)

    await api.publishAll('Review summary')

    expect(client.MergeRequestNotes.create).toHaveBeenCalledWith('group/project', 7, 'Review summary')
    expect(client.MergeRequestDraftNotes.publishBulk).toHaveBeenCalledWith('group/project', 7)
  })
})
