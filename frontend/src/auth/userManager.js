import { UserManager } from 'oidc-client-ts'
import { SplitTokenStore } from './splitTokenStore'

const required = ['VITE_OIDC_AUTHORITY', 'VITE_OIDC_CLIENT_ID', 'VITE_OIDC_REDIRECT_URI']
for (const key of required) {
  if (!import.meta.env[key]) {
    throw new Error(`Missing required env var: ${key}. Check .env.example for setup instructions.`)
  }
}

const userManager = new UserManager({
  authority: import.meta.env.VITE_OIDC_AUTHORITY,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI,
  response_type: 'code',
  scope: 'openid profile email',
  userStore: new SplitTokenStore(),
  automaticSilentRenew: false,
})

export default userManager
