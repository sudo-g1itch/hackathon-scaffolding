'use client'

// Enforces the navigation config on the current URL.
//
// Hiding a menu item is a courtesy; this is what makes it true. Wrapped around
// every dashboard page so a screen that is not offered to a role also cannot be
// reached by typing its address. The API enforces the same rules independently
// — this exists so the user meets a clear explanation instead of a raw 403.
import type { ReactNode } from 'react'

import { usePathname, useRouter } from 'next/navigation'

import Alert from '@mui/material/Alert'
import AlertTitle from '@mui/material/AlertTitle'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CircularProgress from '@mui/material/CircularProgress'

import { canAccessPath, landingPathFor } from '@/configs/navigation'
import { useAuth } from '@/contexts/AuthContext'

const RouteGuard = ({ children }: { children: ReactNode }) => {
  const { user, loading } = useAuth()
  const pathname = usePathname()
  const router = useRouter()

  if (loading) {
    return (
      <Box className='flex items-center justify-center min-bs-[50vh]'>
        <CircularProgress />
      </Box>
    )
  }

  // AuthGuard handles the signed-out case; this component only decides between
  // roles, so an absent user here means auth is still settling.
  if (!user) {
    return null
  }

  if (!canAccessPath(pathname, user.role)) {
    const home = landingPathFor(user.role)

    return (
      <Card variant='outlined' className='m-4'>
        <CardContent className='flex flex-col items-start gap-4'>
          <Alert severity='warning' variant='outlined'>
            <AlertTitle>This screen is not part of your account</AlertTitle>
            You are signed in as a <strong>{user.role}</strong>, and this page belongs to a different part of
            AnchorOne. Nothing has gone wrong — it simply is not yours to open.
          </Alert>
          <Button variant='contained' onClick={() => router.replace(home)} startIcon={<i className='ri-arrow-left-line' />}>
            Go to my home screen
          </Button>
        </CardContent>
      </Card>
    )
  }

  return <>{children}</>
}

export default RouteGuard
