package test_test

import (
	"testing"

	"github.com/chenjianmei111/go-state-types/rt"
)

type TestLogger struct {
	TB testing.TB
}

func (t TestLogger) Log(_ rt.LogLevel, msg string, args ...interface{}) {
	t.TB.Logf(msg, args...)
}
