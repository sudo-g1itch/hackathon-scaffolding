'use client'

// A caregiver's detail view of one person they support.
//
// Three things a caregiver actually needs and could not previously see: how
// this person is trending, what is on their recovery plan, and a way to say
// something to them. Check-in narrative appears here only when the person has
// switched sharing on; the page says so plainly when they have not, rather than
// showing blanks.
import { use, useCallback, useEffect, useState } from 'react'

import { useRouter } from 'next/navigation'

import Alert from '@mui/material/Alert'
import Avatar from '@mui/material/Avatar'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Chip from '@mui/material/Chip'
import CircularProgress from '@mui/material/CircularProgress'
import Grid from '@mui/material/Grid'
import Snackbar from '@mui/material/Snackbar'
import Tab from '@mui/material/Tab'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableContainer from '@mui/material/TableContainer'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Tabs from '@mui/material/Tabs'
import Typography from '@mui/material/Typography'
import { CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'

import EmptyState from '@components/EmptyState'
import StatCard from '@components/StatCard'
import GoalCard from '@components/anchor-one/GoalCard'
import GoalDetailDialog from '@components/anchor-one/GoalDetailDialog'
import GoalFormDialog from '@components/anchor-one/GoalFormDialog'
import RiskChip from '@components/anchor-one/RiskChip'
import SupportChat from '@components/anchor-one/SupportChat'
import { anchorOneService } from '@/services/anchorOneService'
import type { GoalInput, PatientOverview, RiskLevel } from '@/types/anchorOneTypes'
import { RISK_SCORES } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

type TabKey = 'signals' | 'plan' | 'alerts' | 'messages'

const RISK_TICKS: Record<number, string> = { 1: 'LOW', 2: 'MEDIUM', 3: 'HIGH' }

const PatientDetailPage = ({ params }: { params: Promise<{ patientId: string }> }) => {
  const { patientId } = use(params)
  const router = useRouter()

  const [overview, setOverview] = useState<PatientOverview | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tab, setTab] = useState<TabKey>('signals')
  const [toast, setToast] = useState<string | null>(null)

  const [suggestOpen, setSuggestOpen] = useState(false)
  const [openGoalId, setOpenGoalId] = useState<string | null>(null)
  const [acknowledging, setAcknowledging] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setOverview(await anchorOneService.getPatientOverview(patientId))
      setError(null)
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not load this person’s record.'))
    } finally {
      setLoading(false)
    }
  }, [patientId])

  useEffect(() => {
    void load()
  }, [load])

  const suggestGoal = async (input: GoalInput) => {
    await anchorOneService.createPatientGoal(patientId, input)
    await load()
    setToast('Goal suggested. They can accept, edit or archive it.')
  }

  // Acknowledging tells the sender somebody is there — the app cannot say that
  // on its own, so this is the only thing that makes it true.
  const acknowledge = async (emergencyId: string) => {
    setAcknowledging(emergencyId)

    try {
      await anchorOneService.acknowledgeEmergency(emergencyId)
      await load()
      setToast('They have been told you have seen it.')
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not send that acknowledgement.'))
    } finally {
      setAcknowledging(null)
    }
  }

  if (loading) {
    return (
      <Box display='flex' justifyContent='center' alignItems='center' height='50vh'>
        <CircularProgress />
      </Box>
    )
  }

  if (error || !overview) {
    return (
      <Box className='flex flex-col gap-4'>
        <Alert severity='error'>{error ?? 'That record is not available.'}</Alert>
        <Box>
          <Button variant='outlined' onClick={() => router.push('/anchor-one/caregiver')}>
            Back to my people
          </Button>
        </Box>
      </Box>
    )
  }

  const { patient, checkins, goals, goal_summary: summary, emergencies } = overview
  const unacknowledged = emergencies.filter(alert => !alert.acknowledged_at)

  // Oldest-first so the trend line reads left to right.
  const trend = [...checkins]
    .reverse()
    .map(checkin => ({
      label: new Date(checkin.occurred_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
      score: RISK_SCORES[checkin.risk as RiskLevel] ?? 1,
      emotion: checkin.emotion
    }))

  return (
    <Box className='flex flex-col gap-6'>
      <Box className='flex flex-wrap items-center justify-between gap-4'>
        <Box className='flex items-center gap-4'>
          <Button
            size='small'
            color='secondary'
            onClick={() => router.push('/anchor-one/caregiver')}
            startIcon={<i className='ri-arrow-left-line' />}
          >
            Back
          </Button>
          <Avatar sx={{ width: 48, height: 48 }}>{patient.name ? patient.name.charAt(0).toUpperCase() : '?'}</Avatar>
          <Box>
            <Typography variant='h4' fontWeight={700}>
              {patient.name || 'Unknown'}
            </Typography>
            <Typography variant='body2' color='text.secondary'>
              {patient.goal || 'No headline goal set'}
              {patient.substance && ` · ${patient.substance}`}
            </Typography>
          </Box>
        </Box>
        <RiskChip risk={patient.risk} prominent />
      </Box>

      {unacknowledged.length > 0 && (
        <Alert
          severity='error'
          sx={{ fontWeight: 600 }}
          action={
            <Button size='small' color='inherit' onClick={() => setTab('alerts')}>
              Open
            </Button>
          }
        >
          {unacknowledged.length === 1
            ? 'They sent an SOS you have not acknowledged yet.'
            : `${unacknowledged.length} SOS alerts are waiting for your acknowledgement.`}
        </Alert>
      )}

      {patient.risk === 'HIGH' && (
        <Alert severity='error' sx={{ fontWeight: 600 }}>
          Their last check-in scored HIGH risk. A short message can matter more than you think.
        </Alert>
      )}

      <Grid container spacing={6}>
        <Grid size={{ xs: 6, md: 3 }}>
          <StatCard label='Streak' value={patient.recovery_streak} caption='days' icon='ri-fire-fill' color='success' />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <StatCard label='Mood' value={patient.emotion} icon='ri-emotion-line' color='primary' animate={false} />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <StatCard
            label='Craving'
            value={`${patient.craving}/10`}
            icon='ri-pulse-line'
            color='warning'
            animate={false}
          />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <StatCard
            label='Plan progress'
            value={`${summary.average_progress}%`}
            caption={`${summary.active} active · ${summary.completed} done`}
            icon='ri-flag-line'
            color='info'
            animate={false}
          />
        </Grid>
      </Grid>

      <Card>
        <CardContent className='flex flex-col gap-6'>
          <Tabs value={tab} onChange={(_event, value: TabKey) => setTab(value)} variant='scrollable' scrollButtons='auto'>
            <Tab value='signals' label='Check-ins' />
            <Tab value='plan' label={`Recovery plan (${goals.length})`} />
            <Tab value='alerts' label={emergencies.length > 0 ? `SOS alerts (${emergencies.length})` : 'SOS alerts'} />
            <Tab
              value='messages'
              label={patient.unread_messages > 0 ? `Messages (${patient.unread_messages})` : 'Messages'}
            />
          </Tabs>

          {tab === 'signals' && (
            <Box className='flex flex-col gap-6'>
              {!overview.shares_checkin_details && (
                <Alert severity='info'>
                  {patient.name || 'This person'} has kept the wording of their check-ins private. You can see how they
                  are doing — risk, mood and craving — but not what they said.
                </Alert>
              )}

              {checkins.length === 0 ? (
                <EmptyState icon='ri-mic-line' title='No check-ins yet' message='Their history appears here once they check in.' />
              ) : (
                <>
                  <Box sx={{ height: 260 }}>
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
                        <Tooltip formatter={value => [RISK_TICKS[value as number] ?? '', 'Risk']} />
                        <Line type='monotone' dataKey='score' stroke='var(--mui-palette-primary-main)' strokeWidth={3} />
                      </LineChart>
                    </ResponsiveContainer>
                  </Box>

                  <TableContainer>
                    <Table size='small'>
                      <TableHead>
                        <TableRow>
                          <TableCell>When</TableCell>
                          <TableCell>Risk</TableCell>
                          <TableCell>Mood</TableCell>
                          <TableCell>Craving</TableCell>
                          <TableCell>How</TableCell>
                          {overview.shares_checkin_details && <TableCell>What they shared</TableCell>}
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {checkins.map(checkin => (
                          <TableRow key={checkin.id} hover>
                            <TableCell>{new Date(checkin.occurred_at).toLocaleString()}</TableCell>
                            <TableCell>
                              <RiskChip risk={checkin.risk} />
                            </TableCell>
                            <TableCell>{checkin.emotion || '—'}</TableCell>
                            <TableCell>{checkin.craving}/10</TableCell>
                            <TableCell>
                              <Chip
                                size='small'
                                variant='tonal'
                                color='secondary'
                                label={checkin.source === 'voice' ? 'Voice' : 'Typed'}
                              />
                            </TableCell>
                            {overview.shares_checkin_details && (
                              <TableCell sx={{ maxWidth: 320 }}>
                                <Typography variant='body2' color='text.secondary'>
                                  {checkin.summary || '—'}
                                </Typography>
                                {checkin.triggers && checkin.triggers.length > 0 && (
                                  <Box className='flex flex-wrap gap-1 mbs-1'>
                                    {checkin.triggers.map(trigger => (
                                      <Chip key={trigger} size='small' variant='tonal' color='warning' label={trigger} />
                                    ))}
                                  </Box>
                                )}
                              </TableCell>
                            )}
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                </>
              )}
            </Box>
          )}

          {tab === 'plan' && (
            <Box className='flex flex-col gap-6'>
              <Box className='flex flex-wrap items-center justify-between gap-3'>
                <Typography variant='body2' color='text.secondary'>
                  You can suggest a goal and leave encouragement on any of them. Only {patient.name || 'they'} can log
                  progress.
                </Typography>
                <Button variant='contained' onClick={() => setSuggestOpen(true)} startIcon={<i className='ri-add-line' />}>
                  Suggest a goal
                </Button>
              </Box>

              {goals.length === 0 ? (
                <EmptyState
                  icon='ri-flag-line'
                  title='No goals yet'
                  message='Suggesting a first, small goal is often the easiest place to start.'
                  action={
                    <Button variant='contained' onClick={() => setSuggestOpen(true)}>
                      Suggest a goal
                    </Button>
                  }
                />
              ) : (
                <Grid container spacing={6}>
                  {goals.map(goal => (
                    <Grid key={goal.id} size={{ xs: 12, md: 6 }}>
                      <GoalCard goal={goal} onOpen={() => setOpenGoalId(goal.id)} />
                    </Grid>
                  ))}
                </Grid>
              )}
            </Box>
          )}

          {tab === 'alerts' && (
            <Box className='flex flex-col gap-4'>
              {emergencies.length === 0 ? (
                <EmptyState
                  icon='ri-shield-check-line'
                  title='No SOS alerts'
                  message='If they press HELP ME and send it to you, it appears here with their location and voice note.'
                />
              ) : (
                emergencies.map(alert => (
                  <Card key={alert.id} variant='outlined' sx={{ borderColor: 'error.main' }}>
                    <CardContent className='flex flex-col gap-3'>
                      <Box className='flex flex-wrap items-center justify-between gap-2'>
                        <Typography variant='subtitle2' color='error.main'>
                          <i className='ri-alarm-warning-fill align-middle mie-1' />
                          SOS · {new Date(alert.shared_at ?? alert.occurred_at).toLocaleString()}
                        </Typography>
                        {alert.acknowledged_at ? (
                          <Chip
                            size='small'
                            color='success'
                            variant='tonal'
                            label={`Acknowledged ${new Date(alert.acknowledged_at).toLocaleString()}`}
                          />
                        ) : (
                          <Button
                            size='small'
                            variant='contained'
                            color='success'
                            disabled={acknowledging === alert.id}
                            onClick={() => void acknowledge(alert.id)}
                            startIcon={<i className='ri-check-line' />}
                          >
                            I&apos;ve seen this
                          </Button>
                        )}
                      </Box>

                      <Typography variant='body1' sx={{ whiteSpace: 'pre-wrap' }}>
                        {alert.message}
                      </Typography>

                      {alert.audio_transcript && (
                        <Box>
                          <Typography variant='caption' color='text.secondary'>
                            <i className='ri-mic-line align-middle mie-1' />
                            Voice note
                          </Typography>
                          <Typography variant='body2' fontStyle='italic' sx={{ whiteSpace: 'pre-wrap' }}>
                            “{alert.audio_transcript}”
                          </Typography>
                        </Box>
                      )}

                      {alert.location_url && (
                        <Box>
                          <Button
                            size='small'
                            variant='outlined'
                            color='error'
                            href={alert.location_url}
                            target='_blank'
                            rel='noopener noreferrer'
                            startIcon={<i className='ri-map-pin-line' />}
                          >
                            Open where they were
                          </Button>
                        </Box>
                      )}
                    </CardContent>
                  </Card>
                ))
              )}
            </Box>
          )}

          {tab === 'messages' && (
            <SupportChat patientId={patientId} height='calc(100vh - 420px)' onSent={() => void load()} />
          )}
        </CardContent>
      </Card>

      <GoalFormDialog
        open={suggestOpen}
        onClose={() => setSuggestOpen(false)}
        onSubmit={suggestGoal}
        title={`Suggest a goal for ${patient.name || 'them'}`}
        submitLabel='Suggest goal'
      />

      <GoalDetailDialog
        goalId={openGoalId}
        onClose={() => setOpenGoalId(null)}
        canLogProgress={false}
        onChanged={load}
      />

      <Snackbar
        open={Boolean(toast)}
        autoHideDuration={4000}
        onClose={() => setToast(null)}
        message={toast ?? ''}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      />
    </Box>
  )
}

export default PatientDetailPage
