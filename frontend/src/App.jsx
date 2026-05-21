import { useState, useEffect, useRef } from 'react'
import { useAuth } from 'react-oidc-context'
import ProfileCompletionForm from './ProfileCompletionForm'

export default function App() {
  const { isAuthenticated, isLoading, user, signinRedirect, signoutRedirect, signinSilent } = useAuth()
  const [message, setMessage] = useState(null)
  const [health, setHealth] = useState(null)
  const [restoringSession, setRestoringSession] = useState(false)
  const sessionRestoreAttempted = useRef(false)
  const [profileStatus, setProfileStatus] = useState('idle') // 'idle' | 'loading' | 'incomplete' | 'complete'

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
    if (!isAuthenticated || isLoading || restoringSession) return
    setProfileStatus('loading')
    fetch(`/api/v1/users/me?sub=${encodeURIComponent(user.profile.sub)}`)
      .then(r => {
        if (r.status === 404) return { profile_complete: false }
        if (!r.ok) throw new Error('profile check failed')
        return r.json()
      })
      .then(data => setProfileStatus(data.profile_complete ? 'complete' : 'incomplete'))
      .catch(() => setProfileStatus('incomplete'))
  }, [isAuthenticated, isLoading, restoringSession, user])

  useEffect(() => {
    if (profileStatus !== 'complete') return
    fetch('/api/hello')
      .then(r => r.json())
      .then(d => setMessage(d.message))
      .catch(() => setMessage('could not reach backend'))

    fetch('/api/health')
      .then(r => r.json())
      .then(d => setHealth(d.status))
      .catch(() => setHealth('unavailable'))
  }, [profileStatus])

  function handleProfileComplete() {
    setProfileStatus('loading')
    fetch(`/api/v1/users/me?sub=${encodeURIComponent(user.profile.sub)}`)
      .then(r => {
        if (r.status === 404) return { profile_complete: false }
        if (!r.ok) throw new Error('profile check failed')
        return r.json()
      })
      .then(data => setProfileStatus(data.profile_complete ? 'complete' : 'incomplete'))
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
        <h1>React + Go</h1>
        <p>Please log in to continue.</p>
        <button onClick={signinRedirect}>Log in</button>
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
        <h1>React + Go</h1>
        <button onClick={signoutRedirect}>Log out</button>
      </div>
      <p>Logged in as <strong>{user.profile?.email ?? user.profile?.sub}</strong></p>
      <p>Message: <strong>{message ?? '...'}</strong></p>
      <p>Health: <strong>{health ?? '...'}</strong></p>
    </div>
  )
}
