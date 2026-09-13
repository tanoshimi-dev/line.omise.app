import 'server-only'
import { cookies } from 'next/headers'

// Server-side session check for the /admin route guard
// (dev-plan-11-frontend-admin 11.1 explicitly asks for this — checking
// role on the server so /admin never even renders for non-admins, with a
// real HTTP redirect that works without JS, rather than a client-side
// flash-then-redirect).
//
// This is a deliberate, narrow exception to the boundary dev-plan-09 set
// (user-specific data only fetched client-side, to avoid forwarding
// HttpOnly cookies through SSR): forwarding the session cookie here is
// unavoidable because the whole point is to gate rendering itself, which a
// client-side check can't do.

const API_URL = process.env.API_URL ?? 'http://localhost:8080'
const SESSION_COOKIE_NAME = 'line_omise_session'

export interface CurrentUser {
  id: string
  provider: 'line' | 'google'
  email: string
  display_name: string
  avatar_url: string
  role: 'admin' | 'reader'
}

/** Returns the logged-in user (via the forwarded session cookie), or null. */
export async function getCurrentUserServer(): Promise<CurrentUser | null> {
  const cookieStore = await cookies()
  const session = cookieStore.get(SESSION_COOKIE_NAME)
  if (!session) {
    return null
  }

  const res = await fetch(`${API_URL}/auth/me`, {
    headers: { Cookie: `${SESSION_COOKIE_NAME}=${session.value}` },
    cache: 'no-store',
  })
  if (!res.ok) {
    return null
  }
  return (await res.json()) as CurrentUser
}
