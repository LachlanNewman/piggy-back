import { createContext, useContext, useEffect, useState } from 'react'
import userManager from './userManager'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const onUserLoaded = u => setUser(u)
    const onUserUnloaded = () => setUser(null)

    userManager.events.addUserLoaded(onUserLoaded)
    userManager.events.addUserUnloaded(onUserUnloaded)

    async function init() {
      let u = await userManager.getUser()
      // After a page reload the in-memory store is empty but the refresh token
      // cookie may exist. getUser() returns an expired stub in that case —
      // exchange it for fresh tokens before showing the app.
      if (u?.expired && u?.refresh_token) {
        try {
          u = await userManager.signinSilent()
        } catch {
          u = null
        }
      }
      setUser(u)
      setIsLoading(false)
    }

    init()

    return () => {
      userManager.events.removeUserLoaded(onUserLoaded)
      userManager.events.removeUserUnloaded(onUserUnloaded)
    }
  }, [])

  const value = {
    user,
    isAuthenticated: user !== null && !user.expired,
    isLoading,
    login: () => userManager.signinRedirect(),
    logout: () => userManager.signoutRedirect(),
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (ctx === null) {
    throw new Error('useAuth must be used inside AuthProvider')
  }
  return ctx
}
