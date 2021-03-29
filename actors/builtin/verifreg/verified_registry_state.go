package verifreg

import (
	addr "github.com/chenjianmei111/go-address"
	"github.com/chenjianmei111/go-state-types/abi"
	cid "github.com/ipfs/go-cid"
	"golang.org/x/xerrors"

	"github.com/chenjianmei111/specs-actors/v3/actors/builtin"
	"github.com/chenjianmei111/specs-actors/v3/actors/util/adt"
)

// DataCap is an integer number of bytes.
// We can introduce policy changes and replace this in the future.
type DataCap = abi.StoragePower

type State struct {
	// Root key holder multisig.
	// Authorize and remove verifiers.
	RootKey addr.Address

	// Verifiers authorize VerifiedClients.
	// Verifiers delegate their DataCap.
	Verifiers cid.Cid // HAMT[addr.Address]DataCap

	// VerifiedClients can add VerifiedClientData, up to DataCap.
	VerifiedClients cid.Cid // HAMT[addr.Address]DataCap
}

var MinVerifiedDealSize = abi.NewStoragePower(1 << 20)

// rootKeyAddress comes from genesis.
func ConstructState(store adt.Store, rootKeyAddress addr.Address) (*State, error) {
	emptyMapCid, err := adt.StoreEmptyMap(store, builtin.DefaultHamtBitwidth)
	if err != nil {
		return nil, xerrors.Errorf("failed to create empty map: %w", err)
	}

	return &State{
		RootKey:         rootKeyAddress,
		Verifiers:       emptyMapCid,
		VerifiedClients: emptyMapCid,
	}, nil
}
