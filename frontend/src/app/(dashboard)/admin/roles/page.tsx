'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'

import Grid from '@mui/material/Grid'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CardActions from '@mui/material/CardActions'
import Button from '@mui/material/Button'
import Typography from '@mui/material/Typography'
import Avatar from '@mui/material/Avatar'
import AvatarGroup from '@mui/material/AvatarGroup'
import IconButton from '@mui/material/IconButton'
import Dialog from '@mui/material/Dialog'
import DialogTitle from '@mui/material/DialogTitle'
import DialogContent from '@mui/material/DialogContent'
import DialogActions from '@mui/material/DialogActions'
import TextField from '@mui/material/TextField'
import Checkbox from '@mui/material/Checkbox'
import FormControlLabel from '@mui/material/FormControlLabel'
import Chip from '@mui/material/Chip'
import Box from '@mui/material/Box'
import Divider from '@mui/material/Divider'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Paper from '@mui/material/Paper'

import AuthGuard from '@components/auth/AuthGuard'
import RoleGuard from '@components/auth/RoleGuard'
import ConfirmDialog from '@components/ConfirmDialog'
import { rbacService } from '@/services/rbacService'
import type { Role, Permission, RoleFormValues } from '@/types/rbacTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

export default function RolesPage() {
  return (
    <AuthGuard>
      <RoleGuard roles={['admin']}>
        <RolesManagementContent />
      </RoleGuard>
    </AuthGuard>
  )
}

function RolesManagementContent() {
  const [roles, setRoles] = useState<Role[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])

  // Modal State
  const [openModal, setOpenModal] = useState<boolean>(false)
  const [selectedRole, setSelectedRole] = useState<Role | null>(null)

  const [roleForm, setRoleForm] = useState<RoleFormValues>({
    name: '',
    description: '',
    permission_ids: []
  })

  const [formError, setFormError] = useState<string>('')
  const [isSaving, setIsSaving] = useState<boolean>(false)

  // Delete State
  const [deleteRole, setDeleteRole] = useState<Role | null>(null)
  const [isDeleting, setIsDeleting] = useState<boolean>(false)

  const loadData = useCallback(async () => {
    try {
      const [fetchedRoles, fetchedPerms] = await Promise.all([
        rbacService.getRoles(),
        rbacService.getPermissions()
      ])

      setRoles(fetchedRoles)
      setPermissions(fetchedPerms)
    } catch {
      // Error handled by interceptor
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [loadData])

  // Group permissions by module
  const groupedPermissions = useMemo(() => {
    const map: Record<string, Permission[]> = {}

    permissions.forEach(p => {
      if (!map[p.module]) {
        map[p.module] = []
      }

      map[p.module]?.push(p)
    })
    
return map
  }, [permissions])

  const handleOpenAdd = () => {
    setSelectedRole(null)
    setRoleForm({
      name: '',
      description: '',
      permission_ids: permissions.map(p => p.id)
    })
    setFormError('')
    setOpenModal(true)
  }

  const handleOpenEdit = (role: Role) => {
    setSelectedRole(role)
    const existingPermIDs = role.permissions ? role.permissions.map(p => p.id) : []

    setRoleForm({
      name: role.name,
      description: role.description,
      permission_ids: existingPermIDs
    })
    setFormError('')
    setOpenModal(true)
  }

  const handleTogglePermission = (permId: string) => {
    setRoleForm(prev => {
      const exists = prev.permission_ids.includes(permId)

      
return {
        ...prev,
        permission_ids: exists
          ? prev.permission_ids.filter(id => id !== permId)
          : [...prev.permission_ids, permId]
      }
    })
  }

  const handleToggleModule = (moduleName: string) => {
    const modulePermIDs = groupedPermissions[moduleName]?.map(p => p.id) ?? []
    const allSelected = modulePermIDs.every(id => roleForm.permission_ids.includes(id))

    setRoleForm(prev => {
      if (allSelected) {
        return {
          ...prev,
          permission_ids: prev.permission_ids.filter(id => !modulePermIDs.includes(id))
        }
      } else {
        const merged = new Set([...prev.permission_ids, ...modulePermIDs])

        
return {
          ...prev,
          permission_ids: Array.from(merged)
        }
      }
    })
  }

  const handleToggleAll = () => {
    const allPermIDs = permissions.map(p => p.id)
    const allSelected = allPermIDs.length === roleForm.permission_ids.length

    setRoleForm(prev => ({
      ...prev,
      permission_ids: allSelected ? [] : allPermIDs
    }))
  }

  const handleSaveRole = async () => {
    setFormError('')

    if (!roleForm.name.trim()) {
      setFormError('Role title is required.')
      
return
    }

    setIsSaving(true)

    try {
      if (selectedRole) {
        await rbacService.updateRole(selectedRole.id, roleForm)
      } else {
        await rbacService.createRole(roleForm)
      }

      setOpenModal(false)
      loadData()
    } catch (err: unknown) {
      setFormError(getApiErrorMessage(err))
    } finally {
      setIsSaving(false)
    }
  }

  const handleDeleteRole = async () => {
    if (!deleteRole) return
    setIsDeleting(true)

    try {
      await rbacService.deleteRole(deleteRole.id)
      setDeleteRole(null)
      loadData()
    } catch (err: unknown) {
      alert(getApiErrorMessage(err))
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <Box className='flex flex-col gap-6 p-6'>
      {/* Header */}
      <Box className='flex items-center justify-between'>
        <Box>
          <Typography variant='h5' className='font-bold text-textPrimary'>
            Roles & Permission Matrix
          </Typography>
          <Typography variant='body2' className='text-textSecondary'>
            Define security roles, configure granular access permissions, and assign users.
          </Typography>
        </Box>
        <Button
          variant='contained'
          color='primary'
          startIcon={<i className='ri-add-line' />}
          onClick={handleOpenAdd}
        >
          Add New Role
        </Button>
      </Box>

      {/* Role Cards Grid */}
      <Grid container spacing={6}>
        {roles.map(role => (
          <Grid size={{ xs: 12, sm: 6, md: 4 }} key={role.id}>
            <Card className='h-full flex flex-col justify-between hover:shadow-lg transition-shadow duration-200 border border-divider'>
              <CardContent className='flex flex-col gap-4 p-5'>
                <Box className='flex items-center justify-between'>
                  <Typography variant='body2' className='text-textSecondary font-medium text-xs'>
                    Total {role.user_count ?? 0} users
                  </Typography>
                  <AvatarGroup max={4} className='pull-up'>
                    {Array.from({ length: Math.min(role.user_count ?? 1, 4) }).map((_, idx) => (
                      <Avatar key={idx} className='w-8 h-8 text-xs bg-primaryLight text-primaryMain font-bold'>
                        U{idx + 1}
                      </Avatar>
                    ))}
                  </AvatarGroup>
                </Box>

                <Box>
                  <Box className='flex items-center gap-2 mb-1'>
                    <Typography variant='h6' className='font-bold text-textPrimary capitalize'>
                      {role.name}
                    </Typography>
                    {role.is_system && (
                      <Chip label='System Role' size='small' variant='tonal' color='secondary' className='text-[10px] h-5 font-semibold' />
                    )}
                  </Box>
                  <Typography variant='body2' className='text-textSecondary text-xs line-clamp-2 min-h-[36px]'>
                    {role.description || 'No description specified for this role.'}
                  </Typography>
                </Box>
              </CardContent>

              <Divider />

              <CardActions className='flex items-center justify-between p-4 bg-backgroundDefault/50'>
                <Button
                  size='small'
                  color='primary'
                  startIcon={<i className='ri-shield-keyhole-line' />}
                  onClick={() => handleOpenEdit(role)}
                >
                  Edit Role Matrix
                </Button>
                {!role.is_system && (
                  <IconButton size='small' color='error' onClick={() => setDeleteRole(role)}>
                    <i className='ri-delete-bin-line' />
                  </IconButton>
                )}
              </CardActions>
            </Card>
          </Grid>
        ))}

        {/* Add Role Card */}
        <Grid size={{ xs: 12, sm: 6, md: 4 }}>
          <Card
            onClick={handleOpenAdd}
            className='h-full flex flex-col items-center justify-center p-8 border-2 border-dashed border-divider hover:border-primaryMain cursor-pointer transition-colors duration-200 min-h-[200px]'
          >
            <Avatar className='w-12 h-12 mb-3 bg-primaryLight text-primaryMain'>
              <i className='ri-add-line text-2xl' />
            </Avatar>
            <Typography variant='h6' className='font-bold text-primaryMain mb-1'>
              Create Custom Role
            </Typography>
            <Typography variant='body2' className='text-textSecondary text-center text-xs'>
              Set up new access levels with custom permission matrices.
            </Typography>
          </Card>
        </Grid>
      </Grid>

      {/* Permission Matrix Dialog */}
      <Dialog open={openModal} onClose={() => setOpenModal(false)} maxWidth='md' fullWidth>
        <DialogTitle className='font-bold text-lg flex items-center justify-between'>
          <span>{selectedRole ? `Edit Role: ${selectedRole.name}` : 'Create New Security Role'}</span>
          <Chip label={`${roleForm.permission_ids.length} Permissions Selected`} color='primary' size='small' />
        </DialogTitle>
        <DialogContent dividers className='flex flex-col gap-6 py-4'>
          {formError && (
            <Box className='p-3 bg-errorLight text-errorMain rounded-md text-sm font-medium'>{formError}</Box>
          )}

          <Grid container spacing={4}>
            <Grid size={{ xs: 12, sm: 6 }}>
              <TextField
                fullWidth
                size='small'
                label='Role Name'
                required
                disabled={selectedRole?.is_system}
                value={roleForm.name}
                onChange={e => setRoleForm(prev => ({ ...prev, name: e.target.value }))}
              />
            </Grid>
            <Grid size={{ xs: 12, sm: 6 }}>
              <TextField
                fullWidth
                size='small'
                label='Role Description'
                value={roleForm.description}
                onChange={e => setRoleForm(prev => ({ ...prev, description: e.target.value }))}
              />
            </Grid>
          </Grid>

          <Box>
            <Box className='flex items-center justify-between mb-3'>
              <Typography variant='subtitle1' className='font-bold text-textPrimary'>
                Role Permissions Matrix
              </Typography>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={permissions.length > 0 && roleForm.permission_ids.length === permissions.length}
                    onChange={handleToggleAll}
                  />
                }
                label={<Typography className='text-sm font-semibold'>Select All Permissions</Typography>}
              />
            </Box>

            <Paper variant='outlined' className='overflow-hidden'>
              <Table size='small'>
                <TableHead className='bg-backgroundDefault'>
                  <TableRow>
                    <TableCell className='font-bold'>Module</TableCell>
                    <TableCell className='font-bold'>Permission</TableCell>
                    <TableCell className='font-bold'>Description</TableCell>
                    <TableCell align='center' className='font-bold'>Granted</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {Object.entries(groupedPermissions).map(([moduleName, modulePerms]) => {
                    const allModuleSelected = modulePerms.every(p => roleForm.permission_ids.includes(p.id))

                    
return modulePerms.map((perm, idx) => (
                      <TableRow key={perm.id} hover>
                        {idx === 0 && (
                          <TableCell rowSpan={modulePerms.length} className='font-bold align-top bg-backgroundPaper border-r'>
                            <Box className='flex flex-col gap-1'>
                              <span>{moduleName}</span>
                              <FormControlLabel
                                control={
                                  <Checkbox
                                    size='small'
                                    checked={allModuleSelected}
                                    onChange={() => handleToggleModule(moduleName)}
                                  />
                                }
                                label={<Typography variant='caption' className='text-xs text-textSecondary'>All {moduleName}</Typography>}
                              />
                            </Box>
                          </TableCell>
                        )}
                        <TableCell className='font-medium text-xs'>
                          <Chip label={perm.slug} size='small' variant='tonal' className='font-mono text-[11px]' />
                        </TableCell>
                        <TableCell className='text-xs text-textSecondary'>{perm.description}</TableCell>
                        <TableCell align='center'>
                          <Checkbox
                            size='small'
                            checked={roleForm.permission_ids.includes(perm.id)}
                            onChange={() => handleTogglePermission(perm.id)}
                          />
                        </TableCell>
                      </TableRow>
                    ))
                  })}
                </TableBody>
              </Table>
            </Paper>
          </Box>
        </DialogContent>

        <DialogActions className='p-4'>
          <Button color='secondary' onClick={() => setOpenModal(false)}>
            Cancel
          </Button>
          <Button variant='contained' color='primary' onClick={handleSaveRole} disabled={isSaving}>
            {isSaving ? 'Saving...' : 'Save Role Matrix'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={Boolean(deleteRole)}
        title='Delete Custom Security Role'
        message={`Are you sure you want to delete role "${deleteRole?.name}"?`}
        confirmText='Delete Role'
        loading={isDeleting}
        onClose={() => setDeleteRole(null)}
        onConfirm={handleDeleteRole}
      />
    </Box>
  )
}
