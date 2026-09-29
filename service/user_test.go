package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newUserService(t *testing.T) (*UserService, sqlmock.Sqlmock) {
	t.Helper()
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)
	return NewUserService(gdb), m
}

func TestUserService_Register(t *testing.T) {
	s, m := newUserService(t)
	m.ExpectBegin()
	m.ExpectQuery(`INSERT INTO "users"`).
		WithArgs("a@example.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(3, time.Now(), time.Now()))
	m.ExpectCommit()

	u, err := s.Register(context.Background(), "  A@Example.com ", "password123")
	require.NoError(t, err)
	assert.Equal(t, int32(3), u.ID)
	assert.NotEqual(t, "password123", u.PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("password123")))
}

func TestUserService_Register_Duplicate(t *testing.T) {
	s, m := newUserService(t)
	m.ExpectBegin()
	m.ExpectQuery(`INSERT INTO "users"`).WillReturnError(&pgconn.PgError{Code: pgUniqueViolation})
	m.ExpectRollback()

	_, err := s.Register(context.Background(), "a@example.com", "password123")
	assert.ErrorIs(t, err, ErrEmailTaken)
}

func TestUserService_Authenticate(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	cols := []string{"id", "email", "password_hash", "created_at", "updated_at"}

	tests := []struct {
		name, password string
		rows           *sqlmock.Rows
		dbErr          error
		wantErr        error
	}{
		{"ok", "password123", sqlmock.NewRows(cols).AddRow(1, "a@example.com", string(hash), time.Now(), time.Now()), nil, nil},
		{"wrong password", "wrong-password", sqlmock.NewRows(cols).AddRow(1, "a@example.com", string(hash), time.Now(), time.Now()), nil, ErrInvalidCredentials},
		{"unknown email", "password123", sqlmock.NewRows(cols), nil, ErrInvalidCredentials},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, m := newUserService(t)
			m.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1`).WithArgs("a@example.com", 1).WillReturnRows(tt.rows)

			u, err := s.Authenticate(context.Background(), "A@example.com", tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, int32(1), u.ID)
		})
	}

	t.Run("db error is not masked", func(t *testing.T) {
		s, m := newUserService(t)
		boom := errors.New("boom")
		m.ExpectQuery(`SELECT`).WillReturnError(boom)
		_, err := s.Authenticate(context.Background(), "a@example.com", "x")
		assert.ErrorIs(t, err, boom)
	})
}
