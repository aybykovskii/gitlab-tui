import type { Theme } from './types.js'

export const presets: Record<string, Theme> = {
  default: {
    primary: 'cyan',
    secondary: 'blue',
    success: 'green',
    warning: 'yellow',
    error: 'red',
    muted: 'gray',
    border: 'gray',
  },
  dracula: {
    primary: '#bd93f9',
    secondary: '#8be9fd',
    success: '#50fa7b',
    warning: '#ffb86c',
    error: '#ff5555',
    muted: '#6272a4',
    border: '#44475a',
  },
  nord: {
    primary: '#88c0d0',
    secondary: '#81a1c1',
    success: '#a3be8c',
    warning: '#ebcb8b',
    error: '#bf616a',
    muted: '#4c566a',
    border: '#3b4252',
  },
}
