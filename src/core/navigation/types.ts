import type { ComponentType } from 'react'

export interface ScreenProps {
  leftWidth: number
  rightWidth: number
}

export interface Screen {
  id: string
  component: ComponentType<ScreenProps>
  props?: Record<string, unknown>
}
