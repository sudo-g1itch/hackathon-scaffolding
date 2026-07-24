import { useState, useCallback } from 'react'

export interface PaginationState {
  page: number
  pageSize: number
  total: number
}

export const usePagination = (initialPageSize: number = 10) => {
  const [pagination, setPagination] = useState<PaginationState>({
    page: 1,
    pageSize: initialPageSize,
    total: 0
  })

  const setPage = useCallback((page: number) => {
    setPagination(prev => ({ ...prev, page }))
  }, [])

  const setPageSize = useCallback((pageSize: number) => {
    setPagination(prev => ({ ...prev, pageSize, page: 1 }))
  }, [])

  const setTotal = useCallback((total: number) => {
    setPagination(prev => ({ ...prev, total }))
  }, [])

  return {
    ...pagination,
    setPage,
    setPageSize,
    setTotal,
    paginationProps: {
      total: pagination.total,
      page: pagination.page,
      limit: pagination.pageSize,
      onPageChange: setPage,
      onLimitChange: setPageSize
    }
  }
}
