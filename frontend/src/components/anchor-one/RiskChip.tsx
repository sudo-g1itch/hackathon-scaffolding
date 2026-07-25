// Canonical relapse-risk chip. Every screen that shows a risk level renders
// this, so LOW/MEDIUM/HIGH are always coloured and labelled identically.
import Chip from '@mui/material/Chip'

import { resolveRiskColor } from '@/types/anchorOneTypes'

type RiskChipProps = {
  risk?: string | null
  size?: 'small' | 'medium'

  /** Larger treatment for the dashboard hero badge. */
  prominent?: boolean
}

const RISK_ICONS: Record<string, string> = {
  LOW: 'ri-shield-check-line',
  MEDIUM: 'ri-alert-line',
  HIGH: 'ri-alarm-warning-line'
}

const RiskChip = ({ risk, size = 'small', prominent = false }: RiskChipProps) => {
  const level = (risk ?? 'LOW').toUpperCase()

  return (
    <Chip
      size={size}
      variant='tonal'
      color={resolveRiskColor(level)}
      icon={<i className={RISK_ICONS[level] ?? RISK_ICONS.LOW} />}
      label={`${level} RISK`}
      sx={
        prominent
          ? { fontSize: '1.25rem', height: 48, borderRadius: 6, px: 2, fontWeight: 700 }
          : { fontWeight: 600 }
      }
    />
  )
}

export default RiskChip
