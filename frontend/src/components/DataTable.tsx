'use client'

// The single list table used by every screen. It owns react-table wiring,
// header rendering + server-sort affordances, loading skeleton, empty state,
// and pagination.
import type { CSSProperties, ReactNode } from 'react'

import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import Divider from '@mui/material/Divider'
import Skeleton from '@mui/material/Skeleton'
import classnames from 'classnames'
import { flexRender, getCoreRowModel, useReactTable, type ColumnDef, type TableOptions } from '@tanstack/react-table'

import type { SortOrder } from '@/types/apiTypes'
import EmptyState from '@components/EmptyState'
import TablePaginationCustom from '@components/TablePaginationCustom'
import tableStyles from '@core/styles/table.module.css'

const SKELETON_WIDTHS = ['80%', '55%', '70%', '45%', '65%']

type DataTableProps<T> = {
  columns: ColumnDef<T>[]
  data: T[]
  total: number
  loading?: boolean

  // Pagination (from useServerTable)
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  onPageSizeChange: (pageSize: number) => void

  // Server sort (optional; from useServerTable)
  sortBy?: string
  sortOrder?: SortOrder
  onToggleSort?: (columnId: string) => void

  // Chrome
  title?: ReactNode

  /** Toolbar rendered between the header and the table (usually <DataTableFilters>). */
  toolbar?: ReactNode
  emptyMessage?: string

  /** Remix icon for the empty state. */
  emptyIcon?: string

  /** Optional call-to-action rendered under the empty-state message. */
  emptyAction?: ReactNode

  /** Extra class on the root Card. */
  className?: string
}

const DataTable = <T,>({
  columns,
  data,
  total,
  loading = false,
  page,
  pageSize,
  onPageChange,
  onPageSizeChange,
  sortBy,
  sortOrder,
  onToggleSort,
  title,
  toolbar,
  emptyMessage = 'No results found',
  emptyIcon = 'ri-inbox-2-line',
  emptyAction,
  className
}: DataTableProps<T>) => {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualSorting: true,
    pageCount: pageSize > 0 ? Math.ceil(total / pageSize) : 0
  } satisfies Partial<TableOptions<T>> as TableOptions<T>)

  const columnCount = columns.length

  return (
    <Card className={className}>
      {title ? <CardHeader title={title} /> : null}
      {toolbar}
      {(title || toolbar) && <Divider />}
      <div className='overflow-x-auto'>
        <table className={tableStyles.table}>
          <thead>
            {table.getHeaderGroups().map(headerGroup => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map(header => {
                  const canSort = Boolean(onToggleSort) && header.column.getCanSort()
                  const isSorted = sortBy === header.column.id
                  const content = flexRender(header.column.columnDef.header, header.getContext())

                  const isActions = header.column.id === 'actions'

                  return (
                    <th key={header.id} className={classnames({ 'rcm-sticky-action': isActions })}>
                      {canSort ? (
                        <div
                          role='button'
                          tabIndex={0}
                          className='group flex items-center gap-1 cursor-pointer select-none'
                          onClick={() => onToggleSort?.(header.column.id)}
                          onKeyDown={event => {
                            if (event.key === 'Enter' || event.key === ' ') {
                              event.preventDefault()
                              onToggleSort?.(header.column.id)
                            }
                          }}
                        >
                          {content}
                          <i
                            className={classnames('text-base transition-opacity duration-150', {
                              'ri-arrow-up-s-line': isSorted && sortOrder === 'asc',
                              'ri-arrow-down-s-line': isSorted && sortOrder === 'desc',
                              'ri-expand-up-down-line opacity-40 group-hover:opacity-70': !isSorted
                            })}
                          />
                        </div>
                      ) : (
                        content
                      )}
                    </th>
                  )
                })}
              </tr>
            ))}
          </thead>
          <tbody key={`${page}-${sortBy ?? ''}-${sortOrder ?? ''}`}>
            {loading ? (
              Array.from({ length: Math.min(pageSize, 8) }).map((_, rowIndex) => (
                <tr key={`skeleton-${rowIndex}`}>
                  {Array.from({ length: columnCount }).map((__, cellIndex) => (
                    <td key={`skeleton-${rowIndex}-${cellIndex}`}>
                      <Skeleton
                        variant='text'
                        animation='wave'
                        width={SKELETON_WIDTHS[(rowIndex + cellIndex) % SKELETON_WIDTHS.length]}
                      />
                    </td>
                  ))}
                </tr>
              ))
            ) : data.length === 0 ? (
              <tr>
                <td colSpan={columnCount} className='p-2'>
                  <EmptyState icon={emptyIcon} message={emptyMessage} action={emptyAction} size='sm' />
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row, rowIndex) => (
                <tr key={row.id} className='dt-row-in' style={{ '--stagger-i': Math.min(rowIndex, 12) } as CSSProperties}>
                  {row.getVisibleCells().map(cell => (
                    <td
                      key={cell.id}
                      className={classnames({ 'rcm-sticky-action': cell.column.id === 'actions' })}
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      <TablePaginationCustom
        total={total}
        page={page}
        limit={pageSize}
        onPageChange={onPageChange}
        onLimitChange={onPageSizeChange}
      />
    </Card>
  )
}

export default DataTable
