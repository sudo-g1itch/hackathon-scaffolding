// The shared friendly empty state: large muted icon over a faint ECG trace,
// message, optional title and action.
import type { ReactNode } from 'react'

import Box from '@mui/material/Box'
import Typography from '@mui/material/Typography'

import EcgLine from './EcgLine'

type EmptyStateProps = {

  /** Remix icon class. */
  icon?: string
  title?: string
  message: string

  /** Optional call-to-action (e.g. an "Add" button). */
  action?: ReactNode
  size?: 'sm' | 'md'
}

const EmptyState = ({ icon = 'ri-inbox-2-line', title, message, action, size = 'md' }: EmptyStateProps) => (
  <Box
    className={`relative flex flex-col items-center justify-center gap-2 text-center overflow-hidden ${
      size === 'sm' ? 'plb-8' : 'plb-10'
    }`}
  >
    <EcgLine
      animate
      baseOpacity={0.07}
      pulseOpacity={0.2}
      className='text-textPrimary pointer-events-none'
      style={{
        position: 'absolute',
        insetInline: 0,
        insetBlockEnd: 0,
        blockSize: size === 'sm' ? 32 : 48,
        inlineSize: '100%',
        zIndex: 0
      }}
    />
    <i className={`${icon} ${size === 'sm' ? 'text-[32px]' : 'text-[40px]'} text-textDisabled relative z-[1]`} />
    {title ? (
      <Typography variant='h6' className='relative z-[1]'>
        {title}
      </Typography>
    ) : null}
    <Typography variant='body2' color='text.secondary' className='relative z-[1] max-is-[420px]'>
      {message}
    </Typography>
    {action ? <div className='relative z-[1] mbs-2'>{action}</div> : null}
  </Box>
)

export default EmptyState
