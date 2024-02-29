package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tabilabs/tabi/x/mint/keeper"
	"github.com/tabilabs/tabi/x/mint/types"
)

// BeginBlocker handles block beginning logic for mint
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	logger := k.Logger(ctx)
	// Get block BFT time and block height
	blockTime := ctx.BlockHeader().Time
	minter := k.GetMinter(ctx)
	if ctx.BlockHeight() <= 1 { // don't inflate token in the first block
		minter.LastUpdate = blockTime
		k.SetMinter(ctx, minter)
		return
	}

	// todo
	// 1. get seedTabi token supply
	// 2. get tabi token supply
	// 3. update minter

	// Calculate block mint amount
	params := k.GetParams(ctx)
	logger.Info(
		"Mint parameters",
		"inflation_rate",
		params.Inflation.String(),
		"mint_denom",
		params.MintDenom,
	)

	mintedCoin := minter.BlockProvision(params)
	logger.Info("Mint result", "block_provisions", mintedCoin.String(), "time", blockTime.String())

	mintedCoins := sdk.NewCoins(mintedCoin)
	// mint coins to submodule account
	if err := k.MintCoins(ctx, mintedCoins); err != nil {
		panic(err)
	}

	// send the minted coins to the fee collector account
	//if err := k.AddCollectedFees(ctx, mintedCoins); err != nil {
	//	panic(err)
	//}

	// Allocate minted coins according to allocation proportions (staking, usage
	if err := k.AllocateExponentialInflation(ctx, mintedCoin, params); err != nil {
		panic(err)
	}

	// Update last block BFT time
	lastInflationTime := minter.LastUpdate
	minter.LastUpdate = blockTime
	k.SetMinter(ctx, minter)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyLastInflationTime, lastInflationTime.String()),
			sdk.NewAttribute(types.AttributeKeyInflationTime, blockTime.String()),
			sdk.NewAttribute(types.AttributeKeyMintCoin, mintedCoin.Amount.String()),
		),
	)
}
