package client

import (
	"context"
	"testing"

	"github.com/gogo/status"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs/rpc"
	"google.golang.org/grpc/codes"
)

func TestCreate(t *testing.T) {
	t.Run("WithoutAdminToken", func(t *testing.T) {
		f, done := setupFfs(t, "")
		defer done()

		id, token, err := f.Create(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, id)
		require.NotEmpty(t, token)
	})

	t.Run("WithAdminToken", func(t *testing.T) {
		authToken := uuid.New().String()
		f, done := setupFfs(t, authToken)
		defer done()

		t.Run("UnauthorizedEmpty", func(t *testing.T) {
			id, token, err := f.Create(ctx)
			require.Error(t, err)
			require.Empty(t, id)
			require.Empty(t, token)
		})

		t.Run("UnauthorizedWrong", func(t *testing.T) {
			wrongAuths := []string{
				"",      // Empty
				"wrong", // Non-empty
			}
			for _, auth := range wrongAuths {
				ctx := context.WithValue(ctx, AuthKey, auth)
				id, token, err := f.Create(ctx)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
				require.Empty(t, id)
				require.Empty(t, token)
			}
		})
		t.Run("Authorized", func(t *testing.T) {
			ctx := context.WithValue(ctx, AuthKey, authToken)
			id, token, err := f.Create(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, id)
			require.NotEmpty(t, token)
		})
	})
}

func setupFfs(t *testing.T, adminAuthToken string) (*FFS, func()) {
	defConfig := defaultServerConfig(t)
	if adminAuthToken != "" {
		defConfig.FFSAdminToken = adminAuthToken
	}
	serverDone := setupServer(t, defConfig)
	conn, done := setupConnection(t)
	return &FFS{client: rpc.NewRPCServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
