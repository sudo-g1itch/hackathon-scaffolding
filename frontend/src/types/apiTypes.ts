export type Role = 'admin' | 'manager' | 'user'

export type User = {
  id: string
  created_at: string
  updated_at: string
  email: string
  first_name: string
  last_name: string
  role: Role
  is_active: boolean
  avatar_url?: string
  last_login_at?: string
}

export type AuthResponse = {
  user: User
  access_token: string
  token_type: string
  expires_in: number
}

export type LoginRequest = {
  email: string
  password: string
}

export type RegisterRequest = {
  email: string
  password: string
  first_name: string
  last_name: string
  role?: Role
}

export type PaginationMetadata = {
  total: number
  page: number
  page_size: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export type ErrorDetail = {
  code: string
  message: string
  fields?: Record<string, string[]>
  request_id?: string
}

export type Meta = {
  pagination?: PaginationMetadata
}

export type StandardResponse<T> = {
  success: boolean
  data?: T
  meta?: Meta
  error?: ErrorDetail
}

export type SortOrder = 'asc' | 'desc'

// ListQueryParams is the canonical client-side shape for a server-paginated,
// searchable, sortable list request. Service files translate it into the
// backend's snake_case query params (page, page_size, q, sort_by, sort_order)
// and it is also spread into React Query keys so the cache is per-query.
export type ListQueryParams = {
  page: number
  pageSize: number
  q?: string
  sortBy?: string
  sortOrder?: SortOrder
}
