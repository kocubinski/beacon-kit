package testdata

//go:generate go run github.com/ferranbt/fastssz/sszgen --path structs.go -objs FixedContainer,ComplexContainer,VariableContainer -output structs.ssz.go
type ComplexContainer struct {
	One               uint64
	Bytes42           [42]byte
	ListOfFixed       []*FixedContainer `ssz-max:"64"`
	TwentyUint32s     [20]uint32        `ssz-size:"20"` // must rendundantly specify size or receive panic
	OneFixedContainer *FixedContainer
	ListOfVariable    []*VariableContainer `ssz-max:"64"`
	OneVariable       *VariableContainer
	// TODO these fields break the generator but are valid in the ssz spec
	// FiftyContainers   [50]*FixedContainer `ssz-size:"50"`
	// ListOfVectors     [][8]uint16          `ssz-max:"12"`
	// VectorOfLists [8][]uint16 `ssz-size:"8,?" ssz-max:"?,12"`
}

type FixedContainer struct {
	Two     uint8
	Three   uint32
	Bytes42 [42]byte
}

type VariableContainer struct {
	SomeBytes20 [][20]byte `ssz-max:"12"`
}
