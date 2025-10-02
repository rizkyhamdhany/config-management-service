package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func Test_Latest(t *testing.T) {
	type exRes struct{ err error }

	cases := []struct {
		name     string
		cfgName  string
		mockFunc func(m sqlmock.Sqlmock)
		ex       exRes
	}{
		{
			name:    "when not found should return ErrNotFound",
			cfgName: "none",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ?
		ORDER BY version DESC
		LIMIT 1`).WithArgs("none").
					WillReturnError(sql.ErrNoRows)
			},
			ex: exRes{err: ErrNotFound},
		},
		{
			name:    "when success",
			cfgName: "key",
			mockFunc: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "type", "version", "data", "created_at"}).
					AddRow("key", "feature_toggle", 7, `{"on":false}`, "2025-10-01T00:00:00Z")
				m.ExpectQuery(`SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ?
		ORDER BY version DESC
		LIMIT 1`).WithArgs("key").
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
			_, err := r.Latest(context.Background(), tc.cfgName)

			assert.Equal(t, tc.ex.err, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
