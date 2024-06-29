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

package merkleizer

// Merkleizer can be used for merkleizing SSZ types.
type Merkleizer[
	RootT ~[32]byte, T SSZObject[RootT],
] interface {
	MerkleizeBasic(value T) (RootT, error)
	MerkleizeVectorBasic(value []T) (RootT, error)
	MerkleizeListBasic(value []T, limit ...uint64) (RootT, error)
	MerkleizeVectorComposite(value []T) (RootT, error)
	MerkleizeListComposite(value []T, limit ...uint64) (RootT, error)
	MerkleizeByteSlice(value []byte) (RootT, error)
	Merkleize(chunks []RootT, limit ...uint64) (RootT, error)

	// TODO: Move to a separate Merkleizer type for container(s).
	MerkleizeContainer(
		value SSZObject[RootT],
	) (RootT, error)
}
