package nv3

import (
	"context"

	"github.com/chenjianmei111/go-address"
	"github.com/chenjianmei111/go-state-types/abi"
	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"

	miner "github.com/chenjianmei111/specs-actors/actors/builtin/miner"
	"github.com/chenjianmei111/specs-actors/actors/states"
)

type minerMigrator struct {
}

func (m *minerMigrator) MigrateState(ctx context.Context, store cbor.IpldStore, head cid.Cid, _ abi.ChainEpoch, _ address.Address, _ *states.Tree) (cid.Cid, error) {
	var st miner.State
	if err := store.Get(ctx, head, &st); err != nil {
		return cid.Undef, err
	}

	// TODO: https://github.com/chenjianmei111/specs-actors/issues/1177
	//  - repair broken partitions, deadline info:
	//  - fix power actor claim with any power delta

	newHead, err := store.Put(ctx, &st)
	return newHead, err
}