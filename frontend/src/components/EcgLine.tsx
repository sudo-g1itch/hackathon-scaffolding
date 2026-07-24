// A decorative ECG heartbeat trace component. SVG line inheriting stroke currentColor.
import type { CSSProperties } from 'react'

type EcgLineProps = {

  /** Continuously sweep a lit pulse along the trace (live animation). */
  animate?: boolean
  className?: string
  style?: CSSProperties

  /** Stroke opacity of the full baseline trace (0–1). */
  baseOpacity?: number

  /** Stroke opacity of the sweeping pulse highlight (0–1; only when `animate`). */
  pulseOpacity?: number
}

// Two QRS complexes over a flat baseline (viewBox 600×100, baseline y=60).
const ECG_PATH =
  'M0 60 H86 Q94 52 102 60 H128 L140 16 L154 88 L164 48 L172 60 H266 ' +
  'Q274 52 282 60 H308 L320 16 L334 88 L344 48 L352 60 H446 Q454 52 462 60 H600'

const EcgLine = ({ animate = false, className, style, baseOpacity = 0.5, pulseOpacity = 1 }: EcgLineProps) => (
  <svg
    viewBox='0 0 600 100'
    preserveAspectRatio='none'
    aria-hidden='true'
    focusable='false'
    className={className}
    style={style}
  >
    <path
      d={ECG_PATH}
      fill='none'
      stroke='currentColor'
      strokeWidth={2}
      strokeLinecap='round'
      strokeLinejoin='round'
      vectorEffect='non-scaling-stroke'
      strokeOpacity={baseOpacity}
    />
    {animate ? (
      <path
        d={ECG_PATH}
        fill='none'
        stroke='currentColor'
        strokeWidth={2}
        strokeLinecap='round'
        strokeLinejoin='round'
        vectorEffect='non-scaling-stroke'
        pathLength={1}
        strokeOpacity={pulseOpacity}
        className='ecg-pulse'
      />
    ) : null}
  </svg>
)

export default EcgLine
