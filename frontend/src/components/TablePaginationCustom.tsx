'use client'

import { useEffect, useState } from 'react'

import TablePagination from '@mui/material/TablePagination'
import { Typography } from '@mui/material'

interface TablePaginationCustomProps {
  total: number
  page: number
  limit: number
  onPageChange: (page: number) => void
  onLimitChange: (limit: number) => void
}

const TablePaginationCustom = ({
  total,
  page,
  limit,
  onPageChange,
  onLimitChange
}: TablePaginationCustomProps) => {
  const [mounted, setMounted] = useState(false)

  useEffect(() => setMounted(true), [])

  return (
    <div className='flex items-center justify-between p-4 border-bs min-bs-[3.25rem]'>
      <Typography variant='body2' color='text.secondary'>
        Showing {total === 0 ? 0 : Math.min((page - 1) * limit + 1, total)} to {Math.min(page * limit, total)} of {total} records
      </Typography>
      {mounted && (
        <TablePagination
          component='div'
          count={total}
          page={page - 1}
          onPageChange={(_, newPage) => onPageChange(newPage + 1)}
          rowsPerPage={limit}
          onRowsPerPageChange={(e) => onLimitChange(parseInt(e.target.value, 10))}
          rowsPerPageOptions={[10, 25, 50, 100]}
        />
      )}
    </div>
  )
}

export default TablePaginationCustom
