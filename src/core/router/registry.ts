import type { Feature } from './types.js'

const registry: Feature[] = []

export function registerFeature (feature: Feature): void {
  if (registry.some((f) => f.id === feature.id)) {
    throw new Error(`Feature "${feature.id}" is already registered`)
  }
  registry.push(feature)
}

export function getRegisteredFeatures (): readonly Feature[] {
  return registry
}

export function clearRegistry (): void {
  registry.length = 0
}
