'use client'

import { useEffect, useRef, useState } from 'react'

/** A change signal from the backend. The payload names what changed; the
 *  component refetches the authoritative data from the REST API. */
export interface RideEvent {
  type: 'connected' | 'ride_request.created' | 'ride_request.updated'
  id?: string
  status?: string
}

export type ConnectionState = 'connecting' | 'live' | 'offline'

type Listener = (event: RideEvent) => void

/**
 * Subscribes to the user's server-sent event stream.
 *
 * `EventSource` reconnects on its own (the server sends a `retry:` hint), so
 * this hook only owns the lifecycle and surfaces the connection state — the
 * components that consume events register listeners via `useRideEvents`.
 */
export function useEventStream(sub: string | null): {
  state: ConnectionState
  subscribe: (listener: Listener) => () => void
} {
  const [state, setState] = useState<ConnectionState>('connecting')
  const listeners = useRef(new Set<Listener>())

  useEffect(() => {
    if (!sub) return

    setState('connecting')
    const source = new EventSource(`/api/v1/events?sub=${encodeURIComponent(sub)}`)

    source.onopen = () => setState('live')

    source.onmessage = e => {
      let event: RideEvent
      try {
        event = JSON.parse(e.data) as RideEvent
      } catch {
        return
      }
      // The handshake frame only confirms the stream is up. Browsers fire
      // `onopen` on connect, but the frame is the server's own confirmation.
      if (event.type === 'connected') {
        setState('live')
        return
      }
      for (const listener of listeners.current) listener(event)
    }

    // EventSource retries by itself; the error just means we are between
    // attempts, so report offline rather than tearing the stream down.
    source.onerror = () => setState(source.readyState === EventSource.CLOSED ? 'offline' : 'connecting')

    return () => {
      source.close()
      setState('offline')
    }
  }, [sub])

  const subscribeRef = useRef((listener: Listener) => {
    listeners.current.add(listener)
    return () => {
      listeners.current.delete(listener)
    }
  })

  return { state, subscribe: subscribeRef.current }
}

/** Runs `onEvent` for every event on the stream. The callback is kept in a ref
 *  so callers need not memoise it. */
export function useRideEvents(
  subscribe: (listener: Listener) => () => void,
  onEvent: Listener
): void {
  const handler = useRef(onEvent)
  handler.current = onEvent

  useEffect(() => subscribe(e => handler.current(e)), [subscribe])
}
