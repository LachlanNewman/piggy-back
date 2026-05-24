import { useState, useEffect, useRef } from 'react'

export default function RideRequestForm({ sub }) {
  const [pickup, setPickup] = useState('')
  const [dropoff, setDropoff] = useState('')
  const [error, setError] = useState(null)
  const [requestId, setRequestId] = useState(null)
  const [accepted, setAccepted] = useState(false)
  const intervalRef = useRef(null)

  useEffect(() => {
    if (!requestId) return
    intervalRef.current = setInterval(() => {
      fetch(`/api/v1/ride-requests/${requestId}`)
        .then(r => r.json())
        .then(data => {
          if (data.status === 'accepted') {
            setAccepted(true)
            clearInterval(intervalRef.current)
          }
        })
        .catch(() => {})
    }, 3000)
    return () => clearInterval(intervalRef.current)
  }, [requestId])

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
    fetch(`/api/v1/ride-requests?sub=${encodeURIComponent(sub)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pickup_address: pickup, dropoff_address: dropoff }),
    })
      .then(r => {
        if (!r.ok) throw new Error('request failed')
        return r.json()
      })
      .then(data => setRequestId(data.id))
      .catch(() => setError('Failed to submit request. Please try again.'))
  }

  if (accepted) {
    return <p>Your driver is on the way!</p>
  }

  if (requestId) {
    return <p>Waiting for a driver...</p>
  }

  return (
    <form onSubmit={handleSubmit}>
      <h2>Request a Ride</h2>
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
      <button type="submit" style={{ marginTop: 12 }}>Request Ride</button>
    </form>
  )
}
