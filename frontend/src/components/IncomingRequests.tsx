'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { backendClient, type IncomingRequest } from '@/lib/api/client'
import { useNow, msUntil, formatCountdown } from '@/lib/time'
import Avatar from './Avatar'
import { useRideEvents, type ConnectionState, type RideEvent } from '@/lib/realtime'

interface Props {
  sub: string
  subscribe: (listener: (e: RideEvent) => void) => () => void
  connection: ConnectionState
  onNewRequest: (request: IncomingRequest) => void
}

/** Safety net for when the live stream is down. While it is up, the server
 *  tells us the instant anything changes, so we never poll. */
const FALLBACK_POLL_MS = 15_000

export default function IncomingRequests({ sub, subscribe, connection, onNewRequest }: Props) {
  const [requests, setRequests] = useState<IncomingRequest[]>([])
  const [acting, setActing] = useState<string | null>(null)
  const seen = useRef(new Set<string>())
  const now = useNow()

  // Held in a ref so a new callback identity from the parent cannot invalidate
  // fetchIncoming and set off a cascade of refetches.
  const notifyNew = useRef(onNewRequest)
  notifyNew.current = onNewRequest

  const fetchIncoming = useCallback(() => {
    backendClient.getIncomingRequests(sub)
      .then(data => {
        const list = Array.isArray(data) ? data : []
        // Anything we have not shown before is worth an alert.
        for (const r of list) {
          if (!seen.current.has(r.id)) {
            seen.current.add(r.id)
            notifyNew.current(r)
          }
        }
        setRequests(list)
      })
      .catch(() => {})
  }, [sub])

  useEffect(() => {
    fetchIncoming()
  }, [fetchIncoming])

  // The live path: a carrier sees a request the moment the rider sends it.
  useRideEvents(subscribe, event => {
    if (event.type === 'ride_request.created') fetchIncoming()
  })

  // The fallback path: only poll while the stream is not carrying events.
  useEffect(() => {
    if (connection === 'live') return
    const id = setInterval(fetchIncoming, FALLBACK_POLL_MS)
    return () => clearInterval(id)
  }, [connection, fetchIncoming])

  function respond(id: string, action: 'accept' | 'decline') {
    setActing(id)
    const call = action === 'accept'
      ? backendClient.acceptRideRequest(id, sub)
      : backendClient.declineRideRequest(id, sub)
    call
      .then(() => setRequests(prev => prev.filter(r => r.id !== id)))
      .catch(() => {})
      .finally(() => setActing(null))
  }

  // Expiry is a deadline the server already told us, so it is enforced here on
  // the clock rather than discovered by polling.
  const live = requests.filter(r => msUntil(r.expires_at, now) > 0)

  if (live.length === 0) return null

  return (
    <div className="card card-accent">
      <div className="card-head">
        <h2 className="card-title">Someone wants a lift</h2>
        <span className="pill">{live.length} waiting</span>
      </div>

      {live.map(rr => {
        const left = msUntil(rr.expires_at, now)
        const total = Math.max(1, Date.parse(rr.expires_at) - Date.parse(rr.requested_at))
        const urgent = left < total * 0.25
        return (
          <div key={rr.id} className="request">
            <div className="person" style={{ padding: 0, borderTop: 'none' }}>
              <Avatar
                firstName={rr.rider_first_name}
                lastName={rr.rider_last_name}
                remaining={left / total}
              />
              <div className="person-body">
                <div className="person-name">{rr.rider_first_name} {rr.rider_last_name}</div>
                <p className="tiny" style={urgent ? { color: 'var(--red)' } : undefined}>
                  {formatCountdown(left)} left to answer
                </p>
              </div>
            </div>

            <div className="trip">
              <div className="trip-leg">
                <span className="trip-rail"><span className="trip-dot" /></span>
                <span className="trip-text">
                  <span className="trip-label">Pickup</span>
                  <span className="trip-value">{rr.pickup_address}</span>
                </span>
              </div>
              <div className="trip-leg">
                <span className="trip-rail"><span className="trip-dot is-dropoff" /></span>
                <span className="trip-text">
                  <span className="trip-label">Drop-off</span>
                  <span className="trip-value">{rr.dropoff_address}</span>
                </span>
              </div>
            </div>

            <div className="btn-row">
              <button
                className="btn btn-decline"
                disabled={acting === rr.id}
                onClick={() => respond(rr.id, 'decline')}
              >
                Decline
              </button>
              <button
                className="btn btn-accept"
                disabled={acting === rr.id}
                onClick={() => respond(rr.id, 'accept')}
              >
                {acting === rr.id ? 'Saving…' : 'Accept'}
              </button>
            </div>
          </div>
        )
      })}
    </div>
  )
}
