import { useState, useEffect } from 'react'
import { useAuth } from './auth/AuthContext'
import SignupForm from './SignupForm'

export default function App() {
  const { isAuthenticated, isLoading, user, login, logout } = useAuth()
  const [message, setMessage] = useState(null)
  const [health, setHealth] = useState(null)

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

  if (isLoading) {
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
        <button onClick={login}>Log in</button>
      </div>
    )
  }

  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>React + Go</h1>
        <button onClick={logout}>Log out</button>
      </div>
      <p>Logged in as <strong>{user.profile?.email ?? user.profile?.sub}</strong></p>
      <p>Message: <strong>{message ?? '...'}</strong></p>
      <p>Health: <strong>{health ?? '...'}</strong></p>
      <hr style={{ margin: '32px 0' }} />
      <SignupForm />
    </div>
  )
}
