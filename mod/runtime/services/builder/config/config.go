// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Berachain Foundation
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

package config

import (
	"time"

	"github.com/berachain/beacon-kit/mod/primitives"
)

const (
	// defaultLocalBuilderEnabled is the default value for local builder.
	defaultLocalBuilderEnabled = true
	// defaultLocalBuildPayloadTimeout is the default value for local build
	// payload timeout.
	defaultLocalBuildPayloadTimeout = 2500 * time.Millisecond
)

// Builder is the configuration for the payload builder.
//
//nolint:lll // struct tags.
type Config struct {
	// Suggested FeeRecipient is the address that will receive the transaction
	// fees
	// produced by any blocks from this node.
	SuggestedFeeRecipient primitives.ExecutionAddress `mapstructure:"suggested-fee-recipient"`

	// Graffiti is the string that will be included in the
	// graffiti field of the beacon block.
	Graffiti string `mapstructure:"graffiti"`

	// LocalBuilderEnabled determines if the local builder is enabled.
	LocalBuilderEnabled bool `mapstructure:"local-builder-enabled"`

	// LocalBuildPayloadTimeout is the timeout parameter for local build
	// payload. This should match, or be slightly less than the configured
	// timeout on your
	// execution client. It also must be less than timeout_proposal in the
	// CometBFT configuration.
	LocalBuildPayloadTimeout time.Duration `mapstructure:"local-build-payload-timeout"`
}

// DefaultBuilderConfig returns the default fork configuration.
func DefaultBuilderConfig() Config {
	return Config{
		SuggestedFeeRecipient:    primitives.ExecutionAddress{},
		Graffiti:                 "",
		LocalBuilderEnabled:      defaultLocalBuilderEnabled,
		LocalBuildPayloadTimeout: defaultLocalBuildPayloadTimeout,
	}
}