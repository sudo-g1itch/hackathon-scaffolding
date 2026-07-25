'use client'

// AI Recovery Coach. The conversation is stored server-side, so it is loaded on
// mount rather than starting empty after every refresh.
import { useEffect, useRef, useState } from 'react'

import Alert from '@mui/material/Alert'
import Avatar from '@mui/material/Avatar'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Chip from '@mui/material/Chip'
import CircularProgress from '@mui/material/CircularProgress'
import Paper from '@mui/material/Paper'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import SpeakButton from '@components/anchor-one/SpeakButton'
import { useAnchorOne } from '@/contexts/AnchorOneContext'
import { anchorOneService } from '@/services/anchorOneService'
import type { CoachMessage } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

const PROMPT_STARTERS = [
  'Help me get through the next 15 minutes.',
  'I relapsed yesterday.',
  'I am craving right now and I am alone.',
  'How do I say no at a party tonight?'
]

const CoachChat = () => {
  const { capabilities } = useAnchorOne()

  const [messages, setMessages] = useState<CoachMessage[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const endRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    const load = async () => {
      try {
        setMessages(await anchorOneService.getCoachHistory())
      } catch (err) {
        setError(getApiErrorMessage(err, 'Could not load your conversation.'))
      } finally {
        setLoading(false)
      }
    }

    void load()
  }, [])

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, sending])

  const send = async (text: string) => {
    const message = text.trim()

    if (!message || sending) return

    setInput('')
    setSending(true)
    setError(null)

    // Show the user's turn immediately; the server returns the canonical
    // history (including the coach's reply) which then replaces this.
    const pending: CoachMessage = {
      id: `pending-${Date.now()}`,
      created_at: new Date().toISOString(),
      role: 'user',
      message
    }

    setMessages(previous => [...previous, pending])

    try {
      setMessages(await anchorOneService.sendCoachMessage(message))
    } catch (err) {
      setError(getApiErrorMessage(err, 'The coach could not reply just now.'))
      setMessages(previous => previous.filter(item => item.id !== pending.id))
      setInput(message)
    } finally {
      setSending(false)
    }
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 160px)' }} className='gap-4'>
      <Box>
        <Typography variant='h4' fontWeight={700}>
          AI Recovery Coach
        </Typography>
        <Typography variant='body2' color='text.secondary'>
          Judgement-free support, any time. This is not a therapist and never gives medical advice.
        </Typography>
      </Box>

      {!capabilities.ai && <Alert severity='warning'>The AI coach is not configured on this server.</Alert>}
      {error && <Alert severity='error'>{error}</Alert>}

      <Paper
        variant='outlined'
        sx={{ flexGrow: 1, overflowY: 'auto', p: 4, display: 'flex', flexDirection: 'column', gap: 4 }}
      >
        {loading ? (
          <Box className='flex items-center justify-center h-full'>
            <CircularProgress />
          </Box>
        ) : messages.length === 0 ? (
          <Box className='flex flex-col items-center justify-center h-full gap-4 text-center'>
            <i className='ri-robot-line text-5xl' />
            <Typography variant='body1' color='text.secondary'>
              Tell your coach what is going on. Try one of these:
            </Typography>
            <Box className='flex flex-wrap justify-center gap-2'>
              {PROMPT_STARTERS.map(prompt => (
                <Chip
                  key={prompt}
                  label={prompt}
                  variant='tonal'
                  color='primary'
                  onClick={() => void send(prompt)}
                  disabled={!capabilities.ai}
                />
              ))}
            </Box>
          </Box>
        ) : (
          messages.map(message => {
            const isUser = message.role === 'user'

            return (
              <Box key={message.id} className='flex gap-3' flexDirection={isUser ? 'row-reverse' : 'row'}>
                <Avatar sx={{ bgcolor: isUser ? 'primary.main' : 'secondary.main' }}>
                  <i className={isUser ? 'ri-user-smile-line' : 'ri-robot-line'} />
                </Avatar>
                <Box sx={{ maxWidth: '75%' }}>
                  <Paper
                    elevation={0}
                    sx={{
                      p: 3,
                      bgcolor: isUser ? 'primary.main' : 'action.hover',
                      color: isUser ? 'primary.contrastText' : 'text.primary',
                      borderRadius: 2
                    }}
                  >
                    <Typography variant='body1' sx={{ whiteSpace: 'pre-wrap' }}>
                      {message.message}
                    </Typography>
                  </Paper>
                  {!isUser && (
                    <Box className='flex justify-start mbs-1'>
                      <SpeakButton text={message.message} iconOnly />
                    </Box>
                  )}
                </Box>
              </Box>
            )
          })
        )}

        {sending && (
          <Box className='flex items-center gap-3'>
            <Avatar sx={{ bgcolor: 'secondary.main' }}>
              <i className='ri-robot-line' />
            </Avatar>
            <CircularProgress size={20} />
            <Typography variant='body2' color='text.secondary'>
              Your coach is thinking…
            </Typography>
          </Box>
        )}

        <div ref={endRef} />
      </Paper>

      <Box className='flex gap-3'>
        <TextField
          fullWidth
          multiline
          maxRows={4}
          placeholder='Type what is going on…'
          value={input}
          onChange={event => setInput(event.target.value)}
          onKeyDown={event => {
            if (event.key === 'Enter' && !event.shiftKey) {
              event.preventDefault()
              void send(input)
            }
          }}
          disabled={sending || !capabilities.ai}
        />
        <Button
          variant='contained'
          onClick={() => void send(input)}
          disabled={sending || !input.trim() || !capabilities.ai}
          endIcon={<i className='ri-send-plane-fill' />}
        >
          Send
        </Button>
      </Box>
    </Box>
  )
}

export default CoachChat
