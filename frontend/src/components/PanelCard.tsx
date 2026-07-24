// The outlined panel component: tinted icon badge + title (+ optional count chip)
// header strip with action slot, divider, and collapsible body.
import type { ReactNode } from 'react'
import { useState } from 'react'

import Box from '@mui/material/Box'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CardHeader from '@mui/material/CardHeader'
import Chip from '@mui/material/Chip'
import Collapse from '@mui/material/Collapse'
import Divider from '@mui/material/Divider'
import IconButton from '@mui/material/IconButton'
import Typography from '@mui/material/Typography'
import { alpha } from '@mui/material/styles'

import type { ThemeColor } from '@core/types'

type PanelCardProps = {

  /** Remix icon class for the badge. */
  icon: string
  title: string
  color?: ThemeColor

  /** Shown as a small tonal chip next to the title when > 0. */
  count?: number

  /** Right-aligned header slot (buttons, toggles…). */
  action?: ReactNode

  /** Render children without the default padded CardContent. */
  disablePadding?: boolean

  /** Add a header chevron that collapses/expands the body (default false). */
  collapsible?: boolean

  /** Start collapsed when collapsible (default false — open). */
  defaultCollapsed?: boolean
  children: ReactNode
}

const PanelCard = ({
  icon,
  title,
  color = 'primary',
  count,
  action,
  disablePadding = false,
  collapsible = false,
  defaultCollapsed = false,
  children
}: PanelCardProps) => {
  const [collapsed, setCollapsed] = useState(collapsible ? defaultCollapsed : false)

  const body = disablePadding ? children : <CardContent className='flex flex-col gap-5'>{children}</CardContent>

  const toggle = collapsible ? (
    <IconButton
      size='small'
      onClick={() => setCollapsed(c => !c)}
      aria-label={collapsed ? 'Expand' : 'Collapse'}
    >
      <i className={collapsed ? 'ri-arrow-down-s-line' : 'ri-arrow-up-s-line'} />
    </IconButton>
  ) : null

  return (
    <Card variant='outlined' className='bs-full'>
      <CardHeader
        avatar={
          <Box
            className='flex items-center justify-center is-9 bs-9 rounded'
            sx={{ bgcolor: theme => alpha(theme.palette[color].main, 0.12), color: `${color}.main` }}
          >
            <i className={`${icon} text-xl`} />
          </Box>
        }
        title={
          <Box className='flex items-center gap-2'>
            <Typography variant='h6'>{title}</Typography>
            {count != null && count > 0 ? <Chip size='small' variant='tonal' color={color} label={count} /> : null}
          </Box>
        }
        action={
          toggle ? (
            <Box className='flex items-center gap-1'>
              {action}
              {toggle}
            </Box>
          ) : (
            action
          )
        }
        sx={{ '& .MuiCardHeader-action': { alignSelf: 'center', mbs: 0, mie: 0 } }}
      />
      <Divider />
      {collapsible ? <Collapse in={!collapsed}>{body}</Collapse> : body}
    </Card>
  )
}

export default PanelCard
