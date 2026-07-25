'use client'

// The AnchorOne home screen: current state at a glance, the PRD's quick
// actions, and the risk trend built from the recovery timeline.
import { useCallback, useEffect, useState } from 'react'

import { useRouter } from 'next/navigation'

import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Chip from '@mui/material/Chip'
import CircularProgress from '@mui/material/CircularProgress'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import { CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'

import EmergencyModal from '@components/anchor-one/EmergencyModal'
import RiskChip from '@components/anchor-one/RiskChip'
import SpeakButton from '@components/anchor-one/SpeakButton'
import VoiceCheckinModal from '@components/anchor-one/VoiceCheckinModal'
import EmptyState from '@components/EmptyState'
import StatCard from '@components/StatCard'
import { useAnchorOne } from '@/contexts/AnchorOneContext'
import { anchorOneService } from '@/services/anchorOneService'
import { RISK_SCORES, type RiskLevel } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

type TrendPoint = {
  label: string
  score: number
  emotion: string
  craving: number
}

const RISK_TICKS: Record<number, string> = { 1: 'LOW', 2: 'MEDIUM', 3: 'HIGH' }

const QUICK_ACTIONS = [
  { label: 'AI Coach', icon: 'ri-robot-line', href: '/anchor-one/coach', color: 'primary' as const },
  { label: 'My Goals', icon: 'ri-flag-line', href: '/anchor-one/goals', color: 'success' as const },
  { label: 'My Caregiver', icon: 'ri-chat-heart-line', href: '/anchor-one/messages', color: 'warning' as const },
  { label: 'Learn', icon: 'ri-book-read-line', href: '/anchor-one/education', color: 'info' as const },
  { label: 'Timeline', icon: 'ri-time-line', href: '/anchor-one/timeline', color: 'secondary' as const },
  { label: 'My Plan', icon: 'ri-user-heart-line', href: '/anchor-one/profile', color: 'success' as const }
]

const AnchorOneDashboard = () => {
  const router = useRouter()
  const { dashboard, loading, error, refreshDashboard } = useAnchorOne()

  const [checkinOpen, setCheckinOpen] = useState(false)
  const [emergencyOpen, setEmergencyOpen] = useState(false)
  const [trend, setTrend] = useState<TrendPoint[]>([])
  const [trendError, setTrendError] = useState<string | null>(null)

  const loadTrend = useCallback(async () => {
    try {
      const events = await anchorOneService.getTimeline()

      const points = events
        .filter(event => event.type === 'checkin' && event.risk)
        .map(event => ({
          label: new Date(event.occurred_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
          score: RISK_SCORES[event.risk as RiskLevel] ?? 1,
          emotion: event.emotion ?? '',
          craving: event.craving ?? 0
        }))
        .reverse()

      setTrend(points)
      setTrendError(null)
    } catch (err) {
      setTrendError(getApiErrorMessage(err, 'Could not load your risk trend.'))
    }
  }, [])

  // Reloads on mount and whenever a new check-in lands.
  useEffect(() => {
    void loadTrend()
  }, [loadTrend, dashboard?.total_checkins])

  if (loading) {
    return (
      <Box display='flex' justifyContent='center' alignItems='center' height='60vh'>
        <CircularProgress />
      </Box>
    )
  }

  const mood = dashboard?.current_mood ?? 'Unknown'
  const streak = dashboard?.recovery_streak ?? 0
  const risk = dashboard?.risk_badge ?? 'LOW'
  const craving = dashboard?.craving_level ?? 0
  const lastCheckin = dashboard?.last_checkin ?? null
  const capabilities = dashboard?.capabilities

  const goals = dashboard?.goals ?? {
    active: 0,
    completed: 0,
    paused: 0,
    archived: 0,
    total: 0,
    average_progress: 0,
    next_goal_title: ''
  }

  const unread = dashboard?.unread_messages ?? 0

  return (
    <Box className='flex flex-col gap-6'>
      <Box className='flex flex-wrap items-center justify-between gap-4'>
        <Box>
          <Typography variant='h4' fontWeight={700}>
            AnchorOne
          </Typography>
          <Typography variant='body2' color='text.secondary'>
            Your recovery copilot. Check in with your voice whenever things feel heavy.
          </Typography>
        </Box>

        <Box className='flex flex-wrap gap-2'>
          <Button
            variant='contained'
            size='large'
            onClick={() => setCheckinOpen(true)}
            startIcon={<i className='ri-mic-fill' />}
          >
            Voice Check-In
          </Button>
          <Button
            variant='contained'
            color='error'
            size='large'
            onClick={() => setEmergencyOpen(true)}
            startIcon={<i className='ri-alarm-warning-fill' />}
            sx={{ fontWeight: 700 }}
          >
            HELP ME
          </Button>
        </Box>
      </Box>

      {error && <Alert severity='error'>{error}</Alert>}

      {capabilities && !capabilities.ai && (
        <Alert severity='warning'>
          AI analysis is not configured on this server (<code>AI_GEMINI_API_KEY</code> is unset), so check-ins, the
          coach and emergency plans are unavailable.
        </Alert>
      )}
      {capabilities?.ai && !capabilities.voice && (
        <Alert severity='info'>
          Voice is not configured (<code>AI_DEEPGRAM_API_KEY</code> is unset). You can still type your check-ins.
        </Alert>
      )}

      {unread > 0 && (
        <Alert
          severity='info'
          icon={<i className='ri-chat-heart-line' />}
          action={
            <Button size='small' onClick={() => router.push('/anchor-one/messages')}>
              Read
            </Button>
          }
        >
          {unread === 1 ? 'Your caregiver sent you a message.' : `Your caregiver sent you ${unread} messages.`}
        </Alert>
      )}

      <Grid container spacing={6}>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard
            label='Recovery Streak'
            value={streak}
            caption='consecutive days'
            icon='ri-fire-fill'
            color='success'
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard label='Current Mood' value={mood} icon='ri-emotion-line' color='primary' animate={false} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard label='Craving Level' value={`${craving}/10`} icon='ri-pulse-line' color='warning' animate={false} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent className='flex flex-col items-center justify-center gap-2 h-full'>
              <Typography variant='body2' color='text.secondary'>
                Relapse Risk
              </Typography>
              <RiskChip risk={risk} prominent />
              <Typography variant='caption' color='text.secondary'>
                {dashboard?.total_checkins ?? 0} check-ins · {dashboard?.emergency_count ?? 0} SOS
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        {/* The recovery plan at a glance. Sits above the quick actions because
            "what am I working on" is the question a person opens the app with. */}
        <Grid size={{ xs: 12 }}>
          <Card>
            <CardContent className='flex flex-wrap items-center justify-between gap-4'>
              <Box className='flex items-center gap-4'>
                <Box
                  sx={{
                    width: 52,
                    height: 52,
                    borderRadius: 1.5,
                    display: 'grid',
                    placeItems: 'center',
                    color: 'success.main',
                    bgcolor: 'var(--mui-palette-success-lightOpacity)'
                  }}
                >
                  <i className='ri-flag-line text-[26px]' />
                </Box>
                <Box>
                  <Typography variant='h6'>
                    {goals.active > 0
                      ? `${goals.active} goal${goals.active === 1 ? '' : 's'} in progress`
                      : 'No goals yet'}
                  </Typography>
                  <Typography variant='body2' color='text.secondary'>
                    {goals.active > 0
                      ? `${goals.average_progress}% average progress${
                          goals.next_goal_title ? ` · next up: ${goals.next_goal_title}` : ''
                        }`
                      : 'Add what you are working towards, and track it here.'}
                  </Typography>
                </Box>
              </Box>

              <Box className='flex items-center gap-3'>
                {goals.completed > 0 && (
                  <Chip
                    variant='tonal'
                    color='success'
                    icon={<i className='ri-trophy-line' />}
                    label={`${goals.completed} achieved`}
                  />
                )}
                <Button variant='contained' onClick={() => router.push('/anchor-one/goals')}>
                  {goals.total > 0 ? 'Open my goals' : 'Add a goal'}
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Quick actions — the PRD's dashboard button row. */}
        <Grid size={{ xs: 12 }}>
          <Card>
            <CardContent className='flex flex-wrap gap-3'>
              {QUICK_ACTIONS.map(action => (
                <Button
                  key={action.href}
                  variant='outlined'
                  color={action.color}
                  onClick={() => router.push(action.href)}
                  startIcon={<i className={action.icon} />}
                >
                  {action.label}
                </Button>
              ))}
            </CardContent>
          </Card>
        </Grid>

        {/* Last check-in detail */}
        <Grid size={{ xs: 12, md: 5 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant='h6' gutterBottom>
                Latest Check-In
              </Typography>

              {!lastCheckin ? (
                <EmptyState
                  icon='ri-mic-line'
                  title='No check-ins yet'
                  message='Tap Voice Check-In and tell AnchorOne how you are doing.'
                  size='sm'
                  action={
                    <Button variant='contained' size='small' onClick={() => setCheckinOpen(true)}>
                      Start check-in
                    </Button>
                  }
                />
              ) : (
                <Box className='flex flex-col gap-3'>
                  <Box className='flex items-center justify-between gap-2'>
                    <Typography variant='caption' color='text.secondary'>
                      {new Date(lastCheckin.created_at).toLocaleString()}
                    </Typography>
                    <RiskChip risk={lastCheckin.risk} />
                  </Box>

                  <Typography variant='body2' color='text.secondary'>
                    {lastCheckin.summary}
                  </Typography>

                  {lastCheckin.triggers && lastCheckin.triggers.length > 0 && (
                    <Box className='flex flex-wrap gap-2'>
                      {lastCheckin.triggers.map(trigger => (
                        <Chip key={trigger} size='small' variant='tonal' color='warning' label={trigger} />
                      ))}
                    </Box>
                  )}

                  {lastCheckin.recommended_actions && lastCheckin.recommended_actions.length > 0 && (
                    <>
                      <Box className='flex items-center justify-between gap-2'>
                        <Typography variant='subtitle2'>Recommended actions</Typography>
                        <SpeakButton text={lastCheckin.recommended_actions.join('. ')} iconOnly />
                      </Box>
                      <Box component='ul' sx={{ pl: 5, m: 0 }}>
                        {lastCheckin.recommended_actions.map(action => (
                          <li key={action}>
                            <Typography variant='body2'>{action}</Typography>
                          </li>
                        ))}
                      </Box>
                    </>
                  )}

                  <Button size='small' color='secondary' onClick={() => void refreshDashboard()}>
                    Refresh
                  </Button>
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* Risk trend */}
        <Grid size={{ xs: 12, md: 7 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant='h6' gutterBottom>
                Risk Trend
              </Typography>
              {trendError && <Alert severity='error'>{trendError}</Alert>}
              <Box sx={{ height: 320 }}>
                {trend.length === 0 ? (
                  <EmptyState
                    icon='ri-line-chart-line'
                    message='Your risk trend appears once you have a couple of check-ins.'
                    size='sm'
                  />
                ) : (
                  <ResponsiveContainer width='100%' height='100%'>
                    <LineChart data={trend} margin={{ top: 8, right: 8, left: 8, bottom: 0 }}>
                      <CartesianGrid strokeDasharray='3 3' vertical={false} />
                      <XAxis dataKey='label' tickLine={false} />
                      <YAxis
                        domain={[0.5, 3.5]}
                        ticks={[1, 2, 3]}
                        tickLine={false}
                        tickFormatter={value => RISK_TICKS[value as number] ?? ''}
                        width={72}
                      />
                      <Tooltip
                        formatter={(value, _name, item) => {
                          const point = item?.payload as TrendPoint | undefined
                          const label = RISK_TICKS[value as number] ?? ''

                          return [point?.emotion ? `${label} — ${point.emotion}` : label, 'Risk']
                        }}
                      />
                      <Line type='monotone' dataKey='score' stroke='var(--mui-palette-primary-main)' strokeWidth={3} />
                    </LineChart>
                  </ResponsiveContainer>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <VoiceCheckinModal open={checkinOpen} onClose={() => setCheckinOpen(false)} />
      <EmergencyModal open={emergencyOpen} onClose={() => setEmergencyOpen(false)} />
    </Box>
  )
}

export default AnchorOneDashboard
