'use client'

import { useState, useEffect, useCallback } from 'react'

import Grid from '@mui/material/Grid'
import Button from '@mui/material/Button'
import Typography from '@mui/material/Typography'
import Avatar from '@mui/material/Avatar'
import IconButton from '@mui/material/IconButton'
import MenuItem from '@mui/material/MenuItem'
import Select from '@mui/material/Select'
import FormControl from '@mui/material/FormControl'
import InputLabel from '@mui/material/InputLabel'
import Dialog from '@mui/material/Dialog'
import DialogTitle from '@mui/material/DialogTitle'
import DialogContent from '@mui/material/DialogContent'
import DialogActions from '@mui/material/DialogActions'
import TextField from '@mui/material/TextField'
import Switch from '@mui/material/Switch'
import FormControlLabel from '@mui/material/FormControlLabel'
import Chip from '@mui/material/Chip'
import Box from '@mui/material/Box'
import Tooltip from '@mui/material/Tooltip'
import type { ColumnDef } from '@tanstack/react-table'

import AuthGuard from '@components/auth/AuthGuard'
import RoleGuard from '@components/auth/RoleGuard'
import StatCard from '@components/StatCard'
import DataTable from '@components/DataTable'
import DataTableFilters from '@components/DataTableFilters'
import StatusChip from '@components/StatusChip'
import ConfirmDialog from '@components/ConfirmDialog'
import { useServerTable } from '@/hooks/useServerTable'
import { rbacService } from '@/services/rbacService'
import type { User, UserFormValues } from '@/types/rbacTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

export default function UsersPage() {
  return (
    <AuthGuard>
      <RoleGuard roles={['admin']}>
        <UserManagementContent />
      </RoleGuard>
    </AuthGuard>
  )
}

function UserManagementContent() {
  const [roleFilter, setRoleFilter] = useState<string>('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [openModal, setOpenModal] = useState<boolean>(false)
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [deleteConfirmUser, setDeleteConfirmUser] = useState<User | null>(null)
  const [isDeleting, setIsDeleting] = useState<boolean>(false)

  // Data & Table State
  const [users, setUsers] = useState<User[]>([])
  const [total, setTotal] = useState<number>(0)
  const [loading, setLoading] = useState<boolean>(true)
  const tableState = useServerTable()

  // Form State
  const [formData, setFormData] = useState<UserFormValues>({
    email: '',
    first_name: '',
    last_name: '',
    role: 'user',
    is_active: true,
    password: ''
  })

  const [formError, setFormError] = useState<string>('')
  const [isSaving, setIsSaving] = useState<boolean>(false)

  // Stats State
  const [stats, setStats] = useState({
    total: 0,
    active: 0,
    admins: 0,
    inactive: 0
  })

  // Load Users Data
  const loadUsersData = useCallback(async () => {
    setLoading(true)

    try {
      const queryParams: Record<string, unknown> = {
        page: tableState.page,
        page_size: tableState.pageSize,
        ...(tableState.search ? { q: tableState.search } : {}),
        ...(tableState.sortBy ? { sort_by: tableState.sortBy, sort_order: tableState.sortOrder } : {}),
        ...(roleFilter ? { role: roleFilter } : {}),
        ...(statusFilter ? { is_active: statusFilter === 'active' } : {})
      }

      const res = await rbacService.getUsers(queryParams)

      setUsers(res.data)
      setTotal(res.meta?.pagination?.total ?? res.data.length)
    } catch {
      // Handled by axios
    } finally {
      setLoading(false)
    }
  }, [tableState.page, tableState.pageSize, tableState.search, tableState.sortBy, tableState.sortOrder, roleFilter, statusFilter])

  // Load KPI Stats
  const loadStats = useCallback(async () => {
    try {
      const res = await rbacService.getUsers({ page_size: 500 })
      const allUsers = res.data

      setStats({
        total: res.meta?.pagination?.total ?? allUsers.length,
        active: allUsers.filter(u => u.is_active).length,
        admins: allUsers.filter(u => u.role === 'admin').length,
        inactive: allUsers.filter(u => !u.is_active).length
      })
    } catch {
      // Fallback
    }
  }, [])

  useEffect(() => {
    loadUsersData()
    loadStats()
  }, [loadUsersData, loadStats])

  const handleOpenAdd = () => {
    setSelectedUser(null)
    setFormData({
      email: '',
      first_name: '',
      last_name: '',
      role: 'user',
      is_active: true,
      password: ''
    })
    setFormError('')
    setOpenModal(true)
  }

  const handleOpenEdit = (user: User) => {
    setSelectedUser(user)
    setFormData({
      email: user.email,
      first_name: user.first_name,
      last_name: user.last_name,
      role: user.role,
      is_active: user.is_active,
      password: ''
    })
    setFormError('')
    setOpenModal(true)
  }

  const handleSaveUser = async () => {
    setFormError('')

    if (!formData.first_name || !formData.last_name || !formData.email) {
      setFormError('Please fill in all required fields.')
      
return
    }

    if (!selectedUser && (!formData.password || formData.password.length < 6)) {
      setFormError('Password must be at least 6 characters long.')
      
return
    }

    setIsSaving(true)

    try {
      if (selectedUser) {
        await rbacService.updateUser(selectedUser.id, {
          first_name: formData.first_name,
          last_name: formData.last_name,
          role: formData.role,
          is_active: formData.is_active,
          ...(formData.password ? { password: formData.password } : {})
        })
      } else {
        await rbacService.createUser(formData)
      }

      setOpenModal(false)
      loadUsersData()
      loadStats()
    } catch (err: unknown) {
      setFormError(getApiErrorMessage(err))
    } finally {
      setIsSaving(false)
    }
  }

  const handleDeleteUser = async () => {
    if (!deleteConfirmUser) return
    setIsDeleting(true)

    try {
      await rbacService.deleteUser(deleteConfirmUser.id)
      setDeleteConfirmUser(null)
      loadUsersData()
      loadStats()
    } catch (err: unknown) {
      alert(getApiErrorMessage(err))
    } finally {
      setIsDeleting(false)
    }
  }

  const columns: ColumnDef<User>[] = [
    {
      id: 'first_name',
      header: 'User',
      cell: ({ row }) => {
        const u = row.original
        const initials = `${u.first_name[0] ?? ''}${u.last_name[0] ?? ''}`.toUpperCase()

        
return (
          <Box className='flex items-center gap-3'>
            <Avatar src={u.avatar_url} className='bg-primaryMain text-white font-medium text-sm'>
              {initials}
            </Avatar>

            <Box className='flex flex-col'>
              <Typography className='font-medium text-textPrimary hover:text-primaryMain cursor-pointer'>
                {u.first_name} {u.last_name}
              </Typography>
              <Typography variant='body2' className='text-textSecondary text-xs'>
                {u.email}
              </Typography>
            </Box>
          </Box>
        )
      }
    },
    {
      id: 'role',
      header: 'Role',
      cell: ({ row }) => {
        const role = row.original.role
        const roleColor = role === 'admin' ? 'error' : role === 'manager' ? 'warning' : 'info'

        
return (
          <Chip
            size='small'
            variant='tonal'
            color={roleColor}
            label={role.toUpperCase()}
            className='font-semibold text-[11px] capitalize'
          />
        )
      }
    },
    {
      id: 'is_active',
      header: 'Status',
      cell: ({ row }) => {
        const isActive = row.original.is_active

        
return <StatusChip status={isActive ? 'active' : 'inactive'} />
      }
    },
    {
      id: 'created_at',
      header: 'Joined',
      cell: ({ row }) => {
        const dateStr = row.original.created_at

        
return (
          <Typography variant='body2' className='text-textSecondary text-xs'>
            {dateStr ? new Date(dateStr).toLocaleDateString() : 'N/A'}
          </Typography>
        )
      }
    },
    {
      id: 'actions',
      header: 'Actions',
      cell: ({ row }) => {
        const user = row.original

        
return (
          <Box className='flex items-center gap-1'>
            <Tooltip title='Edit User'>
              <IconButton size='small' onClick={() => handleOpenEdit(user)}>
                <i className='ri-pencil-line text-actionActive' />
              </IconButton>
            </Tooltip>
            <Tooltip title='Delete User'>
              <IconButton size='small' color='error' onClick={() => setDeleteConfirmUser(user)}>
                <i className='ri-delete-bin-7-line' />
              </IconButton>
            </Tooltip>
          </Box>
        )
      }
    }
  ]

  const activeFilterCount = (roleFilter ? 1 : 0) + (statusFilter ? 1 : 0)

  return (
    <Box className='flex flex-col gap-6 p-6'>
      {/* Page Title Header */}
      <Box className='flex items-center justify-between'>
        <Box>
          <Typography variant='h5' className='font-bold text-textPrimary'>
            User Management
          </Typography>
          <Typography variant='body2' className='text-textSecondary'>
            Manage users, assign security roles, and monitor account access.
          </Typography>
        </Box>
      </Box>

      {/* KPI Stat Cards */}
      <Grid container spacing={4}>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard label='Total Users' value={stats.total} icon='ri-user-line' color='primary' />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard label='Active Accounts' value={stats.active} icon='ri-user-check-line' color='success' />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard label='Administrators' value={stats.admins} icon='ri-shield-user-line' color='warning' />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <StatCard label='Inactive Accounts' value={stats.inactive} icon='ri-user-unfollow-line' color='error' />
        </Grid>
      </Grid>

      {/* Main Users Table */}
      <DataTable
        title={
          <Box className='flex items-center gap-2'>
            <i className='ri-team-line text-primaryMain text-xl' />
            <Typography variant='h6' className='font-bold'>
              All User Accounts
            </Typography>
          </Box>
        }
        columns={columns}
        data={users}
        total={total}
        loading={loading}
        page={tableState.page}
        pageSize={tableState.pageSize}
        onPageChange={tableState.setPage}
        onPageSizeChange={tableState.setPageSize}
        sortBy={tableState.sortBy}
        sortOrder={tableState.sortOrder}
        onToggleSort={tableState.toggleSort}
        toolbar={
          <DataTableFilters
            searchValue={tableState.search}
            onSearchChange={tableState.setSearch}
            searchPlaceholder='Search by name or email...'
            activeFilterCount={activeFilterCount}
            onReset={() => {
              setRoleFilter('')
              setStatusFilter('')
            }}
            action={
              <Button
                variant='contained'
                color='primary'
                startIcon={<i className='ri-user-add-line' />}
                onClick={handleOpenAdd}
              >
                Add New User
              </Button>
            }
          >
            <FormControl size='small' className='min-w-[150px]'>
              <InputLabel>Role Filter</InputLabel>
              <Select value={roleFilter} label='Role Filter' onChange={e => setRoleFilter(e.target.value)}>
                <MenuItem value=''>All Roles</MenuItem>
                <MenuItem value='admin'>Admin</MenuItem>
                <MenuItem value='manager'>Manager</MenuItem>
                <MenuItem value='user'>User</MenuItem>
              </Select>
            </FormControl>

            <FormControl size='small' className='min-w-[150px]'>
              <InputLabel>Status Filter</InputLabel>
              <Select value={statusFilter} label='Status Filter' onChange={e => setStatusFilter(e.target.value)}>
                <MenuItem value=''>All Statuses</MenuItem>
                <MenuItem value='active'>Active</MenuItem>
                <MenuItem value='inactive'>Inactive</MenuItem>
              </Select>
            </FormControl>
          </DataTableFilters>
        }
      />

      {/* Add / Edit User Dialog */}
      <Dialog open={openModal} onClose={() => setOpenModal(false)} maxWidth='sm' fullWidth>
        <DialogTitle className='font-bold text-lg'>
          {selectedUser ? `Edit User: ${selectedUser.email}` : 'Add New User Account'}
        </DialogTitle>
        <DialogContent dividers className='flex flex-col gap-4 py-4'>
          {formError && (
            <Box className='p-3 bg-errorLight text-errorMain rounded-md text-sm font-medium'>{formError}</Box>
          )}

          <Grid container spacing={4}>
            <Grid size={{ xs: 12, sm: 6 }}>
              <TextField
                fullWidth
                size='small'
                label='First Name'
                required
                value={formData.first_name}
                onChange={e => setFormData(prev => ({ ...prev, first_name: e.target.value }))}
              />
            </Grid>
            <Grid size={{ xs: 12, sm: 6 }}>
              <TextField
                fullWidth
                size='small'
                label='Last Name'
                required
                value={formData.last_name}
                onChange={e => setFormData(prev => ({ ...prev, last_name: e.target.value }))}
              />
            </Grid>
          </Grid>

          <TextField
            fullWidth
            size='small'
            label='Email Address'
            type='email'
            required
            disabled={Boolean(selectedUser)}
            value={formData.email}
            onChange={e => setFormData(prev => ({ ...prev, email: e.target.value }))}
          />

          <TextField
            fullWidth
            size='small'
            label={selectedUser ? 'Password (leave blank to keep current)' : 'Password'}
            type='password'
            required={!selectedUser}
            value={formData.password}
            onChange={e => setFormData(prev => ({ ...prev, password: e.target.value }))}
          />

          <Grid container spacing={4} className='items-center'>
            <Grid size={{ xs: 12, sm: 6 }}>
              <FormControl fullWidth size='small'>
                <InputLabel>User Role</InputLabel>
                <Select
                  value={formData.role}
                  label='User Role'
                  onChange={e => setFormData(prev => ({ ...prev, role: e.target.value }))}
                >
                  <MenuItem value='user'>User (Standard)</MenuItem>
                  <MenuItem value='manager'>Manager</MenuItem>
                  <MenuItem value='admin'>Administrator</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid size={{ xs: 12, sm: 6 }}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.is_active}
                    onChange={e => setFormData(prev => ({ ...prev, is_active: e.target.checked }))}
                    color='primary'
                  />
                }
                label={formData.is_active ? 'Account Active' : 'Account Disabled'}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions className='p-4'>
          <Button color='secondary' onClick={() => setOpenModal(false)}>
            Cancel
          </Button>
          <Button variant='contained' color='primary' onClick={handleSaveUser} disabled={isSaving}>
            {isSaving ? 'Saving...' : selectedUser ? 'Update User' : 'Create User'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete User Confirmation */}
      <ConfirmDialog
        open={Boolean(deleteConfirmUser)}
        title='Deactivate / Delete User Account'
        message={`Are you sure you want to delete account ${deleteConfirmUser?.email}? This action cannot be undone.`}
        confirmText='Delete User'
        loading={isDeleting}
        onClose={() => setDeleteConfirmUser(null)}
        onConfirm={handleDeleteUser}
      />
    </Box>
  )
}
