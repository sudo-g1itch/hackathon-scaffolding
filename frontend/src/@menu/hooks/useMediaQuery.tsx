'use client'

// React Imports
import { useEffect, useState } from 'react'

const useMediaQuery = (breakpoint?: string): boolean => {
  // States
  const [matches, setMatches] = useState(breakpoint === 'always')

  useEffect(() => {
    if (!breakpoint || breakpoint === 'always') return

    const media = window.matchMedia(`(max-width: ${breakpoint})`)

    if (media.matches !== matches) {
      setMatches(media.matches)
    }

    const listener = (event: MediaQueryListEvent) => setMatches(event.matches)

    media.addEventListener('change', listener)

    return () => media.removeEventListener('change', listener)
  }, [matches, breakpoint])

  return matches
}

export default useMediaQuery
