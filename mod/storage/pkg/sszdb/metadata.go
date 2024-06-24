package sszdb

import (
	"strconv"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/types"
)

// this file contains functions to parse object paths into a gindex.
// it would be generated by tool which ingests plain old go structs and spits out the gindex calculation

type schemaNode struct {
	gindex   uint64
	order    uint64
	height   uint64
	size     uint64 // size of the object in bytes
	list     bool
	vector   uint64
	children lookupTable
}

type lookupTable map[string]*schemaNode
type objectPath []string

// getSchemaNode returns the schema node for a given object path
// and sets the gindex for the node
func getSchemaNode(path objectPath) *schemaNode {
	// todo: pass in root for partial traversal from subtree
	prev := beaconStateLookup["."]
	var (
		node *schemaNode
		ok   bool
	)
	for _, p := range path {
		switch {
		case p == "__len__": // "__len__" reserved
			if !prev.list {
				return nil
			}
			return &schemaNode{
				gindex: 2*prev.gindex + 1,
				size:   8,
			}
		case prev.list || prev.vector > 0:
			// TODO this incorrectly asssumes all list elements are 32 bytes, that is, either container types
			// or basic types with length <= 32.  in other words it doesn't handle packing as specified in ssz
			i, err := strconv.ParseUint(p, 10, 64)
			if err != nil {
				return nil
			}
			node = &schemaNode{
				gindex:   powerTwo(prev.height-1)*prev.gindex + i,
				size:     prev.size,     // list specifies the element length
				children: prev.children, // children specifies the element schema
				height:   ceilLog2(uint64(len(prev.children))) + 1,
			}
			// special case for vector of bytes
			if prev.vector > 0 && prev.size == 1 {
				node.size = prev.vector
			}

		default:
			node, ok = prev.children[p]
			if !ok {
				return nil
			}
			node.gindex = powerTwo(prev.height-1)*prev.gindex + node.order
		}
		prev = node
	}
	return node
}

// beaconStateLookup is a lookup table for the beacon state schema
// it would be generated by a tool that ingests the beacon state struct
var beaconStateLookup = map[string]*schemaNode{
	".": {
		gindex: 1, // root node gindex=1
		height: 5,
		children: lookupTable{
			"genesis_validators_root": {
				size:  32,
				order: 0,
			},
			"slot": {
				size:  8,
				order: 1,
			},
			"fork": {
				height: 3,
				order:  2,
				children: lookupTable{
					"previous_version": {
						order: 0,
						size:  4,
					},
					"current_version": {
						order: 1,
						size:  4,
					},
					"epoch": {
						order: 2,
						size:  8,
					},
				},
			},
			"latest_block_header": {
				height: 4,
				order:  3,
				children: lookupTable{
					"slot": {
						order: 0,
						size:  8,
					},
					"proposer_index": {
						order: 1,
						size:  8,
					},
					"parent_block_root": {
						order: 2,
						size:  32,
					},
					"state_root": {
						order: 3,
						size:  32,
					},
					"body_root": {
						order: 4,
						size:  32,
					},
				},
			},
			"block_roots": {
				// calc height for perfect binary tree with 8192 leaves, +1 for length node
				height: ceilLog2(nextPowerOfTwo(8192)) + 2,
				order:  4,
				list:   true,
				size:   32,
			},
			"validators": {
				height: ceilLog2(nextPowerOfTwo(1099511627776)) + 2,
				order:  9,
				list:   true,
				size:   uint64((&types.Validator{}).SizeSSZ()),
				children: lookupTable{
					"pubkey": {
						order:  0,
						size:   1,
						vector: 48,
						height: 2,
					},
					"withdrawal_credentials": {
						order: 1,
						size:  32,
					},
					"effective_balance": {
						order: 2,
						size:  8,
					},
					"slashed": {
						order: 3,
						size:  1,
					},
					"activation_eligibility_epoch": {
						order: 4,
						size:  8,
					},
					"activation_epoch": {
						order: 5,
						size:  8,
					},
					"exit_epoch": {
						order: 6,
						size:  8,
					},
					"withdrawable_epoch": {
						order: 7,
						size:  8,
					},
				},
			},
		},
	},
}
