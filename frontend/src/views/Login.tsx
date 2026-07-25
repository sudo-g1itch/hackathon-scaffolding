'use client'

import type { FormEvent } from 'react'
import { useState } from 'react'

import { useRouter, useSearchParams } from 'next/navigation'

import Alert from '@mui/material/Alert'
import Button from '@mui/material/Button'
import Checkbox from '@mui/material/Checkbox'
import CircularProgress from '@mui/material/CircularProgress'
import FormControlLabel from '@mui/material/FormControlLabel'
import IconButton from '@mui/material/IconButton'
import InputAdornment from '@mui/material/InputAdornment'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import Radio from '@mui/material/Radio'
import RadioGroup from '@mui/material/RadioGroup'
import classnames from 'classnames'

import axios from '@/libs/axios'

import Link from '@components/Link'
import Logo from '@components/layout/shared/Logo'
import themeConfig from '@configs/themeConfig'
import { landingPathFor } from '@/configs/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { useImageVariant } from '@core/hooks/useImageVariant'
import { useSettings } from '@core/hooks/useSettings'
import type { Mode } from '@core/types'
import { getApiErrorMessage } from '@/utils/handleApiError'

const LoginV2 = ({ mode }: { mode: Mode }) => {
  const [isLoginView, setIsLoginView] = useState(true)
  const [email, setEmail] = useState('admin@hackathon.local')
  const [password, setPassword] = useState('Admin123!')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [role, setRole] = useState('user')

  const [isPasswordShown, setIsPasswordShown] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const router = useRouter()
  const searchParams = useSearchParams()
  const { login } = useAuth()
  const { settings } = useSettings()

  const darkIllustration = '/images/illustrations/auth/v2-login-dark.png'
  const lightIllustration = '/images/illustrations/auth/v2-login-light.png'

  const characterIllustration = useImageVariant(
    mode,
    lightIllustration,
    darkIllustration
  )

  const handleClickShowPassword = () => setIsPasswordShown(show => !show)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setErrorMessage(null)
    setSuccessMessage(null)
    setIsSubmitting(true)

    if (isLoginView) {
      try {
        const signedIn = await login({ email, password })
        const returnUrl = searchParams.get('returnUrl')

        // Land on the first screen this role can actually use. Sending a
        // caregiver to the personal dashboard would only bounce them straight
        // into the route guard.
        router.push(returnUrl ? decodeURIComponent(returnUrl) : landingPathFor(signedIn.role))
      } catch (err: unknown) {
        setErrorMessage(getApiErrorMessage(err, 'Login failed. Please check your credentials.'))
      } finally {
        setIsSubmitting(false)
      }
    } else {
      try {
        await axios.post('/auth/register', {
          email,
          password,
          first_name: firstName,
          last_name: lastName,
          role
        })
        setSuccessMessage('Registration successful! Logging you in...')

        const signedIn = await login({ email, password })

        router.push(landingPathFor(signedIn.role))
      } catch (err: unknown) {
        setErrorMessage(getApiErrorMessage(err, 'Registration failed.'))
      } finally {
        setIsSubmitting(false)
      }
    }
  }

  return (
    <div className='flex bs-full justify-center'>
      <div
        className={classnames(
          'flex flex-col bs-full items-center justify-center flex-1 min-bs-[100dvh] relative p-6 max-md:hidden',
          'bg-[var(--mui-palette-primary-lightOpacity)]',
          {
            'border-ie': settings.skin === 'bordered'
          }
        )}
      >
        <div className='flex flex-col items-center max-is-[600px] text-center'>
          <img
            src={characterIllustration}
            alt='care-illustration'
            className='max-bs-[500px] max-is-full bs-auto mbe-8'
          />
          <Typography variant='h3' className='text-primary mbe-3 font-semibold'>
            You are not alone.
          </Typography>
          <Typography variant='h6' color='text.secondary'>
            AnchorOne provides a safe, supportive, and professional environment for your recovery journey. We are here with you every step of the way.
          </Typography>
        </div>
      </div>
      <div className='flex justify-center items-center bs-full bg-backgroundPaper !min-is-full p-6 md:!min-is-[unset] md:p-12 md:is-[480px]'>
        <Link className='absolute block-start-5 sm:block-start-[38px] inline-start-6 sm:inline-start-[38px]'>
          <Logo />
        </Link>
        <div className='flex flex-col gap-5 is-full sm:is-auto md:is-full sm:max-is-[400px] md:max-is-[unset] mbs-11 sm:mbs-14 md:mbs-0'>
          <div>
            <Typography variant='h4'>
              {isLoginView ? `Welcome to ${themeConfig.templateName}! 👋🏻` : 'Create an Account'}
            </Typography>
            <Typography className='mbs-1'>
              {isLoginView ? 'Please sign-in to your account to continue' : 'Join the platform today'}
            </Typography>
          </div>
          {errorMessage && (
            <Alert severity='error' variant='outlined' className='mbs-2'>
              {errorMessage}
            </Alert>
          )}
          {successMessage && (
            <Alert severity='success' variant='outlined' className='mbs-2'>
              {successMessage}
            </Alert>
          )}
          <form noValidate autoComplete='off' onSubmit={handleSubmit} className='flex flex-col gap-5'>
            {!isLoginView && (
              <div className='flex gap-4'>
                <TextField
                  fullWidth
                  label='First Name'
                  value={firstName}
                  onChange={e => setFirstName(e.target.value)}
                  disabled={isSubmitting}
                />
                <TextField
                  fullWidth
                  label='Last Name'
                  value={lastName}
                  onChange={e => setLastName(e.target.value)}
                  disabled={isSubmitting}
                />
              </div>
            )}
            {!isLoginView && (
              <div>
                <Typography variant="body2" color="text.secondary">I am joining as a:</Typography>
                <RadioGroup row value={role} onChange={(e) => setRole(e.target.value)}>
                  <FormControlLabel value="user" control={<Radio />} label="Patient / User" />
                  <FormControlLabel value="caregiver" control={<Radio />} label="Caregiver" />
                </RadioGroup>
              </div>
            )}
            
            <TextField
              autoFocus
              fullWidth
              label='Email'
              value={email}
              onChange={e => setEmail(e.target.value)}
              disabled={isSubmitting}
            />
            <TextField
              fullWidth
              label='Password'
              type={isPasswordShown ? 'text' : 'password'}
              value={password}
              onChange={e => setPassword(e.target.value)}
              disabled={isSubmitting}
              slotProps={{
                input: {
                  endAdornment: (
                    <InputAdornment position='end'>
                      <IconButton
                        size='small'
                        edge='end'
                        onClick={handleClickShowPassword}
                        onMouseDown={e => e.preventDefault()}
                      >
                        <i className={isPasswordShown ? 'ri-eye-off-line' : 'ri-eye-line'} />
                      </IconButton>
                    </InputAdornment>
                  )
                }
              }}
            />
            
            {isLoginView && (
              <div className='flex items-center gap-x-3 gap-y-1 mbe-2'>
                <FormControlLabel control={<Checkbox defaultChecked />} label='Remember me' />
              </div>
            )}
            
            <Button fullWidth variant='contained' type='submit' disabled={isSubmitting}>
              {isSubmitting ? <CircularProgress size={24} color='inherit' /> : isLoginView ? 'Log In' : 'Register'}
            </Button>
            
            <div className='flex justify-center items-center flex-wrap gap-2'>
              <Typography>{isLoginView ? 'New on our platform?' : 'Already have an account?'}</Typography>
              <Typography 
                color='primary.main' 
                className='cursor-pointer font-medium'
                onClick={() => {
                  setIsLoginView(!isLoginView)
                  setErrorMessage(null)
                  setSuccessMessage(null)
                }}
              >
                {isLoginView ? 'Create an account' : 'Sign in instead'}
              </Typography>
            </div>
            
          </form>
        </div>
      </div>
    </div>
  )
}

export default LoginV2
