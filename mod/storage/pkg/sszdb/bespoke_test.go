package sszdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_powerOfTwo(t *testing.T) {
	require.Equal(t, uint64(1), nextPowerOfTwo(1))
	require.Equal(t, uint64(4), nextPowerOfTwo(3))
	require.Equal(t, uint64(16), nextPowerOfTwo(16))
	require.Equal(t, uint64(8192), nextPowerOfTwo(8192))
	require.Equal(t, uint64(32), nextPowerOfTwo(17))

	require.Equal(t, uint64(16), prevPowerOfTwo(18))
}
