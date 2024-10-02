package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type HandlerOptions struct {
	AccountKeeper ante.AccountKeeper
	BankKeeper    types.BankKeeper
}

func NewAnteHandler(opts HandlerOptions) (sdk.AnteHandler, error) {
	anteDecorators := []sdk.AnteDecorator{
		ante.NewConsumeGasForTxSizeDecorator(opts.AccountKeeper),
		NewDeductFeeDecorator(opts.AccountKeeper, opts.BankKeeper, nil),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
