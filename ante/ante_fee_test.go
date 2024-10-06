package ante

import (
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/golang/mock/gomock"
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	antetestutil "github.com/cosmos/cosmos-sdk/x/auth/ante/testutil"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	txtestutil "github.com/cosmos/cosmos-sdk/x/auth/tx/testutil"
)

type TestAccount struct {
	acc  sdk.AccountI
	priv cryptotypes.PrivKey
}

// AnteTestSuite is a test suite to be used with ante handler tests.
type AnteTestSuite struct {
	anteHandler    sdk.AnteHandler
	ctx            sdk.Context
	clientCtx      client.Context
	txBuilder      client.TxBuilder
	accountKeeper  keeper.AccountKeeper
	bankKeeper     *authtestutil.MockBankKeeper
	txBankKeeper   *txtestutil.MockBankKeeper
	feeGrantKeeper *antetestutil.MockFeegrantKeeper
	encCfg         moduletestutil.TestEncodingConfig
}

func (suite *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
		_ = acc.SetAccountNumber(uint64(i + 1000))
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

func TestCustomFeeDecorator(t *testing.T) {
	s := SetupSuite(t, false)

	// keys and addresses
	accs := s.CreateTestAccounts(1)

	// msg and signatures
	msg := testdata.NewTestMsg(accs[0].acc.GetAddress())
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{accs[0].priv}, []uint64{0}, []uint64{0}
	testTx, err := s.CreateTestTx(s.ctx, privs, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	// Create the custom fee decorator
	dfd := NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper)
	antehandler := sdk.ChainAnteDecorators(dfd)

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any()).
		Return(sdkerrors.ErrInsufficientFunds)
	_, err = antehandler(s.ctx, testTx, false)

	require.NotNil(t, err, "Tx did not error when fee payer had insufficient funds")

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(nil)
	_, err = antehandler(s.ctx, testTx, false)

	require.Nil(t, err, "Tx errored after account has been set with sufficient funds")

	//// Define test cases
	//testCases := []struct {
	//	txSize int
	//	fee    sdk.Coins
	//}{
	//	{txSize: 4000, fee: sdk.NewCoins(sdk.NewInt64Coin("atom", 100))},
	//	{txSize: 6000, fee: sdk.NewCoins(sdk.NewInt64Coin("atom", 200))},
	//}
	//
	//for _, _ = range testCases {
	//	//// Create a mock transaction with the specified size and fee
	//	//testTx := mockFeeTx{size: tc.txSize, fee: tc.fee}
	//	//
	//	//// Call the AnteHandle method
	//	//_, err := decorator.AnteHandle(ctx, testTx, false, nil)
	//	//
	//	//// Verify the fee calculation
	//	//if tc.txSize > LargeTransactionThreshold {
	//	//	require.Equal(t, tc.fee.MulInt(sdk.NewInt(LargeTransactionFeeMultiplier)), testTx.GetFee())
	//	//} else {
	//	//	require.Equal(t, tc.fee, testTx.GetFee())
	//	//}
	//}
}

func SetupSuite(t *testing.T, isCheckTx bool) *AnteTestSuite {
	suite := &AnteTestSuite{}
	ctrl := gomock.NewController(t)
	suite.bankKeeper = authtestutil.NewMockBankKeeper(ctrl)
	suite.txBankKeeper = txtestutil.NewMockBankKeeper(ctrl)
	suite.feeGrantKeeper = antetestutil.NewMockFeegrantKeeper(ctrl)

	// Setup context, account keeper, bank keeper, and other dependencies
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithIsCheckTx(isCheckTx).WithBlockHeight(1) // app.BaseApp.NewContext(isCheckTx, cmtproto.Header{}).WithBlockHeight(1)
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})

	maccPerms := map[string][]string{
		"fee_collector":          nil,
		"mint":                   {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		"multiPerm":              {"burner", "minter", "staking"},
		"random":                 {"random"},
	}

	suite.accountKeeper = keeper.NewAccountKeeper(
		suite.encCfg.Codec,
		runtime.NewKVStoreService(key),
		types.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec("cosmos"),
		sdk.Bech32MainPrefix,
		types.NewModuleAddress("gov").String(),
	)

	suite.accountKeeper.GetModuleAccount(suite.ctx, types.FeeCollectorName)
	err := suite.accountKeeper.Params.Set(suite.ctx, types.DefaultParams())
	require.NoError(t, err)

	suite.clientCtx = client.Context{}.
		WithTxConfig(suite.encCfg.TxConfig).
		WithClient(clitestutil.NewMockCometRPC(abci.ResponseQuery{}))

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   suite.accountKeeper,
			BankKeeper:      suite.bankKeeper,
			FeegrantKeeper:  suite.feeGrantKeeper,
			SignModeHandler: suite.encCfg.TxConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
	)

	require.NoError(t, err)
	suite.anteHandler = anteHandler
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	return suite
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(
	ctx sdk.Context, privs []cryptotypes.PrivKey,
	accNums, accSeqs []uint64,
	chainID string, signMode signing.SignMode,
) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			Address:       sdk.AccAddress(priv.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        priv.PubKey(),
		}
		sigV2, err := tx.SignWithPrivKey(
			ctx, signMode, signerData,
			suite.txBuilder, priv, suite.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}

	err = suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return suite.txBuilder.GetTx(), nil
}
