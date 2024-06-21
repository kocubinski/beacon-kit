package sszdb

import (
	"fmt"
	"math"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/types"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/common"
	pmath "github.com/berachain/beacon-kit/mod/primitives/pkg/math"
	ssz "github.com/ferranbt/fastssz"
)

// versioning

func (d *DB) GetGenesisValidatorsRoot() (common.Root, error) {
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

func (d *DB) GetSlot() (pmath.Slot, error) {
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

func (d *DB) GetFork() (*types.Fork, error) {
	const parentNumFields = 3
	const rootGindex = 18 // field index 2 in parent, 16 + 2 = 18

	depth := ceilLog2(parentNumFields)
	gindex := powerTwo(depth) * rootGindex

	f := &types.Fork{}
	bz, err := d.getNodeBytes(gindex, 4)
	if err != nil {
		return nil, err
	}
	copy(f.PreviousVersion[:], bz)
	gindex++
	bz, err = d.getNodeBytes(gindex, 4)
	if err != nil {
		return nil, err
	}
	copy(f.CurrentVersion[:], bz)
	gindex++
	bz, err = d.getNodeBytes(gindex, 8)
	if err != nil {
		return nil, err
	}
	f.Epoch = pmath.Epoch(ssz.UnmarshallUint64(bz))
	return f, nil
}

func (d *DB) GetLatestBlockHeader() (*types.BeaconBlockHeader, error) {
	const parentNumFields = 5
	const rootGindex = 19

	relativeHeight := ceilLog2(nextPowerOfTwo(parentNumFields) * 2)               // 4
	totalHeight := ceilLog2(nextPowerOfTwo(parentNumFields*2)*2) + relativeHeight // 8
	leafOffset := powerTwo(totalHeight - relativeHeight - 1)
	gindex := powerTwo(totalHeight-1) + leafOffset // 64 + 8 = 72

	fmt.Println("gindex", gindex)

	return nil, nil
}

// registry

func (d *DB) AddValidator(v *types.Validator) {
	d.monolith.Validators = append(d.monolith.Validators, v)
}

func (d *DB) UpdateValidatorAtIndex(index int, v *types.Validator) {
	d.monolith.Validators[index] = v
}

func (d *DB) RemoveValidatorAtIndex(index int) {
}

func (d *DB) GetValidators() []*types.Validator {
	return d.monolith.Validators
}

// util

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
