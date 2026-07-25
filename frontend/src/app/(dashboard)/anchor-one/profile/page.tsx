'use client'

// Recovery plan & caregiver setup.
//
// This screen is what makes the AI personal: the goal and substance are fed
// into every prompt, and the caregiver details are what the emergency script is
// addressed to. Without it the plan can only ever be generic.
import { useEffect, useState } from 'react'

import { useRouter } from 'next/navigation'

import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CircularProgress from '@mui/material/CircularProgress'
import Divider from '@mui/material/Divider'
import FormControlLabel from '@mui/material/FormControlLabel'
import Grid from '@mui/material/Grid'
import MenuItem from '@mui/material/MenuItem'
import Snackbar from '@mui/material/Snackbar'
import Switch from '@mui/material/Switch'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'


import { useAnchorOne } from '@/contexts/AnchorOneContext'
import { anchorOneService } from '@/services/anchorOneService'
import type { CaregiverOption, ProfileInput } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

const SUBSTANCES = ['Alcohol', 'Nicotine', 'Cannabis', 'Opioids', 'Stimulants', 'Gambling', 'Other']

const EMPTY_FORM: ProfileInput = {
  goal: '',
  substance: '',
  caregiver_name: '',
  caregiver_phone: '',
  emergency_contact: '',
  share_checkin_details: false
}

const ProfilePage = () => {
  const { refreshDashboard } = useAnchorOne()
  const router = useRouter()

  const [form, setForm] = useState<ProfileInput>(EMPTY_FORM)
  const [caregiverId, setCaregiverId] = useState('')
  const [options, setOptions] = useState<CaregiverOption[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)

  useEffect(() => {
    const load = async () => {
      try {
        const [profile, caregivers] = await Promise.all([
          anchorOneService.getProfile(),
          anchorOneService.getAvailableCaregivers()
        ])

        setForm({
          goal: profile.goal,
          substance: profile.substance,
          caregiver_name: profile.caregiver_name,
          caregiver_phone: profile.caregiver_phone,
          emergency_contact: profile.emergency_contact,
          share_checkin_details: profile.share_checkin_details
        })
        setCaregiverId(profile.caregiver_id ?? '')
        setOptions(caregivers)
      } catch (err) {
        setError(getApiErrorMessage(err, 'Could not load your recovery plan.'))
      } finally {
        setLoading(false)
      }
    }

    void load()
  }, [])

  const update = (field: keyof ProfileInput) => (event: { target: { value: string } }) =>
    setForm(previous => ({ ...previous, [field]: event.target.value }))

  const handleSave = async () => {
    setSaving(true)
    setError(null)

    try {
      await anchorOneService.updateProfile(form)
      await refreshDashboard()
      setToast('Recovery plan saved.')
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not save your recovery plan.'))
    } finally {
      setSaving(false)
    }
  }

  const handleCaregiverChange = async (value: string) => {
    const previous = caregiverId

    setCaregiverId(value)
    setError(null)

    try {
      const profile = await anchorOneService.setCaregiver(value === '' ? null : value)

      setCaregiverId(profile.caregiver_id ?? '')
      await refreshDashboard()
      setToast(value === '' ? 'Caregiver unlinked.' : 'Caregiver linked.')
    } catch (err) {
      setCaregiverId(previous)
      setError(getApiErrorMessage(err, 'Could not update your caregiver.'))
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
      <Box>
        <Typography variant='h4' fontWeight={700}>
          My Recovery Plan
        </Typography>
        <Typography variant='body2' color='text.secondary'>
          These details personalise your check-in analysis, coaching and emergency script.
        </Typography>
      </Box>

      {error && <Alert severity='error'>{error}</Alert>}

      <Grid container spacing={6}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent className='flex flex-col gap-5'>
              <Typography variant='h6'>Recovery goal</Typography>

              <TextField
                fullWidth
                label='My goal'
                placeholder='e.g. Stay sober for 90 days'
                value={form.goal}
                onChange={update('goal')}
                inputProps={{ maxLength: 255 }}
              />

              <TextField
                select
                fullWidth
                label='Substance or behaviour'
                value={form.substance}
                onChange={update('substance')}
              >
                <MenuItem value=''>
                  <em>Not specified</em>
                </MenuItem>
                {SUBSTANCES.map(substance => (
                  <MenuItem key={substance} value={substance}>
                    {substance}
                  </MenuItem>
                ))}
              </TextField>

              <Button
                variant='contained'
                onClick={handleSave}
                disabled={saving}
                startIcon={saving ? <CircularProgress size={16} color='inherit' /> : <i className='ri-save-line' />}
              >
                Save plan
              </Button>

              <Divider />

              {/* The headline goal above is one sentence for the AI to work
                  from. The measurable commitments live on their own screen,
                  because there is usually more than one of them. */}
              <Box>
                <Typography variant='subtitle2' gutterBottom>
                  Your goals
                </Typography>
                <Typography variant='body2' color='text.secondary' className='mbe-3'>
                  Track the specific things you are working towards — each with its own target and progress.
                </Typography>
                <Button
                  variant='outlined'
                  onClick={() => router.push('/anchor-one/goals')}
                  startIcon={<i className='ri-flag-line' />}
                >
                  Open my goals
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent className='flex flex-col gap-5'>
              <Typography variant='h6'>Support network</Typography>

              <TextField
                select
                fullWidth
                label='Linked caregiver account'
                value={caregiverId}
                onChange={event => void handleCaregiverChange(event.target.value)}
                helperText='A linked caregiver sees your risk level, can message you, and is who your SOS reaches.'
              >
                <MenuItem value=''>
                  <em>No caregiver</em>
                </MenuItem>
                {options.map(option => (
                  <MenuItem key={option.id} value={option.id}>
                    {option.name}
                  </MenuItem>
                ))}
              </TextField>

              <TextField
                fullWidth
                label='Who should we address messages to?'
                placeholder='e.g. Mom'
                value={form.caregiver_name}
                onChange={update('caregiver_name')}
                inputProps={{ maxLength: 150 }}
              />

              {/* No phone field: AnchorOne reaches a caregiver in-app, and the
                  app does not place calls or send SMS. Asking for a number it
                  will never dial would be a promise it cannot keep. */}

              <TextField
                fullWidth
                label='Backup emergency contact'
                placeholder='e.g. Sister — +1 555 987 6543'
                value={form.emergency_contact}
                onChange={update('emergency_contact')}
                inputProps={{ maxLength: 150 }}
              />

              <Divider />

              {/* Consent, stated in the user's own words rather than as a
                  setting. Off by default; nothing else in the app widens what
                  a caregiver can read. */}
              <Box>
                <FormControlLabel
                  control={
                    <Switch
                      checked={form.share_checkin_details}
                      onChange={event =>
                        setForm(previous => ({ ...previous, share_checkin_details: event.target.checked }))
                      }
                    />
                  }
                  label='Let my caregiver read my check-in summaries'
                />
                <Typography variant='caption' color='text.secondary' component='p'>
                  {form.share_checkin_details
                    ? 'Your caregiver can see the summary and triggers of each check-in. Your recordings and your AI coach chats stay private either way.'
                    : 'Your caregiver sees only your risk level, mood, craving and streak — never what you said. Turn this on to share the summaries too.'}
                </Typography>
              </Box>

              <Button
                variant='outlined'
                onClick={handleSave}
                disabled={saving}
                startIcon={<i className='ri-save-line' />}
              >
                Save contacts
              </Button>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

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

export default ProfilePage
