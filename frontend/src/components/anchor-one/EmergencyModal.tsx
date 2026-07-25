'use client'

// Emergency Mode. One tap produces the crisis package the PRD specifies —
// immediate actions, a grounding exercise, an encouraging message — and then a
// message the user can actually send to their caregiver.
//
// Honesty rules this screen:
//   - Before sending, nothing claims anyone has been contacted.
//   - Sending delivers into the caregiver's conversation, so "sent" is true.
//   - "They have seen it" appears only once the caregiver acknowledges.
//   - With no caregiver linked there is no send button, and the screen says why.
//
// SMS and phone calls are deliberately absent: the app cannot place them, and a
// button that hands off to the OS and hopes is not a feature we can stand behind.
import { useCallback, useEffect, useRef, useState } from 'react'

import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Chip from '@mui/material/Chip'
import CircularProgress from '@mui/material/CircularProgress'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogTitle from '@mui/material/DialogTitle'
import Divider from '@mui/material/Divider'
import FormControlLabel from '@mui/material/FormControlLabel'
import Paper from '@mui/material/Paper'
import Snackbar from '@mui/material/Snackbar'
import Switch from '@mui/material/Switch'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import SpeakButton from '@components/anchor-one/SpeakButton'
import { useAnchorOne } from '@/contexts/AnchorOneContext'
import { anchorOneService } from '@/services/anchorOneService'
import type { EmergencyResult } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

type Props = {
  open: boolean
  onClose: () => void
}

/** Browser geolocation, wrapped so a refusal is a value rather than a throw. */
const readLocation = (): Promise<GeolocationPosition | null> =>
  new Promise(resolve => {
    if (typeof navigator === 'undefined' || !navigator.geolocation) {
      resolve(null)

      return
    }

    navigator.geolocation.getCurrentPosition(
      position => resolve(position),
      () => resolve(null),
      { enableHighAccuracy: true, timeout: 10_000, maximumAge: 60_000 }
    )
  })

const pickMimeType = (): string | undefined => {
  if (typeof MediaRecorder === 'undefined') return undefined

  return ['audio/webm;codecs=opus', 'audio/webm', 'audio/mp4', 'audio/ogg;codecs=opus'].find(type =>
    MediaRecorder.isTypeSupported(type)
  )
}

const EmergencyModal = ({ open, onClose }: Props) => {
  const { dashboard, refreshDashboard, capabilities } = useAnchorOne()

  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<EmergencyResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)

  const [message, setMessage] = useState('')
  const [shareLocation, setShareLocation] = useState(true)
  const [sending, setSending] = useState(false)
  const [recording, setRecording] = useState(false)
  const [transcribing, setTranscribing] = useState(false)

  const recorderRef = useRef<MediaRecorder | null>(null)
  const streamRef = useRef<MediaStream | null>(null)
  const chunksRef = useRef<Blob[]>([])

  const caregiverLabel =
    dashboard?.profile?.caregiver_name?.trim() || result?.caregiver_name?.trim() || 'your caregiver'

  const log = result?.log ?? null
  const sent = Boolean(log?.shared_at)
  const acknowledged = Boolean(log?.acknowledged_at)

  const releaseStream = useCallback(() => {
    streamRef.current?.getTracks().forEach(track => track.stop())
    streamRef.current = null
    recorderRef.current = null
    chunksRef.current = []
  }, [])

  // Never leave the microphone open if the dialog is dismissed mid-recording.
  useEffect(() => () => releaseStream(), [releaseStream])

  const handleTrigger = async () => {
    setLoading(true)
    setError(null)

    try {
      const emergency = await anchorOneService.triggerEmergency()

      setResult(emergency)

      // Pre-fill with the AI's draft when there is one, else the first preset,
      // so there is always something ready to send without typing.
      setMessage(emergency.plan?.emergency_sms || emergency.presets[0]?.body || '')
      await refreshDashboard()
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not build your emergency plan.'))
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    releaseStream()
    setResult(null)
    setMessage('')
    setError(null)
    setRecording(false)
    onClose()
  }

  const transcribeNote = useCallback(
    async (audio: Blob) => {
      if (!log) return

      setTranscribing(true)
      setError(null)

      try {
        setResult(await anchorOneService.attachEmergencyNote(log.id, audio))
        setToast('Voice note attached.')
      } catch (err) {
        setError(getApiErrorMessage(err, 'Could not attach that voice note.'))
      } finally {
        setTranscribing(false)
      }
    },
    [log]
  )

  const startRecording = async () => {
    setError(null)

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const mimeType = pickMimeType()
      const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined)

      chunksRef.current = []

      recorder.ondataavailable = event => {
        if (event.data.size > 0) chunksRef.current.push(event.data)
      }

      recorder.onstop = () => {
        const audio = new Blob(chunksRef.current, { type: recorder.mimeType || 'audio/webm' })

        releaseStream()
        if (audio.size > 0) void transcribeNote(audio)
      }

      streamRef.current = stream
      recorderRef.current = recorder
      recorder.start()
      setRecording(true)
    } catch {
      setError('Microphone access was blocked. You can still send your message without a voice note.')
    }
  }

  const stopRecording = () => {
    setRecording(false)
    if (recorderRef.current?.state === 'recording') recorderRef.current.stop()
  }

  const handleSend = async () => {
    if (!log) return

    setSending(true)
    setError(null)

    try {
      let latitude: number | undefined
      let longitude: number | undefined

      if (shareLocation) {
        const position = await readLocation()

        if (position) {
          latitude = position.coords.latitude
          longitude = position.coords.longitude
        } else {
          setToast('Location unavailable — sending your message without it.')
        }
      }

      setResult(
        await anchorOneService.sendEmergencyAlert(log.id, {
          message,
          share_location: shareLocation && latitude !== undefined,
          latitude,
          longitude
        })
      )
      await refreshDashboard()
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not send that alert.'))
    } finally {
      setSending(false)
    }
  }

  return (
    <>
      <Dialog open={open} onClose={loading || sending ? undefined : handleClose} maxWidth='sm' fullWidth>
        <DialogTitle sx={{ color: 'error.main', fontWeight: 700 }} className='flex items-center gap-2'>
          <i className='ri-alarm-warning-fill text-2xl' />
          {result ? 'Your Emergency Plan' : 'Emergency SOS'}
        </DialogTitle>

        <DialogContent dividers>
          {error && (
            <Alert severity='error' sx={{ mb: 3 }}>
              {error}
            </Alert>
          )}

          {!result ? (
            <Box textAlign='center' className='plb-6'>
              <Typography variant='body1' mb={1}>
                We will build you a plan for the next few minutes: immediate steps, a grounding exercise, and a message
                you can send to {caregiverLabel}.
              </Typography>
              <Typography variant='body2' color='text.secondary' mb={4}>
                If you are in medical danger, call your local emergency number instead.
              </Typography>
              <Button
                variant='contained'
                color='error'
                size='large'
                onClick={handleTrigger}
                disabled={loading}
                sx={{ borderRadius: 8, px: 5, py: 2, fontSize: '1.1rem', fontWeight: 700 }}
                startIcon={loading ? undefined : <i className='ri-alarm-warning-fill' />}
              >
                {loading ? <CircularProgress size={26} color='inherit' /> : 'HELP ME'}
              </Button>
            </Box>
          ) : (
            <Box className='flex flex-col gap-5'>
              {result.plan && (
                <>
                  <Box>
                    <Box className='flex items-center justify-between gap-2 mbe-2'>
                      <Typography variant='h6' fontWeight={700}>
                        Do these now
                      </Typography>
                      <SpeakButton text={result.plan.immediate_actions.join('. ')} label='Read aloud' />
                    </Box>
                    <Box component='ol' sx={{ pl: 5, m: 0 }}>
                      {result.plan.immediate_actions.map(action => (
                        <li key={action}>
                          <Typography variant='body1' sx={{ mb: 0.5 }}>
                            {action}
                          </Typography>
                        </li>
                      ))}
                    </Box>
                  </Box>

                  <Divider />
                </>
              )}

              <Box>
                <Typography variant='h6' fontWeight={700} gutterBottom>
                  Message for {caregiverLabel}
                </Typography>

                {sent ? (
                  <Alert severity={acknowledged ? 'success' : 'info'} sx={{ mb: 2 }}>
                    {acknowledged
                      ? `${caregiverLabel} has seen your message. Help is on the way.`
                      : `Sent to ${caregiverLabel}. They will see it in your conversation — you will be told here the moment they acknowledge it.`}
                  </Alert>
                ) : !result.caregiver_linked ? (
                  <Alert severity='warning' sx={{ mb: 2 }}>
                    You have not linked a caregiver yet, so there is nobody to send this to. You can choose one on your
                    recovery plan — the message below is still yours to use however you like.
                  </Alert>
                ) : (
                  <Alert severity='info' sx={{ mb: 2 }}>
                    Nothing has been sent yet. Edit the words if you want to, then press send.
                  </Alert>
                )}

                {!sent && result.presets.length > 0 && (
                  <Box className='flex flex-wrap gap-2 mbe-3'>
                    {result.presets.map(preset => (
                      <Chip
                        key={preset.id}
                        label={preset.label}
                        variant={message === preset.body ? 'filled' : 'tonal'}
                        color='error'
                        size='small'
                        onClick={() => setMessage(preset.body)}
                      />
                    ))}
                  </Box>
                )}

                {sent ? (
                  <Paper variant='outlined' sx={{ p: 2, whiteSpace: 'pre-wrap', bgcolor: 'action.hover' }}>
                    <Typography variant='body1'>{log?.sent_message}</Typography>
                  </Paper>
                ) : (
                  <TextField
                    fullWidth
                    multiline
                    minRows={3}
                    value={message}
                    onChange={event => setMessage(event.target.value)}
                    inputProps={{ maxLength: 2000 }}
                    disabled={sending}
                  />
                )}

                {log?.audio_transcript && (
                  <Paper variant='outlined' sx={{ p: 2, mt: 2, borderColor: 'primary.main' }}>
                    <Typography variant='caption' color='text.secondary'>
                      <i className='ri-mic-line align-middle mie-1' />
                      Voice note {sent ? 'sent' : 'attached'} — {caregiverLabel} reads this text
                    </Typography>
                    <Typography variant='body2' sx={{ whiteSpace: 'pre-wrap', mt: 1 }}>
                      {log.audio_transcript}
                    </Typography>
                  </Paper>
                )}

                {!sent && (
                  <Box className='flex flex-col gap-3 mbs-3'>
                    <FormControlLabel
                      control={
                        <Switch
                          checked={shareLocation}
                          onChange={event => setShareLocation(event.target.checked)}
                          disabled={sending}
                        />
                      }
                      label='Share my current location'
                    />
                    <Typography variant='caption' color='text.secondary' sx={{ mt: -2 }}>
                      {shareLocation
                        ? 'A Google Maps link to where you are now is included, so they can come and find you. Your browser will ask permission.'
                        : 'No location is shared. Only your message is sent.'}
                    </Typography>

                    <Box className='flex flex-wrap items-center gap-2'>
                      {capabilities.voice &&
                        (recording ? (
                          <Button
                            variant='contained'
                            color='error'
                            size='small'
                            onClick={stopRecording}
                            startIcon={<i className='ri-stop-circle-line' />}
                          >
                            Stop recording
                          </Button>
                        ) : (
                          <Button
                            variant='outlined'
                            size='small'
                            onClick={() => void startRecording()}
                            disabled={sending || transcribing}
                            startIcon={
                              transcribing ? <CircularProgress size={14} /> : <i className='ri-mic-line' />
                            }
                          >
                            {transcribing
                              ? 'Transcribing…'
                              : log?.audio_transcript
                                ? 'Re-record voice note'
                                : 'Add a voice note'}
                          </Button>
                        ))}
                      <SpeakButton text={message} label='Speak' />
                    </Box>

                    {result.caregiver_linked && (
                      <Button
                        variant='contained'
                        color='error'
                        size='large'
                        onClick={() => void handleSend()}
                        disabled={sending || recording || !message.trim()}
                        startIcon={
                          sending ? <CircularProgress size={18} color='inherit' /> : <i className='ri-send-plane-fill' />
                        }
                        sx={{ fontWeight: 700 }}
                      >
                        {sending ? 'Sending…' : `Send to ${caregiverLabel}`}
                      </Button>
                    )}
                  </Box>
                )}
              </Box>

              {result.plan && (
                <>
                  <Divider />

                  <Box>
                    <Box className='flex items-center justify-between gap-2 mbe-1'>
                      <Typography variant='h6' fontWeight={700}>
                        Grounding exercise
                      </Typography>
                      <SpeakButton text={result.plan.grounding_exercise} label='Guide me' />
                    </Box>
                    <Typography variant='body1' sx={{ whiteSpace: 'pre-wrap' }}>
                      {result.plan.grounding_exercise}
                    </Typography>
                  </Box>

                  <Paper variant='outlined' sx={{ p: 2, borderColor: 'success.main' }}>
                    <Typography variant='body1' fontStyle='italic' color='text.secondary'>
                      {result.plan.encouraging_message}
                    </Typography>
                  </Paper>
                </>
              )}
            </Box>
          )}
        </DialogContent>

        <DialogActions>
          <Button onClick={handleClose} disabled={loading || sending}>
            Close
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={Boolean(toast)}
        autoHideDuration={4000}
        onClose={() => setToast(null)}
        message={toast ?? ''}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      />
    </>
  )
}

export default EmergencyModal
