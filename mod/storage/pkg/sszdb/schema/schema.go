package schema

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

type SSZType interface {
	Size() uint64
	Chunks() uint64
	Position(p pathSegment) (uint64, uint8, error)

	child(p pathSegment) SSZType
}

// Basic Type

type Basic struct {
	size uint64
}

func (b Basic) Size() uint64 { return b.size }

func (b Basic) Chunks() uint64 { return 1 }

func (b Basic) child(_ pathSegment) SSZType { return b }

func (b Basic) Position(p pathSegment) (uint64, uint8, error) {
	return 0, 0, errors.New("basic type has children")
}

// Container Type

type Container struct {
	Fields     map[string]SSZType
	FieldIndex map[string]uint64
}

func (c Container) Size() uint64 { return 32 }

func (c Container) Length() uint64 { return uint64(len(c.Fields)) }

func (c Container) Chunks() uint64 { return uint64(len(c.Fields)) }

func (c Container) child(p pathSegment) SSZType { return c.Fields[p.s] }

func (c Container) Position(p pathSegment) (uint64, uint8, error) {
	pos, ok := c.FieldIndex[p.s]
	if !ok {
		return 0, 0, fmt.Errorf("field %s not found", p.s)
	}
	return pos, 0, nil
}

// Enumerable Type (vectors and lists)

type Enumerable struct {
	Element   SSZType
	length    uint64
	maxLength uint64
}

func (e Enumerable) Size() uint64 { return 32 }

func (e Enumerable) Chunks() uint64 {
	x := float64(e.Length()*e.Element.Size()) / 32
	return uint64(math.Ceil(x))
}

func (e Enumerable) child(_ pathSegment) SSZType {
	return e.Element
}

func (e Enumerable) Length() uint64 {
	if e.length == 0 {
		return e.maxLength
	}
	return e.length
}

func (e Enumerable) Position(p pathSegment) (uint64, uint8, error) {
	if p.s != "" {
		return 0, 0, fmt.Errorf("expected index, got name %s", p.s)
	}
	start := p.i * e.Element.Size()
	return uint64(math.Floor(float64(start) / 32)),
		uint8(start % 32),
		nil
}

func (e Enumerable) IsByteVector() bool {
	return e.Element.Size() == 1 && e.length > 0
}

// Object Path

type pathSegment struct {
	s string
	i uint64
}

type ObjectPath []pathSegment

func Path(names ...string) ObjectPath {
	path := make(ObjectPath, len(names))
	for i, name := range names {
		path[i] = pathSegment{s: name}
	}
	return path
}

func (o ObjectPath) AppendIndex(i uint64) ObjectPath {
	return append(o, pathSegment{i: i})
}

func (o ObjectPath) AppendName(name string) ObjectPath {
	return append(o, pathSegment{s: name})
}

const (
	uint8Size  = 1
	uint16Size = 2
	uint32Size = 4
	uint64Size = 8
)

type Node struct {
	SSZType
	GIndex uint64
	Offset uint8
}

// API

func CreateSchema(obj any) (SSZType, error) {
	typ := reflect.TypeOf(obj)
	return traverse(typ, nil)
}

func GetTreeNode(typ SSZType, path ObjectPath) (Node, error) {
	var (
		gindex = uint64(1)
		offset uint8
	)
	for _, p := range path {
		if p.s == "__len__" {
			if _, ok := typ.(Enumerable); !ok {
				return Node{}, fmt.Errorf("type %T is not enumerable", typ)
			}
			gindex = 2*gindex + 1
			offset = 0
		} else {
			pos, off, err := typ.Position(p)
			if err != nil {
				return Node{}, err
			}
			i := uint64(1)
			if e, ok := typ.(Enumerable); ok && e.maxLength > 0 {
				// list case
				i = 2
			}
			gindex = gindex*i*nextPowerOfTwo(typ.Chunks()) + pos
			typ = typ.child(p)
			offset = off
		}
	}
	return Node{SSZType: typ, GIndex: gindex, Offset: offset}, nil
}

func traverse(typ reflect.Type, field *reflect.StructField) (SSZType, error) {
	kind := typ.Kind()

	switch kind {
	case reflect.Ptr:
		return traverse(typ.Elem(), field)
	case reflect.Bool:
		return Basic{size: 1}, nil
	case reflect.Uint8:
		return Basic{size: uint8Size}, nil
	case reflect.Uint16:
		return Basic{size: uint16Size}, nil
	case reflect.Uint32:
		return Basic{size: uint32Size}, nil
	case reflect.Uint64:
		return Basic{size: uint64Size}, nil
	case reflect.Slice:
		// hack: slices with an `ssz-size` tag to be treated as vectors.
		// I'd prefer to not support this and change the struct definition instead.
		length, ok, err := getTagVal(field, "ssz-size")
		if ok {
			// vector
			elemType, err := traverse(typ.Elem(), nil)
			if err != nil {
				return nil, err
			}
			return Enumerable{Element: elemType, length: length}, nil
		} else {
			// list
			length, ok, err = getTagVal(field, "ssz-max")
			if !ok {
				return nil, err
			}
			elemType, err := traverse(typ.Elem(), nil)
			if err != nil {
				return nil, err
			}
			return Enumerable{Element: elemType, maxLength: length}, nil
		}
	case reflect.Array:
		// vector
		elemType, err := traverse(typ.Elem(), nil)
		if err != nil {
			return nil, err
		}
		return Enumerable{Element: elemType, length: uint64(typ.Len())}, nil
	case reflect.Struct:
		container := Container{
			Fields:     make(map[string]SSZType),
			FieldIndex: make(map[string]uint64),
		}
		for i, field := range flattenStructFields(typ) {
			sszType, err := traverse(field.Type, &field)
			if err != nil {
				return nil, err
			}
			container.Fields[field.Name] = sszType
			container.FieldIndex[field.Name] = uint64(i)
		}
		return container, nil
	default:
		return nil, fmt.Errorf("unsupported type: %v", kind)
	}
}

func getTagVal(field *reflect.StructField, tag string) (uint64, bool, error) {
	str := field.Tag.Get(tag)
	if str == "" {
		return 0, false, nil
	}
	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf(
			"tag %s value %s not an integer: %w", tag, str, err)
	}
	return i, true, nil
}

func flattenStructFields(typ reflect.Type) []reflect.StructField {
	var fields []reflect.StructField
	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.Anonymous {
			// flatten embedded struct fields
			embedded := flattenStructFields(field.Type)
			fields = append(fields, embedded...)
		} else {
			fields = append(fields, field)
		}
	}
	return fields
}

func nextPowerOfTwo(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return uint64(v)
}
