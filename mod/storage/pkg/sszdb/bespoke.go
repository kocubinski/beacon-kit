package sszdb

import (
	"math"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/types"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/common"
	pmath "github.com/berachain/beacon-kit/mod/primitives/pkg/math"
	ssz "github.com/ferranbt/fastssz"
)

type BespokeDBReader struct {
	*DB
}

func (d *BespokeDBReader) GetGenesisValidatorsRoot() (common.Root, error) {
	const parentNumFields = 16
	const fieldIndex = 0
	const length = 32

	gindex := powerTwo(ceilLog2(parentNumFields)) + fieldIndex
	bz, err := d.getNodeBytes(gindex, length)
	if err != nil {
		return common.Root{}, err
	}
	return common.Root(bz), nil
}

func (d *BespokeDBReader) GetSlot() (pmath.Slot, error) {
	const parentNumFields = 16
	const fieldIndex = 1
	const length = 8

	gindex := powerTwo(ceilLog2(parentNumFields)) + fieldIndex
	n, err := d.getNodeBytes(gindex, length)
	if err != nil {
		return 0, err
	}
	slot := ssz.UnmarshallUint64(n)
	return pmath.Slot(slot), nil
}

func (d *BespokeDBReader) GetFork() (*types.Fork, error) {
	const parentNumFields = 3
	const rootGindex = 18 // field index 2 in parent, 16 + 2 = 18

	depth := ceilLog2(parentNumFields)
	gindex := powerTwo(depth) * rootGindex

	f := &types.Fork{}
	// field 0
	bz, err := d.getNodeBytes(gindex, 4)
	if err != nil {
		return nil, err
	}
	copy(f.PreviousVersion[:], bz)

	// field 1
	gindex++
	bz, err = d.getNodeBytes(gindex, 4)
	if err != nil {
		return nil, err
	}
	copy(f.CurrentVersion[:], bz)

	// field 2
	gindex++
	bz, err = d.getNodeBytes(gindex, 8)
	if err != nil {
		return nil, err
	}
	f.Epoch = pmath.Epoch(ssz.UnmarshallUint64(bz))

	return f, nil
}

func (d *BespokeDBReader) GetLatestBlockHeader() (*types.BeaconBlockHeader, error) {
	const parentNumFields = 5
	const rootGindex = 19

	depth := ceilLog2(parentNumFields)
	gindex := powerTwo(depth) * rootGindex

	h := &types.BeaconBlockHeader{}
	// field 0
	bz, err := d.getNodeBytes(gindex, 8)
	if err != nil {
		return nil, err
	}
	h.Slot = ssz.UnmarshallUint64(bz)
	gindex++

	// field 1
	bz, err = d.getNodeBytes(gindex, 8)
	if err != nil {
		return nil, err
	}
	h.ProposerIndex = ssz.UnmarshallUint64(bz)
	gindex++

	// field 2
	bz, err = d.getNodeBytes(gindex, 32)
	if err != nil {
		return nil, err
	}
	copy(h.ParentBlockRoot[:], bz)
	gindex++

	// field 3
	bz, err = d.getNodeBytes(gindex, 32)
	if err != nil {
		return nil, err
	}
	copy(h.StateRoot[:], bz)
	gindex++

	// field 4
	bz, err = d.getNodeBytes(gindex, 32)
	if err != nil {
		return nil, err
	}
	copy(h.BodyRoot[:], bz)

	return h, nil
}

func (d *BespokeDBReader) GetBlockRoots() ([]common.Root, error) {
	const parentNumFields = 8192
	const rootGindex = 20

	depth := ceilLog2(parentNumFields) + 1 // +1 because of the length field
	gindex := powerTwo(depth) * rootGindex

	bz, err := d.getNodeBytes(2*rootGindex+1, 8)
	if err != nil {
		return nil, err
	}
	length := ssz.UnmarshallUint64(bz)

	var roots []common.Root
	for i := uint64(0); i < length; i++ {
		bz, err = d.getNodeBytes(gindex+i, 32)
		if err != nil {
			return nil, err
		}
		roots = append(roots, common.Root(bz))
	}
	return roots, nil
}

func floorLog2(n uint64) uint64 {
	return uint64(math.Floor(math.Log2(float64(n))))
}

func ceilLog2(n uint64) uint64 {
	return uint64(math.Ceil(math.Log2(float64(n))))
}

func powerTwo(n uint64) uint64 {
	return uint64(math.Pow(2, float64(n)))
}

func nextPowerOfTwo(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return uint64(v)
}

func prevPowerOfTwo(v uint64) uint64 {
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	return uint64(v) - (v >> 1)
}
