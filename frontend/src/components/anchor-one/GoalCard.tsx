'use client'

// One goal on a recovery plan: what it is, how far along it is, and the two
// actions that matter most — nudge the number, or open the full history.
//
// Shared by the user's own plan and by the caregiver's view of someone else's,
// so `readOnly` is the only difference between the two renderings.
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Chip from '@mui/material/Chip'
import IconButton from '@mui/material/IconButton'
import LinearProgress from '@mui/material/LinearProgress'
import Tooltip from '@mui/material/Tooltip'
import Typography from '@mui/material/Typography'
import { alpha, useTheme } from '@mui/material/styles'

import type { Goal } from '@/types/anchorOneTypes'
import { resolveGoalCategory, resolveGoalStatus } from '@/types/anchorOneTypes'

type GoalCardProps = {
  goal: Goal

  /** Log a step of progress. Omitted for a viewer who may not move the number. */
  onLogProgress?: (goal: Goal) => void
  onOpen: (goal: Goal) => void
  onEdit?: (goal: Goal) => void
  onDelete?: (goal: Goal) => void
}

/** Human phrasing for a deadline, including the overdue case. */
const deadlineLabel = (goal: Goal): string | null => {
  if (goal.days_remaining === null) return null
  if (goal.status === 'completed') return 'Completed'
  if (goal.days_remaining < 0) return `${Math.abs(goal.days_remaining)} days overdue`
  if (goal.days_remaining === 0) return 'Due today'
  if (goal.days_remaining === 1) return '1 day left'

  return `${goal.days_remaining} days left`
}

const GoalCard = ({ goal, onLogProgress, onOpen, onEdit, onDelete }: GoalCardProps) => {
  const theme = useTheme()
  const category = resolveGoalCategory(goal.category)
  const status = resolveGoalStatus(goal.status)
  const accent = theme.palette[category.color].main
  const deadline = deadlineLabel(goal)
  const isOverdue = goal.status === 'active' && goal.days_remaining !== null && goal.days_remaining < 0

  return (
    <Card
      sx={{
        height: '100%',
        position: 'relative',
        overflow: 'hidden',
        '&::before': {
          content: '""',
          position: 'absolute',
          insetBlock: 0,
          insetInlineStart: 0,
          inlineSize: 4,
          background: `linear-gradient(to bottom, ${accent}, ${alpha(accent, 0.4)})`
        }
      }}
    >
      <CardContent className='flex flex-col gap-4'>
        <Box className='flex items-start justify-between gap-3'>
          <Box className='flex items-start gap-3'>
            <Box
              sx={{
                width: 44,
                height: 44,
                flexShrink: 0,
                borderRadius: 1.5,
                display: 'grid',
                placeItems: 'center',
                color: accent,
                background: `linear-gradient(135deg, ${alpha(accent, 0.2)}, ${alpha(accent, 0.07)})`
              }}
            >
              <i className={`${category.icon} text-[22px]`} />
            </Box>
            <Box>
              <Typography variant='h6' sx={{ lineHeight: 1.3 }}>
                {goal.title}
              </Typography>
              <Typography variant='caption' color='text.secondary'>
                {category.label}
                {goal.created_by_role === 'caregiver' && ' · suggested by your caregiver'}
              </Typography>
            </Box>
          </Box>

          <Chip size='small' variant='tonal' color={status.color} label={status.label} />
        </Box>

        {goal.description && (
          <Typography variant='body2' color='text.secondary'>
            {goal.description}
          </Typography>
        )}

        <Box className='flex flex-col gap-2'>
          <Box className='flex items-baseline justify-between gap-2'>
            <Typography variant='body2' sx={{ fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
              {goal.current_value} / {goal.target_value} {goal.unit}
            </Typography>
            <Typography variant='body2' color='text.secondary'>
              {goal.progress_percent}%
            </Typography>
          </Box>
          <LinearProgress
            variant='determinate'
            value={goal.progress_percent}
            color={goal.status === 'completed' ? 'success' : category.color}
            sx={{ height: 8, borderRadius: 4 }}
          />
          {deadline && (
            <Typography variant='caption' color={isOverdue ? 'error.main' : 'text.secondary'}>
              <i className='ri-calendar-line align-middle mie-1' />
              {deadline}
            </Typography>
          )}
        </Box>

        <Box className='flex flex-wrap items-center gap-2'>
          {onLogProgress && goal.status !== 'completed' && (
            <Button
              size='small'
              variant='contained'
              onClick={() => onLogProgress(goal)}
              startIcon={<i className='ri-add-line' />}
            >
              Log progress
            </Button>
          )}
          <Button size='small' color='secondary' onClick={() => onOpen(goal)} startIcon={<i className='ri-history-line' />}>
            History
          </Button>

          <Box className='flex-grow' />

          {onEdit && (
            <Tooltip title='Edit goal'>
              <IconButton size='small' onClick={() => onEdit(goal)}>
                <i className='ri-pencil-line' />
              </IconButton>
            </Tooltip>
          )}
          {onDelete && (
            <Tooltip title='Delete goal'>
              <IconButton size='small' color='error' onClick={() => onDelete(goal)}>
                <i className='ri-delete-bin-line' />
              </IconButton>
            </Tooltip>
          )}
        </Box>
      </CardContent>
    </Card>
  )
}

export default GoalCard
