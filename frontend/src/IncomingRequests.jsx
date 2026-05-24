import { useState, useEffect, useRef } from 'react'

export default function IncomingRequests({ sub, pollIntervalMs }) {
  const [requests, setRequests] = useState([])
  const intervalRef = useRef(null)

  function fetchIncoming() {
    fetch(`/api/v1/ride-requests/incoming?sub=${encodeURIComponent(sub)}`)
      .then(r => r.ok ? r.json() : [])
      .then(data => setRequests(Array.isArray(data) ? data : []))
      .catch(() => {})
  }

  useEffect(() => {
    fetchIncoming()
    intervalRef.current = setInterval(fetchIncoming, pollIntervalMs)
    return () => clearInterval(intervalRef.current)
  }, [sub, pollIntervalMs])

  function handleAccept(id) {
    fetch(`/api/v1/ride-requests/${id}/accept?sub=${encodeURIComponent(sub)}`, { method: 'PATCH' })
      .then(() => setRequests(prev => prev.filter(r => r.id !== id)))
      .catch(() => {})
  }

  function handleDecline(id) {
    fetch(`/api/v1/ride-requests/${id}/decline?sub=${encodeURIComponent(sub)}`, { method: 'PATCH' })
      .then(() => setRequests(prev => prev.filter(r => r.id !== id)))
      .catch(() => {})
  }

  if (requests.length === 0) return null

  return (
    <div style={{ border: '1px solid #f0a', borderRadius: 6, padding: '12px 16px', marginBottom: 24 }}>
      <h3 style={{ margin: '0 0 12px' }}>Incoming Ride Requests</h3>
      {requests.map(rr => (
        <div key={rr.id} style={{ marginBottom: 12, paddingBottom: 12, borderBottom: '1px solid #eee' }}>
          <p style={{ margin: '0 0 4px' }}>
            <strong>From:</strong> {rr.rider_first_name} {rr.rider_last_name}
          </p>
          <p style={{ margin: '0 0 4px' }}>
            <strong>Pickup:</strong> {rr.pickup_address}
          </p>
          <p style={{ margin: '0 0 8px' }}>
            <strong>Dropoff:</strong> {rr.dropoff_address}
          </p>
          <button onClick={() => handleAccept(rr.id)} style={{ marginRight: 8 }}>Accept</button>
          <button onClick={() => handleDecline(rr.id)}>Decline</button>
        </div>
      ))}
    </div>
  )
}
