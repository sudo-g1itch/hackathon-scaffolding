'use client'

// Create or edit one goal. The same dialog serves both, and also serves a
// caregiver suggesting a goal for someone they support — the only difference is
// the copy, which the caller passes in.
import { useEffect, useState } from 'react'

import Alert from '@mui/material/Alert'
import Button from '@mui/material/Button'
import CircularProgress from '@mui/material/CircularProgress'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogTitle from '@mui/material/DialogTitle'
import Grid from '@mui/material/Grid'
import MenuItem from '@mui/material/MenuItem'
import TextField from '@mui/material/TextField'

import type { Goal, GoalCategory, GoalInput } from '@/types/anchorOneTypes'
import { GOAL_CATEGORIES, GOAL_CATEGORY_ORDER, GOAL_UNITS } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

type GoalFormDialogProps = {
  open: boolean
  onClose: () => void

  /** Present when editing; absent when creating. */
  goal?: Goal | null
  onSubmit: (input: GoalInput) => Promise<void>
  title?: string
  submitLabel?: string
}

const EMPTY: GoalInput = {
  title: '',
  description: '',
  category: 'sobriety',
  target_value: 30,
  unit: 'days',
  target_date: null
}

/** <input type="date"> wants YYYY-MM-DD; the API speaks RFC3339. */
const toDateInput = (iso: string | null): string => (iso ? iso.slice(0, 10) : '')

const fromDateInput = (value: string): string | null =>
  value ? new Date(`${value}T00:00:00Z`).toISOString() : null

const GoalFormDialog = ({ open, onClose, goal, onSubmit, title, submitLabel }: GoalFormDialogProps) => {
  const [form, setForm] = useState<GoalInput>(EMPTY)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Reset every time the dialog opens, so a cancelled edit never leaks into the
  // next one.
  useEffect(() => {
    if (!open) return

    setError(null)
    setForm(
      goal
        ? {
            title: goal.title,
            description: goal.description,
            category: goal.category,
            target_value: goal.target_value,
            unit: goal.unit,
            target_date: goal.target_date
          }
        : EMPTY
    )
  }, [open, goal])

  const update =
    <K extends keyof GoalInput>(field: K) =>
    (value: GoalInput[K]) =>
      setForm(previous => ({ ...previous, [field]: value }))

  const handleSubmit = async () => {
    if (!form.title.trim()) {
      setError('Give your goal a name.')

      return
    }

    setSaving(true)
    setError(null)

    try {
      await onSubmit({ ...form, title: form.title.trim() })
      onClose()
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not save that goal.'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onClose={saving ? undefined : onClose} fullWidth maxWidth='sm'>
      <DialogTitle>{title ?? (goal ? 'Edit goal' : 'New goal')}</DialogTitle>
      <DialogContent>
        <Grid container spacing={5} sx={{ pt: 2 }}>
          {error && (
            <Grid size={{ xs: 12 }}>
              <Alert severity='error'>{error}</Alert>
            </Grid>
          )}

          <Grid size={{ xs: 12 }}>
            <TextField
              fullWidth
              autoFocus
              label='What are you working towards?'
              placeholder='e.g. 90 days sober'
              value={form.title}
              onChange={event => update('title')(event.target.value)}
              inputProps={{ maxLength: 200 }}
            />
          </Grid>

          <Grid size={{ xs: 12 }}>
            <TextField
              fullWidth
              multiline
              minRows={2}
              label='Why does it matter? (optional)'
              placeholder='The reason you will read back to yourself on a hard day.'
              value={form.description}
              onChange={event => update('description')(event.target.value)}
              inputProps={{ maxLength: 2000 }}
            />
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              select
              fullWidth
              label='Category'
              value={form.category}
              onChange={event => update('category')(event.target.value as GoalCategory)}
            >
              {GOAL_CATEGORY_ORDER.map(category => (
                <MenuItem key={category} value={category}>
                  {GOAL_CATEGORIES[category].label}
                </MenuItem>
              ))}
            </TextField>
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              fullWidth
              type='date'
              label='Target date (optional)'
              value={toDateInput(form.target_date)}
              onChange={event => update('target_date')(fromDateInput(event.target.value))}
              InputLabelProps={{ shrink: true }}
            />
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              fullWidth
              type='number'
              label='Target'
              value={form.target_value}
              onChange={event => update('target_value')(Math.max(1, Number(event.target.value) || 1))}
              inputProps={{ min: 1, max: 100000 }}
              helperText='How many before this is done?'
            />
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              select
              fullWidth
              label='Measured in'
              value={form.unit}
              onChange={event => update('unit')(event.target.value)}
            >
              {GOAL_UNITS.map(unit => (
                <MenuItem key={unit} value={unit}>
                  {unit}
                </MenuItem>
              ))}
            </TextField>
          </Grid>
        </Grid>
      </DialogContent>
      <DialogActions>
        <Button color='secondary' onClick={onClose} disabled={saving}>
          Cancel
        </Button>
        <Button
          variant='contained'
          onClick={() => void handleSubmit()}
          disabled={saving}
          startIcon={saving ? <CircularProgress size={16} color='inherit' /> : <i className='ri-save-line' />}
        >
          {submitLabel ?? (goal ? 'Save changes' : 'Add goal')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default GoalFormDialog
