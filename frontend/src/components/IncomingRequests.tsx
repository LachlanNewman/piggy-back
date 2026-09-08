'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { backendClient, type IncomingRequest } from '@/lib/api/client'
import { initials } from '@/lib/format'

interface Props {
  sub: string
  pollIntervalMs: number
}

export default function IncomingRequests({ sub, pollIntervalMs }: Props) {
  const [requests, setRequests] = useState<IncomingRequest[]>([])
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchIncoming = useCallback(() => {
    backendClient.getIncomingRequests(sub)
      .then(data => setRequests(Array.isArray(data) ? data : []))
      .catch(() => {})
  }, [sub])

  useEffect(() => {
    fetchIncoming()
    intervalRef.current = setInterval(fetchIncoming, pollIntervalMs)
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [fetchIncoming, pollIntervalMs])

  function handleAccept(id: string) {
    backendClient.acceptRideRequest(id, sub)
      .then(() => setRequests(prev => prev.filter(r => r.id !== id)))
      .catch(() => {})
  }

  function handleDecline(id: string) {
    backendClient.declineRideRequest(id, sub)
      .then(() => setRequests(prev => prev.filter(r => r.id !== id)))
      .catch(() => {})
  }

  if (requests.length === 0) return null

  return (
    <div className="card">
      <div className="card-head">
        <h2 className="card-title">Someone wants a lift</h2>
        <span className="pill">{requests.length} waiting</span>
      </div>

      {requests.map(rr => (
        <div key={rr.id} className="request">
          <div className="person" style={{ padding: 0, borderTop: 'none' }}>
            <div className="avatar" aria-hidden="true">
              {initials(rr.rider_first_name, rr.rider_last_name)}
            </div>
            <div className="person-body">
              <div className="person-name">{rr.rider_first_name} {rr.rider_last_name}</div>
              <p className="tiny">Wants a piggyback</p>
            </div>
          </div>

          <div className="trip">
            <div className="trip-leg">
              <span className="trip-dot" />
              <span className="trip-leg-label">Pickup</span>
              <span>{rr.pickup_address}</span>
            </div>
            <div className="trip-leg">
              <span className="trip-dot is-dropoff" />
              <span className="trip-leg-label">Drop-off</span>
              <span>{rr.dropoff_address}</span>
            </div>
          </div>

          <div className="btn-row">
            <button className="btn btn-decline" onClick={() => handleDecline(rr.id)}>Decline</button>
            <button className="btn btn-accept" onClick={() => handleAccept(rr.id)}>Accept</button>
          </div>
        </div>
      ))}
    </div>
  )
}
