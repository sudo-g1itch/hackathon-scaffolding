'use client'

// useServerTable owns the client-side state for a server-paginated list:
// pagination, sort, and search term. It produces a memoized `params` object
// (ListQueryParams) passed to API hooks and query keys. <DataTable> consumes
// the returned state & handlers.
import { useCallback, useMemo, useState } from 'react'

import type { ListQueryParams, SortOrder } from '@/types/apiTypes'

type UseServerTableOptions = {
  initialPageSize?: number
  initialSortBy?: string
  initialSortOrder?: SortOrder
}

export type ServerTable = {
  params: ListQueryParams
  page: number
  pageSize: number
  search: string
  sortBy?: string
  sortOrder: SortOrder
  setPage: (page: number) => void
  setPageSize: (pageSize: number) => void
  setSearch: (value: string) => void

  /** Toggle/set sort for a column id (asc ⇄ desc); resets to page 1. */
  toggleSort: (columnId: string) => void
}

export const useServerTable = (options: UseServerTableOptions = {}): ServerTable => {
  const { initialPageSize = 10, initialSortBy, initialSortOrder = 'asc' } = options

  const [page, setPageState] = useState(1)
  const [pageSize, setPageSizeState] = useState(initialPageSize)
  const [search, setSearchState] = useState('')
  const [sortBy, setSortBy] = useState<string | undefined>(initialSortBy)
  const [sortOrder, setSortOrder] = useState<SortOrder>(initialSortOrder)

  const setPage = useCallback((next: number) => setPageState(next), [])

  const setPageSize = useCallback((next: number) => {
    setPageSizeState(next)
    setPageState(1)
  }, [])

  const setSearch = useCallback((value: string) => {
    setSearchState(value)
    setPageState(1)
  }, [])

  const toggleSort = useCallback(
    (columnId: string) => {
      setPageState(1)

      if (sortBy === columnId) {
        setSortOrder(prev => (prev === 'asc' ? 'desc' : 'asc'))
      } else {
        setSortBy(columnId)
        setSortOrder('asc')
      }
    },
    [sortBy]
  )

  const params = useMemo<ListQueryParams>(
    () => ({
      page,
      pageSize,
      q: search.trim() || undefined,
      sortBy,
      sortOrder: sortBy ? sortOrder : undefined
    }),
    [page, pageSize, search, sortBy, sortOrder]
  )

  return { params, page, pageSize, search, sortBy, sortOrder, setPage, setPageSize, setSearch, toggleSort }
}
