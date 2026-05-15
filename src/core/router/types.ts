import type { ComponentType } from 'react'

export interface Route {
  pattern: string
  screen: ComponentType
}

export interface NavigationItem {
  key: string
  label: string
  route: string
}

export interface Feature {
  id: string
  name: string
  routes: Route[]
  navigationItems: NavigationItem[]
}
