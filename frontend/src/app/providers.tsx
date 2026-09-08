'use client'

import { useEffect, useState } from 'react'
import { AuthProvider } from 'react-oidc-context'
import { SplitTokenStore } from '@/lib/auth/splitTokenStore'

const authority = process.env.NEXT_PUBLIC_OIDC_AUTHORITY
const clientId = process.env.NEXT_PUBLIC_OIDC_CLIENT_ID
const redirectUri = process.env.NEXT_PUBLIC_OIDC_REDIRECT_URI

export default function Providers({ children }: { children: React.ReactNode }) {
  // oidc-client-ts touches browser storage as soon as it is constructed, so the
  // provider is mounted only after hydration rather than during SSR.
  const [mounted, setMounted] = useState(false)
  // One store for the lifetime of the tab — a new instance would drop the
  // in-memory access/id tokens on every re-render.
  const [userStore] = useState(() => new SplitTokenStore())
  useEffect(() => setMounted(true), [])

  if (!mounted) return null

  if (!authority || !clientId || !redirectUri) {
    return (
      <main className="shell">
        <div className="card">
          <h1 className="card-title">Configuration missing</h1>
          <p className="muted">
            Set <code>NEXT_PUBLIC_OIDC_AUTHORITY</code>, <code>NEXT_PUBLIC_OIDC_CLIENT_ID</code> and{' '}
            <code>NEXT_PUBLIC_OIDC_REDIRECT_URI</code>. See <code>frontend/.env.example</code>.
          </p>
        </div>
      </main>
    )
  }

  return (
    <AuthProvider
      authority={authority}
      client_id={clientId}
      redirect_uri={redirectUri}
      response_type="code"
      scope="openid profile email"
      userStore={userStore}
      automaticSilentRenew={false}
    >
      {children}
    </AuthProvider>
  )
}
