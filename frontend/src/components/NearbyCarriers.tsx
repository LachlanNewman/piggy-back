'use client'

import { useState, useEffect, useCallback } from 'react'
import { backendClient, ApiError, type NearbyUser } from '@/lib/api/client'
import { initials } from '@/lib/format'

export interface Carrier {
  sub: string
  name: string
}

interface Props {
  sub: string
  onRequestRide: (carrier: Carrier) => void
}

type ErrorKind = 'location' | 'fetch'

export default function NearbyCarriers({ sub, onRequestRide }: Props) {
  const [carriers, setCarriers] = useState<NearbyUser[] | null>(null)
  const [error, setError] = useState<ErrorKind | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchNearby = useCallback(() => {
    setLoading(true)
    backendClient.getNearbyUsers(sub)
      .then(data => { setCarriers(data); setError(null) })
      .catch(err => setError(err instanceof ApiError && err.status === 404 ? 'location' : 'fetch'))
      .finally(() => setLoading(false))
  }, [sub])

  useEffect(() => {
    fetchNearby()
  }, [fetchNearby])

  return (
    <div className="card">
      <div className="card-head">
        <h2 className="card-title">Carriers nearby</h2>
        <button className="btn btn-ghost" onClick={fetchNearby} disabled={loading}>
          {loading ? 'Looking…' : 'Refresh'}
        </button>
      </div>

      {loading && <p className="empty">Looking for backs nearby<span className="dots" /></p>}

      {!loading && error === 'location' && (
        <p className="banner banner-info">
          We need your location to find carriers around you. Allow location access, then refresh.
        </p>
      )}

      {!loading && error === 'fetch' && (
        <p className="banner banner-error">Could not load nearby carriers. Try again in a moment.</p>
      )}

      {!loading && !error && carriers?.length === 0 && (
        <p className="empty">Nobody around right now. Stay put — carriers appear as they come online.</p>
      )}

      {!loading && !error && carriers && carriers.length > 0 && (
        <div className="people">
          {carriers.map(c => {
            const name = `${c.first_name} ${c.last_name}`.trim()
            return (
              <div key={c.id} className="person">
                <div className="avatar" aria-hidden="true">{initials(c.first_name, c.last_name)}</div>
                <div className="person-body">
                  <div className="person-name">{name}</div>
                  <p className="tiny">Available for a lift</p>
                </div>
                <button className="btn" onClick={() => onRequestRide({ sub: c.auth_subject, name })}>
                  Request
                </button>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
