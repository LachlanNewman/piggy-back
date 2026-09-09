'use client'

import { useState, useEffect, useCallback } from 'react'
import { backendClient, ApiError, type RideRequest } from '@/lib/api/client'
import { useNow, msUntil, formatCountdown } from '@/lib/time'
import PiggybackMark from './PiggybackMark'
import { useRideEvents, type ConnectionState, type RideEvent } from '@/lib/realtime'
import type { Carrier } from './NearbyCarriers'

interface Props {
  sub: string
  carrier: Carrier
  subscribe: (listener: (e: RideEvent) => void) => () => void
  connection: ConnectionState
  onSettled: (status: SettledStatus, carrierName: string) => void
  onDone: () => void
}

type SettledStatus = 'accepted' | 'declined' | 'expired'

/** Safety net only — while the stream is live the server pushes the answer. */
const FALLBACK_POLL_MS = 10_000

export default function RideRequestFlow({
  sub, carrier, subscribe, connection, onSettled, onDone,
}: Props) {
  const [pickup, setPickup] = useState('')
  const [dropoff, setDropoff] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [requestId, setRequestId] = useState<string | null>(null)
  const [request, setRequest] = useState<RideRequest | null>(null)
  const [announced, setAnnounced] = useState(false)
  const now = useNow()

  const refresh = useCallback(() => {
    if (!requestId) return
    backendClient.getRideRequest(requestId).then(setRequest).catch(() => {})
  }, [requestId])

  useEffect(() => {
    refresh()
  }, [refresh])

  // The live path: the carrier's answer lands here as it happens.
  useRideEvents(subscribe, event => {
    if (event.type === 'ride_request.updated' && event.id === requestId) refresh()
  })

  useEffect(() => {
    if (connection === 'live' || !requestId) return
    const id = setInterval(refresh, FALLBACK_POLL_MS)
    return () => clearInterval(id)
  }, [connection, requestId, refresh])

  // The server marks a request expired lazily, on read. The deadline is known
  // up front, so the countdown settles it on the exact second instead.
  const remaining = request ? msUntil(request.expires_at, now) : 0
  const totalMs = request
    ? Math.max(1, Date.parse(request.expires_at) - Date.parse(request.requested_at))
    : 1
  const status: SettledStatus | 'pending' | null = !request
    ? null
    : request.status === 'accepted' || request.status === 'declined' || request.status === 'expired'
      ? request.status
      : remaining > 0
        ? 'pending'
        : 'expired'

  useEffect(() => {
    if (!announced && status && status !== 'pending') {
      setAnnounced(true)
      onSettled(status, carrier.name)
    }
  }, [status, announced, onSettled, carrier.name])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!pickup.trim()) {
      setError('Where should they pick you up?')
      return
    }
    if (!dropoff.trim()) {
      setError('Where are you being carried to?')
      return
    }
    setError(null)
    setSubmitting(true)
    backendClient.createRideRequest(sub, { pickupAddress: pickup, dropoffAddress: dropoff, driverID: carrier.sub })
      .then(data => setRequestId(data.id))
      .catch(err => {
        if (err instanceof ApiError && err.status === 409) {
          setError('You already have a request in flight. Wait for it to be answered or expire.')
        } else {
          setError('Could not send the request. Please try again.')
        }
      })
      .finally(() => setSubmitting(false))
  }

  if (status === 'accepted') {
    return (
      <Outcome
        emoji="🎉"
        title={`${carrier.name} is on the way`}
        body="Find them at the pickup point and hop on. Hold the shoulders, not the neck."
        tone="success"
        onDone={onDone}
        doneLabel="Done"
      />
    )
  }

  if (status === 'declined') {
    return (
      <Outcome
        emoji="🙅"
        title="Request declined"
        body={`${carrier.name} can't carry you right now. Try another carrier nearby.`}
        onDone={onDone}
        doneLabel="Back to carriers"
      />
    )
  }

  if (status === 'expired') {
    return (
      <Outcome
        emoji="⌛"
        title="Request timed out"
        body="Nobody answered in time. Send a fresh request when you're ready."
        onDone={onDone}
        doneLabel="Back to carriers"
      />
    )
  }

  if (requestId) {
    return (
      <div className="card card-accent status-card">
        <div className="status-timer">
          <CountdownRing remaining={request ? remaining / totalMs : 1} />
          <div className="status-mark status-mark-accent"><PiggybackMark /></div>
        </div>
        <h2>Waiting on {carrier.name}<span className="dots" /></h2>
        <p className="muted">
          {request
            ? `They have ${formatCountdown(remaining)} to answer.`
            : 'Sending your request…'}
        </p>
        <div className="btn-row">
          <button className="btn btn-secondary btn-block" onClick={onDone}>Cancel</button>
        </div>
      </div>
    )
  }

  return (
    <div className="card">
      <div className="card-head">
        <h2 className="card-title">Ride with {carrier.name}</h2>
        <span className="pill">Piggyback</span>
      </div>

      <form onSubmit={handleSubmit} className="form">
        <label className="field">
          <span>Pickup</span>
          <input
            className="input"
            type="text"
            value={pickup}
            onChange={e => setPickup(e.target.value)}
            placeholder="123 Main St"
          />
        </label>

        <label className="field">
          <span>Drop-off</span>
          <input
            className="input"
            type="text"
            value={dropoff}
            onChange={e => setDropoff(e.target.value)}
            placeholder="456 Oak Ave"
          />
        </label>

        {error && <p className="banner banner-error">{error}</p>}

        <div className="btn-row btn-row-split">
          <button type="button" className="btn btn-secondary" onClick={onDone}>Back</button>
          <button type="submit" className="btn" disabled={submitting}>
            {submitting ? 'Sending…' : 'Request piggyback'}
          </button>
        </div>
      </form>
    </div>
  )
}

function Outcome({
  emoji,
  title,
  body,
  tone,
  onDone,
  doneLabel,
}: {
  emoji: string
  title: string
  body: string
  tone?: 'success'
  onDone: () => void
  doneLabel: string
}) {
  return (
    <div className="card status-card">
      <div className={`status-mark${tone === 'success' ? ' status-mark-success' : ''}`}>{emoji}</div>
      <h2 style={tone === 'success' ? { color: 'var(--green)' } : undefined}>{title}</h2>
      <p className="muted">{body}</p>
      <div className="btn-row">
        <button className="btn btn-block" onClick={onDone}>{doneLabel}</button>
      </div>
    </div>
  )
}

const RING_R = 47
const RING_C = 2 * Math.PI * RING_R

/** Large ring that drains as the request nears expiry. */
function CountdownRing({ remaining }: { remaining: number }) {
  const clamped = Math.min(1, Math.max(0, remaining))
  return (
    <svg className="ring" viewBox="0 0 100 100" aria-hidden="true">
      <circle className="ring-track" cx="50" cy="50" r={RING_R} />
      <circle
        className={`ring-progress${clamped < 0.25 ? ' is-urgent' : ''}`}
        cx="50"
        cy="50"
        r={RING_R}
        strokeDasharray={RING_C}
        strokeDashoffset={RING_C * (1 - clamped)}
      />
    </svg>
  )
}
