'use client'

// Emergency Mode. One tap produces the crisis package the PRD specifies:
// immediate actions, a personalised caregiver script (copy / speak / share), a
// grounding exercise and an encouraging message.
//
// Note: nothing here claims the caregiver has been contacted. The app cannot
// send an SMS on the user's behalf, so it hands them a ready-to-send message and
// says so — telling someone in crisis that help is already coming when it is not
// would be actively dangerous.
import { useState } from 'react'

import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import CircularProgress from '@mui/material/CircularProgress'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogTitle from '@mui/material/DialogTitle'
import Divider from '@mui/material/Divider'
import Paper from '@mui/material/Paper'
import Snackbar from '@mui/material/Snackbar'
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

const EmergencyModal = ({ open, onClose }: Props) => {
  const { dashboard, refreshDashboard } = useAnchorOne()

  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<EmergencyResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)

  const caregiverPhone = dashboard?.profile?.caregiver_phone?.trim() ?? ''

  const caregiverLabel =
    dashboard?.profile?.caregiver_name?.trim() || dashboard?.profile?.linked_caregiver_name?.trim() || 'your caregiver'

  const handleTrigger = async () => {
    setLoading(true)
    setError(null)

    try {
      setResult(await anchorOneService.triggerEmergency())
      await refreshDashboard()
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not build your emergency plan.'))
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    setResult(null)
    setError(null)
    onClose()
  }

  const handleCopy = async () => {
    if (!result) return

    try {
      await navigator.clipboard.writeText(result.plan.emergency_sms)
      setToast('Message copied — paste it to ' + caregiverLabel + '.')
    } catch {
      setToast('Could not copy automatically. Select the text and copy it manually.')
    }
  }

  const handleShare = async () => {
    if (!result) return

    const text = result.plan.emergency_sms

    // Native share sheet where available (mobile), else open the SMS composer
    // pre-filled, else fall back to copying.
    if (typeof navigator !== 'undefined' && navigator.share) {
      try {
        await navigator.share({ text })

        return
      } catch {
        // User dismissed the sheet — fall through to the SMS composer.
      }
    }

    if (caregiverPhone) {
      window.location.href = `sms:${encodeURIComponent(caregiverPhone)}?&body=${encodeURIComponent(text)}`

      return
    }

    await handleCopy()
  }

  return (
    <>
      <Dialog open={open} onClose={loading ? undefined : handleClose} maxWidth='sm' fullWidth>
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

              <Box>
                <Typography variant='h6' fontWeight={700} gutterBottom>
                  Message for {caregiverLabel}
                </Typography>
                <Alert severity='info' sx={{ mb: 2 }}>
                  We have written this for you — it is not sent yet. Send it to reach out.
                </Alert>
                <Paper variant='outlined' sx={{ p: 2, whiteSpace: 'pre-wrap', bgcolor: 'action.hover' }}>
                  <Typography variant='body1'>{result.plan.emergency_sms}</Typography>
                </Paper>
                <Box className='flex flex-wrap gap-2 mbs-3'>
                  <Button
                    variant='contained'
                    size='small'
                    onClick={handleCopy}
                    startIcon={<i className='ri-file-copy-line' />}
                  >
                    Copy
                  </Button>
                  <SpeakButton text={result.plan.emergency_sms} label='Speak' />
                  <Button
                    variant='outlined'
                    size='small'
                    onClick={handleShare}
                    startIcon={<i className='ri-share-forward-line' />}
                  >
                    {caregiverPhone ? 'Send as SMS' : 'Share'}
                  </Button>
                  {caregiverPhone && (
                    <Button
                      variant='outlined'
                      color='success'
                      size='small'
                      href={`tel:${caregiverPhone}`}
                      startIcon={<i className='ri-phone-line' />}
                    >
                      Call {caregiverPhone}
                    </Button>
                  )}
                </Box>
              </Box>

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
            </Box>
          )}
        </DialogContent>

        <DialogActions>
          <Button onClick={handleClose} disabled={loading}>
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
