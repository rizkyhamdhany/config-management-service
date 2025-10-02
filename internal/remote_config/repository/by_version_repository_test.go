package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func newMockRepoEq(t *testing.T) (*repo, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	return &repo{db: db}, mock, db
}

func Test_ByVersion(t *testing.T) {
	type exRes struct {
		err error
	}

	cases := []struct {
		name     string
		cfgName  string
		version  int
		mockFunc func(m sqlmock.Sqlmock)
		ex       exRes
	}{
		{
			name:    "when not found",
			cfgName: "missing",
			version: 9,
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ? AND version = ?
		LIMIT 1`).WithArgs("missing", 9).
					WillReturnError(sql.ErrNoRows)
			},
			ex: exRes{err: ErrNotFound},
		},
		{
			name:    "when success",
			cfgName: "key",
			version: 2,
			mockFunc: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "type", "version", "data", "created_at"}).
					AddRow("key", "feature_toggle", 2, `{"on":true}`, "2025-10-01T00:00:00Z")
				m.ExpectQuery(`SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ? AND version = ?
		LIMIT 1`).WithArgs("key", 2).
					WillReturnRows(rows)
			},
			ex: exRes{err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, mock, db := newMockRepoEq(t)
			defer db.Close()

			tc.mockFunc(mock)

			_, err := r.ByVersion(context.Background(), tc.cfgName, tc.version)

			assert.Equal(t, tc.ex.err, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
