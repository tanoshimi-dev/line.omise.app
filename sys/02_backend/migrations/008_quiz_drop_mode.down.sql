ALTER TABLE quizzes ADD COLUMN mode TEXT NOT NULL DEFAULT 'practice'
    CHECK (mode IN ('practice', 'exam'));
