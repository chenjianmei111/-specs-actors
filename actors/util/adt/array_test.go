package adt_test

import (
	"context"
	"testing"

	"github.com/chenjianmei111/go-address"
	"github.com/stretchr/testify/require"

	"github.com/chenjianmei111/specs-actors/v3/actors/util/adt"
	"github.com/chenjianmei111/specs-actors/v3/support/mock"
)

func TestArrayNotFound(t *testing.T) {
	rt := mock.NewBuilder(context.Background(), address.Undef).Build(t)
	store := adt.AsStore(rt)
	arr, err := adt.MakeEmptyArray(store, 3)
	require.NoError(t, err)

	found, err := arr.Get(7, nil)
	require.NoError(t, err)
	require.False(t, found)
}
