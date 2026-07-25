'use client'

// A goal's full history, and the box for adding to it.
//
// This is the one screen where the two sides of the relationship meet on a
// goal: the person in recovery logs progress and writes notes; their caregiver
// can only add encouragement. That asymmetry is enforced by the API — here it
// decides which controls are even shown, so nobody is offered a button that
// would come back 403.
import { useCallback, useEffect, useState } from 'react'

import Alert from '@mui/material/Alert'
import Avatar from '@mui/material/Avatar'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Chip from '@mui/material/Chip'
import CircularProgress from '@mui/material/CircularProgress'
import Dialog from '@mui/material/Dialog'
import DialogContent from '@mui/material/DialogContent'
import DialogTitle from '@mui/material/DialogTitle'
import Divider from '@mui/material/Divider'
import IconButton from '@mui/material/IconButton'
import LinearProgress from '@mui/material/LinearProgress'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import EmptyState from '@components/EmptyState'
import { anchorOneService } from '@/services/anchorOneService'
import type { GoalDetail, GoalUpdate } from '@/types/anchorOneTypes'
import { resolveGoalCategory, resolveGoalStatus } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

type GoalDetailDialogProps = {
  goalId: string | null
  onClose: () => void

  /** Whether the viewer owns this goal. Owners move the number; others cheer. */
  canLogProgress: boolean

  /** Called after any change, so the list behind the dialog can refresh. */
  onChanged?: () => void
}

const UPDATE_ICONS: Record<GoalUpdate['kind'], string> = {
  progress: 'ri-arrow-up-line',
  note: 'ri-sticky-note-line',
  encouragement: 'ri-heart-line',
  status: 'ri-flag-line'
}

const GoalDetailDialog = ({ goalId, onClose, canLogProgress, onChanged }: GoalDetailDialogProps) => {
  const [detail, setDetail] = useState<GoalDetail | null>(null)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [note, setNote] = useState('')
  const [step, setStep] = useState(1)

  const load = useCallback(async (id: string) => {
    setLoading(true)

    try {
      setDetail(await anchorOneService.getGoal(id))
      setError(null)
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not load that goal.'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!goalId) {
      setDetail(null)
      setNote('')
      setStep(1)

      return
    }

    void load(goalId)
  }, [goalId, load])

  const submit = async (delta?: number) => {
    if (!goalId) return

    const trimmed = note.trim()

    // The API requires movement or words; asking for nothing is a no-op.
    if (delta === undefined && !trimmed) {
      setError('Write a note, or log some progress.')

      return
    }

    setSaving(true)
    setError(null)

    try {
      const updated = await anchorOneService.logGoalProgress(goalId, {
        ...(delta !== undefined ? { delta } : {}),
        note: trimmed,
        kind: delta !== undefined ? 'progress' : canLogProgress ? 'note' : 'encouragement'
      })

      setDetail(updated)
      setNote('')
      onChanged?.()
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not save that update.'))
    } finally {
      setSaving(false)
    }
  }

  const goal = detail?.goal
  const category = resolveGoalCategory(goal?.category)
  const status = resolveGoalStatus(goal?.status)

  return (
    <Dialog open={Boolean(goalId)} onClose={saving ? undefined : onClose} fullWidth maxWidth='sm'>
      <DialogTitle className='flex items-start justify-between gap-3'>
        <Box className='flex items-center gap-3'>
          <i className={`${category.icon} text-[22px]`} />
          <Box>
            <Typography variant='h6'>{goal?.title ?? 'Goal'}</Typography>
            {goal && (
              <Chip size='small' variant='tonal' color={status.color} label={status.label} sx={{ mt: 1 }} />
            )}
          </Box>
        </Box>
        <IconButton size='small' onClick={onClose} disabled={saving}>
          <i className='ri-close-line' />
        </IconButton>
      </DialogTitle>

      <DialogContent className='flex flex-col gap-5'>
        {error && <Alert severity='error'>{error}</Alert>}

        {loading || !goal ? (
          <Box className='flex justify-center plb-10'>
            <CircularProgress />
          </Box>
        ) : (
          <>
            {goal.description && (
              <Typography variant='body2' color='text.secondary'>
                {goal.description}
              </Typography>
            )}

            <Box className='flex flex-col gap-2'>
              <Box className='flex items-baseline justify-between'>
                <Typography variant='body2' sx={{ fontWeight: 600 }}>
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
            </Box>

            <Divider />

            <Box className='flex flex-col gap-3'>
              <TextField
                fullWidth
                multiline
                minRows={2}
                size='small'
                placeholder={
                  canLogProgress ? 'How did it go? (optional with a progress step)' : 'Send a word of encouragement…'
                }
                value={note}
                onChange={event => setNote(event.target.value)}
                disabled={saving}
                inputProps={{ maxLength: 1000 }}
              />

              <Box className='flex flex-wrap items-center gap-2'>
                {canLogProgress && goal.status !== 'completed' && (
                  <>
                    <TextField
                      type='number'
                      size='small'
                      label='Step'
                      value={step}
                      onChange={event => setStep(Math.max(1, Number(event.target.value) || 1))}
                      inputProps={{ min: 1, max: 1000 }}
                      sx={{ width: 96 }}
                    />
                    <Button
                      variant='contained'
                      disabled={saving}
                      onClick={() => void submit(step)}
                      startIcon={<i className='ri-add-line' />}
                    >
                      Add {step} {goal.unit}
                    </Button>
                  </>
                )}
                <Button
                  variant='outlined'
                  disabled={saving || !note.trim()}
                  onClick={() => void submit()}
                  startIcon={<i className={canLogProgress ? 'ri-sticky-note-line' : 'ri-heart-line'} />}
                >
                  {canLogProgress ? 'Add note' : 'Send encouragement'}
                </Button>
              </Box>
            </Box>

            <Divider />

            <Typography variant='subtitle2'>History</Typography>

            {detail.updates.length === 0 ? (
              <EmptyState
                icon='ri-history-line'
                message='Nothing logged yet. The first step counts.'
                size='sm'
              />
            ) : (
              <Box className='flex flex-col gap-4'>
                {detail.updates.map(update => (
                  <Box key={update.id} className='flex items-start gap-3'>
                    <Avatar
                      sx={{
                        width: 32,
                        height: 32,
                        bgcolor: update.author_role === 'caregiver' ? 'warning.main' : 'primary.main'
                      }}
                    >
                      <i className={`${UPDATE_ICONS[update.kind]} text-[16px]`} />
                    </Avatar>
                    <Box className='flex-grow'>
                      <Box className='flex flex-wrap items-baseline gap-2'>
                        <Typography variant='body2' sx={{ fontWeight: 600 }}>
                          {update.author_name || (update.author_role === 'caregiver' ? 'Caregiver' : 'You')}
                        </Typography>
                        {update.delta !== 0 && (
                          <Chip
                            size='small'
                            variant='tonal'
                            color={update.delta > 0 ? 'success' : 'warning'}
                            label={`${update.delta > 0 ? '+' : ''}${update.delta} → ${update.value}`}
                          />
                        )}
                        <Typography variant='caption' color='text.secondary'>
                          {new Date(update.created_at).toLocaleString()}
                        </Typography>
                      </Box>
                      {update.note && (
                        <Typography variant='body2' color='text.secondary' sx={{ whiteSpace: 'pre-wrap' }}>
                          {update.note}
                        </Typography>
                      )}
                    </Box>
                  </Box>
                ))}
              </Box>
            )}
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}

export default GoalDetailDialog
