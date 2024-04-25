package logger

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestCreateLogger(t *testing.T) {
	logger := CreateLogger()

	if logger == nil {
		t.Errorf("CreateLogger() returned nil")
	}

	assert.NotNil(t, logger, "Logger должен быть создан")
}
