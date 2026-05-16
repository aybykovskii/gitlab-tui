// eslint-disable-next-line @typescript-eslint/no-explicit-any
import type { ComponentType } from 'react'

export interface ScreenProps {
  leftWidth: number
  rightWidth: number
}

export interface Screen {
  id: string
  // any: Screen props extend ScreenProps with feature-specific fields passed via Screen.props
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  component: ComponentType<any>
  props?: Record<string, unknown>
}
