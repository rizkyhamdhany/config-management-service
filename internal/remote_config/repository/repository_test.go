package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewRepo(t *testing.T) {
	assert.NotPanics(t, func() { NewRepo(nil) })
}
