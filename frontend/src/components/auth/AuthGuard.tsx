'use client'

import type { ReactNode } from 'react'
import { useEffect } from 'react'

import { usePathname, useRouter } from 'next/navigation'

import Box from '@mui/material/Box'
import CircularProgress from '@mui/material/CircularProgress'

import { useAuth } from '@/contexts/AuthContext'

type AuthGuardProps = {
  children: ReactNode
  fallback?: ReactNode
}

const AuthGuard = ({ children, fallback }: AuthGuardProps) => {
  const { isAuthenticated, loading } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    if (!loading && !isAuthenticated) {
      const returnUrl = encodeURIComponent(pathname)

      router.replace(`/login?returnUrl=${returnUrl}`)
    }
  }, [loading, isAuthenticated, router, pathname])

  if (loading) {
    return (
      fallback || (
        <Box className='flex items-center justify-center min-bs-screen'>
          <CircularProgress />
        </Box>
      )
    )
  }

  if (!isAuthenticated) {
    return null
  }

  return <>{children}</>
}

export default AuthGuard
