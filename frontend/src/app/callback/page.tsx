'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from 'react-oidc-context'

export default function CallbackPage() {
  const router = useRouter()
  const { isLoading, isAuthenticated, error } = useAuth()

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace('/')
    }
  }, [isLoading, isAuthenticated, router])

  if (error) {
    return (
      <main className="shell">
        <div className="card status-card">
          <div className="status-emoji">😵</div>
          <h2>Login failed</h2>
          <p className="muted">{error.message}</p>
        </div>
      </main>
    )
  }

  return (
    <main className="shell">
      <div className="card status-card">
        <div className="status-emoji">🐷</div>
        <p className="muted">Completing login<span className="dots" /></p>
      </div>
    </main>
  )
}
