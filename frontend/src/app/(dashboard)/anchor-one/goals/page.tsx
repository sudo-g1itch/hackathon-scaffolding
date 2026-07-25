'use client'

// My Goals — the recovery plan itself.
//
// A person in recovery holds several commitments at once, so this screen is a
// list of goals rather than a single field: each with its own target, its own
// progress and its own history. The summary strip on top is the same roll-up
// the dashboard and the caregiver's view read, so all three agree.
import { useCallback, useEffect, useMemo, useState } from 'react'

import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CircularProgress from '@mui/material/CircularProgress'
import Grid from '@mui/material/Grid'
import Snackbar from '@mui/material/Snackbar'
import Tab from '@mui/material/Tab'
import Tabs from '@mui/material/Tabs'
import Typography from '@mui/material/Typography'

import ConfirmDialog from '@components/ConfirmDialog'
import EmptyState from '@components/EmptyState'
import StatCard from '@components/StatCard'
import GoalCard from '@components/anchor-one/GoalCard'
import GoalDetailDialog from '@components/anchor-one/GoalDetailDialog'
import GoalFormDialog from '@components/anchor-one/GoalFormDialog'
import { useAnchorOne } from '@/contexts/AnchorOneContext'
import { anchorOneService } from '@/services/anchorOneService'
import type { Goal, GoalInput, GoalStatus } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

type FilterKey = 'active' | 'completed' | 'all'

const FILTERS: { key: FilterKey; label: string }[] = [
  { key: 'active', label: 'In progress' },
  { key: 'completed', label: 'Achieved' },
  { key: 'all', label: 'All' }
]

const matchesFilter = (goal: Goal, filter: FilterKey): boolean => {
  if (filter === 'all') return true
  if (filter === 'completed') return goal.status === 'completed'

  return goal.status === 'active' || goal.status === 'paused'
}

const GoalsPage = () => {
  const { refreshDashboard } = useAnchorOne()

  const [goals, setGoals] = useState<Goal[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)
  const [filter, setFilter] = useState<FilterKey>('active')

  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Goal | null>(null)
  const [openGoalId, setOpenGoalId] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<Goal | null>(null)

  const load = useCallback(async () => {
    try {
      setGoals(await anchorOneService.getGoals())
      setError(null)
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not load your goals.'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  // The dashboard's plan summary is derived from these goals, so it is
  // refreshed alongside rather than going stale until the next page load.
  const reload = useCallback(async () => {
    await load()
    await refreshDashboard()
  }, [load, refreshDashboard])

  const visible = useMemo(() => goals.filter(goal => matchesFilter(goal, filter)), [goals, filter])

  const counts = useMemo(() => {
    const byStatus = (status: GoalStatus) => goals.filter(goal => goal.status === status).length
    const open = goals.filter(goal => goal.status === 'active')

    return {
      active: byStatus('active'),
      completed: byStatus('completed'),
      progress: open.length
        ? Math.round(open.reduce((total, goal) => total + goal.progress_percent, 0) / open.length)
        : 0
    }
  }, [goals])

  const handleCreate = async (input: GoalInput) => {
    await anchorOneService.createGoal(input)
    await reload()
    setToast('Goal added to your plan.')
  }

  const handleEdit = async (input: GoalInput) => {
    if (!editing) return

    await anchorOneService.updateGoal(editing.id, {
      ...input,
      clear_target_date: input.target_date === null
    })
    await reload()
    setToast('Goal updated.')
  }

  const handleDelete = async () => {
    if (!deleting) return

    try {
      await anchorOneService.deleteGoal(deleting.id)
      await reload()
      setToast('Goal removed.')
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not remove that goal.'))
    } finally {
      setDeleting(null)
    }
  }

  if (loading) {
    return (
      <Box display='flex' justifyContent='center' alignItems='center' height='50vh'>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box className='flex flex-col gap-6'>
      <Box className='flex flex-wrap items-center justify-between gap-4'>
        <Box>
          <Typography variant='h4' fontWeight={700}>
            My Goals
          </Typography>
          <Typography variant='body2' color='text.secondary'>
            Your recovery plan, one commitment at a time. Small steps count — log them.
          </Typography>
        </Box>
        <Button
          variant='contained'
          size='large'
          onClick={() => {
            setEditing(null)
            setFormOpen(true)
          }}
          startIcon={<i className='ri-add-line' />}
        >
          New goal
        </Button>
      </Box>

      {error && <Alert severity='error'>{error}</Alert>}

      <Grid container spacing={6}>
        <Grid size={{ xs: 12, sm: 4 }}>
          <StatCard label='In progress' value={counts.active} icon='ri-flag-line' color='primary' />
        </Grid>
        <Grid size={{ xs: 12, sm: 4 }}>
          <StatCard label='Achieved' value={counts.completed} icon='ri-trophy-line' color='success' />
        </Grid>
        <Grid size={{ xs: 12, sm: 4 }}>
          <StatCard
            label='Average progress'
            value={`${counts.progress}%`}
            caption='across your open goals'
            icon='ri-line-chart-line'
            color='info'
            animate={false}
          />
        </Grid>
      </Grid>

      <Card>
        <CardContent className='flex flex-col gap-6'>
          <Tabs
            value={filter}
            onChange={(_event, value: FilterKey) => setFilter(value)}
            variant='scrollable'
            scrollButtons='auto'
          >
            {FILTERS.map(item => (
              <Tab key={item.key} value={item.key} label={item.label} />
            ))}
          </Tabs>

          {visible.length === 0 ? (
            <EmptyState
              icon='ri-flag-line'
              title={filter === 'completed' ? 'Nothing achieved yet' : 'No goals here yet'}
              message={
                filter === 'completed'
                  ? 'Finished goals will collect here. The first one is coming.'
                  : 'Add what you are working towards — "90 days sober", "gym twice a week", "call my sponsor daily".'
              }
              action={
                filter !== 'completed' ? (
                  <Button
                    variant='contained'
                    onClick={() => {
                      setEditing(null)
                      setFormOpen(true)
                    }}
                  >
                    Add my first goal
                  </Button>
                ) : undefined
              }
            />
          ) : (
            <Grid container spacing={6}>
              {visible.map(goal => (
                <Grid key={goal.id} size={{ xs: 12, md: 6 }}>
                  <GoalCard
                    goal={goal}
                    onOpen={() => setOpenGoalId(goal.id)}
                    onLogProgress={() => setOpenGoalId(goal.id)}
                    onEdit={() => {
                      setEditing(goal)
                      setFormOpen(true)
                    }}
                    onDelete={() => setDeleting(goal)}
                  />
                </Grid>
              ))}
            </Grid>
          )}
        </CardContent>
      </Card>

      <GoalFormDialog
        open={formOpen}
        goal={editing}
        onClose={() => setFormOpen(false)}
        onSubmit={editing ? handleEdit : handleCreate}
      />

      <GoalDetailDialog goalId={openGoalId} onClose={() => setOpenGoalId(null)} canLogProgress onChanged={reload} />

      <ConfirmDialog
        open={Boolean(deleting)}
        title='Remove this goal?'
        message={`"${deleting?.title ?? ''}" and its history will be removed from your plan. This cannot be undone.`}
        confirmText='Remove'
        onConfirm={() => void handleDelete()}
        onClose={() => setDeleting(null)}
      />

      <Snackbar
        open={Boolean(toast)}
        autoHideDuration={3000}
        onClose={() => setToast(null)}
        message={toast ?? ''}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      />
    </Box>
  )
}

export default GoalsPage
