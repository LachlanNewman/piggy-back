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

const oidcConfig = {
  authority: import.meta.env.VITE_OIDC_AUTHORITY,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI,
  response_type: 'code',
  scope: 'openid profile email',
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
