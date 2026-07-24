'use client'

import type { ReactNode } from 'react'
import { createContext, useContext, useEffect, useState } from 'react'

import { useRouter } from 'next/navigation'

import axios from '@/libs/axios'
import type { AuthResponse, LoginRequest, RegisterRequest, Role, StandardResponse, User } from '@/types/apiTypes'

type AuthContextType = {
  user: User | null
  token: string | null
  loading: boolean
  isAuthenticated: boolean
  login: (credentials: LoginRequest) => Promise<User>
  register: (data: RegisterRequest) => Promise<User>
  logout: () => void
  hasRole: (allowed: Role | Role[]) => boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const router = useRouter()

  useEffect(() => {
    const initAuth = async () => {
      const storedToken = localStorage.getItem('accessToken')

      if (!storedToken) {
        setLoading(false)

        return
      }

      try {
        setToken(storedToken)
        const response = await axios.get<StandardResponse<User>>('/auth/me')

        if (response.data.success && response.data.data) {
          setUser(response.data.data)
        } else {
          localStorage.removeItem('accessToken')
          setToken(null)
        }
      } catch {
        localStorage.removeItem('accessToken')
        setToken(null)
      } finally {
        setLoading(false)
      }
    }

    initAuth()
  }, [])

  const login = async (credentials: LoginRequest): Promise<User> => {
    const response = await axios.post<StandardResponse<AuthResponse>>('/auth/login', credentials)

    if (response.data.success && response.data.data) {
      const { user: userData, access_token } = response.data.data

      localStorage.setItem('accessToken', access_token)
      setToken(access_token)
      setUser(userData)

      return userData
    }

    throw new Error(response.data.error?.message || 'Login failed')
  }

  const register = async (data: RegisterRequest): Promise<User> => {
    const response = await axios.post<StandardResponse<AuthResponse>>('/auth/register', data)

    if (response.data.success && response.data.data) {
      const { user: userData, access_token } = response.data.data

      localStorage.setItem('accessToken', access_token)
      setToken(access_token)
      setUser(userData)

      return userData
    }

    throw new Error(response.data.error?.message || 'Registration failed')
  }

  const logout = () => {
    localStorage.removeItem('accessToken')
    setToken(null)
    setUser(null)
    router.push('/login')
  }

  const hasRole = (allowed: Role | Role[]): boolean => {
    if (!user) return false
    const allowedRoles = Array.isArray(allowed) ? allowed : [allowed]

    return allowedRoles.includes(user.role)
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        loading,
        isAuthenticated: Boolean(user && token),
        login,
        register,
        logout,
        hasRole
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext)

  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }

  return context
}

export default AuthContext
