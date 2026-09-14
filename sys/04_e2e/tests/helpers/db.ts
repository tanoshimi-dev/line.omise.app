import crypto from 'node:crypto'
import { Client } from 'pg'
import 'dotenv/config'

// dev-plan-12-test-phase1 12.3 decision: E2E tests seed a valid session
// directly in Postgres and sign the cookie the same way
// sys/02_backend/internal/session/cookie.go does, rather than driving a real
// LINE/Google OAuth handshake. Real third-party login has no stable test
// account or sandbox we can drive headlessly and reproducibly in CI; the
// authorization-code exchange itself is already covered by
// dev-plan-04-auth's own verification (real accounts, done manually) and by
// the Go handler tests (dev-plan-12 12.1). What's left for E2E to prove is
// "does the app behave correctly once a session exists", which seeding
// answers directly — plus one test that confirms the login buttons at least
// navigate to the correct provider URL, without completing the handshake.
//
// Every row this file creates is deleted by TestUser.cleanup(), so tests are
// safe to run against the shared dev database without touching real seed
// data or other users' state.

function signSession(id: string, secret: string): string {
  const mac = crypto.createHmac('sha256', secret).update(id).digest('hex')
  return `${id}.${mac}`
}

export interface TestUser {
  id: number
  cookie: string
  cleanup: () => Promise<void>
}

export async function seedUser(role: 'admin' | 'reader'): Promise<TestUser> {
  const client = new Client({ connectionString: process.env.DATABASE_URL })
  await client.connect()

  const providerUserId = `e2e-${role}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  const userRes = await client.query<{ id: number }>(
    `INSERT INTO users (provider, provider_user_id, email, display_name, role)
     VALUES ('google', $1, $2, $3, $4) RETURNING id`,
    [providerUserId, `${providerUserId}@example.com`, `E2E ${role}`, role],
  )
  const userId = userRes.rows[0].id

  const sessionId = crypto.randomBytes(32).toString('base64url')
  const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000)
  await client.query(`INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`, [sessionId, userId, expiresAt])

  const secret = process.env.SESSION_SECRET
  if (!secret) {
    throw new Error('SESSION_SECRET is not set — copy sys/04_e2e/.env.example to .env and fill it in')
  }

  return {
    id: userId,
    cookie: signSession(sessionId, secret),
    cleanup: async () => {
      // ON DELETE CASCADE removes the session and any progress/exam rows.
      await client.query(`DELETE FROM users WHERE id = $1`, [userId])
      await client.end()
    },
  }
}

/** Runs fn with a raw DB client, for tests that need to create/clean up their own content rows (e.g. a course). */
export async function withDB<T>(fn: (client: Client) => Promise<T>): Promise<T> {
  const client = new Client({ connectionString: process.env.DATABASE_URL })
  await client.connect()
  try {
    return await fn(client)
  } finally {
    await client.end()
  }
}

// dev-plan-2-6-test 2-6.3: a published quiz + one two-choice question, seeded
// directly (like seedUser above) rather than through the Admin UI — the one
// test that exercises the Admin UI itself (quiz.spec.ts) creates its own
// quiz that way instead of using this helper.
export interface TestQuiz {
  id: number
  slug: string
  title: string
  questionId: number
  correctChoiceId: number
  wrongChoiceId: number
  cleanup: () => Promise<void>
}

export async function seedQuiz(mode: 'practice' | 'exam', options?: { passingScore?: number }): Promise<TestQuiz> {
  const client = new Client({ connectionString: process.env.DATABASE_URL })
  await client.connect()

  const slug = `e2e-quiz-${mode}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  const title = `E2E ${mode} quiz ${Date.now()}`
  const quizRes = await client.query<{ id: number }>(
    `INSERT INTO quizzes (slug, title, mode, passing_score, published)
     VALUES ($1, $2, $3, $4, true) RETURNING id`,
    [slug, title, mode, options?.passingScore ?? null],
  )
  const quizId = quizRes.rows[0].id

  const questionRes = await client.query<{ id: number }>(
    `INSERT INTO quiz_questions (quiz_id, question_text, allow_multiple, explanation, sort_order)
     VALUES ($1, '2+2?', false, 'Because 2+2=4.', 1) RETURNING id`,
    [quizId],
  )
  const questionId = questionRes.rows[0].id

  const correctRes = await client.query<{ id: number }>(
    `INSERT INTO quiz_choices (question_id, choice_text, is_correct, sort_order) VALUES ($1, '4', true, 1) RETURNING id`,
    [questionId],
  )
  const wrongRes = await client.query<{ id: number }>(
    `INSERT INTO quiz_choices (question_id, choice_text, is_correct, sort_order) VALUES ($1, '5', false, 2) RETURNING id`,
    [questionId],
  )

  return {
    id: quizId,
    slug,
    title,
    questionId,
    correctChoiceId: correctRes.rows[0].id,
    wrongChoiceId: wrongRes.rows[0].id,
    cleanup: async () => {
      // ON DELETE CASCADE removes the question/choices and any answers/attempts.
      await client.query(`DELETE FROM quizzes WHERE id = $1`, [quizId])
      await client.end()
    },
  }
}
