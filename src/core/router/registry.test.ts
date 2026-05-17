import { beforeEach, describe, expect, it } from 'vitest'

import { clearRegistry, getRegisteredFeatures, registerFeature } from './registry.js'
import type { Feature } from './types.js'

const makeFeature = (id: string): Feature => ({
  id,
  name: `Feature ${id}`,
  routes: [],
  navigationItems: [],
})

beforeEach(() => {
  clearRegistry()
})

describe('getRegisteredFeatures', () => {
  it('returns empty array before any feature is registered', () => {
    expect(getRegisteredFeatures()).toEqual([])
  })

  it('returns a feature after it is registered', () => {
    const feature = makeFeature('mrs')
    registerFeature(feature)
    expect(getRegisteredFeatures()).toEqual([feature])
  })

  it('returns multiple features in registration order', () => {
    const mrs = makeFeature('mrs')
    const pipelines = makeFeature('pipelines')
    registerFeature(mrs)
    registerFeature(pipelines)
    expect(getRegisteredFeatures()).toEqual([mrs, pipelines])
  })
})

describe('registerFeature', () => {
  it('throws when registering a duplicate feature id', () => {
    registerFeature(makeFeature('mrs'))
    expect(() => registerFeature(makeFeature('mrs'))).toThrow('Feature "mrs" is already registered')
  })
})
