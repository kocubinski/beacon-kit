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

package merkle

import (
	"runtime"
	"unsafe"

	"github.com/berachain/beacon-kit/mod/primitives/pkg/math"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/merkle/zero"
	"github.com/prysmaticlabs/gohashtree"
	"golang.org/x/sync/errgroup"
)

const (
	// MinParallelizationSize is the minimum size of the input list that
	// should be hashed using the default method. If the input list is smaller
	// than this size, the overhead of parallelizing the hashing process is.
	//
	// TODO: This value is arbitrary and should be benchmarked to find the
	// optimal value.
	MinParallelizationSize = 5000
	// two is a constant to make the linter happy.
	two = 2
)

// NewRootWithMaxLeaves constructs a Merkle tree root from a set of.
func NewRootWithMaxLeaves[U64T U64[U64T], RootT ~[32]byte](
	leaves []RootT,
	length U64T,
) (RootT, error) {
	return NewRootWithDepth(
		leaves, math.U64(length).NextPowerOfTwo().ILog2Ceil(),
	)
}

// NewRootWithDepth constructs a Merkle tree root from a set of leaves.
func NewRootWithDepth[RootT ~[32]byte](
	leaves []RootT,
	depth uint8,
) (RootT, error) {
	// Return zerohash at depth
	if len(leaves) == 0 {
		return zero.Hashes[depth], nil
	}

	// Preallocate a single buffer large enough for the maximum layer size
	// TODO: It seems that BuildParentTreeRoots has different behaviour
	// when we pass leaves in directly.
	buffer := make([]RootT, (len(leaves)+1)/two)

	var err error
	for i := range depth {
		layerLen := len(leaves)
		if layerLen%two == 1 {
			leaves = append(leaves, zero.Hashes[i])
		}

		newLayerSize := (layerLen + 1) / two
		if err = BuildParentTreeRoots(buffer[:newLayerSize], leaves); err != nil {
			return zero.Hashes[depth], err
		}
		leaves, buffer = buffer[:newLayerSize], leaves
	}
	if len(leaves) != 1 {
		return zero.Hashes[depth], nil
	}
	return leaves[0], nil
}

// BuildParentTreeRoots calls BuildParentTreeRootsWithNRoutines with the
// number of routines set to runtime.GOMAXPROCS(0)-1.
func BuildParentTreeRoots[RootT ~[32]byte](
	outputList, inputList []RootT,
) error {
	err := BuildParentTreeRootsWithNRoutines(
		//#nosec:G103 // on purpose.
		*(*[][32]byte)(unsafe.Pointer(&outputList)),
		//#nosec:G103 // on purpose.
		*(*[][32]byte)(unsafe.Pointer(&inputList)),
		runtime.GOMAXPROCS(0)-1,
	)

	// Convert out back to []RootT using unsafe pointer cas
	return err
}

// BuildParentTreeRootsWithNRoutines optimizes hashing of a list of roots
// using CPU-specific vector instructions and parallel processing. This
// method adapts to the host machine's hardware for potential performance
// gains over sequential hashing.
//
// TODO: We do not use generics here due to the gohashtree library not
// supporting
// generics.
func BuildParentTreeRootsWithNRoutines(
	outputList, inputList [][32]byte, n int,
) error {
	// Validate input list length.
	inputLength := len(inputList)
	if inputLength%2 != 0 {
		return ErrOddLengthTreeRoots
	}
	// Build output variables
	outputLength := inputLength / two

	// If the input list is small, hash it using the default method since
	// the overhead of parallelizing the hashing process is not worth it.
	if inputLength < MinParallelizationSize {
		//#nosec:G103 // used of unsafe calls should be audited.
		return gohashtree.Hash(outputList, inputList)
	}

	// Otherwise parallelize the hashing process for large inputs.
	// Take the max(n, 1) to prevent division by 0.
	groupSize := inputLength / (two * max(n, 1))
	twiceGroupSize := two * groupSize
	eg := new(errgroup.Group)

	// if n is 0 the parallelization is disabled and the whole inputList is
	// hashed in the main goroutine at the end of this function.
	for j := 0; j <= n; j++ {
		eg.Go(func() error {
			// inputList:  [-------------------2*groupSize-------------------]
			//              ^                  ^                    ^        ^
			//              |                  |                    |        |
			// j*2*groupSize   (j+1)*2*groupSize    (j+2)*2*groupSize   End
			//
			// outputList: [---------groupSize---------]
			//              ^                         ^
			//              |                         |
			//             j*groupSize         (j+1)*groupSize
			//
			// Each goroutine processes a segment of inputList that is twice as
			// large as the segment it fills in outputList. This is because the
			// hash
			// operation reduces the
			// size of the input by half.
			// Define the segment of the inputList each goroutine will process.
			segmentStart := j * twiceGroupSize
			segmentEnd := min((j+1)*twiceGroupSize, inputLength)

			return gohashtree.Hash(
				outputList[j*groupSize:min((j+1)*groupSize, outputLength)],
				inputList[segmentStart:segmentEnd],
			)
		})
	}

	// Wait for all goroutines to complete.
	return eg.Wait()
}
