import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import userManager from '../auth/userManager'

export default function CallbackPage() {
  const navigate = useNavigate()
  const [error, setError] = useState(null)

  useEffect(() => {
    userManager.signinRedirectCallback()
      .then(() => navigate('/', { replace: true }))
      .catch(err => setError(err.message ?? 'Login failed'))
  }, [navigate])

  if (error) {
    return (
      <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
        <h2>Login failed</h2>
        <p>{error}</p>
      </div>
    )
  }

  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
      <p>Completing login...</p>
    </div>
  )
}
