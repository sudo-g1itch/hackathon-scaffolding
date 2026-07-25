'use client'

// People I Support — a caregiver's list of everyone who chose them.
//
// This screen stays a signals-only overview: risk, mood, streak, plan progress,
// waiting messages. It never shows a transcript, and the backend does not send
// one. Opening a person leads to their detail view, where check-in wording
// appears only if they have chosen to share it.
import { useCallback, useEffect, useRef, useState } from 'react'

import { useRouter } from 'next/navigation'

import Alert from '@mui/material/Alert'
import Avatar from '@mui/material/Avatar'
import Badge from '@mui/material/Badge'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CircularProgress from '@mui/material/CircularProgress'
import Grid from '@mui/material/Grid'
import LinearProgress from '@mui/material/LinearProgress'
import Paper from '@mui/material/Paper'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableContainer from '@mui/material/TableContainer'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Typography from '@mui/material/Typography'
import { Bar, BarChart, Cell, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'

import RiskChip from '@components/anchor-one/RiskChip'
import EmptyState from '@components/EmptyState'
import StatCard from '@components/StatCard'
import type { CaregiverPatient, RiskLevel } from '@/types/anchorOneTypes'
import { anchorOneService } from '@/services/anchorOneService'
import { getApiErrorMessage } from '@/utils/handleApiError'

const POLL_INTERVAL_MS = 15_000

const RISK_FILLS: Record<RiskLevel, string> = {
  LOW: 'var(--mui-palette-success-main)',
  MEDIUM: 'var(--mui-palette-warning-main)',
  HIGH: 'var(--mui-palette-error-main)'
}

const countByRisk = (patients: CaregiverPatient[], level: RiskLevel) =>
  patients.filter(patient => patient.risk === level).length

const CaregiverPage = () => {
  const router = useRouter()
  const [patients, setPatients] = useState<CaregiverPatient[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Remembers the previous high-risk count so a *new* escalation can chime.
  const previousHighRisk = useRef<number | null>(null)

  const playAlert = useCallback(() => {
    try {
      const AudioContextCtor = window.AudioContext

      if (!AudioContextCtor) return

      const context = new AudioContextCtor()
      const oscillator = context.createOscillator()
      const gain = context.createGain()

      oscillator.connect(gain)
      gain.connect(context.destination)
      oscillator.type = 'sine'
      oscillator.frequency.setValueAtTime(880, context.currentTime)
      gain.gain.setValueAtTime(0.25, context.currentTime)
      gain.gain.exponentialRampToValueAtTime(0.0001, context.currentTime + 0.6)
      oscillator.start()
      oscillator.stop(context.currentTime + 0.6)
      oscillator.onended = () => void context.close()
    } catch {
      // Audio alerts are a nicety; a blocked AudioContext must never break the page.
    }
  }, [])

  const load = useCallback(
    async (isPoll: boolean) => {
      try {
        const data = await anchorOneService.getCaregiverData()

        setPatients(data)
        setError(null)

        const highRisk = countByRisk(data, 'HIGH')

        if (isPoll && previousHighRisk.current !== null && highRisk > previousHighRisk.current) {
          playAlert()
        }

        previousHighRisk.current = highRisk
      } catch (err) {
        setError(getApiErrorMessage(err, 'Could not load your patients.'))
      } finally {
        if (!isPoll) setLoading(false)
      }
    },
    [playAlert]
  )

  useEffect(() => {
    void load(false)

    const interval = setInterval(() => void load(true), POLL_INTERVAL_MS)

    return () => clearInterval(interval)
  }, [load])

  if (loading) {
    return (
      <Box display='flex' justifyContent='center' alignItems='center' height='50vh'>
        <CircularProgress />
      </Box>
    )
  }

  const highCount = countByRisk(patients, 'HIGH')
  const unreadTotal = patients.reduce((total, patient) => total + patient.unread_messages, 0)

  const chartData = (['LOW', 'MEDIUM', 'HIGH'] as RiskLevel[]).map(level => ({
    name: level,
    count: countByRisk(patients, level)
  }))

  return (
    <Box className='flex flex-col gap-6'>
      <Box>
        <Typography variant='h4' fontWeight={700}>
          People I Support
        </Typography>
        <Typography variant='body2' color='text.secondary'>
          Everyone who chose you as their caregiver. Open someone to see their plan and message them.
        </Typography>
      </Box>

      {error && <Alert severity='error'>{error}</Alert>}

      {highCount > 0 && (
        <Alert severity='error' sx={{ fontWeight: 600 }}>
          {highCount} {highCount === 1 ? 'person is' : 'people are'} at HIGH risk right now. Consider reaching out.
        </Alert>
      )}

      {unreadTotal > 0 && (
        <Alert severity='info'>
          {unreadTotal} unread {unreadTotal === 1 ? 'message is' : 'messages are'} waiting for you.
        </Alert>
      )}

      {patients.length === 0 ? (
        <EmptyState
          icon='ri-group-line'
          title='No one linked yet'
          message='When someone assigns you as their caregiver, they appear here.'
        />
      ) : (
        <Grid container spacing={6}>
          <Grid size={{ xs: 12, sm: 4 }}>
            <StatCard label='People supported' value={patients.length} icon='ri-group-line' color='primary' />
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <StatCard label='At high risk' value={highCount} icon='ri-alarm-warning-line' color='error' />
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <StatCard
              label='Emergencies logged'
              value={patients.reduce((total, patient) => total + patient.emergency_count, 0)}
              icon='ri-first-aid-kit-line'
              color='warning'
            />
          </Grid>

          <Grid size={{ xs: 12, md: 4 }}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Typography variant='h6' gutterBottom>
                  Risk distribution
                </Typography>
                <Box sx={{ height: 280 }}>
                  <ResponsiveContainer width='100%' height='100%'>
                    <BarChart data={chartData}>
                      <CartesianGrid strokeDasharray='3 3' vertical={false} />
                      <XAxis dataKey='name' tickLine={false} />
                      <YAxis allowDecimals={false} tickLine={false} />
                      <Tooltip />
                      <Bar dataKey='count' radius={[4, 4, 0, 0]}>
                        {chartData.map(entry => (
                          <Cell key={entry.name} fill={RISK_FILLS[entry.name as RiskLevel]} />
                        ))}
                      </Bar>
                    </BarChart>
                  </ResponsiveContainer>
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid size={{ xs: 12, md: 8 }}>
            <TableContainer component={Paper} variant='outlined'>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Person</TableCell>
                    <TableCell>Mood</TableCell>
                    <TableCell>Streak</TableCell>
                    <TableCell>Plan</TableCell>
                    <TableCell>Last check-in</TableCell>
                    <TableCell>Risk</TableCell>
                    <TableCell align='right'>Open</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {patients.map(patient => (
                    <TableRow
                      key={patient.user_id}
                      hover
                      sx={{ cursor: 'pointer' }}
                      onClick={() => router.push(`/anchor-one/caregiver/${patient.user_id}`)}
                    >
                      <TableCell>
                        <Box className='flex items-center gap-3'>
                          <Badge
                            color='error'
                            badgeContent={patient.unread_messages}
                            invisible={patient.unread_messages === 0}
                          >
                            <Avatar>{patient.name ? patient.name.charAt(0).toUpperCase() : '?'}</Avatar>
                          </Badge>
                          <Box>
                            <Typography fontWeight={500}>{patient.name || 'Unknown'}</Typography>
                            <Typography variant='caption' color='text.secondary'>
                              {patient.goal || patient.substance || 'No goal set'}
                            </Typography>
                          </Box>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Typography variant='body2'>{patient.emotion}</Typography>
                        <Typography variant='caption' color='text.secondary'>
                          craving {patient.craving}/10
                        </Typography>
                      </TableCell>
                      <TableCell>{patient.recovery_streak} d</TableCell>
                      <TableCell sx={{ minWidth: 140 }}>
                        <Typography variant='caption' color='text.secondary'>
                          {patient.active_goals} active · {patient.completed_goals} done
                        </Typography>
                        <LinearProgress
                          variant='determinate'
                          value={patient.average_progress}
                          sx={{ height: 6, borderRadius: 3, mt: 1 }}
                        />
                      </TableCell>
                      <TableCell>
                        {patient.last_checkin_at ? (
                          <Typography variant='body2'>{new Date(patient.last_checkin_at).toLocaleString()}</Typography>
                        ) : (
                          <Typography variant='body2' color='text.secondary'>
                            Never
                          </Typography>
                        )}
                      </TableCell>
                      <TableCell>
                        <RiskChip risk={patient.risk} />
                      </TableCell>
                      <TableCell align='right'>
                        <i className='ri-arrow-right-s-line text-textSecondary' />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>

            <Box className='flex justify-end mbs-3'>
              <Button size='small' color='secondary' onClick={() => void load(false)}>
                Refresh now
              </Button>
            </Box>
          </Grid>
        </Grid>
      )}
    </Box>
  )
}

export default CaregiverPage
