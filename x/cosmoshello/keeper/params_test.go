package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "cosmoshello/testutil/keeper"
	"cosmoshello/x/cosmoshello/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := keepertest.CosmoshelloKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
