package sszdb

import (
	"fmt"
	"strconv"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/types"
	"github.com/berachain/beacon-kit/mod/primitives/pkg/common"
	pmath "github.com/berachain/beacon-kit/mod/primitives/pkg/math"
	ssz "github.com/ferranbt/fastssz"
)

type MetadataDB struct {
	*DB
}

func (d *MetadataDB) getLeafBytes(path objectPath) ([]byte, error) {
	schemaNode := getSchemaNode(path)
	if schemaNode == nil {
		return nil, fmt.Errorf("path %v not found", path)
	}
	return d.getNodeBytes(schemaNode.gindex, schemaNode.length)
}

func (d *MetadataDB) GetGenesisValidatorsRoot() (common.Root, error) {
	path := objectPath{"genesis_validators_root"}
	bz, err := d.getLeafBytes(path)
	if err != nil {
		return common.Root{}, err
	}
	return common.Root(bz), nil
}

func (d *MetadataDB) GetSlot() (pmath.Slot, error) {
	path := objectPath{"slot"}
	n, err := d.getLeafBytes(path)
	if err != nil {
		return 0, err
	}
	slot := ssz.UnmarshallUint64(n)
	return pmath.Slot(slot), nil
}

func (d *MetadataDB) GetFork() (*types.Fork, error) {
	f := &types.Fork{}
	bz, err := d.getLeafBytes(objectPath{"fork", "previous_version"})
	if err != nil {
		return nil, err
	}
	copy(f.PreviousVersion[:], bz)

	bz, err = d.getLeafBytes(objectPath{"fork", "current_version"})
	if err != nil {
		return nil, err
	}
	copy(f.CurrentVersion[:], bz)

	bz, err = d.getLeafBytes(objectPath{"fork", "epoch"})
	if err != nil {
		return nil, err
	}
	f.Epoch = pmath.Epoch(ssz.UnmarshallUint64(bz))

	return f, nil
}

func (d *MetadataDB) GetLatestBlockHeader() (*types.BeaconBlockHeader, error) {
	bh := &types.BeaconBlockHeader{}
	bz, err := d.getLeafBytes(objectPath{"latest_block_header", "slot"})
	if err != nil {
		return nil, err
	}
	bh.Slot = ssz.UnmarshallUint64(bz)

	bz, err = d.getLeafBytes(objectPath{"latest_block_header", "proposer_index"})
	if err != nil {
		return nil, err
	}
	bh.ProposerIndex = ssz.UnmarshallUint64(bz)

	bz, err = d.getLeafBytes(objectPath{"latest_block_header", "parent_block_root"})
	if err != nil {
		return nil, err
	}
	copy(bh.ParentBlockRoot[:], bz)

	bz, err = d.getLeafBytes(objectPath{"latest_block_header", "state_root"})
	if err != nil {
		return nil, err
	}
	copy(bh.StateRoot[:], bz)

	bz, err = d.getLeafBytes(objectPath{"latest_block_header", "body_root"})
	if err != nil {
		return nil, err
	}
	copy(bh.BodyRoot[:], bz)

	return bh, nil
}

func (d *MetadataDB) GetBlockRoots() ([]common.Root, error) {
	path := objectPath{"block_roots", "__len__"}
	schemaNode := getSchemaNode(path)
	bz, err := d.getNodeBytes(schemaNode.gindex, schemaNode.length)
	if err != nil {
		return nil, err
	}

	length := ssz.UnmarshallUint64(bz)
	roots := make([]common.Root, length)
	for i := uint64(0); i < length; i++ {
		path = objectPath{"block_roots", strconv.FormatInt(int64(i), 10)}
		bz, err = d.getLeafBytes(path)
		if err != nil {
			return nil, err
		}
		roots[i] = common.Root(bz)
	}

	return roots, nil
}
