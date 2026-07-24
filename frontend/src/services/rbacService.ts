import axios from '@/libs/axios'
import type { StandardResponse, ListQueryParams, PaginationMetadata } from '@/types/apiTypes'
import type { User, Role, Permission, UserFormValues, RoleFormValues } from '@/types/rbacTypes'

export const rbacService = {
  // Users
  async getUsers(params?: ListQueryParams | Record<string, unknown>): Promise<{ data: User[]; meta?: { pagination?: PaginationMetadata } }> {
    const res = await axios.get<StandardResponse<User[]>>('/users', { params })

    return {
      data: res.data.data ?? [],
      meta: res.data.meta
    }
  },

  async getUser(id: string): Promise<User> {
    const res = await axios.get<StandardResponse<User>>(`/users/${id}`)

    if (!res.data.data) throw new Error('User not found')

    return res.data.data
  },

  async createUser(data: UserFormValues): Promise<User> {
    const res = await axios.post<StandardResponse<User>>('/users', data)

    if (!res.data.data) throw new Error('Failed to create user')

    return res.data.data
  },

  async updateUser(id: string, data: Partial<UserFormValues>): Promise<User> {
    const res = await axios.put<StandardResponse<User>>(`/users/${id}`, data)

    if (!res.data.data) throw new Error('Failed to update user')

    return res.data.data
  },

  async deleteUser(id: string): Promise<void> {
    await axios.delete(`/users/${id}`)
  },

  // Roles
  async getRoles(): Promise<Role[]> {
    const res = await axios.get<StandardResponse<Role[]>>('/roles')

    return res.data.data ?? []
  },

  async getRole(id: string): Promise<Role> {
    const res = await axios.get<StandardResponse<Role>>(`/roles/${id}`)

    if (!res.data.data) throw new Error('Role not found')

    return res.data.data
  },

  async createRole(data: RoleFormValues): Promise<Role> {
    const res = await axios.post<StandardResponse<Role>>('/roles', data)

    if (!res.data.data) throw new Error('Failed to create role')

    return res.data.data
  },

  async updateRole(id: string, data: RoleFormValues): Promise<Role> {
    const res = await axios.put<StandardResponse<Role>>(`/roles/${id}`, data)

    if (!res.data.data) throw new Error('Failed to update role')

    return res.data.data
  },

  async deleteRole(id: string): Promise<void> {
    await axios.delete(`/roles/${id}`)
  },

  // Permissions
  async getPermissions(): Promise<Permission[]> {
    const res = await axios.get<StandardResponse<Permission[]>>('/permissions')

    return res.data.data ?? []
  }
}
