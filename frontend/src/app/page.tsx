'use client'

import { useState, useEffect, useRef } from 'react'
import { useAuth } from 'react-oidc-context'
import ProfileCompletionForm from '@/components/ProfileCompletionForm'
import NearbyCarriers, { type Carrier } from '@/components/NearbyCarriers'
import RideRequestFlow from '@/components/RideRequestFlow'
import IncomingRequests from '@/components/IncomingRequests'
import { backendClient } from '@/lib/api/client'
import { hasRefreshCookie } from '@/lib/auth/splitTokenStore'

const POLL_INTERVAL_MS = 30_000

type ProfileStatus = 'idle' | 'loading' | 'incomplete' | 'complete'

export default function HomePage() {
  const { isAuthenticated, isLoading, user, signinRedirect, signoutRedirect, signinSilent } = useAuth()
  const [restoringSession, setRestoringSession] = useState(false)
  const sessionRestoreAttempted = useRef(false)
  const [profileStatus, setProfileStatus] = useState<ProfileStatus>('idle')
  const [locationDenied, setLocationDenied] = useState(false)
  const [selectedCarrier, setSelectedCarrier] = useState<Carrier | null>(null)

  // Restore a session from the refresh-token cookie after a hard reload.
  useEffect(() => {
    if (isLoading || sessionRestoreAttempted.current) return
    if (!isAuthenticated && hasRefreshCookie()) {
      sessionRestoreAttempted.current = true
      setRestoringSession(true)
      signinSilent().finally(() => setRestoringSession(false))
    }
  }, [isLoading, isAuthenticated, signinSilent])

  useEffect(() => {
    if (!isAuthenticated || isLoading || restoringSession || !user) return
    setProfileStatus('loading')
    backendClient.getUserMe(user.profile.sub)
      .then(data => setProfileStatus(data?.profile_complete ? 'complete' : 'incomplete'))
      .catch(() => setProfileStatus('incomplete'))
  }, [isAuthenticated, isLoading, restoringSession, user])

  // Share location so nearby carriers (and riders) can find each other.
  useEffect(() => {
    if (!isAuthenticated || profileStatus !== 'complete' || !user) return
    if (!navigator.geolocation) return

    const sub = user.profile.sub

    function pushLocation() {
      navigator.geolocation.getCurrentPosition(
        pos => {
          backendClient.pushLocation(sub, pos.coords.latitude, pos.coords.longitude).catch(() => {})
          setLocationDenied(false)
        },
        () => setLocationDenied(true)
      )
    }

    pushLocation()
    const interval = setInterval(pushLocation, POLL_INTERVAL_MS)
    return () => clearInterval(interval)
  }, [isAuthenticated, profileStatus, user])

  function refreshProfileStatus() {
    if (!user) return
    setProfileStatus('loading')
    backendClient.getUserMe(user.profile.sub)
      .then(data => setProfileStatus(data?.profile_complete ? 'complete' : 'incomplete'))
      .catch(() => setProfileStatus('incomplete'))
  }

  if (isLoading || restoringSession || (isAuthenticated && profileStatus === 'loading')) {
    return (
      <main className="shell">
        <Brand />
        <div className="card status-card">
          <div className="status-emoji">🐷</div>
          <p className="muted">Saddling up<span className="dots" /></p>
        </div>
      </main>
    )
  }

  if (!isAuthenticated) {
    return (
      <main className="shell">
        <Brand />
        <div className="hero">
          <div className="hero-emoji">🐷</div>
          <h1>Piggyback rides, on demand</h1>
          <p className="muted">
            Find someone nearby with a free back, tell them where you&apos;re going, and hop on.
          </p>
          <button className="btn btn-block" style={{ marginTop: 22 }} onClick={() => signinRedirect()}>
            Log in to ride
          </button>
        </div>
        <p className="tiny" style={{ textAlign: 'center' }}>
          Every rider is also a carrier. Someone carries you, you carry someone.
        </p>
      </main>
    )
  }

  if (profileStatus === 'incomplete') {
    return (
      <main className="shell">
        <Brand />
        <ProfileCompletionForm onComplete={refreshProfileStatus} />
      </main>
    )
  }

  const sub = user!.profile.sub

  return (
    <main className="shell">
      <div className="topbar">
        <Brand bare />
        <button className="btn btn-ghost" onClick={() => signoutRedirect()}>Log out</button>
      </div>

      <p className="tiny">
        Riding as <strong>{user?.profile?.email ?? sub}</strong>
      </p>

      {locationDenied && (
        <p className="banner banner-info">
          Location is off, so nobody can find you. Enable location access in your browser to ride or carry.
        </p>
      )}

      <IncomingRequests sub={sub} pollIntervalMs={POLL_INTERVAL_MS} />

      {selectedCarrier ? (
        <RideRequestFlow
          sub={sub}
          carrier={selectedCarrier}
          pollIntervalMs={POLL_INTERVAL_MS}
          onDone={() => setSelectedCarrier(null)}
        />
      ) : (
        <NearbyCarriers sub={sub} onRequestRide={setSelectedCarrier} />
      )}
    </main>
  )
}

function Brand({ bare = false }: { bare?: boolean }) {
  const brand = (
    <h1 className="brand">
      <span className="brand-mark" aria-hidden="true">🐷</span>
      Piggy Back
    </h1>
  )
  return bare ? brand : <div className="topbar">{brand}</div>
}
