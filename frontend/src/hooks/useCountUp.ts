'use client'

// Animates a number from its previous value to `target` with an ease-out
// requestAnimationFrame loop (no dependencies). Returns `target` immediately
// when disabled or when the user prefers reduced motion.
import { useEffect, useRef, useState } from 'react'

type UseCountUpOptions = {

  /** Animation length in ms. */
  duration?: number

  /** Disable to render the target instantly (e.g. for string values). */
  enabled?: boolean
}

const prefersReducedMotion = () =>
  typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches

const useCountUp = (target: number, options?: UseCountUpOptions): number => {
  const { duration = 900, enabled = true } = options ?? {}

  const [value, setValue] = useState(enabled ? 0 : target)
  const valueRef = useRef(value)

  valueRef.current = value

  useEffect(() => {
    if (!enabled || !Number.isFinite(target) || prefersReducedMotion()) {
      setValue(target)

      return
    }

    const from = valueRef.current

    if (from === target) return

    let raf = 0
    const start = performance.now()

    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / duration)
      const eased = 1 - Math.pow(1 - t, 3)

      setValue(t < 1 ? from + (target - from) * eased : target)
      if (t < 1) raf = requestAnimationFrame(tick)
    }

    raf = requestAnimationFrame(tick)

    return () => cancelAnimationFrame(raf)
  }, [target, duration, enabled])

  return value
}

export default useCountUp
