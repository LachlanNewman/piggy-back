'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from 'react-oidc-context'
import PiggybackMark from '@/components/PiggybackMark'

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
          <div className="status-mark">😵</div>
          <h2>Login failed</h2>
          <p className="muted">{error.message}</p>
        </div>
      </main>
    )
  }

  return (
    <main className="shell">
      <div className="card status-card">
        <div className="status-mark status-mark-accent"><PiggybackMark /></div>
        <p className="muted" style={{ marginTop: 16 }}>Completing login<span className="dots" /></p>
      </div>
    </main>
  )
}
