'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { backendClient, ApiError, type NearbyUser } from '@/lib/api/client'
import Avatar from './Avatar'
import Radar from './Radar'
import RefreshButton from './RefreshButton'

export interface Carrier {
  sub: string
  name: string
}

interface Props {
  sub: string
  /** Flips true once our own location has reached the backend. Nearby lookups
   *  404 until then, so the first successful push is the cue to try again. */
  locationReady: boolean
  onRequestRide: (carrier: Carrier) => void
}

type ErrorKind = 'location' | 'fetch'

/** Who is nearby changes as people move, and there is no event for it — the
 *  backend only knows when someone next pushes their location. */
const REFRESH_MS = 10_000

export default function NearbyCarriers({ sub, locationReady, onRequestRide }: Props) {
  const [carriers, setCarriers] = useState<NearbyUser[] | null>(null)
  const [error, setError] = useState<ErrorKind | null>(null)
  const [loading, setLoading] = useState(true)
  const loaded = useRef(false)

  const fetchNearby = useCallback((background = false) => {
    // Background refreshes must not flash a spinner over a list the user is
    // reading.
    if (!background) setLoading(true)
    return backendClient.getNearbyUsers(sub)
      .then(data => { setCarriers(data); setError(null) })
      .catch(err => {
        if (!background || !loaded.current) {
          setError(err instanceof ApiError && err.status === 404 ? 'location' : 'fetch')
        }
      })
      .finally(() => {
        loaded.current = true
        if (!background) setLoading(false)
      })
  }, [sub])

  useEffect(() => {
    fetchNearby()
    const id = setInterval(() => fetchNearby(true), REFRESH_MS)
    return () => clearInterval(id)
  }, [fetchNearby])

  // Retry the moment our location lands rather than leaving the user staring
  // at "we need your location" until the next refresh tick.
  useEffect(() => {
    if (locationReady) fetchNearby(true)
  }, [locationReady, fetchNearby])

  return (
    <div className="card">
      <div className="card-head">
        <h2 className="card-title">Carriers nearby</h2>
        <div className="card-head-actions">
          {carriers && carriers.length > 0 && (
            <span className="pill">{carriers.length} in range</span>
          )}
          <RefreshButton busy={loading} onClick={() => fetchNearby()} />
        </div>
      </div>

      {loading && (
        <>
          <Radar />
          <p className="empty">Scanning for free backs nearby<span className="dots" /></p>
        </>
      )}

      {!loading && error === 'location' && (
        <p className="banner banner-info">
          We need your location to find carriers around you. Allow location access, then refresh.
        </p>
      )}

      {!loading && error === 'fetch' && (
        <p className="banner banner-error">Could not load nearby carriers. Try again in a moment.</p>
      )}

      {!loading && !error && carriers?.length === 0 && (
        <>
          <Radar />
          <p className="empty">No backs in range. Stay put — carriers appear here the moment they come online.</p>
        </>
      )}

      {!loading && !error && carriers && carriers.length > 0 && (
        <div className="people">
          {carriers.map(c => {
            const name = `${c.first_name} ${c.last_name}`.trim()
            return (
              <div key={c.id} className="person">
                <Avatar firstName={c.first_name} lastName={c.last_name} />
                <div className="person-body">
                  <div className="person-name">{name}</div>
                  <p className="tiny">Ready to carry</p>
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
