export interface Permission {
  id: string
  slug: string
  name: string
  module: string
  description: string
  created_at: string
  updated_at: string
}

export interface Role {
  id: string
  name: string
  description: string
  is_system: boolean
  permissions?: Permission[]
  user_count?: number
  created_at: string
  updated_at: string
}

export interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  role: string
  is_active: boolean
  avatar_url?: string
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface UserFormValues {
  email: string
  first_name: string
  last_name: string
  role: string
  is_active: boolean
  password?: string
}

export interface RoleFormValues {
  name: string
  description: string
  permission_ids: string[]
}
