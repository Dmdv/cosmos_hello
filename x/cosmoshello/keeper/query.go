package keeper

import (
	"cosmoshello/x/cosmoshello/types"
)

var _ types.QueryServer = Keeper{}
