'use client'

import { useEffect, useState } from 'react'

/** Current time, re-rendered every `intervalMs`. One timer drives every
 *  countdown and expiry check on the page. */
export function useNow(intervalMs = 1000): number {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])

  return now
}

/** Milliseconds remaining until `iso`, floored at zero. */
export function msUntil(iso: string, now: number): number {
  const at = Date.parse(iso)
  if (Number.isNaN(at)) return 0
  return Math.max(0, at - now)
}

/** Formats a remaining duration as `m:ss`, or `0:00` once elapsed. */
export function formatCountdown(ms: number): string {
  const total = Math.floor(ms / 1000)
  const minutes = Math.floor(total / 60)
  const seconds = total % 60
  return `${minutes}:${String(seconds).padStart(2, '0')}`
}
