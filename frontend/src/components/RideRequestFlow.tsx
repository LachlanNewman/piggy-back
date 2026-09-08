'use client'

import { useState, useEffect, useRef } from 'react'
import { backendClient, ApiError } from '@/lib/api/client'
import type { Carrier } from './NearbyCarriers'

interface Props {
  sub: string
  carrier: Carrier
  pollIntervalMs: number
  onDone: () => void
}

type RideStatus = 'accepted' | 'declined' | 'expired'

const TERMINAL: RideStatus[] = ['accepted', 'declined', 'expired']

export default function RideRequestFlow({ sub, carrier, pollIntervalMs, onDone }: Props) {
  const [pickup, setPickup] = useState('')
  const [dropoff, setDropoff] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [requestId, setRequestId] = useState<string | null>(null)
  const [status, setStatus] = useState<RideStatus | null>(null)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    if (!requestId) return
    intervalRef.current = setInterval(() => {
      backendClient.getRideRequest(requestId)
        .then(data => {
          if (TERMINAL.includes(data.status as RideStatus)) {
            setStatus(data.status as RideStatus)
            if (intervalRef.current) clearInterval(intervalRef.current)
          }
        })
        .catch(() => {})
    }, pollIntervalMs)
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [requestId, pollIntervalMs])

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
      <div className="card status-card">
        <div className="status-emoji">🐷</div>
        <h2>Waiting on {carrier.name}<span className="dots" /></h2>
        <p className="muted">We&apos;ll let you know the moment they accept.</p>
        <div className="btn-row">
          <button className="btn btn-secondary" onClick={onDone}>Cancel</button>
        </div>
      </div>
    )
  }

  return (
    <div className="card">
      <div className="card-head">
        <h2 className="card-title">Ride with {carrier.name}</h2>
        <span className="pill">🐷 Piggyback</span>
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

        <div className="btn-row">
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
      <div className="status-emoji">{emoji}</div>
      <h2 style={tone === 'success' ? { color: 'var(--green)' } : undefined}>{title}</h2>
      <p className="muted">{body}</p>
      <div className="btn-row">
        <button className="btn btn-block" onClick={onDone}>{doneLabel}</button>
      </div>
    </div>
  )
}
