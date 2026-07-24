'use client'

import type { ReactNode } from 'react'

import Alert from '@mui/material/Alert'
import AlertTitle from '@mui/material/AlertTitle'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'

import type { Role } from '@/types/apiTypes'
import { useAuth } from '@/contexts/AuthContext'

type RoleGuardProps = {
  roles: Role | Role[]
  children: ReactNode
  fallback?: ReactNode
}

const RoleGuard = ({ roles, children, fallback }: RoleGuardProps) => {
  const { hasRole, loading } = useAuth()

  if (loading) {
    return null
  }

  if (!hasRole(roles)) {
    if (fallback) {
      return <>{fallback}</>
    }

    return (
      <Card variant='outlined' className='m-4'>
        <CardContent>
          <Alert severity='error' variant='outlined'>
            <AlertTitle>Access Denied</AlertTitle>
            You do not have the required permissions to access this feature.
          </Alert>
        </CardContent>
      </Card>
    )
  }

  return <>{children}</>
}

export default RoleGuard
