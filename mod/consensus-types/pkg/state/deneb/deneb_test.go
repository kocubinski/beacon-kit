// SPDX-License-Identifier: MIT
//
// # Copyright (c) 2024 Berachain Foundation
//
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the "Software"), to deal in the Software without
// restriction, including without limitation the rights to use,
// copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following
// conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
//

package deneb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/berachain/beacon-kit/mod/primitives/pkg/common"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/math"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/state/deneb"
	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/types"
	"github.com/berachain/beacon-kit/mod/primitives"
	ssz "github.com/ferranbt/fastssz"
	"github.com/stretchr/testify/require"
)

// emptyValidBeaconState generates a valid beacon state for the Deneb.
func emptyValidBeaconState() *deneb.BeaconState {
	var byteArray [256]byte
	return &deneb.BeaconState{
		BlockRoots:  []primitives.Root{},
		StateRoots:  []primitives.Root{},
		Validators:  []*types.Validator{},
		Balances:    []uint64{},
		RandaoMixes: []primitives.Bytes32{},
		Slashings:   []uint64{},
		LatestExecutionPayloadHeader: &types.ExecutionPayloadHeaderDeneb{
			LogsBloom: byteArray[:],
			ExtraData: []byte{},
		},
	}
}

// TODO for fuzzing
func generateValidBeaconState() *deneb.BeaconState {
	return emptyValidBeaconState()
}

func arbitraryValidBeaconState() *deneb.BeaconState {
	var bytes256 [256]byte
	return &deneb.BeaconState{
		GenesisValidatorsRoot: primitives.Root{},
		Slot:                  1,
		Fork: &types.Fork{
			PreviousVersion: common.Version{},
			CurrentVersion:  common.Version{},
			Epoch:           0,
		},
		LatestBlockHeader: nil,
		BlockRoots:        []primitives.Root{{1, 2, 3}},
		StateRoots:        []primitives.Root{{4, 5, 6}},
		Eth1Data:          nil,
		Eth1DepositIndex:  0,
		LatestExecutionPayloadHeader: &types.ExecutionPayloadHeaderDeneb{
			ParentHash:       common.ExecutionHash{},
			FeeRecipient:     common.ExecutionAddress{},
			StateRoot:        primitives.Bytes32{},
			ReceiptsRoot:     primitives.Bytes32{},
			LogsBloom:        bytes256[:],
			Random:           primitives.Bytes32{},
			Number:           0,
			GasLimit:         0,
			GasUsed:          0,
			Timestamp:        0,
			ExtraData:        nil,
			BaseFeePerGas:    math.Wei{},
			BlockHash:        common.ExecutionHash{},
			TransactionsRoot: primitives.Root{},
			WithdrawalsRoot:  primitives.Root{},
			BlobGasUsed:      0,
			ExcessBlobGas:    0,
		},
		Validators:                   nil,
		Balances:                     []uint64{1, 2, 3},
		RandaoMixes:                  []primitives.Bytes32{{1, 2, 3}},
		NextWithdrawalIndex:          0,
		NextWithdrawalValidatorIndex: 0,
		Slashings:                    []uint64{7, 8, 9},
		TotalSlashing:                0,
	}
}

func devBeaconState() (*deneb.BeaconState, error) {
	bz, err := os.ReadFile("/tmp/beacon.ssz")
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

func TestBeaconStateMarshalUnmarshalSSZ(t *testing.T) {
	state := emptyValidBeaconState()

	data, fastSSZMarshalErr := state.MarshalSSZ()
	require.NoError(t, fastSSZMarshalErr)
	require.NotNil(t, data)

	newState := &deneb.BeaconState{}
	err := newState.UnmarshalSSZ(data)
	require.NoError(t, err)

	require.Equal(t, state, newState)

	// Check if the state size is greater than 0
	require.Positive(t, state.SizeSSZ())
}

func TestHashTreeRoot(t *testing.T) {
	state := arbitraryValidBeaconState()
	_, err := state.HashTreeRoot()
	require.NoError(t, err)
}

func TestGetTree(t *testing.T) {
	state, err := devBeaconState()
	require.NoError(t, err)
	//state := emptyValidBeaconState()
	rootHash, err := state.HashTreeRoot()
	require.NoError(t, err)
	fmt.Printf("hash=%x\n", rootHash)

	tree, err := state.GetTree()
	require.NoError(t, err)
	require.NotNil(t, tree)
	f, err := os.Create("/tmp/tree.dot")
	require.NoError(t, err)
	tree.Draw(f)
	pn := func(tr *ssz.Node, i int) {
		n, err := tr.Get(i)
		require.NoError(t, err)
		fmt.Printf("gidx=%d hash=%x\n", i, n.Hash())
	}
	pn(tree, 1)
	pn(tree, 5)
	pn(tree, 11)
	// pn(tree, 3489660928)
}

func TestBeaconState_UnmarshalSSZ_Error(t *testing.T) {
	state := &deneb.BeaconState{}
	err := state.UnmarshalSSZ([]byte{0x01, 0x02, 0x03}) // Invalid data
	require.ErrorIs(t, err, ssz.ErrSize)
}

func TestBeaconState_MarshalSSZTo(t *testing.T) {
	state := emptyValidBeaconState()
	data, err := state.MarshalSSZ()
	require.NoError(t, err)
	require.NotNil(t, data)

	var buf []byte
	buf, err = state.MarshalSSZTo(buf)
	require.NoError(t, err)

	// The two byte slices should be equal
	require.Equal(t, data, buf)
}

func TestBeaconState_MarshalSSZFields(t *testing.T) {
	state := generateValidBeaconState()

	// Test BlockRoots field
	state.BlockRoots = make([]primitives.Root, 8193) // Exceeding the limit
	_, err := state.MarshalSSZ()
	require.Error(t, err)
	state.BlockRoots = make([]primitives.Root, 8192) // Within the limit
	_, err = state.MarshalSSZ()
	require.NoError(t, err)

	// Test StateRoots field
	state.StateRoots = make([]primitives.Root, 8193) // Exceeding the limit
	_, err = state.MarshalSSZ()
	require.Error(t, err)
	state.StateRoots = make([]primitives.Root, 8192) // Within the limit
	_, err = state.MarshalSSZ()
	require.NoError(t, err)

	// Test LatestExecutionPayloadHeader field
	state.LatestExecutionPayloadHeader = &types.ExecutionPayloadHeaderDeneb{
		LogsBloom: make([]byte, 256), // Initialize LogsBloom with 256 bytes
	}
	_, err = state.MarshalSSZ()
	require.NoError(t, err)

	// Test RandaoMixes field
	state.RandaoMixes = make([]primitives.Bytes32, 65537) // Exceeding the limit
	_, err = state.MarshalSSZ()
	require.Error(t, err)
	state.RandaoMixes = make([]primitives.Bytes32, 65536) // Within the limit
	_, err = state.MarshalSSZ()
	require.NoError(t, err)
}
