package account_test

import (
	"context"
	"testing"

	"github.com/chenjianmei111/go-address"
	"github.com/chenjianmei111/go-state-types/exitcode"
	"github.com/stretchr/testify/assert"

	builtin "github.com/chenjianmei111/specs-actors/actors/builtin"
	account "github.com/chenjianmei111/specs-actors/actors/builtin/account"
	mock "github.com/chenjianmei111/specs-actors/support/mock"
	tutil "github.com/chenjianmei111/specs-actors/support/testing"
)

type constructorTestCase struct {
	desc     string
	addr     address.Address
	exitCode exitcode.ExitCode
}

func TestExports(t *testing.T) {
	mock.CheckActorExports(t, account.Actor{})
}

func TestAccountactor(t *testing.T) {
	actor := account.Actor{}

	receiver := tutil.NewIDAddr(t, 100)
	builder := mock.NewBuilder(context.Background(), receiver).WithCaller(builtin.SystemActorAddr, builtin.SystemActorCodeID)

	testCases := []constructorTestCase{
		{
			desc:     "happy path construct SECP256K1 address",
			addr:     tutil.NewSECP256K1Addr(t, "secpaddress"),
			exitCode: exitcode.Ok,
		},
		{
			desc:     "happy path construct BLS address",
			addr:     tutil.NewBLSAddr(t, 1),
			exitCode: exitcode.Ok,
		},
		{
			desc:     "fail to construct account actor using ID address",
			addr:     tutil.NewIDAddr(t, 1),
			exitCode: exitcode.ErrIllegalArgument,
		},
		{
			desc:     "fail to construct account actor using Actor address",
			addr:     tutil.NewActorAddr(t, "actoraddress"),
			exitCode: exitcode.ErrIllegalArgument,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			rt := builder.Build(t)
			rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

			if tc.exitCode.IsSuccess() {
				rt.Call(actor.Constructor, &tc.addr)

				var st account.State
				rt.GetState(&st)
				assert.Equal(t, tc.addr, st.Address)

				rt.ExpectValidateCallerAny()
				pubkeyAddress := rt.Call(actor.PubkeyAddress, nil).(*address.Address)
				assert.Equal(t, &tc.addr, pubkeyAddress)
			} else {
				rt.ExpectAbort(tc.exitCode, func() {
					rt.Call(actor.Constructor, &tc.addr)
				})
			}
			rt.Verify()
		})
	}
}
