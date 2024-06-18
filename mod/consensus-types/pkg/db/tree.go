package db

import ssz "github.com/ferranbt/fastssz"

func Tree(v ssz.HashRoot) (*ssz.Node, error) {
	w := &TreeView{}
	if err := v.HashTreeRootWith(w); err != nil {
		return nil, err
	}
	return w.Node(), nil
}
