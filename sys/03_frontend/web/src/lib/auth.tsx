'use client'

// Client-side auth state (dev-plan-07-frontend-base, task 7.4).
//
// Decision: plain React Context that fetches `GET /auth/me` on mount, rather
// than a server-component-per-request approach — simplest option that still
// works once dev-plan-04-auth's session cookie exists. Until that step
// lands, `/auth/me` 404s and is treated the same as "not logged in" (same
// stub pattern used for the DB check in dev-plan-03-backend-base).

import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { api, ApiError } from './api'

export interface CurrentUser {
  id: string
  provider: 'line' | 'google'
  email: string
  display_name: string
  avatar_url: string
  role: 'admin' | 'reader'
}

interface AuthState {
  user: CurrentUser | null
  loading: boolean
  refresh: () => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [loading, setLoading] = useState(true)

  const refresh = async () => {
    try {
      const me = await api.get<CurrentUser>('/auth/me')
      setUser(me)
    } catch (err) {
      // Not logged in, or /auth/me doesn't exist yet (pre dev-plan-04-auth).
      if (!(err instanceof ApiError)) {
        console.error('failed to fetch current user', err)
      }
      setUser(null)
    } finally {
      setLoading(false)
    }
  }

  const logout = async () => {
    try {
      await api.post('/auth/logout')
    } catch {
      // Ignore — logging out is best-effort until dev-plan-04-auth exists.
    }
    setUser(null)
    // Hard navigation, not router.push: logging out while on an
    // admin-guarded page (dev-plan-11-frontend-admin) must force a fresh
    // server-side re-check of the session on whatever page loads next.
    // The admin guard only runs in the Server Component layout, which
    // client-side state changes here don't re-invoke, and even a
    // router.push could still serve a cached client-side render of the
    // admin route on a subsequent back-navigation. A full reload sidesteps
    // both.
    window.location.href = '/'
  }

  useEffect(() => {
    void refresh()
  }, [])

  return <AuthContext.Provider value={{ user, loading, refresh, logout }}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return ctx
}
