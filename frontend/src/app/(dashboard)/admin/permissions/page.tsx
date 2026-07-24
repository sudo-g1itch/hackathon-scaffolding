'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'

import Box from '@mui/material/Box'
import Typography from '@mui/material/Typography'
import Chip from '@mui/material/Chip'
import TextField from '@mui/material/TextField'
import InputAdornment from '@mui/material/InputAdornment'
import Paper from '@mui/material/Paper'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'

import AuthGuard from '@components/auth/AuthGuard'
import RoleGuard from '@components/auth/RoleGuard'
import { rbacService } from '@/services/rbacService'
import type { Permission } from '@/types/rbacTypes'

export default function PermissionsPage() {
  return (
    <AuthGuard>
      <RoleGuard roles={['admin']}>
        <PermissionsContent />
      </RoleGuard>
    </AuthGuard>
  )
}

function PermissionsContent() {
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [search, setSearch] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(true)

  const loadPermissions = useCallback(async () => {
    setLoading(true)

    try {
      const data = await rbacService.getPermissions()

      setPermissions(data)
    } catch {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadPermissions()
  }, [loadPermissions])

  const filteredPermissions = useMemo(() => {
    if (!search.trim()) return permissions
    const term = search.toLowerCase()

    
return permissions.filter(
      p =>
        p.name.toLowerCase().includes(term) ||
        p.slug.toLowerCase().includes(term) ||
        p.module.toLowerCase().includes(term) ||
        p.description.toLowerCase().includes(term)
    )
  }, [permissions, search])

  return (
    <Box className='flex flex-col gap-6 p-6'>
      {/* Title Header */}
      <Box className='flex items-center justify-between'>
        <Box>
          <Typography variant='h5' className='font-bold text-textPrimary'>
            System Permissions Reference
          </Typography>
          <Typography variant='body2' className='text-textSecondary'>
            Granular capabilities registered across system modules and APIs.
          </Typography>
        </Box>
      </Box>

      {/* Toolbar Search */}
      <Paper variant='outlined' className='p-4 flex items-center justify-between bg-backgroundPaper'>
        <TextField
          size='small'
          placeholder='Search permissions by name, slug, or module...'
          value={search}
          onChange={e => setSearch(e.target.value)}
          className='w-80'
          InputProps={{
            startAdornment: (
              <InputAdornment position='start'>
                <i className='ri-search-line text-textDisabled' />
              </InputAdornment>
            )
          }}
        />
        <Chip label={`Total ${filteredPermissions.length} Permissions`} color='primary' size='small' variant='tonal' />
      </Paper>

      {/* Table */}
      <Paper variant='outlined' className='overflow-hidden'>
        <Table>
          <TableHead className='bg-backgroundDefault'>
            <TableRow>
              <TableCell className='font-bold'>Permission Name</TableCell>
              <TableCell className='font-bold'>Slug Identifier</TableCell>
              <TableCell className='font-bold'>System Module</TableCell>
              <TableCell className='font-bold'>Description</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={4} align='center' className='py-8 text-textSecondary'>
                  Loading permissions...
                </TableCell>
              </TableRow>
            ) : filteredPermissions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} align='center' className='py-8 text-textSecondary'>
                  No permissions matching search criteria.
                </TableCell>
              </TableRow>
            ) : (
              filteredPermissions.map(p => (
                <TableRow key={p.id} hover>
                  <TableCell className='font-semibold text-sm text-textPrimary'>{p.name}</TableCell>
                  <TableCell>
                    <Chip label={p.slug} size='small' variant='tonal' color='secondary' className='font-mono text-xs' />
                  </TableCell>
                  <TableCell>
                    <Chip label={p.module} size='small' variant='tonal' color='info' className='font-medium text-xs' />
                  </TableCell>
                  <TableCell className='text-xs text-textSecondary'>{p.description}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Paper>
    </Box>
  )
}
