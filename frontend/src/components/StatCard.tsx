'use client'

// Canonical KPI stat card — tinted icon badge, large tabular figure, label,
// optional caption and trend, with a color-coded accent bar down the start edge.
import type { CSSProperties } from 'react'

import Box from '@mui/material/Box'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Skeleton from '@mui/material/Skeleton'
import Typography from '@mui/material/Typography'
import { alpha, useTheme } from '@mui/material/styles'
import classnames from 'classnames'

import type { ThemeColor } from '@core/types'
import useCountUp from '@/hooks/useCountUp'

export type StatCardTrend = {

  /** Pre-formatted delta, e.g. "12%" or "+3,420". */
  value: string
  direction: 'up' | 'down'

  /** Whether an upward move is good (default true). */
  positiveIsGood?: boolean
}

export type StatCardProps = {
  label: string

  /** Numbers animate (count-up) and format; strings render as-is. */
  value: number | string

  /** Remix icon class; omit to hide badge. */
  icon?: string
  color?: ThemeColor
  caption?: string
  trend?: StatCardTrend
  loading?: boolean
  formatter?: (value: number) => string

  /** Disable count-up animation. */
  animate?: boolean

  /** Row index for entrance animation stagger. */
  index?: number
}

const defaultFormatter = (value: number) => Math.round(value).toLocaleString()

const StatCard = ({
  label,
  value,
  icon,
  color = 'primary',
  caption,
  trend,
  loading = false,
  formatter = defaultFormatter,
  animate = true,
  index = 0
}: StatCardProps) => {
  const theme = useTheme()
  const accent = theme.palette[color].main
  const isNumeric = typeof value === 'number'
  const animated = useCountUp(isNumeric ? value : 0, { enabled: animate && isNumeric && !loading })
  const display = isNumeric ? formatter(animated) : value

  const trendIsGood = trend ? (trend.direction === 'up') === (trend.positiveIsGood ?? true) : true
  const trendColor = trendIsGood ? theme.palette.success.main : theme.palette.error.main

  return (
    <Card
      className={classnames({ 'stat-card-in': !loading })}
      style={{ '--stagger-i': index } as CSSProperties}
      sx={{
        height: '100%',
        position: 'relative',
        overflow: 'hidden',

        '&::before': {
          content: '""',
          position: 'absolute',
          insetBlock: 0,
          insetInlineStart: 0,
          inlineSize: 4,
          background: `linear-gradient(to bottom, ${accent}, ${alpha(accent, 0.45)})`
        }
      }}
    >
      <CardContent className='flex flex-col gap-3'>
        <div className='flex items-center justify-between gap-2'>
          {loading ? (
            <Skeleton variant='text' width='55%' />
          ) : (
            <Typography variant='body2' color='text.secondary' sx={{ fontWeight: 500 }}>
              {label}
            </Typography>
          )}
          {icon ? (
            <Box
              sx={{
                width: 52,
                height: 52,
                flexShrink: 0,
                borderRadius: 1.5,
                display: 'grid',
                placeItems: 'center',
                color: accent,
                background: `linear-gradient(135deg, ${alpha(accent, 0.2)}, ${alpha(accent, 0.07)})`,
                boxShadow: `inset 0 0 0 1px ${alpha(accent, 0.18)}`
              }}
            >
              <i className={`${icon} text-[28px]`} />
            </Box>
          ) : null}
        </div>
        {loading ? (
          <Skeleton variant='text' width='70%' sx={{ fontSize: theme.typography.h4.fontSize }} />
        ) : (
          <Typography
            variant='h4'
            sx={{ fontWeight: 600, fontVariantNumeric: 'tabular-nums', lineHeight: 1.1 }}
          >
            {display}
          </Typography>
        )}
        {!loading && (caption || trend) ? (
          <div className='flex items-center gap-2'>
            {trend ? (
              <Typography
                variant='caption'
                component='span'
                sx={{ color: trendColor, fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: 0.5 }}
              >
                <i
                  className={classnames('text-sm', {
                    'ri-arrow-up-line': trend.direction === 'up',
                    'ri-arrow-down-line': trend.direction === 'down'
                  })}
                />
                {trend.value}
              </Typography>
            ) : null}
            {caption ? (
              <Typography variant='caption' color='text.secondary'>
                {caption}
              </Typography>
            ) : null}
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}

export default StatCard
