import axios from 'axios'

const instance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:20080/api/v1',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Attach access token from localStorage (or authorization header) if present.
instance.interceptors.request.use(
  config => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('token') || localStorage.getItem('accessToken')

      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
    }

    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// Global response interceptor for 401 handling.
instance.interceptors.response.use(
  response => response,
  error => {
    const status = error?.response?.status

    if (status === 401 && typeof window !== 'undefined') {
      const isAuthRoute = window.location.pathname.includes('/login')

      if (!isAuthRoute) {
        localStorage.removeItem('token')
        localStorage.removeItem('accessToken')
        window.location.href = '/login'
      }
    }

    return Promise.reject(error)
  }
)

export default instance
