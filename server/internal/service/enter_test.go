package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServicesShareRootContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	services := New(ctx, nil, nil)
	require.Equal(t, ctx, services.Delivery.root)
	require.Equal(t, ctx, services.Message.root)
	require.Equal(t, ctx, services.WebSocket.root)
	require.Equal(t, ctx, services.QRLogin.manager.root)
}
