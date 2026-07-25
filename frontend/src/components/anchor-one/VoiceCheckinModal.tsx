'use client'

// Voice check-in: record → Deepgram → Gemini → risk score, with every stage
// visible. The PRD's flow ends at "Risk Score Generated", so the result is
// shown here rather than silently closing the dialog.
import { useState } from 'react'

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
import LinearProgress from '@mui/material/LinearProgress'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import { keyframes } from '@mui/material/styles'

import RiskChip from '@components/anchor-one/RiskChip'
import SpeakButton from '@components/anchor-one/SpeakButton'
import { useAnchorOne } from '@/contexts/AnchorOneContext'

type Props = {
  open: boolean
  onClose: () => void
}

const pulse = keyframes`
  0%   { transform: scale(1);    opacity: 1; }
  70%  { transform: scale(1.35); opacity: 0; }
  100% { transform: scale(1.35); opacity: 0; }
`

const formatTime = (seconds: number) => {
  const m = Math.floor(seconds / 60)
    .toString()
    .padStart(2, '0')

  const s = (seconds % 60).toString().padStart(2, '0')

  return `${m}:${s}`
}

const VoiceCheckinModal = ({ open, onClose }: Props) => {
  const {
    phase,
    recordingSeconds,
    result,
    checkinError,
    startRecording,
    stopRecording,
    cancelRecording,
    submitText,
    resetCheckin,
    capabilities
  } = useAnchorOne()

  const [typedMode, setTypedMode] = useState(false)
  const [transcript, setTranscript] = useState('')

  const recording = phase === 'recording'
  const processing = phase === 'processing'

  const handleClose = () => {
    if (recording) cancelRecording()
    resetCheckin()
    setTypedMode(false)
    setTranscript('')
    onClose()
  }

  const handleSubmitText = async () => {
    if (!transcript.trim()) return
    await submitText(transcript.trim())
    setTranscript('')
  }

  const handleRetry = () => {
    resetCheckin()
    setTranscript('')
  }

  return (
    <Dialog open={open} onClose={recording || processing ? undefined : handleClose} maxWidth='sm' fullWidth>
      <DialogTitle className='flex items-center gap-2'>
        <i className='ri-mic-fill text-xl' />
        Voice Check-In
      </DialogTitle>

      <DialogContent dividers>
        {/* --- result --- */}
        {phase === 'done' && result ? (
          <Box className='flex flex-col gap-4'>
            <Box className='flex items-center justify-between gap-2'>
              <RiskChip risk={result.risk} size='medium' />
              <Chip size='small' variant='tonal' color='info' label={`Craving ${result.craving}/10`} />
              <Chip size='small' variant='tonal' label={result.emotion} />
            </Box>

            <Box>
              <Typography variant='caption' color='text.secondary'>
                What we heard
              </Typography>
              <Typography variant='body2' fontStyle='italic'>
                &ldquo;{result.transcript}&rdquo;
              </Typography>
            </Box>

            <Divider />

            <Box>
              <Typography variant='subtitle2' gutterBottom>
                Summary
              </Typography>
              <Typography variant='body2' color='text.secondary'>
                {result.summary}
              </Typography>
            </Box>

            {result.triggers && result.triggers.length > 0 && (
              <Box>
                <Typography variant='subtitle2' gutterBottom>
                  Triggers detected
                </Typography>
                <Box className='flex flex-wrap gap-2'>
                  {result.triggers.map(trigger => (
                    <Chip key={trigger} size='small' variant='tonal' color='warning' label={trigger} />
                  ))}
                </Box>
              </Box>
            )}

            {result.recommended_actions && result.recommended_actions.length > 0 && (
              <Box>
                <Box className='flex items-center justify-between gap-2'>
                  <Typography variant='subtitle2'>Do this next</Typography>
                  <SpeakButton text={result.recommended_actions.join('. ')} label='Read aloud' />
                </Box>
                <Box component='ol' sx={{ pl: 5, m: 0, mt: 1 }}>
                  {result.recommended_actions.map(action => (
                    <li key={action}>
                      <Typography variant='body2'>{action}</Typography>
                    </li>
                  ))}
                </Box>
              </Box>
            )}

            {result.risk === 'HIGH' && (
              <Alert severity='error'>
                This check-in scored HIGH risk. Consider opening Emergency Mode for a full crisis plan.
              </Alert>
            )}
          </Box>
        ) : processing ? (

          /* --- processing --- */
          <Box className='flex flex-col items-center gap-3 plb-8'>
            <CircularProgress />
            <Typography variant='h6'>Analysing your check-in…</Typography>
            <Typography variant='body2' color='text.secondary'>
              Transcribing your voice, then scoring your relapse risk.
            </Typography>
            <LinearProgress className='is-full' />
          </Box>
        ) : recording ? (

          /* --- recording --- */
          <Box className='flex flex-col items-center gap-3 plb-6'>
            <Box sx={{ position: 'relative', display: 'grid', placeItems: 'center', mb: 2 }}>
              <Box
                sx={{
                  position: 'absolute',
                  inlineSize: 88,
                  blockSize: 88,
                  borderRadius: '50%',
                  bgcolor: 'error.main',
                  animation: `${pulse} 1.5s ease-out infinite`
                }}
              />
              <Box
                sx={{
                  inlineSize: 72,
                  blockSize: 72,
                  borderRadius: '50%',
                  bgcolor: 'error.main',
                  color: 'common.white',
                  display: 'grid',
                  placeItems: 'center',
                  fontSize: 32
                }}
              >
                <i className='ri-mic-fill' />
              </Box>
            </Box>

            <Typography variant='h4' sx={{ fontVariantNumeric: 'tabular-nums' }}>
              {formatTime(recordingSeconds)}
            </Typography>
            <Typography variant='body2' color='text.secondary' textAlign='center'>
              Speak naturally — how you feel, and whether you are having cravings.
            </Typography>

            <Button
              variant='contained'
              color='error'
              size='large'
              onClick={stopRecording}
              startIcon={<i className='ri-stop-circle-fill' />}
              sx={{ borderRadius: 8, px: 4 }}
            >
              Stop &amp; Analyse
            </Button>
          </Box>
        ) : (

          /* --- idle / error --- */
          <Box className='flex flex-col items-center gap-4 plb-4'>
            {checkinError && (
              <Alert severity='error' className='is-full'>
                {checkinError}
              </Alert>
            )}

            {!capabilities.voice && (
              <Alert severity='warning' className='is-full'>
                Voice transcription is not configured on this server. You can still type your check-in below.
              </Alert>
            )}

            {typedMode ? (
              <>
                <TextField
                  fullWidth
                  multiline
                  minRows={4}
                  autoFocus
                  label='How are you feeling right now?'
                  placeholder='I have been stressed all day and I do not think I can resist tonight…'
                  value={transcript}
                  onChange={event => setTranscript(event.target.value)}
                />
                <Box className='flex gap-2'>
                  <Button
                    variant='contained'
                    onClick={handleSubmitText}
                    disabled={transcript.trim().length < 2}
                    startIcon={<i className='ri-sparkling-line' />}
                  >
                    Analyse Check-In
                  </Button>
                  <Button color='secondary' onClick={() => setTypedMode(false)}>
                    Back
                  </Button>
                </Box>
              </>
            ) : (
              <>
                <Typography variant='body1' textAlign='center'>
                  Tap the microphone and tell us how you are doing. No typing needed.
                </Typography>
                <Button
                  variant='contained'
                  size='large'
                  onClick={startRecording}
                  disabled={!capabilities.voice}
                  startIcon={<i className='ri-mic-fill' />}
                  sx={{ borderRadius: 8, px: 4, py: 2 }}
                >
                  Start Recording
                </Button>
                <Button color='secondary' size='small' onClick={() => setTypedMode(true)}>
                  I would rather type it
                </Button>
              </>
            )}
          </Box>
        )}
      </DialogContent>

      <DialogActions>
        {phase === 'done' && (
          <Button color='secondary' onClick={handleRetry}>
            New check-in
          </Button>
        )}
        {recording ? (
          <Button color='secondary' onClick={cancelRecording}>
            Discard
          </Button>
        ) : (
          <Button onClick={handleClose} disabled={processing}>
            {phase === 'done' ? 'Done' : 'Close'}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  )
}

export default VoiceCheckinModal
