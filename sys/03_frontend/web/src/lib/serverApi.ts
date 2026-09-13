import 'server-only'

// Server-only fetch helper for Server Components (dev-plan-09-frontend-learn).
//
// Uses API_URL (no NEXT_PUBLIC_ prefix, so it's never bundled to the client)
// rather than src/lib/api.ts's NEXT_PUBLIC_API_URL: inside docker-compose the
// Next.js server must reach the backend via the internal "line-api" hostname,
// while the browser reaches it via the host-mapped port instead. `server-only`
// makes importing this from a Client Component a build error, so the two
// never get mixed up.
//
// Always cache: 'no-store' (always dynamic, always fresh) rather than ISR
// with `next.revalidate`: ISR makes Next.js attempt to statically prerender
// the page at `next build` time, but this project's Dockerfile builds
// line-web without line-api reachable (no docker-compose network exists
// during an isolated image build) — that prerender attempt fails the build.
// dev-plan-10-frontend-usecase hit this concretely and reverted; see that
// step's result doc.
//
// Only public (no-auth) content endpoints are called this way. Anything
// user-specific (progress, exam submission, lesson completion) goes through
// the browser-side src/lib/api.ts instead, which carries the session cookie
// — see the dev-plan-09-frontend-learn result doc for why that split was
// chosen over forwarding cookies through server-side fetches.

const API_URL = process.env.API_URL ?? 'http://localhost:8080'

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, { cache: 'no-store' })
  if (!res.ok) {
    throw new ApiError(`GET ${path} failed`, res.status)
  }
  return (await res.json()) as T
}

/** Like get, but returns null instead of throwing on a 404. */
async function getOrNull<T>(path: string): Promise<T | null> {
  try {
    return await get<T>(path)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      return null
    }
    throw err
  }
}

export const serverApi = { get, getOrNull }
