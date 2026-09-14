CREATE TABLE quizzes (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT,
    mode TEXT NOT NULL DEFAULT 'practice'
        CHECK (mode IN ('practice', 'exam')),
    passing_score INTEGER,
    published BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE quiz_questions (
    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT NOT NULL REFERENCES quizzes (id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    allow_multiple BOOLEAN NOT NULL DEFAULT false,
    explanation TEXT NOT NULL,
    reference_url TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE quiz_choices (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES quiz_questions (id) ON DELETE CASCADE,
    choice_text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT false,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE user_quiz_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    quiz_id BIGINT NOT NULL REFERENCES quizzes (id) ON DELETE CASCADE,
    score INTEGER NOT NULL,
    total_questions INTEGER NOT NULL,
    passed BOOLEAN,
    started_at TIMESTAMPTZ NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_quiz_answers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES quiz_questions (id) ON DELETE CASCADE,
    attempt_id BIGINT REFERENCES user_quiz_attempts (id) ON DELETE CASCADE,
    selected_choice_ids BIGINT[] NOT NULL,
    is_correct BOOLEAN NOT NULL,
    answered_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
