'use client'

import { useCallback, useEffect, useState } from 'react'

export type AlertPermission = 'unsupported' | 'default' | 'granted' | 'denied'

/**
 * Opt-in browser notifications.
 *
 * Permission is only ever requested from a click — browsers reject (and users
 * resent) prompts fired on page load.
 */
export function useAlerts(): {
  permission: AlertPermission
  request: () => void
  notify: (title: string, body: string, tag?: string) => void
} {
  const [permission, setPermission] = useState<AlertPermission>('unsupported')

  useEffect(() => {
    if (typeof Notification === 'undefined') return
    setPermission(Notification.permission as AlertPermission)
  }, [])

  const request = useCallback(() => {
    if (typeof Notification === 'undefined') return
    Notification.requestPermission().then(p => setPermission(p as AlertPermission))
  }, [])

  const notify = useCallback(
    (title: string, body: string, tag?: string) => {
      if (typeof Notification === 'undefined' || permission !== 'granted') return
      // Only interrupt when the tab isn't already in front of the user; the UI
      // updates live either way.
      if (typeof document !== 'undefined' && document.visibilityState === 'visible') return
      try {
        // `tag` collapses repeats for the same ride request into one alert.
        new Notification(title, { body, tag, icon: '/icon.svg' })
      } catch {
        // Some browsers only allow notifications from a service worker.
      }
    },
    [permission]
  )

  return { permission, request, notify }
}
