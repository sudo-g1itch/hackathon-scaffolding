'use client'

// The conversation between a person in recovery and their caregiver.
//
// One component serves both ends: the thread is identified by the patient's id,
// and whose bubble sits on which side is decided by comparing the sender to the
// signed-in user — not by role, so the same code renders correctly for either
// party.
import { useCallback, useEffect, useRef, useState } from 'react'

import Alert from '@mui/material/Alert'
import Avatar from '@mui/material/Avatar'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import CircularProgress from '@mui/material/CircularProgress'
import Paper from '@mui/material/Paper'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import EmptyState from '@components/EmptyState'
import { useAuth } from '@/contexts/AuthContext'
import { anchorOneService } from '@/services/anchorOneService'
import type { SupportThread } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

// New messages arrive by poll. A conversation is not a chat room, so this is
// deliberately gentle.
const POLL_INTERVAL_MS = 15_000

type SupportChatProps = {

  /** Whose thread this is — the user's own id, or the patient's. */
  patientId: string

  /** Rendered when no caregiver is linked yet. */
  unlinkedMessage?: string
  unlinkedAction?: React.ReactNode
  height?: number | string

  /** Called after a send, so a parent badge can refresh. */
  onSent?: () => void
}

const SupportChat = ({
  patientId,
  unlinkedMessage = 'Link a caregiver on your recovery plan and you can talk to them here.',
  unlinkedAction,
  height = 'calc(100vh - 320px)',
  onSent
}: SupportChatProps) => {
  const { user } = useAuth()

  const [thread, setThread] = useState<SupportThread | null>(null)
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const endRef = useRef<HTMLDivElement | null>(null)

  const load = useCallback(
    async (initial: boolean) => {
      try {
        setThread(await anchorOneService.getSupportThread(patientId))
        setError(null)
      } catch (err) {
        setError(getApiErrorMessage(err, 'Could not load this conversation.'))
      } finally {
        if (initial) setLoading(false)
      }
    },
    [patientId]
  )

  useEffect(() => {
    setLoading(true)
    void load(true)

    const interval = setInterval(() => void load(false), POLL_INTERVAL_MS)

    return () => clearInterval(interval)
  }, [load])

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [thread?.messages.length])

  const send = async () => {
    const body = input.trim()

    if (!body || sending) return

    setInput('')
    setSending(true)
    setError(null)

    try {
      setThread(await anchorOneService.sendSupportMessage(patientId, body))
      onSent?.()
    } catch (err) {
      setError(getApiErrorMessage(err, 'That message could not be sent.'))
      setInput(body)
    } finally {
      setSending(false)
    }
  }

  if (loading) {
    return (
      <Box className='flex items-center justify-center' sx={{ height }}>
        <CircularProgress />
      </Box>
    )
  }

  if (thread && !thread.linked) {
    return (
      <EmptyState icon='ri-user-add-line' title='No caregiver linked' message={unlinkedMessage} action={unlinkedAction} />
    )
  }

  return (
    <Box className='flex flex-col gap-4' sx={{ height }}>
      {error && <Alert severity='error'>{error}</Alert>}

      <Paper
        variant='outlined'
        sx={{ flexGrow: 1, overflowY: 'auto', p: 4, display: 'flex', flexDirection: 'column', gap: 4 }}
      >
        {thread && thread.messages.length === 0 ? (
          <Box className='flex items-center justify-center h-full'>
            <EmptyState
              icon='ri-chat-heart-line'
              title='No messages yet'
              message='Say hello. A short check-in from either side is enough to start.'
              size='sm'
            />
          </Box>
        ) : (
          thread?.messages.map(message => {
            const isMine = message.sender_id === user?.id

            return (
              <Box key={message.id} className='flex gap-3' flexDirection={isMine ? 'row-reverse' : 'row'}>
                <Avatar sx={{ bgcolor: isMine ? 'primary.main' : 'warning.main', width: 36, height: 36 }}>
                  <i className={message.sender_role === 'caregiver' ? 'ri-heart-pulse-line' : 'ri-user-smile-line'} />
                </Avatar>
                <Box sx={{ maxWidth: '75%' }}>
                  <Paper
                    elevation={0}
                    sx={{
                      p: 3,
                      bgcolor: isMine ? 'primary.main' : 'action.hover',
                      color: isMine ? 'primary.contrastText' : 'text.primary',
                      borderRadius: 2
                    }}
                  >
                    <Typography variant='body1' sx={{ whiteSpace: 'pre-wrap' }}>
                      {message.body}
                    </Typography>
                  </Paper>
                  <Typography
                    variant='caption'
                    color='text.secondary'
                    className={isMine ? 'flex justify-end mbs-1' : 'flex mbs-1'}
                  >
                    {new Date(message.created_at).toLocaleString()}
                    {isMine && message.read_at && ' · read'}
                  </Typography>
                </Box>
              </Box>
            )
          })
        )}

        <div ref={endRef} />
      </Paper>

      <Box className='flex gap-3'>
        <TextField
          fullWidth
          multiline
          maxRows={4}
          placeholder='Write a message…'
          value={input}
          onChange={event => setInput(event.target.value)}
          onKeyDown={event => {
            if (event.key === 'Enter' && !event.shiftKey) {
              event.preventDefault()
              void send()
            }
          }}
          disabled={sending}
          inputProps={{ maxLength: 2000 }}
        />
        <Button
          variant='contained'
          onClick={() => void send()}
          disabled={sending || !input.trim()}
          endIcon={<i className='ri-send-plane-fill' />}
        >
          Send
        </Button>
      </Box>
    </Box>
  )
}

export default SupportChat
