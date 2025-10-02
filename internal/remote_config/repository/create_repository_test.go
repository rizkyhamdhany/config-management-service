package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func Test_Create(t *testing.T) {
	type exRes struct {
		err error
	}

	cases := []struct {
		name       string
		schemaType string
		cfgName    string
		data       json.RawMessage
		mockFunc   func(m sqlmock.Sqlmock)
		ex         exRes
	}{
		{
			name:       "when unique violation should return ErrAlreadyExists",
			schemaType: "feature_toggle",
			cfgName:    "dup",
			data:       json.RawMessage(`{}`),
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO configs(name, type, version, data)
		VALUES(?, ?, 1, ?)`).
					WithArgs("dup", "feature_toggle", "{}").
					WillReturnError(errors.New("UNIQUE constraint failed: configs.name"))
			},
			ex: exRes{err: ErrAlreadyExists},
		},
		{
			name:       "when other exec error should return error",
			schemaType: "feature_toggle",
			cfgName:    "x",
			data:       json.RawMessage(`{}`),
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO configs(name, type, version, data)
		VALUES(?, ?, 1, ?)`).
					WithArgs("x", "feature_toggle", "{}").
					WillReturnError(errors.New("boom"))
			},
			ex: exRes{err: errors.New("boom")},
		},
		{
			name:       "when success",
			schemaType: "feature_toggle",
			cfgName:    "qris",
			data:       json.RawMessage(`{"enabled":true}`),
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO configs(name, type, version, data)
		VALUES(?, ?, 1, ?)`).
					WithArgs("qris", "feature_toggle", `{"enabled":true}`).
					WillReturnResult(sqlmock.NewResult(1, 1))

				rows := sqlmock.NewRows([]string{"name", "type", "version", "data", "created_at"}).
					AddRow("qris", "feature_toggle", 1, `{"enabled":true}`, "2025-10-01T00:00:00Z")
				m.ExpectQuery(`SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ? AND version = ?
		LIMIT 1`).WithArgs("qris", 1).
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
			_, err := r.Create(context.Background(), tc.schemaType, tc.cfgName, tc.data)

			if tc.ex.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if errors.Is(tc.ex.err, ErrAlreadyExists) {
					assert.ErrorIs(t, err, ErrAlreadyExists)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
