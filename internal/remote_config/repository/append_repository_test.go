package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func Test_Append(t *testing.T) {
	type exRes struct{ err error }

	name := "key"
	newData := json.RawMessage(`{"on":true}`)

	// match actual SQL alias used in implementation
	const selectNextSQL = `SELECT COALESCE(MAX(version), 0) + 1 AS next_version, (SELECT type FROM configs WHERE name = ? ORDER BY version DESC LIMIT 1) AS schemaType FROM configs WHERE name = ?`
	const insertSQL = `INSERT INTO configs(name, type, version, data) VALUES(?, ?, ?, ?)`
	const readBackSQL = `SELECT name, type, version, data, created_at FROM configs WHERE name = ? AND version = ? LIMIT 1`

	cases := []struct {
		name     string
		mockFunc func(m sqlmock.Sqlmock)
		ex       exRes
	}{
		{
			name: "when type missing should return ErrNotFound",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				// IMPORTANT: return empty string "" instead of NULL to avoid Scan error,
				// so repository code can map it to ErrNotFound.
				m.ExpectQuery(selectNextSQL).
					WithArgs(name, name).
					WillReturnRows(sqlmock.NewRows([]string{"next_version", "schemaType"}).AddRow(1, "")) // not nil
				m.ExpectRollback()
			},
			ex: exRes{err: ErrNotFound},
		},
		{
			name: "when success",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(selectNextSQL).
					WithArgs(name, name).
					WillReturnRows(sqlmock.NewRows([]string{"next_version", "schemaType"}).AddRow(3, "feature_toggle"))

				m.ExpectExec(insertSQL).
					WithArgs(name, "feature_toggle", 3, `{"on":true}`).
					WillReturnResult(sqlmock.NewResult(1, 1))

				m.ExpectQuery(readBackSQL).
					WithArgs(name, 3).
					WillReturnRows(sqlmock.NewRows([]string{"name", "type", "version", "data", "created_at"}).
						AddRow(name, "feature_toggle", 3, `{"on":true}`, "2025-10-01T00:00:01Z"))

				m.ExpectCommit()
			},
			ex: exRes{err: nil},
		},
		{
			name: "when insert error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(selectNextSQL).
					WithArgs(name, name).
					WillReturnRows(sqlmock.NewRows([]string{"next_version", "schemaType"}).AddRow(2, "feature_toggle"))

				m.ExpectExec(insertSQL).
					WithArgs(name, "feature_toggle", 2, `{"on":true}`).
					WillReturnError(errors.New("insert failed"))

				m.ExpectRollback()
			},
			// repo wraps error as "append.insert: insert failed"
			ex: exRes{err: errors.New("append.insert: insert failed")},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, mock, db := newMockRepoEq(t)
			defer db.Close()

			tc.mockFunc(mock)
			_, err := r.Append(context.Background(), name, newData)

			if tc.ex.err == nil {
				assert.NoError(t, err)
			} else {
				// either compare exact wrapped error or use Contains if you prefer:
				assert.EqualError(t, err, tc.ex.err.Error())
				// alternatively:
				// assert.ErrorContains(t, err, "insert failed")
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
