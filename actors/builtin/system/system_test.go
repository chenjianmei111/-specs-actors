package system_test

import (
	"testing"

	"github.com/chenjianmei111/specs-actors/actors/builtin/system"
	"github.com/chenjianmei111/specs-actors/support/mock"
)

func TestExports(t *testing.T) {
	mock.CheckActorExports(t, system.Actor{})
}
