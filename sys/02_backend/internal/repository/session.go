package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/session"
)

// SessionTTL is how long an issued session stays valid.
const SessionTTL = 30 * 24 * time.Hour

// ErrSessionNotFound is returned when a session id doesn't exist or has
// expired.
var ErrSessionNotFound = errors.New("session not found")

// Session mirrors the `sessions` table (dev-plan-04-auth 4.3).
type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SessionRepository queries the `sessions` table.
type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create issues a new session for userID with a cryptographically random id.
func (r *SessionRepository) Create(ctx context.Context, userID int64) (*Session, error) {
	id, err := session.RandomToken()
	if err != nil {
		return nil, err
	}

	sess := &Session{
		ID:        id,
		UserID:    userID,
		ExpiresAt: time.Now().Add(SessionTTL),
	}

	err = r.db.QueryRow(ctx, `
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`, sess.ID, sess.UserID, sess.ExpiresAt).Scan(&sess.CreatedAt)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// GetValid returns the session for id, or ErrSessionNotFound if it doesn't
// exist or has expired.
func (r *SessionRepository) GetValid(ctx context.Context, id string) (*Session, error) {
	var sess Session
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = $1 AND expires_at > now()
	`, id).Scan(&sess.ID, &sess.UserID, &sess.ExpiresAt, &sess.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// Delete removes a session (logout). Deleting an already-gone session is not
// an error.
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}
