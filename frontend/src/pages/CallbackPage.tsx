import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from 'react-oidc-context'

export default function CallbackPage() {
  const navigate = useNavigate()
  const { isLoading, isAuthenticated, error } = useAuth()

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      navigate('/', { replace: true })
    }
  }, [isLoading, isAuthenticated, navigate])

  if (error) {
    return (
      <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
        <h2>Login failed</h2>
        <p>{error.message}</p>
      </div>
    )
  }

  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 560, margin: '80px auto', padding: '0 24px' }}>
      <p>Completing login...</p>
    </div>
  )
}
