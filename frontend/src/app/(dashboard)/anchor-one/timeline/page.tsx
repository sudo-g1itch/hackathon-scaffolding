'use client'

// Recovery Timeline — one chronological thread of check-ins and emergencies,
// newest first, grouped by day.
import { useEffect, useMemo, useState } from 'react'

import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Chip from '@mui/material/Chip'
import CircularProgress from '@mui/material/CircularProgress'
import Divider from '@mui/material/Divider'
import Typography from '@mui/material/Typography'

import RiskChip from '@components/anchor-one/RiskChip'
import SpeakButton from '@components/anchor-one/SpeakButton'
import EmptyState from '@components/EmptyState'
import { anchorOneService } from '@/services/anchorOneService'
import type { TimelineEvent } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

const dayLabel = (iso: string) => {
  const date = new Date(iso)
  const today = new Date()
  const yesterday = new Date(today)

  yesterday.setDate(today.getDate() - 1)

  if (date.toDateString() === today.toDateString()) return 'Today'
  if (date.toDateString() === yesterday.toDateString()) return 'Yesterday'

  return date.toLocaleDateString(undefined, { weekday: 'long', month: 'short', day: 'numeric' })
}

const timeLabel = (iso: string) => new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })

const TimelinePage = () => {
  const [events, setEvents] = useState<TimelineEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const load = async () => {
      try {
        setEvents(await anchorOneService.getTimeline())
      } catch (err) {
        setError(getApiErrorMessage(err, 'Could not load your timeline.'))
      } finally {
        setLoading(false)
      }
    }

    void load()
  }, [])

  // Group the flat, already-sorted list into days for readability.
  const groups = useMemo(() => {
    const byDay = new Map<string, TimelineEvent[]>()

    events.forEach(event => {
      const key = dayLabel(event.occurred_at)
      const bucket = byDay.get(key)

      if (bucket) {
        bucket.push(event)
      } else {
        byDay.set(key, [event])
      }
    })

    return Array.from(byDay.entries())
  }, [events])

  if (loading) {
    return (
      <Box display='flex' justifyContent='center' alignItems='center' height='50vh'>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box className='flex flex-col gap-6'>
      <Box>
        <Typography variant='h4' fontWeight={700}>
          Recovery Timeline
        </Typography>
        <Typography variant='body2' color='text.secondary'>
          Every check-in and every emergency, in order.
        </Typography>
      </Box>

      {error && <Alert severity='error'>{error}</Alert>}

      {events.length === 0 ? (
        <EmptyState
          icon='ri-time-line'
          title='Nothing here yet'
          message='Your recovery history builds itself as you check in.'
        />
      ) : (
        groups.map(([day, dayEvents]) => (
          <Box key={day} className='flex flex-col gap-3'>
            <Box className='flex items-center gap-3'>
              <Typography variant='subtitle1' fontWeight={700}>
                {day}
              </Typography>
              <Divider className='flex-grow' />
            </Box>

            {dayEvents.map(event => {
              const isEmergency = event.type === 'emergency'

              return (
                <Card
                  key={`${event.type}-${event.id}`}
                  variant='outlined'
                  sx={{
                    borderInlineStartWidth: 4,
                    borderInlineStartStyle: 'solid',
                    borderInlineStartColor: isEmergency ? 'error.main' : 'primary.main'
                  }}
                >
                  <CardContent className='flex flex-col gap-3'>
                    <Box className='flex flex-wrap items-center justify-between gap-2'>
                      <Box className='flex items-center gap-2'>
                        <i className={isEmergency ? 'ri-alarm-warning-fill' : 'ri-mic-fill'} />
                        <Typography variant='subtitle2' fontWeight={700}>
                          {isEmergency ? 'Emergency plan generated' : 'Check-in'}
                        </Typography>
                        <Typography variant='caption' color='text.secondary'>
                          {timeLabel(event.occurred_at)}
                        </Typography>
                      </Box>

                      <Box className='flex items-center gap-2'>
                        {event.source === 'text' && <Chip size='small' variant='tonal' label='typed' />}
                        {event.craving ? (
                          <Chip size='small' variant='tonal' color='info' label={`Craving ${event.craving}/10`} />
                        ) : null}
                        {event.emotion && <Chip size='small' variant='tonal' label={event.emotion} />}
                        <RiskChip risk={event.risk} />
                      </Box>
                    </Box>

                    {event.summary && (
                      <Typography variant='body2' color='text.secondary'>
                        {event.summary}
                      </Typography>
                    )}

                    {event.triggers && event.triggers.length > 0 && (
                      <Box className='flex flex-wrap gap-2'>
                        {event.triggers.map(trigger => (
                          <Chip key={trigger} size='small' variant='tonal' color='warning' label={trigger} />
                        ))}
                      </Box>
                    )}

                    {event.actions && event.actions.length > 0 && (
                      <Box>
                        <Box className='flex items-center justify-between gap-2'>
                          <Typography variant='caption' color='text.secondary'>
                            {isEmergency ? 'Immediate actions' : 'Recommended actions'}
                          </Typography>
                          <SpeakButton text={event.actions.join('. ')} iconOnly />
                        </Box>
                        <Box component='ul' sx={{ pl: 5, m: 0 }}>
                          {event.actions.map(action => (
                            <li key={action}>
                              <Typography variant='body2'>{action}</Typography>
                            </li>
                          ))}
                        </Box>
                      </Box>
                    )}

                    {event.generated_script && (
                      <Box>
                        <Typography variant='caption' color='text.secondary'>
                          Caregiver message
                        </Typography>
                        <Typography variant='body2' fontStyle='italic' sx={{ whiteSpace: 'pre-wrap' }}>
                          {event.generated_script}
                        </Typography>
                      </Box>
                    )}

                    {event.grounding_exercise && (
                      <Box>
                        <Typography variant='caption' color='text.secondary'>
                          Grounding exercise
                        </Typography>
                        <Typography variant='body2'>{event.grounding_exercise}</Typography>
                      </Box>
                    )}
                  </CardContent>
                </Card>
              )
            })}
          </Box>
        ))
      )}
    </Box>
  )
}

export default TimelinePage
