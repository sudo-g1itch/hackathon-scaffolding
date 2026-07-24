// Single error funnel for mutations/queries. Field-level validation
// errors from the backend envelope (`error.fields`) are mapped back onto
// react-hook-form via setError if provided.
import { isAxiosError } from 'axios'
import type { FieldValues, Path, UseFormSetError } from 'react-hook-form'

import type { StandardResponse } from '@/types/apiTypes'

const DEFAULT_MESSAGE = 'Something went wrong. Please try again.'

export const getApiErrorMessage = (error: unknown, fallback: string = DEFAULT_MESSAGE): string => {
  if (isAxiosError(error)) {
    const apiError = (error.response?.data as StandardResponse<unknown> | undefined)?.error

    if (apiError?.message) return apiError.message
  } else if (error instanceof Error && error.message) {
    return error.message
  }

  return fallback
}

export const handleApiError = <T extends FieldValues = FieldValues>(
  error: unknown,
  setError?: UseFormSetError<T>,
  fallbackMessage: string = DEFAULT_MESSAGE
): string => {
  const message = getApiErrorMessage(error, fallbackMessage)

  if (isAxiosError(error) && setError) {
    const apiError = (error.response?.data as StandardResponse<unknown> | undefined)?.error

    if (apiError?.fields) {
      for (const [field, messages] of Object.entries(apiError.fields)) {
        setError(field as Path<T>, {
          type: 'server',
          message: Array.isArray(messages) ? messages.join(', ') : String(messages)
        })
      }
    }
  }

  return message
}
