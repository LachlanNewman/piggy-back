import { useState, useEffect } from 'react'
import SignupForm from './SignupForm'

export default function App() {
  const [message, setMessage] = useState(null)
  const [health, setHealth] = useState(null)

  useEffect(() => {
    fetch('/api/hello')
      .then(r => r.json())
      .then(d => setMessage(d.message))
      .catch(() => setMessage('could not reach backend'))

    fetch('/api/health')
      .then(r => r.json())
      .then(d => setHealth(d.status))
      .catch(() => setHealth('unavailable'))
  }, [])

  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
      <h1>React + Go</h1>
      <p>Message: <strong>{message ?? '...'}</strong></p>
      <p>Health: <strong>{health ?? '...'}</strong></p>
      <hr style={{ margin: '32px 0' }} />
      <SignupForm />
    </div>
  )
}
