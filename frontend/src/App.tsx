import { useState, useEffect, useRef } from 'react'
import { useAuth } from 'react-oidc-context'
import ProfileCompletionForm from './ProfileCompletionForm'
import NearbyUsersList from './NearbyUsersList'
import RideRequestForm from './RideRequestForm'
import IncomingRequests from './IncomingRequests'
import { backendClient } from './api/client'

const POLL_INTERVAL_MS = 30_000

type ProfileStatus = 'idle' | 'loading' | 'incomplete' | 'complete'

interface Driver {
  sub: string
  name: string
}

export default function App() {
  const { isAuthenticated, isLoading, user, signinRedirect, signoutRedirect, signinSilent } = useAuth()
  const [restoringSession, setRestoringSession] = useState(false)
  const sessionRestoreAttempted = useRef(false)
  const [profileStatus, setProfileStatus] = useState<ProfileStatus>('idle')
  const [locationDenied, setLocationDenied] = useState(false)
  const [selectedDriver, setSelectedDriver] = useState<Driver | null>(null)

  useEffect(() => {
    if (isLoading || sessionRestoreAttempted.current) return
    if (!isAuthenticated) {
      const hasCookie = document.cookie.split('; ').some(c => c.startsWith('oidc_rt='))
      if (hasCookie) {
        sessionRestoreAttempted.current = true
        setRestoringSession(true)
        signinSilent().finally(() => setRestoringSession(false))
      }
    }
  }, [isLoading, isAuthenticated, signinSilent])

  useEffect(() => {
    if (!isAuthenticated || isLoading || restoringSession || !user) return
    setProfileStatus('loading')
    backendClient.getUserMe(user.profile.sub)
      .then(data => setProfileStatus(data?.profile_complete ? 'complete' : 'incomplete'))
      .catch(() => setProfileStatus('incomplete'))
  }, [isAuthenticated, isLoading, restoringSession, user])

  useEffect(() => {
    if (!isAuthenticated || profileStatus !== 'complete' || !user) return
    if (!navigator.geolocation) return

    function pushLocation() {
      navigator.geolocation.getCurrentPosition(
        pos => {
          backendClient.pushLocation(user!.profile.sub, pos.coords.latitude, pos.coords.longitude).catch(() => {})
        },
        () => setLocationDenied(true)
      )
    }

    pushLocation()
    const interval = setInterval(pushLocation, POLL_INTERVAL_MS)
    return () => clearInterval(interval)
  }, [isAuthenticated, profileStatus, user])

  function handleProfileComplete() {
    if (!user) return
    setProfileStatus('loading')
    backendClient.getUserMe(user.profile.sub)
      .then(data => setProfileStatus(data?.profile_complete ? 'complete' : 'incomplete'))
      .catch(() => setProfileStatus('incomplete'))
  }

  if (isLoading || restoringSession || (isAuthenticated && profileStatus === 'loading')) {
    return (
      <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
        <p>Loading...</p>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
        <h1>Piggy Back</h1>
        <p>Please log in to continue.</p>
        <button onClick={() => signinRedirect()}>Log in</button>
      </div>
    )
  }

  if (profileStatus === 'incomplete') {
    return (
      <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
        <ProfileCompletionForm onComplete={handleProfileComplete} />
      </div>
    )
  }

  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>Piggy Back</h1>
        <button onClick={() => signoutRedirect()}>Log out</button>
      </div>
      <p>Logged in as <strong>{user?.profile?.email ?? user?.profile?.sub}</strong></p>

      {locationDenied && (
        <p style={{ color: '#888', fontSize: 13 }}>
          Location permission denied. Enable it in your browser to find nearby users.
        </p>
      )}

      <IncomingRequests sub={user!.profile.sub} pollIntervalMs={POLL_INTERVAL_MS} />

      <hr />

      {selectedDriver ? (
        <RideRequestForm
          sub={user!.profile.sub}
          driverID={selectedDriver.sub}
          driverName={selectedDriver.name}
          pollIntervalMs={POLL_INTERVAL_MS}
          onCancel={() => setSelectedDriver(null)}
        />
      ) : (
        <NearbyUsersList
          sub={user!.profile.sub}
          onRequestRide={driver => setSelectedDriver(driver)}
        />
      )}
    </div>
  )
}
