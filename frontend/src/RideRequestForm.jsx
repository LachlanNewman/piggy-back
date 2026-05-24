import { useState, useEffect, useRef } from 'react'
import { backendClient, ApiError } from './api/client'

export default function RideRequestForm({ sub, driverID, driverName, pollIntervalMs, onCancel }) {
  const [pickup, setPickup] = useState('')
  const [dropoff, setDropoff] = useState('')
  const [error, setError] = useState(null)
  const [requestId, setRequestId] = useState(null)
  const [status, setStatus] = useState(null) // null | 'accepted' | 'declined' | 'expired'
  const intervalRef = useRef(null)

  useEffect(() => {
    if (!requestId) return
    intervalRef.current = setInterval(() => {
      backendClient.getRideRequest(requestId)
        .then(data => {
          if (['accepted', 'declined', 'expired'].includes(data.status)) {
            setStatus(data.status)
            clearInterval(intervalRef.current)
          }
        })
        .catch(() => {})
    }, pollIntervalMs)
    return () => clearInterval(intervalRef.current)
  }, [requestId, pollIntervalMs])

  function handleSubmit(e) {
    e.preventDefault()
    if (!pickup.trim()) {
      setError('Pickup address is required')
      return
    }
    if (!dropoff.trim()) {
      setError('Dropoff address is required')
      return
    }
    setError(null)
    backendClient.createRideRequest(sub, { pickupAddress: pickup, dropoffAddress: dropoff, driverID })
      .then(data => setRequestId(data.id))
      .catch(err => {
        if (err instanceof ApiError && err.status === 409) {
          setError('You already have an active request. Wait for it to expire before requesting again.')
        } else {
          setError('Failed to submit request. Please try again.')
        }
      })
  }

  if (status === 'accepted') {
    return (
      <div>
        <p style={{ color: 'green' }}>Your driver is on the way!</p>
        <button onClick={onCancel}>Back</button>
      </div>
    )
  }

  if (status === 'declined') {
    return (
      <div>
        <p>Your request was declined. Try another nearby user.</p>
        <button onClick={onCancel}>Back</button>
      </div>
    )
  }

  if (status === 'expired') {
    return (
      <div>
        <p>Your request timed out. Try again.</p>
        <button onClick={onCancel}>Back</button>
      </div>
    )
  }

  if (requestId) {
    return (
      <div>
        <p>Waiting for {driverName ?? 'driver'} to respond...</p>
        <button onClick={onCancel} style={{ marginTop: 8 }}>Cancel</button>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit}>
      <h2>Request a Ride from {driverName ?? 'driver'}</h2>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <div>
        <label>
          Pickup address<br />
          <input
            type="text"
            value={pickup}
            onChange={e => setPickup(e.target.value)}
            placeholder="123 Main St"
          />
        </label>
      </div>
      <div style={{ marginTop: 8 }}>
        <label>
          Dropoff address<br />
          <input
            type="text"
            value={dropoff}
            onChange={e => setDropoff(e.target.value)}
            placeholder="456 Oak Ave"
          />
        </label>
      </div>
      <div style={{ marginTop: 12 }}>
        <button type="submit">Request Ride</button>
        <button type="button" onClick={onCancel} style={{ marginLeft: 8 }}>Cancel</button>
      </div>
    </form>
  )
}
