// Thin fetch-based client for the line-api backend (dev-plan-07-frontend-base).
//
// NEXT_PUBLIC_API_URL points at the Go/Gin backend (dev-plan-03-backend-base).
// `credentials: 'include'` is set on every request so the session cookie
// issued by dev-plan-04-auth is sent/received once that step lands.

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })

  if (!res.ok) {
    throw new ApiError(`${init?.method ?? 'GET'} ${path} failed`, res.status)
  }

  if (res.status === 204) {
    return undefined as T
  }

  return (await res.json()) as T
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
}

export function loginUrl(provider: 'line' | 'google'): string {
  return `${API_URL}/auth/${provider}/login`
}
