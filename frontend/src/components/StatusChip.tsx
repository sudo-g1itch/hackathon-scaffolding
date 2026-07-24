// Shared lifecycle-status chip with leading status dot.
import Chip from '@mui/material/Chip'

import type { ThemeColor } from '@core/types'
import { resolveStatusColor, formatStatus } from '@configs/statusColors'

type StatusChipProps = {

  /** Raw backend status key (e.g. "Partially_Paid") — drives color. */
  status: string

  /** Display label; falls back to humanized status key. */
  label?: string
  size?: 'small' | 'medium'

  /** Screen-local semantic exceptions. */
  overrides?: Record<string, ThemeColor>
}

const Dot = () => (
  <span
    style={{
      display: 'inline-block',
      blockSize: 6,
      inlineSize: 6,
      borderRadius: '50%',
      backgroundColor: 'currentColor',
      flexShrink: 0
    }}
  />
)

const StatusChip = ({ status, label, size = 'small', overrides }: StatusChipProps) => (
  <Chip
    size={size}
    variant='tonal'
    color={resolveStatusColor(status, overrides)}
    label={
      <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, lineHeight: 1 }}>
        <Dot />
        {label ?? formatStatus(status)}
      </span>
    }
  />
)

export default StatusChip
