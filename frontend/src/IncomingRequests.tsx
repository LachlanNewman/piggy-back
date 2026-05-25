import { useState, useEffect, useRef } from 'react'
import { backendClient, IncomingRequest } from './api/client'

interface Props {
  sub: string
  pollIntervalMs: number
}

export default function IncomingRequests({ sub, pollIntervalMs }: Props) {
  const [requests, setRequests] = useState<IncomingRequest[]>([])
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  function fetchIncoming() {
    backendClient.getIncomingRequests(sub)
      .then(data => setRequests(Array.isArray(data) ? data : []))
      .catch(() => {})
  }

  useEffect(() => {
    fetchIncoming()
    intervalRef.current = setInterval(fetchIncoming, pollIntervalMs)
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [sub, pollIntervalMs])

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
