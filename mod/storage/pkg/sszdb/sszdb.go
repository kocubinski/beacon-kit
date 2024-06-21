package sszdb

import (
	"encoding/binary"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/state/deneb"
	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/types"
	"github.com/berachain/beacon-kit/mod/errors"
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

func New() (*DB, error) {
	db, err := pebble.Open(devDBPath, &pebble.Options{})
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
	return d.save(root, 1)
}

func (d *DB) save(node *ssz.Node, gindex uint64) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}

	// Save the node
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, gindex)
	val := node.Hash()
	if err := d.Set(key, val); err != nil {
		return err
	}

	left := getLeftNode(node)
	right := getRightNode(node)

	switch {
	case left != nil && right != nil:
		return nil
	case left == nil && right == nil:
		if err := d.save(left, 2*gindex); err != nil {
			return err
		}
		if err := d.save(right, 2*gindex+1); err != nil {
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

/*
func drawTree(n *ssz.Node, w io.Writer) {
	g := dot.NewGraph(dot.Directed)
	drawNode(n, 1, g)
	g.Write(w)
}

func drawNode(n *ssz.Node, levelOrder int, g *dot.Graph) dot.Node {
	var h string
	left := getLeftNode(n)
	right := getRightNode(n)
	if left != nil || right != nil {
		h = hex.EncodeToString(n.Hash())
	}
	if n.value != nil {
		h = hex.EncodeToString(n.value)
	}
	dn := g.Node(fmt.Sprintf("n%d", levelOrder)).
		Label(fmt.Sprintf("%d\n%s..%s", levelOrder, h[:3], h[len(h)-3:]))

	if n.left != nil {
		ln := n.left.draw(2*levelOrder, g)
		g.Edge(dn, ln).Label("0")
	}
	if n.right != nil {
		rn := n.right.draw(2*levelOrder+1, g)
		g.Edge(dn, rn).Label("1")
	}
	return dn
}
*/

// copy tree
// TODO replace with ssz/v2 impl

// versioning

func (d *DB) SetGenesisValidatorsRoot(root [32]byte) {
	d.monolith.GenesisValidatorsRoot = root
}

func (d *DB) GetGenesisValidatorsRoot() [32]byte {
	return d.monolith.GenesisValidatorsRoot
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
