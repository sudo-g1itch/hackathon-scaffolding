'use client'

// Standard toolbar above a <DataTable>: a debounced search box, optional extra
// filter controls (children), and an optional action slot (e.g. an "Add" button).
import type { ReactNode } from 'react'

import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Chip from '@mui/material/Chip'
import InputAdornment from '@mui/material/InputAdornment'

import DebouncedInput from '@components/DebouncedInput'

type DataTableFiltersProps = {
  searchValue: string
  onSearchChange: (value: string) => void
  searchLabel?: string
  searchPlaceholder?: string

  /** Hide the search box. */
  hideSearch?: boolean

  /** Extra filter controls rendered next to the search box. */
  children?: ReactNode

  /** Pinned to the end of the toolbar (typically a primary action button). */
  action?: ReactNode

  /** Number of active (non-default) filters. */
  activeFilterCount?: number

  /** Clears custom screen filters and search value. */
  onReset?: () => void
}

const DataTableFilters = ({
  searchValue,
  onSearchChange,
  searchLabel,
  searchPlaceholder = 'Search...',
  hideSearch = false,
  children,
  action,
  activeFilterCount = 0,
  onReset
}: DataTableFiltersProps) => {
  const showReset = searchValue.trim() !== '' || activeFilterCount > 0

  const handleReset = () => {
    onSearchChange('')
    onReset?.()
  }

  const hasFilters = Boolean(children)

  return (
    <Box sx={{ p: 4, display: 'flex', flexDirection: 'column', gap: 4 }}>
      {hasFilters && action ? (
        <Box sx={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'flex-end', gap: 2 }}>
          {action}
        </Box>
      ) : null}
      <Box
        sx={{
          display: 'flex',
          flexWrap: 'wrap',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 4
        }}
      >
        <Box sx={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 4, flex: 1 }}>
          {!hideSearch && (
            <DebouncedInput
              value={searchValue}
              onChange={value => onSearchChange(String(value))}
              label={searchLabel}
              placeholder={searchPlaceholder}
              size='small'
              sx={{ minInlineSize: 240 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position='start'>
                    <i className='ri-search-line text-xl text-textDisabled' />
                  </InputAdornment>
                )
              }}
            />
          )}
          {children}
          {activeFilterCount > 0 ? (
            <Chip
              size='small'
              variant='tonal'
              color='primary'
              icon={<i className='ri-filter-3-line' />}
              label={activeFilterCount}
            />
          ) : null}
          {showReset ? (
            <Button
              size='small'
              color='secondary'
              startIcon={<i className='ri-refresh-line' />}
              onClick={handleReset}
            >
              Reset Filters
            </Button>
          ) : null}
        </Box>
        {!hasFilters && action ? <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>{action}</Box> : null}
      </Box>
    </Box>
  )
}

export default DataTableFilters
