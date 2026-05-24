import { useState, useEffect } from 'react'

export default function NearbyUsersList({ sub, onRequestRide }) {
  const [users, setUsers] = useState(null)
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(true)

  function fetchNearby() {
    setLoading(true)
    fetch(`/api/v1/users/nearby?sub=${encodeURIComponent(sub)}`)
      .then(r => {
        if (r.status === 404) return { noLocation: true }
        if (!r.ok) throw new Error('fetch failed')
        return r.json().then(data => ({ users: data }))
      })
      .then(result => {
        if (result.noLocation) {
          setError('location')
        } else {
          setUsers(result.users)
          setError(null)
        }
      })
      .catch(() => setError('fetch'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    fetchNearby()
  }, [sub])

  if (loading) return <p>Finding nearby users...</p>

  if (error === 'location') {
    return (
      <div>
        <h2>Nearby Users</h2>
        <p style={{ color: '#888' }}>
          Location sharing required to find nearby users. Make sure you've granted location permission.
        </p>
        <button onClick={fetchNearby}>Retry</button>
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <h2>Nearby Users</h2>
        <p style={{ color: 'red' }}>Could not load nearby users.</p>
        <button onClick={fetchNearby}>Retry</button>
      </div>
    )
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>Nearby Users</h2>
        <button onClick={fetchNearby}>Refresh</button>
      </div>
      {users && users.length === 0 && (
        <p style={{ color: '#888' }}>No nearby users found.</p>
      )}
      {users && users.map(u => (
        <div key={u.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '8px 0', borderBottom: '1px solid #eee' }}>
          <span>{u.first_name} {u.last_name}</span>
          <button onClick={() => onRequestRide({ sub: u.auth_subject, name: `${u.first_name} ${u.last_name}` })}>
            Request ride
          </button>
        </div>
      ))}
    </div>
  )
}
