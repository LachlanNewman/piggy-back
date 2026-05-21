import { useState, useEffect, useRef } from 'react'
import { useAuth } from 'react-oidc-context'
import SignupForm from './SignupForm'

export default function App() {
  const { isAuthenticated, isLoading, user, signinRedirect, signoutRedirect, signinSilent } = useAuth()
  const [message, setMessage] = useState(null)
  const [health, setHealth] = useState(null)
  const [restoringSession, setRestoringSession] = useState(false)
  const sessionRestoreAttempted = useRef(false)

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
    if (!isAuthenticated) return
    fetch('/api/hello')
      .then(r => r.json())
      .then(d => setMessage(d.message))
      .catch(() => setMessage('could not reach backend'))

    fetch('/api/health')
      .then(r => r.json())
      .then(d => setHealth(d.status))
      .catch(() => setHealth('unavailable'))
  }, [isAuthenticated])

  if (isLoading || restoringSession) {
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

  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>React + Go</h1>
        <button onClick={signoutRedirect}>Log out</button>
      </div>
      <p>Logged in as <strong>{user.profile?.email ?? user.profile?.sub}</strong></p>
      <p>Message: <strong>{message ?? '...'}</strong></p>
      <p>Health: <strong>{health ?? '...'}</strong></p>
      <hr style={{ margin: '32px 0' }} />
      <SignupForm />
    </div>
  )
}
