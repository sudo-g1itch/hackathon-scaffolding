export type PrimaryColorConfig = {
  name?: string
  light?: string
  main: string
  dark?: string
}

// Primary color config object
const primaryColorConfig: PrimaryColorConfig[] = [
  {
    name: 'primary-1', // Calming Teal (Default)
    light: '#5EEAD4',
    main: '#0D9488',
    dark: '#0F766E'
  },
  {
    name: 'primary-2', // Soft Medical Blue
    light: '#93C5FD',
    main: '#3B82F6',
    dark: '#1D4ED8'
  },
  {
    name: 'primary-3', // Gentle Lavender
    light: '#C4B5FD',
    main: '#8B5CF6',
    dark: '#6D28D9'
  },
  {
    name: 'primary-4', // Soft Sage Green
    light: '#A7F3D0',
    main: '#10B981',
    dark: '#047857'
  },
  {
    name: 'primary-5', // Warm Coral/Peach
    light: '#FDA4AF',
    main: '#F43F5E',
    dark: '#BE123C'
  }
]

export default primaryColorConfig
