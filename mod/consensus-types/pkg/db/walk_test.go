package db_test

import (
	"os"
	"testing"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/db"
	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/state/deneb"
	"github.com/stretchr/testify/require"
)

func testBeaconState() (*deneb.BeaconState, error) {
	bz, err := os.ReadFile("./testdata/beacon.ssz")
	if err != nil {
		return nil, err
	}
	state := &deneb.BeaconState{}
	err = state.UnmarshalSSZ(bz)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func TestBeaconState_Storage(t *testing.T) {
	state, err := testBeaconState()
	require.NoError(t, err)
	rootHash, err := state.HashTreeRoot()
	require.NoError(t, err)
	require.NotNil(t, rootHash)

	root, err := db.Tree(state)
	require.NoError(t, err)
	require.NotNil(t, root)
}
