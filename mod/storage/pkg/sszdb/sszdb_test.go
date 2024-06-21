package sszdb_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/state/deneb"
	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/types"
	"github.com/berachain/beacon-kit/mod/storage/pkg/sszdb"
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

func TestDB_Save(t *testing.T) {
	dir := t.TempDir()

	t.Logf("temp dir: %s", dir)
	dbPath := dir + "/sszdb.db"

	db, err := sszdb.New(sszdb.Config{Path: dbPath})
	require.NoError(t, err)
	st, err := testBeaconState()
	require.NoError(t, err)

	// remarshal
	_, err = st.MarshalSSZ()
	require.NoError(t, err)

	err = db.SaveMonolith(st)
	require.NoError(t, err)

	loaded := &deneb.BeaconState{}
	require.NoError(t, db.Load(loaded))

	require.NoError(t, db.Close())
}

func TestDB_Bespoke(t *testing.T) {
	dir := t.TempDir() + "/sszdb.db"
	db, err := sszdb.New(sszdb.Config{Path: dir})
	require.NoError(t, err)
	beacon, err := testBeaconState()
	require.NoError(t, err)

	beacon.GenesisValidatorsRoot = [32]byte{7, 7, 7, 7}
	beacon.Slot = 777
	beacon.Fork = &types.Fork{
		PreviousVersion: [4]byte{1, 2, 3, 4},
		CurrentVersion:  [4]byte{5, 6, 7, 8},
		Epoch:           123,
	}
	err = db.SaveMonolith(beacon)
	require.NoError(t, err)

	bz, err := db.GetGenesisValidatorsRoot()
	require.NoError(t, err)
	require.True(t, bytes.Equal(bz[:], beacon.GenesisValidatorsRoot[:]))

	slot, err := db.GetSlot()
	require.NoError(t, err)
	require.Equal(t, beacon.Slot, slot)

	fork, err := db.GetFork()
	require.NoError(t, err)
	require.Equal(t, beacon.Fork, fork)
}
