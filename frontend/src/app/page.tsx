'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useAuth } from 'react-oidc-context'
import ProfileCompletionForm from '@/components/ProfileCompletionForm'
import NearbyCarriers, { type Carrier } from '@/components/NearbyCarriers'
import RideRequestFlow from '@/components/RideRequestFlow'
import IncomingRequests from '@/components/IncomingRequests'
import { backendClient, type IncomingRequest } from '@/lib/api/client'
import { hasRefreshCookie } from '@/lib/auth/splitTokenStore'
import { useEventStream, type ConnectionState } from '@/lib/realtime'
import { useAlerts } from '@/lib/notifications'
import PiggybackMark from '@/components/PiggybackMark'

/** Location is pushed on movement, but never more often than this — the
 *  browser can fire watchPosition many times a second. */
const LOCATION_THROTTLE_MS = 10_000

type ProfileStatus = 'idle' | 'loading' | 'incomplete' | 'complete'

export default function HomePage() {
  const { isAuthenticated, isLoading, user, signinRedirect, signoutRedirect, signinSilent } = useAuth()
  const [restoringSession, setRestoringSession] = useState(false)
  const sessionRestoreAttempted = useRef(false)
  const [profileStatus, setProfileStatus] = useState<ProfileStatus>('idle')
  const [locationDenied, setLocationDenied] = useState(false)
  const [locationReady, setLocationReady] = useState(false)
  const [selectedCarrier, setSelectedCarrier] = useState<Carrier | null>(null)

  const sub = user?.profile.sub ?? null
  const ready = isAuthenticated && profileStatus === 'complete' && !!sub

  const { state: connection, subscribe } = useEventStream(ready ? sub : null)
  const { permission, request: requestAlerts, notify } = useAlerts()

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
    if (!isAuthenticated || isLoading || restoringSession || !sub) return
    setProfileStatus('loading')
    backendClient.getUserMe(sub)
      .then(data => setProfileStatus(data?.profile_complete ? 'complete' : 'incomplete'))
      .catch(() => setProfileStatus('incomplete'))
  }, [isAuthenticated, isLoading, restoringSession, sub])

  // Share location continuously so the nearby list reflects where people
  // actually are. watchPosition lets the device push updates as they happen
  // rather than waking the GPS on a fixed timer.
  useEffect(() => {
    if (!ready || !sub || !navigator.geolocation) return

    let lastPush = 0
    const watchId = navigator.geolocation.watchPosition(
      pos => {
        setLocationDenied(false)
        const now = Date.now()
        if (now - lastPush < LOCATION_THROTTLE_MS) return
        lastPush = now
        backendClient.pushLocation(sub, pos.coords.latitude, pos.coords.longitude)
          .then(() => setLocationReady(true))
          .catch(() => {})
      },
      () => setLocationDenied(true),
      { enableHighAccuracy: true, maximumAge: 10_000, timeout: 20_000 }
    )

    return () => {
      navigator.geolocation.clearWatch(watchId)
      setLocationReady(false)
    }
  }, [ready, sub])

  const handleNewRequest = useCallback((rr: IncomingRequest) => {
    notify(
      'Someone wants a piggyback',
      `${rr.rider_first_name} ${rr.rider_last_name} — ${rr.pickup_address} to ${rr.dropoff_address}`,
      rr.id
    )
  }, [notify])

  const handleSettled = useCallback((status: 'accepted' | 'declined' | 'expired', carrierName: string) => {
    const message = {
      accepted: `${carrierName} accepted — they're on the way.`,
      declined: `${carrierName} can't carry you right now.`,
      expired: 'Your request timed out.',
    }[status]
    notify('Piggy Back', message, 'ride-status')
  }, [notify])

  function refreshProfileStatus() {
    if (!sub) return
    setProfileStatus('loading')
    backendClient.getUserMe(sub)
      .then(data => setProfileStatus(data?.profile_complete ? 'complete' : 'incomplete'))
      .catch(() => setProfileStatus('incomplete'))
  }

  if (isLoading || restoringSession || (isAuthenticated && profileStatus === 'loading')) {
    return (
      <main className="shell">
        <Brand />
        <div className="card status-card">
          <div className="status-mark status-mark-accent"><PiggybackMark /></div>
          <p className="muted" style={{ marginTop: 16 }}>Saddling up<span className="dots" /></p>
        </div>
      </main>
    )
  }

  if (!isAuthenticated) {
    return (
      <main className="shell">
        <Brand />
        <div className="hero">
          <div className="hero-mark"><PiggybackMark /></div>
          <h1>Piggyback rides,<br />on demand</h1>
          <p className="muted">
            Find someone nearby with a free back, tell them where you&apos;re going, and hop on.
          </p>
          <button className="btn btn-block" style={{ marginTop: 26 }} onClick={() => signinRedirect()}>
            Log in to ride
          </button>
        </div>
        <p className="section-note">
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

  return (
    <main className="shell">
      <div className="topbar">
        <Brand bare />
        <div className="topbar-actions">
          <LiveIndicator state={connection} />
          <button className="btn btn-ghost" onClick={() => signoutRedirect()}>Log out</button>
        </div>
      </div>

      <p className="tiny">
        Riding as <strong>{user?.profile?.email ?? sub}</strong>
      </p>

      {locationDenied && (
        <p className="banner banner-info">
          Location is off, so nobody can find you. Enable location access in your browser to ride or carry.
        </p>
      )}

      {permission === 'default' && (
        <div className="banner banner-info nudge">
          <span>Get alerted when someone asks you for a lift.</span>
          <button className="btn btn-ghost" onClick={requestAlerts}>Turn on alerts</button>
        </div>
      )}

      <IncomingRequests
        sub={sub!}
        subscribe={subscribe}
        connection={connection}
        onNewRequest={handleNewRequest}
      />

      {selectedCarrier ? (
        <RideRequestFlow
          // Remount per carrier so a finished ride never leaks its outcome
          // into the next request.
          key={selectedCarrier.sub}
          sub={sub!}
          carrier={selectedCarrier}
          subscribe={subscribe}
          connection={connection}
          onSettled={handleSettled}
          onDone={() => setSelectedCarrier(null)}
        />
      ) : (
        <NearbyCarriers sub={sub!} locationReady={locationReady} onRequestRide={setSelectedCarrier} />
      )}
    </main>
  )
}

function LiveIndicator({ state }: { state: ConnectionState }) {
  const label = { live: 'Live', connecting: 'Connecting', offline: 'Offline' }[state]
  return (
    <span className={`live live-${state}`} title={`Realtime updates: ${label.toLowerCase()}`}>
      <span className="live-dot" aria-hidden="true" />
      {label}
    </span>
  )
}

function Brand({ bare = false }: { bare?: boolean }) {
  const brand = (
    <h1 className="brand">
      <span className="brand-mark"><PiggybackMark /></span>
      Piggy Back
    </h1>
  )
  return bare ? brand : <div className="topbar">{brand}</div>
}
