import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError, loginUrl } from './api'

describe('api client', () => {
  const originalFetch = global.fetch

  beforeEach(() => {
    global.fetch = vi.fn()
  })

  afterEach(() => {
    global.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('get() sends credentials and parses JSON on success', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ hello: 'world' }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
    )

    const result = await api.get<{ hello: string }>('/api/courses')

    expect(result).toEqual({ hello: 'world' })
    const [, init] = vi.mocked(fetch).mock.calls[0]
    expect(init?.credentials).toBe('include')
  })

  it('post() serializes the body and uses the POST method', async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(JSON.stringify({ id: '1' }), { status: 201 }))

    await api.post('/api/admin/courses', { slug: 'x', title: 'X' })

    const [, init] = vi.mocked(fetch).mock.calls[0]
    expect(init?.method).toBe('POST')
    expect(init?.body).toBe(JSON.stringify({ slug: 'x', title: 'X' }))
  })

  it('put() uses the PUT method', async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(JSON.stringify({}), { status: 200 }))
    await api.put('/api/admin/courses/1', { title: 'Updated' })
    const [, init] = vi.mocked(fetch).mock.calls[0]
    expect(init?.method).toBe('PUT')
  })

  it('delete() uses the DELETE method and tolerates a 204 with no body', async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 204 }))
    const result = await api.delete('/api/admin/courses/1')
    expect(result).toBeUndefined()
    const [, init] = vi.mocked(fetch).mock.calls[0]
    expect(init?.method).toBe('DELETE')
  })

  it('throws ApiError with the response status on a non-2xx response', async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(JSON.stringify({ error: 'not found' }), { status: 404 }))

    await expect(api.get('/api/courses/missing')).rejects.toMatchObject({
      name: 'ApiError',
      status: 404,
    })
  })

  it('ApiError is an instanceof Error', async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 403 }))
    try {
      await api.get('/api/admin/courses')
      expect.unreachable('expected api.get to throw')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      expect(err).toBeInstanceOf(Error)
    }
  })
})

describe('loginUrl', () => {
  it('builds a provider-specific login URL', () => {
    expect(loginUrl('line')).toMatch(/\/auth\/line\/login$/)
    expect(loginUrl('google')).toMatch(/\/auth\/google\/login$/)
  })
})
