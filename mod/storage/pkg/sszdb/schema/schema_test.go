package schema_test

import (
	"fmt"
	"testing"

	"github.com/berachain/beacon-kit/mod/consensus-types/pkg/state/deneb"
	"github.com/berachain/beacon-kit/mod/storage/pkg/sszdb/schema"
	"github.com/stretchr/testify/require"
)

func Test_CreateSchema(t *testing.T) {
	root, err := schema.CreateSchema(deneb.BeaconState{})
	require.NoError(t, err)
	fmt.Println(root)
}
