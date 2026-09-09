import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from 'react-oidc-context'
import { SplitTokenStore } from './auth/splitTokenStore'
import App from './App'
import CallbackPage from './pages/CallbackPage'

const required = ['VITE_OIDC_AUTHORITY', 'VITE_OIDC_CLIENT_ID', 'VITE_OIDC_REDIRECT_URI', 'VITE_OIDC_AUDIENCE'] as const
for (const key of required) {
  if (!import.meta.env[key]) {
    throw new Error(`Missing required env var: ${key}. Check .env.example for setup instructions.`)
  }
}

const oidcConfig = {
  authority: import.meta.env.VITE_OIDC_AUTHORITY,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI,
  response_type: 'code',
  // offline_access is what makes Auth0 issue a refresh token; SplitTokenStore
  // needs one to restore the session after a page reload.
  scope: 'openid profile email offline_access',
  // Without an audience Auth0 returns an opaque access token that no backend
  // can verify. This asks for a JWT minted for our API instead.
  extraQueryParams: { audience: import.meta.env.VITE_OIDC_AUDIENCE },
  userStore: new SplitTokenStore(),
  automaticSilentRenew: false,
}

const root = document.getElementById('root')!

createRoot(root).render(
  <StrictMode>
    <AuthProvider {...oidcConfig}>
      <BrowserRouter>
        <Routes>
          <Route path="/callback" element={<CallbackPage />} />
          <Route path="/*" element={<App />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  </StrictMode>
)
