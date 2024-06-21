package sszdb

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/state/deneb"
	"github.com/berachain/beacon-kit/mod/errors"
	"github.com/berachain/beacon-kit/mod/storage/pkg/sszdb/tree"
	"github.com/cockroachdb/pebble"
	ssz "github.com/ferranbt/fastssz"
)

const devDBPath = "./.tmp/sszdb.db"

type DB struct {
	db *pebble.DB

	// TODO: tightly couple here for now to develop logic
	// Will decouple via the use of generics
	monolith *deneb.BeaconState
}

type Config struct {
	Path string
}

func New(cfg Config) (*DB, error) {
	if cfg.Path == "" {
		cfg.Path = devDBPath
	}
	db, err := pebble.Open(cfg.Path, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &DB{
		db:       db,
		monolith: &deneb.BeaconState{},
	}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}

	res, closer, err := d.db.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	defer closer.Close()
	ret := make([]byte, len(res))
	copy(ret, res)
	return ret, nil
}

func (d *DB) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}

	wopts := pebble.NoSync
	err := d.db.Set(key, value, wopts)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) Save() error {
	root, err := d.monolith.GetTree()
	if err != nil {
		return err
	}
	return d.hackyFastSszSave(root, 1)
}

func keyBytes(gindex uint64) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, gindex)
	return key
}

func (d *DB) SaveMonolith(mono ssz.HashRoot) error {
	treeRoot, err := tree.NewTreeFromFastSSZ(mono)
	if err != nil {
		return err
	}
	treeRoot.Hash()
	return d.save(treeRoot, 1)
}

func (d *DB) save(node *tree.Node, gindex uint64) error {
	// Save the node
	key := keyBytes(gindex)
	if err := d.Set(key, node.Encode()); err != nil {
		return err
	}

	switch {
	case node.Left == nil && node.Right == nil:
		return nil
	case node.Left != nil && node.Right != nil:
		if err := d.save(node.Left, 2*gindex); err != nil {
			return err
		}
		if err := d.save(node.Right, 2*gindex+1); err != nil {
			return err
		}
	default:
		return errors.New("node has only one child")
	}
	return nil
}

func (d *DB) hackyFastSszSave(node *ssz.Node, gindex uint64) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}

	// Save the node
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, gindex)
	// TODO this rehashes the entire subtree
	val := node.Hash()
	if err := d.Set(key, val); err != nil {
		return err
	}

	left := getLeftNode(node)
	right := getRightNode(node)

	switch {
	case left == nil && right == nil:
		return nil
	case left != nil && right != nil:
		if err := d.hackyFastSszSave(left, 2*gindex); err != nil {
			return err
		}
		if err := d.hackyFastSszSave(right, 2*gindex+1); err != nil {
			return err
		}
	default:
		return errors.New("node has only one child")
	}
	return nil
}

func (d *DB) getNode(gindex uint64) (*tree.Node, error) {
	key := keyBytes(gindex)
	bz, err := d.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, nil
	}
	return tree.DecodeNode(bz)
}

func (d *DB) mustGetNode(gindex uint64) (*tree.Node, error) {
	key := keyBytes(gindex)
	bz, err := d.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, errors.New("node not found")
	}
	return tree.DecodeNode(bz)
}

func (d *DB) getNodeBytes(gindex uint64, lenBz uint) ([]byte, error) {
	const chunksize = 32

	numNodes := int(math.Ceil(float64(lenBz) / chunksize))
	rem := lenBz % chunksize
	var (
		buf bytes.Buffer
	)
	for i := 0; i < numNodes; i++ {
		n, err := d.mustGetNode(gindex + uint64(i))
		if err != nil {
			return nil, err
		}
		// last node
		if i == numNodes-1 && rem != 0 {
			buf.Write(n.Value[:rem])
		} else {
			buf.Write(n.Value)
		}
	}

	return buf.Bytes(), nil
}

func (d *DB) Load(um ssz.Unmarshaler) error {
	root, err := d.getNode(1)
	var buf bytes.Buffer
	err = d.leafBytes(root, &buf, 1)
	if err != nil {
		return err
	}
	return um.UnmarshalSSZ(buf.Bytes())
}

func (d *DB) leafBytes(node *tree.Node, w io.Writer, gindex uint64) error {
	li := 2 * gindex
	ri := 2*gindex + 1
	left, err := d.getNode(li)
	if err != nil {
		return err
	}
	right, err := d.getNode(ri)
	if err != nil {
		return err
	}
	switch {
	case left == nil && right == nil:
		if node.IsEmpty {
			return nil
		}
		_, err = w.Write(node.Value)
		return err
	case left != nil && right != nil:
		if err = d.leafBytes(left, w, li); err != nil {
			return err
		}
		if err = d.leafBytes(right, w, ri); err != nil {
			return err
		}
	default:
		return errors.New("node has only one child")
	}
	return nil
}

func getLeftNode(node *ssz.Node) *ssz.Node {
	left, err := node.Get(2)
	if err != nil {
		return nil
	}
	return left
}

func getRightNode(node *ssz.Node) *ssz.Node {
	right, err := node.Get(3)
	if err != nil {
		return nil
	}
	return right
}
