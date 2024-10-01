package cosmoshello_test

import (
	"testing"

	keepertest "cosmoshello/testutil/keeper"
	"cosmoshello/testutil/nullify"
	cosmoshello "cosmoshello/x/cosmoshello/module"
	"cosmoshello/x/cosmoshello/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.CosmoshelloKeeper(t)
	cosmoshello.InitGenesis(ctx, k, genesisState)
	got := cosmoshello.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
