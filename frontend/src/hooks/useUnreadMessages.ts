'use client'

// How many caregiver messages are waiting for the signed-in user, whichever
// side of the conversation they are on. The navigation badge reads this, so it
// deliberately lives outside AnchorOneContext — the sidebar renders on admin
// screens too, where that provider is not mounted.
import { useCallback, useEffect, useState } from 'react'

import { useAuth } from '@/contexts/AuthContext'
import { anchorOneService } from '@/services/anchorOneService'

const POLL_INTERVAL_MS = 30_000

const useUnreadMessages = () => {
  const { isAuthenticated } = useAuth()
  const [unread, setUnread] = useState(0)

  const refresh = useCallback(async () => {
    if (!isAuthenticated) {
      setUnread(0)

      return
    }

    try {
      setUnread(await anchorOneService.getUnreadCount())
    } catch {
      // A badge is decoration. If the count cannot be fetched the menu still
      // works, and there is nothing useful to tell the user about it.
    }
  }, [isAuthenticated])

  useEffect(() => {
    void refresh()

    if (!isAuthenticated) return undefined

    const interval = setInterval(() => void refresh(), POLL_INTERVAL_MS)

    return () => clearInterval(interval)
  }, [refresh, isAuthenticated])

  return { unread, refresh }
}

export default useUnreadMessages
