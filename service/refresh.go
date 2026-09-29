package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrInvalidRefreshToken covers unknown, expired, revoked and reused tokens.
// Callers must not distinguish them to the client.
var ErrInvalidRefreshToken = errors.New("invalid refresh token")

// RefreshService issues and rotates opaque refresh tokens stored as SHA-256
// hashes. Every login starts a family; each refresh replaces the token with a
// new one in the same family. Presenting an already-used token means it leaked
// (or the client raced itself), so the whole family is revoked.
type RefreshService struct {
	db  *gorm.DB
	ttl time.Duration
	now func() time.Time
}

func NewRefreshService(db *gorm.DB, ttl time.Duration) *RefreshService {
	return &RefreshService{db: db, ttl: ttl, now: time.Now}
}

func newSecret(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *RefreshService) insert(ctx context.Context, db *gorm.DB, userID int32, family string) (string, time.Time, error) {
	token, err := newSecret(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expires := s.now().Add(s.ttl)
	err = db.WithContext(ctx).Exec(
		`INSERT INTO refresh_token (user_id, family_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`,
		userID, family, hashToken(token), expires,
	).Error
	return token, expires, err
}

// Issue starts a new token family for the user (called on login and register).
func (s *RefreshService) Issue(ctx context.Context, userID int32) (string, time.Time, error) {
	family, err := newSecret(12) // 16 chars
	if err != nil {
		return "", time.Time{}, err
	}
	// Opportunistic cleanup keeps the table from growing with dead tokens.
	if err := s.db.WithContext(ctx).Exec(
		`DELETE FROM refresh_token WHERE user_id = ? AND expires_at < ?`, userID, s.now().Add(-24*time.Hour),
	).Error; err != nil {
		return "", time.Time{}, err
	}
	return s.insert(ctx, s.db, userID, family)
}

// Rotate consumes token and returns the user ID with a fresh token of the same family.
func (s *RefreshService) Rotate(ctx context.Context, token string) (int32, string, time.Time, error) {
	var (
		userID  int32
		newTok  string
		expires time.Time
	)
	hash := hashToken(token)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Atomically claim the token: only one concurrent caller can win.
		var claimed struct {
			UserID   int32
			FamilyID string
		}
		res := tx.Raw(
			`UPDATE refresh_token SET revoked_at = ?
			 WHERE token_hash = ? AND revoked_at IS NULL AND expires_at > ?
			 RETURNING user_id, family_id`,
			s.now(), hash, s.now(),
		).Scan(&claimed)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errNotClaimed
		}

		var err error
		userID = claimed.UserID
		newTok, expires, err = s.insert(ctx, tx, claimed.UserID, claimed.FamilyID)
		return err
	})

	if errors.Is(err, errNotClaimed) {
		// Known but already revoked => reuse: burn the whole family.
		if rerr := s.revokeFamilyOf(ctx, hash); rerr != nil {
			return 0, "", time.Time{}, rerr
		}
		return 0, "", time.Time{}, ErrInvalidRefreshToken
	}
	if err != nil {
		return 0, "", time.Time{}, err
	}
	return userID, newTok, expires, nil
}

var errNotClaimed = errors.New("token not claimable")

// revokeFamilyOf revokes every live token in the family of the token with this
// hash, if the token exists and was already revoked (i.e. it is being reused).
func (s *RefreshService) revokeFamilyOf(ctx context.Context, hash string) error {
	return s.db.WithContext(ctx).Exec(
		`UPDATE refresh_token SET revoked_at = ?
		 WHERE revoked_at IS NULL AND family_id = (
		     SELECT family_id FROM refresh_token WHERE token_hash = ? AND revoked_at IS NOT NULL)`,
		s.now(), hash,
	).Error
}

// Revoke ends the session the token belongs to. Unknown tokens are ignored so
// logout is idempotent.
func (s *RefreshService) Revoke(ctx context.Context, token string) error {
	return s.db.WithContext(ctx).Exec(
		`UPDATE refresh_token SET revoked_at = ?
		 WHERE revoked_at IS NULL AND family_id = (SELECT family_id FROM refresh_token WHERE token_hash = ?)`,
		s.now(), hashToken(token),
	).Error
}

// RevokeAll ends every session of the user.
func (s *RefreshService) RevokeAll(ctx context.Context, userID int32) error {
	return s.db.WithContext(ctx).Exec(
		`UPDATE refresh_token SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`, s.now(), userID,
	).Error
}
