package abci

import (
	"cosmoshello/x/cosmoshello/keeper"
	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// In the new ABCI 2.0, these functions are encapsulated into a single ABCI method called FinalizeBlock.
// This method delivers a decided block to the application, which must execute the transactions in the
// block deterministically and update its state accordingly.
// As for the consensus, it's typically handled by the underlying consensus engine (like Tendermint in Cosmos),
// not directly in these functions. The consensus engine ensures that at least 2/3 of the validators agree on
// the transactions in a block before it's finalized.

// https://docs.cosmos.network/main/learn/advanced/baseapp#finalizeblock
// https://docs.cosmos.network/main/learn/advanced/baseapp#commit
// https://docs.cosmos.network/main/learn/advanced/baseapp#main-abci-20-messages

// The Commit ABCI message is sent from the underlying CometBFT engine after the full-node has
// received precommits from 2/3+ of validators (weighted by voting power).
// This is the final step where nodes commit the block and state changes.
// Validator nodes perform the previous step of executing state transitions to validate the transactions,
// then sign the block to confirm it.
// Full nodes that are not validators do not participate in consensus but listen for votes to understand
// whether they should commit the state changes.

type ProposalHandler struct {
	logger   log.Logger
	keeper   keeper.Keeper
	valStore baseapp.ValidatorStore
}

func NewProposalHandler(
	logger log.Logger,
	keeper keeper.Keeper,
	valStore baseapp.ValidatorStore,
) *ProposalHandler {
	return &ProposalHandler{
		logger:   logger,
		keeper:   keeper,
		valStore: valStore,
	}
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		return &abci.ResponsePrepareProposal{
			Txs: req.Txs,
		}, nil
	}
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		if len(req.Txs) == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		// Check transactions and process them, but it's not required.

		// Check this one
		//return h.BaseApp.ProcessProposal(req)
		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}
