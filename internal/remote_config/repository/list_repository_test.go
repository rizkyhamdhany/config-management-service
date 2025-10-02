package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func Test_List(t *testing.T) {
	type exRes struct {
		count int
		err   error
	}

	cases := []struct {
		name     string
		cfgName  string
		mockFunc func(m sqlmock.Sqlmock)
		ex       exRes
	}{
		{
			name:    "when query error should return error",
			cfgName: "key",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ?
		ORDER BY version ASC`).WithArgs("key").
					WillReturnError(errors.New("query err"))
			},
			ex: exRes{count: 0, err: errors.New("query err")},
		},
		{
			name:    "when success empty should return empty",
			cfgName: "key",
			mockFunc: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "type", "version", "data", "created_at"})
				m.ExpectQuery(`SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ?
		ORDER BY version ASC`).WithArgs("key").
					WillReturnRows(rows)
			},
			ex: exRes{count: 0, err: nil},
		},
		{
			name:    "when success with rows should return rows",
			cfgName: "key",
			mockFunc: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "type", "version", "data", "created_at"}).
					AddRow("key", "feature_toggle", 1, `{"on":true}`, "2025-10-01T00:00:00Z").
					AddRow("key", "feature_toggle", 2, `{"on":false}`, "2025-10-01T00:01:00Z")
				m.ExpectQuery(`SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ?
		ORDER BY version ASC`).WithArgs("key").
					WillReturnRows(rows)
			},
			ex: exRes{count: 2, err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, mock, db := newMockRepoEq(t)
			defer db.Close()

			tc.mockFunc(mock)
			got, err := r.List(context.Background(), tc.cfgName)

			if tc.ex.err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Len(t, got, tc.ex.count)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
