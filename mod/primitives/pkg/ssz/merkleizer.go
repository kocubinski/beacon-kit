// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2024, Berachain Foundation. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package ssz

import (
	"reflect"

	"github.com/berachain/beacon-kit/mod/errors"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/bytes"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/math"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/merkle"
)

// merkleizer can be used for merkleizing SSZ types.
type merkleizer[
	SpecT any, RootT ~[32]byte, T Basic[SpecT, RootT],
] struct {
	hasher      *merkle.Hasher[RootT]
	bytesBuffer bytes.Buffer[RootT]
}

// NewMerkleizer creates a new merkleizer with reusable buffers.
func NewMerkleizer[
	SpecT any, RootT ~[32]byte, T Basic[SpecT, RootT],
]() Merkleizer[SpecT, RootT, T] {
	return &merkleizer[SpecT, RootT, T]{
		hasher: merkle.NewHasher(
			bytes.NewReusableBuffer[RootT](),
			merkle.BuildParentTreeRoots[RootT],
		),
		bytesBuffer: bytes.NewReusableBuffer[RootT](),
	}
}

// MerkleizeBasic hashes the packed value and returns the HTR.
func (m *merkleizer[SpecT, RootT, T]) MerkleizeBasic(
	value T,
) (RootT, error) {
	return m.MerkleizeVecBasic([]T{value})
}

// MerkleizeVecBasic implements the SSZ merkleization algorithm
// for a vector of basic types.
func (m *merkleizer[SpecT, RootT, T]) MerkleizeVecBasic(
	value []T,
) (RootT, error) {
	packed, err := Pack[SpecT](value)
	if err != nil {
		return [32]byte{}, err
	}
	return m.Merkleize(packed)
}

// MerkleizeListBasic implements the SSZ merkleization algorithm for a list of
// basic types.
func (m *merkleizer[SpecT, RootT, T]) MerkleizeListBasic(
	value []T,
	limit ...uint64,
) (RootT, error) {
	packed, err := Pack[SpecT](value)
	if err != nil {
		return [32]byte{}, err
	}

	var effectiveLimit uint64
	if len(limit) > 0 {
		effectiveLimit = limit[0]
	} else {
		effectiveLimit = uint64(len(packed))
	}

	root, err := m.Merkleize(
		packed, ChunkCountBasicList[SpecT](value, effectiveLimit),
	)
	if err != nil {
		return [32]byte{}, err
	}
	return merkle.MixinLength(root, uint64(len(value))), nil
}

// TODO: MerkleizeBitlist

// MerkleizeContainer implements the SSZ merkleization algorithm for a
// container.
//
// TODO: Make a separate merkleizer for container and list of containers.
func (m *merkleizer[SpecT, RootT, T]) MerkleizeContainer(
	value Container[SpecT, RootT], _ ...SpecT,
) (RootT, error) {
	rValue := reflect.ValueOf(value)
	if rValue.Kind() == reflect.Ptr {
		rValue = rValue.Elem()
	}
	numFields := rValue.NumField()
	htrs := make([]RootT, numFields)
	var err error
	for i := range numFields {
		fieldValue := rValue.Field(i)
		if !fieldValue.CanInterface() {
			return RootT{}, errors.Newf(
				"cannot interface with field %v",
				fieldValue,
			)
		}

		// TODO: handle different types.
		field, ok := fieldValue.Interface().(Basic[SpecT, RootT])
		if !ok {
			return RootT{}, errors.Newf(
				"field %d does not implement Hashable",
				i,
			)
		}
		htrs[i], err = field.HashTreeRoot( /*args...*/ )
		if err != nil {
			return RootT{}, err
		}
	}
	return m.Merkleize(htrs)
}

// MerkleizeVecComposite implements the SSZ merkleization algorithm for a vector
// of composite types.
func (m *merkleizer[SpecT, RootT, T]) MerkleizeVecComposite(
	value []T,
) (RootT, error) {
	var (
		err  error
		htrs = m.bytesBuffer.Get(len(value))
	)

	for i, el := range value {
		htrs[i], err = el.HashTreeRoot()
		if err != nil {
			return RootT{}, err
		}
	}
	return m.Merkleize(htrs)
}

// MerkleizeListComposite implements the SSZ merkleization algorithm for a list
// of composite types.
func (m *merkleizer[SpecT, RootT, T]) MerkleizeListComposite(
	value []T,
	limit ...uint64,
) (RootT, error) {
	var (
		err  error
		htrs = m.bytesBuffer.Get(len(value))
	)

	for i, el := range value {
		htrs[i], err = el.HashTreeRoot()
		if err != nil {
			return RootT{}, err
		}
	}

	var effectiveLimit uint64
	if len(limit) > 0 {
		effectiveLimit = limit[0]
	} else {
		effectiveLimit = uint64(len(value))
	}

	root, err := m.Merkleize(
		htrs, ChunkCountCompositeList[SpecT](value, effectiveLimit),
	)
	if err != nil {
		return RootT{}, err
	}

	return merkle.MixinLength(root, uint64(len(value))), nil
}

// MerkleizeByteSlice hashes a byteslice by chunkifying it and returning the
// corresponding HTR as if it were a fixed vector of bytes of the given length.
func (m *merkleizer[SpecT, RootT, T]) MerkleizeByteSlice(
	input []byte,
) (RootT, error) {
	chunks, numChunks, err := PartitionBytes[RootT](input)
	if err != nil {
		return RootT{}, err
	}
	return m.Merkleize(
		chunks,
		numChunks,
	)
}

// Merkleize hashes a list of chunks and returns the HTR of the list of.
//
// merkleize(chunks, limit=None): Given ordered BYTES_PER_CHUNK-byte chunks,
// merkleize the chunks, and return the root: The merkleization depends on the
// effective input, which must be padded/limited:
//
//	if no limit:
//		pad the chunks with zeroed chunks to next_pow_of_two(len(chunks))
//
// (virtually for memory efficiency).
//
//	if limit >= len(chunks):
//		pad the chunks with zeroed chunks to next_pow_of_two(limit) (virtually for
//
// memory efficiency).
//
//	if limit < len(chunks):
//		do not merkleize, input exceeds limit. Raise an error instead.
//	  Then, merkleize the chunks (empty input is padded to 1 zero chunk):
//	 If 1 chunk: the root is the chunk itself.
//	If > 1 chunks: merkleize as binary tree.
func (m *merkleizer[SpecT, RootT, T]) Merkleize(
	chunks []RootT,
	limit ...uint64,
) (RootT, error) {
	var (
		effectiveLimit  math.U64
		effectiveChunks []RootT
		lenChunks       = uint64(len(chunks))
	)

	switch {
	case len(limit) == 0:
		effectiveLimit = math.U64(lenChunks).NextPowerOfTwo()
	case limit[0] >= lenChunks:
		effectiveLimit = math.U64(limit[0]).NextPowerOfTwo()
	default:
		if limit[0] < lenChunks {
			return RootT{}, errors.New("input exceeds limit")
		}
		effectiveLimit = math.U64(limit[0])
	}

	effectiveChunks = PadTo(chunks, effectiveLimit)
	if len(effectiveChunks) == 1 {
		return effectiveChunks[0], nil
	}

	return m.hasher.NewRootWithMaxLeaves(effectiveChunks, effectiveLimit)
}
