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

package engineprimitives

import (
	"github.com/berachain/beacon-kit/mod/primitives/pkg/common"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/constants"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/math"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/ssz"
)

// Withdrawal represents a validator withdrawal from the consensus layer.
//
//go:generate go run github.com/ferranbt/fastssz/sszgen -path withdrawal.go -objs Withdrawal -include ../../../primitives/pkg/math,../../../primitives/pkg/common,$GETH_PKG_INCLUDE/common,$GETH_PKG_INCLUDE/common/hexutil -output withdrawal.ssz.go
type Withdrawal struct {
	// Index is the unique identifier for the withdrawal.
	Index math.U64 `json:"index"`
	// Validator is the index of the validator initiating the withdrawal.
	Validator math.ValidatorIndex `json:"validatorIndex"`
	// Address is the execution address where the withdrawal will be sent.
	// It has a fixed size of 20 bytes.
	Address common.ExecutionAddress `json:"address"        ssz-size:"20"`
	// Amount is the amount of Gwei to be withdrawn.
	Amount math.Gwei `json:"amount"`
}

// Equals returns true if the Withdrawal is equal to the other.
func (w *Withdrawal) Equals(other *Withdrawal) bool {
	return w.Index == other.Index &&
		w.Validator == other.Validator &&
		w.Address == other.Address &&
		w.Amount == other.Amount
}

// GetIndex returns the unique identifier for the withdrawal.
func (w *Withdrawal) GetIndex() math.U64 {
	return w.Index
}

// GetValidatorIndex returns the index of the validator initiating the
// withdrawal.
func (w *Withdrawal) GetValidatorIndex() math.ValidatorIndex {
	return w.Validator
}

// GetAddress returns the execution address where the withdrawal will be sent.
func (w *Withdrawal) GetAddress() common.ExecutionAddress {
	return w.Address
}

// GetAmount returns the amount of Gwei to be withdrawn.
func (w *Withdrawal) GetAmount() math.Gwei {
	return w.Amount
}

// Withdrawals represents a slice of withdrawals.
type Withdrawals []*Withdrawal

// HashTreeRoot returns the hash tree root of the Withdrawals list.
func (w Withdrawals) HashTreeRoot() (common.Root, error) {
	// TODO: read max withdrawals from the chain spec.
	merkleizer := ssz.NewMerkleizer[common.ChainSpec, [32]byte, *Withdrawal]()
	return merkleizer.MerkleizeListComposite(
		w,
		constants.MaxWithdrawalsPerPayload,
	)
}
