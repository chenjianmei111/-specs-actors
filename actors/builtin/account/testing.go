package account

import (
	"github.com/chenjianmei111/go-address"

	"github.com/chenjianmei111/specs-actors/v3/actors/builtin"
)

type StateSummary struct {
	PubKeyAddr address.Address
}

// Checks internal invariants of account state.
func CheckStateInvariants(st *State, idAddr address.Address) (*StateSummary, *builtin.MessageAccumulator) {
	acc := &builtin.MessageAccumulator{}
	accountSummary := &StateSummary{
		PubKeyAddr: st.Address,
	}

	if id, err := address.IDFromAddress(idAddr); err != nil {
		acc.Addf("error extracting actor ID from address: %v", err)
	} else if id >= builtin.FirstNonSingletonActorId {
		acc.Require(st.Address.Protocol() == address.BLS || st.Address.Protocol() == address.SECP256K1,
			"actor address %v must be BLS or SECP256K1 protocol", st.Address)
	}

	return accountSummary, acc
}
