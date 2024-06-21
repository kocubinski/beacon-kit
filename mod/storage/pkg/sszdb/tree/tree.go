package tree

import (
	"crypto/sha256"
	"reflect"
	"unsafe"

	ssz "github.com/ferranbt/fastssz"
)

type Node struct {
	Left    *Node
	Right   *Node
	IsEmpty bool
	Value   []byte
}

func NewTreeFromFastSSZ(r ssz.HashRoot) (*Node, error) {
	root, err := ssz.ProofTree(r)
	if err != nil {
		return nil, err
	}
	return copyTree(root), nil
}

// TODO this is a big hack to speed up development
// to be replaced with either a custom walker or simply ssz/v2
// It can also be used for regression testing against the fastssz implementation
func copyTree(node *ssz.Node) *Node {
	if node == nil {
		return nil
	}
	reflectNode := reflect.Indirect(reflect.ValueOf(node))

	f := reflectNode.FieldByIndex([]int{0})
	left := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Interface().(*ssz.Node)

	f = reflectNode.FieldByIndex([]int{1})
	right := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Interface().(*ssz.Node)

	f = reflectNode.FieldByIndex([]int{2})
	isEmpty := f.Bool()

	f = reflectNode.FieldByIndex([]int{3})
	value := f.Bytes()

	return &Node{
		Left:    copyTree(left),
		Right:   copyTree(right),
		IsEmpty: isEmpty,
		Value:   value,
	}
}

func (n *Node) CachedHash() []byte {
	if (n.Left == nil && n.Right == nil) || n.Value != nil {
		return n.Value
	}
	h := sha256.Sum256(append(n.Left.CachedHash(), n.Right.CachedHash()...))
	n.Value = h[:]
	return n.Value
}
