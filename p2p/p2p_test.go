package p2p

import (
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/stretchr/testify/require"
)

func TestLibp2p1(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)

	t.Logf("peer-id=%s", h.ID())
}
