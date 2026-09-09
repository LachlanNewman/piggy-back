import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from 'react-oidc-context'
import { SplitTokenStore } from './auth/splitTokenStore'
import App from './App'
import CallbackPage from './pages/CallbackPage'

const required = ['VITE_OIDC_AUTHORITY', 'VITE_OIDC_CLIENT_ID', 'VITE_OIDC_REDIRECT_URI'] as const
for (const key of required) {
  if (!import.meta.env[key]) {
    throw new Error(`Missing required env var: ${key}. Check .env.example for setup instructions.`)
  }
}

// Optional so the app can boot before an API is registered in Auth0. When it is
// unset Auth0 issues an *opaque* access token, which the backend cannot verify,
// so login succeeds but every /api/v1 call returns 401.
const audience = import.meta.env.VITE_OIDC_AUDIENCE
if (!audience) {
  console.warn(
    'VITE_OIDC_AUDIENCE is not set: Auth0 will issue an opaque access token and ' +
    'every /api/v1 request will fail with 401. Set it to your Auth0 API identifier ' +
    '(it must match the backend OIDC_AUDIENCE) and rebuild.'
  )
}

const oidcConfig = {
  authority: import.meta.env.VITE_OIDC_AUTHORITY,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI,
  response_type: 'code',
  // offline_access is what makes Auth0 issue a refresh token; SplitTokenStore
  // needs one to restore the session after a page reload.
  scope: 'openid profile email offline_access',
  // Omitted entirely when unset — sending audience="" is not the same as not
  // asking for one.
  ...(audience ? { extraQueryParams: { audience } } : {}),
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
