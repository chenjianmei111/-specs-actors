package system

import (
	abi "github.com/chenjianmei111/specs-actors/actors/abi"
	builtin "github.com/chenjianmei111/specs-actors/actors/builtin"
	runtime "github.com/chenjianmei111/specs-actors/actors/runtime"
	adt "github.com/chenjianmei111/specs-actors/actors/util/adt"
)

type Actor struct{}

func (a Actor) Exports() []interface{} {
	return []interface{}{
		builtin.MethodConstructor: a.Constructor,
	}
}

var _ abi.Invokee = Actor{}

func (a Actor) Constructor(rt runtime.Runtime, _ *adt.EmptyValue) *adt.EmptyValue {
	rt.ValidateImmediateCallerIs(builtin.SystemActorAddr)

	rt.State().Create(&State{})
	return nil
}

type State struct{}
