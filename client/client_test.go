package client

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"testing"

	"github.com/rafaelescrich/go-defi-1/binding/erc20"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var key *ecdsa.PrivateKey
var ethClient *ethclient.Client
var publicKey *ecdsa.PublicKey
var fromAddr common.Address
var defiClient *DefiClient

func init() {
	var err error
	key, err = crypto.HexToECDSA("b8c1b5c1d81f9475fdf2e334517d29f733bdfa40682207571b12fc1142cbf329")
	if err != nil {
		log.Fatalf("Failed to create private key: %v", err)
	}
	ethClient, err = ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to ETH: %v", err)
	}

	publicKeyECDSA, ok := (key.Public()).(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type")
	}
	fromAddr = crypto.PubkeyToAddress(*publicKeyECDSA)
	defiClient = NewClient(bind.NewKeyedTransactor(key), ethClient)
	//_ = fromAddr
	//_ = ethClient
	//_ = defiClient
}

func TestInteractWithCompound(t *testing.T) {

	beforeETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)

	err = defiClient.Compound().Supply(int64(1e18), ETH)
	if err != nil {
		t.Errorf("Failed to supply in compound: %v", err)
	}

	cETH, err := defiClient.Compound().BalanceOf(ETH)
	if err != nil {
		t.Errorf("Failed to get balance: %v", err)
	}

	if cETH.Cmp(big.NewInt(0)) == 0 {
		t.Errorf("CETH minting is not successful")
	}

	afterETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)

	if beforeETH.Cmp(afterETH) != 1 {
		t.Errorf("ETH balance not decreasing.")
	}
}

func TestInteractWithUniswap(t *testing.T) {
	beforeETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)
	beforeDAI, err := defiClient.BalanceOf(DAI)

	err = defiClient.Uniswap().Swap(1e18, DAI, ETH, fromAddr)
	if err != nil {
		t.Errorf("Failed to swap in uniswap: %v", err)
	}

	afterETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)
	afterDAI, err := defiClient.BalanceOf(DAI)

	if beforeETH.Cmp(afterETH) != 1 {
		t.Errorf("ETH balance not decreasing.")
	}
	if afterDAI.Cmp(beforeDAI) != 1 {
		t.Errorf("Dai hasn't increased!")
	}

	err = defiClient.Compound().Supply(int64(1e18), DAI)
	if err != nil {
		t.Errorf("Failed to supply Dai in compound: %v", err)
	}

	cDai, err := defiClient.Compound().BalanceOf(DAI)
	if err != nil {
		t.Errorf("Failed to get balance: %v", err)
	}

	if cDai.Cmp(big.NewInt(0)) != 1 {
		t.Errorf("cDAI minting is not successful: %v", cDai)
	}

}

func TestInteractWithCompoundInDai(t *testing.T) {

	beforeDAI, err := defiClient.BalanceOf(DAI)

	err = defiClient.Compound().Supply(int64(1e18), DAI)
	if err != nil {
		t.Errorf("Failed to supply in compound: %v", err)
	}

	cDAI, err := defiClient.Compound().BalanceOf(DAI)
	if err != nil {
		t.Errorf("Failed to get balance: %v", err)
	}

	if cDAI.Cmp(big.NewInt(0)) == 0 {
		t.Errorf("CDai minting is not successful")
	}

	afterDAI, err := defiClient.BalanceOf(DAI)

	if beforeDAI.Cmp(afterDAI) != 1 {
		t.Errorf("Dai balance not decreasing.")
	}
}

func TestMintSomeUSDC(t *testing.T) {
	_, err := erc20.NewErc20(CoinToAddressMap[USDC], ethClient)
	if err != nil {
		t.Errorf("Error getting USDC Contract")
	}
	beforeUSDC, err := defiClient.BalanceOf(USDC)
	if err != nil {
		t.Errorf("Error getting USDC balance")
	}

	if beforeUSDC.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Before USDC not 0, %v", beforeUSDC)
	}
}

func TestInteractWithYearn(t *testing.T) {
	beforeETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)

	err = defiClient.Yearn().addLiquidity(big.NewInt(1e18), ETH)
	if err != nil {
		t.Errorf("Failed to add liquidity in yearn: %v", err)
	}

	afterETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)
	if beforeETH.Cmp(afterETH) != 1 {
		t.Errorf("ETH balance not decreasing.")
	}
}

func TestInteractWithFurucomboWithYearn(t *testing.T) {
	beforeETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)

	actions := new(Actions)
	actions.Add(
		defiClient.Yearn().AddLiquidityActions(big.NewInt(1e18), ETH),
	)

	err = defiClient.ExecuteActions(actions)
	if err != nil {
		t.Errorf("Failed to add liquidity in yearn: %v", err)
	}

	afterETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)
	if beforeETH.Cmp(afterETH) != 1 {
		t.Errorf("ETH balance not decreasing.")
	}

	actions = new(Actions)
	actions.Add(
		defiClient.Yearn().RemoveLiquidityActions(big.NewInt(1e18), ETH),
	)
	Approve(defiClient, yWETH, common.HexToAddress(ProxyAddr), big.NewInt(1e18))
	err = defiClient.ExecuteActions(actions)
	if err != nil {
		t.Errorf("Failed to remove liquidity in yearn: %v", err)
	}

	// Not sure why this is not increasing...
	afterafterETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)

	if afterETH.Cmp(afterafterETH) != -1 {
		t.Errorf("ETH balance not increasing. %v %v, %v", afterETH, afterafterETH, actions.Actions[0].data)
	}
}

func TestInteractWithFurucomboWithCompoundNew(t *testing.T) {

	beforeCETH, err := defiClient.Compound().BalanceOf(ETH)

	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	actions := new(Actions)

	actions.Add(
		defiClient.Compound().SupplyActions(big.NewInt(1e18), ETH),
	)

	defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

	afterCETH, err := defiClient.Compound().BalanceOf(ETH)
	if err != nil {
		t.Errorf("Failed to get balance: %v", err)
	}

	if afterCETH.Cmp(beforeCETH) != 1 {
		t.Errorf("cETH minting is not successful via Furucombo: %v %v", afterCETH, beforeCETH)
	}

}

func TestInteractWithFurucomboWithCompoundERC20New(t *testing.T) {
	Approve(defiClient, DAI, common.HexToAddress(ProxyAddr), big.NewInt(1e18))
	beforeCDai, err := defiClient.Compound().BalanceOf(DAI)

	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	actions := new(Actions)

	actions.Add(
		defiClient.Compound().SupplyActions(big.NewInt(1e18), DAI),
	)

	defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

	afterCDai, err := defiClient.Compound().BalanceOf(DAI)
	if err != nil {
		t.Errorf("Failed to get balance: %v", err)
	}

	if afterCDai.Cmp(beforeCDai) != 1 {
		t.Errorf("cDai minting is not successful via Furucombo: %v %v", afterCDai, beforeCDai)
	}

}

func TestInteractWithFurucomboWithCompoundERC20withRedeem(t *testing.T) {
	Approve(defiClient, DAI, common.HexToAddress(ProxyAddr), big.NewInt(1e18))
	Approve(defiClient, cDAI, common.HexToAddress(ProxyAddr), big.NewInt(1e18))

	beforeCDai, err := defiClient.Compound().BalanceOf(DAI)

	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	actions := new(Actions)

	actions.Add(
		defiClient.Compound().SupplyActions(big.NewInt(1e18), DAI),
		defiClient.Compound().RedeemActions(big.NewInt(100000), DAI),
	)

	defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

	afterCDai, err := defiClient.Compound().BalanceOf(DAI)
	if err != nil {
		t.Errorf("Failed to get balance: %v", err)
	}

	if afterCDai.Cmp(beforeCDai) != 1 {
		t.Errorf("cDai minting is not successful via Furucombo: %v %v", afterCDai, beforeCDai)
	}

}

func TestInteractWithFurucomboFlashLoan(t *testing.T) {
	Approve(defiClient, DAI, common.HexToAddress(ProxyAddr), big.NewInt(1e18))

	actions := new(Actions)
	flashLoanActions := new(Actions)

	actions.Add(
		defiClient.SupplyFundActions(big.NewInt(1e18), DAI),
		defiClient.Aave().FlashLoanActions(
			big.NewInt(5e18),
			DAI,
			flashLoanActions,
		),
	)

	err := defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

}

// func TestInteractWithFurucomboFlashSwap(t *testing.T) {
// 	Approve(defiClient, DAI, common.HexToAddress(FurucomboAddr), big.NewInt(1e18))

// 	actions := new(Actions)
// 	flashSwapActions := new(Actions)

// 	actions.Add(
// 		defiClient.SupplyFundActions(big.NewInt(1e18), DAI),
// 		defiClient.Uniswap().FlashSwapActions(
// 			big.NewInt(2e18),
// 			DAI,
// 			DAI,
// 			flashSwapActions,
// 		),
// 	)

// 	err := defiClient.ExecuteActions(actions)

// 	if err != nil {
// 		t.Errorf("Failed to interact with Furucombo..: %v", err)
// 	}

// }

func TestInteractWithFurucomboUniswap(t *testing.T) {
	beforeETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		t.Errorf("Error getting ETH balance")
	}
	beforeDAI, err := defiClient.BalanceOf(DAI)
	if err != nil {
		t.Errorf("Error getting DAI balance")
	}

	actions := new(Actions)

	actions.Add(
		defiClient.Uniswap().SwapActions(big.NewInt(1e18), DAI, ETH),
	)

	err = defiClient.ExecuteActions(actions)

	afterETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)
	afterDAI, err := defiClient.BalanceOf(DAI)

	if beforeETH.Cmp(afterETH) != 1 {
		t.Errorf("ETH balance not decreasing.")
	}
	if afterDAI.Cmp(beforeDAI) != 1 {
		t.Errorf("DAI hasn't increased!")
	}

}

func TestInteractWithFurucomboKyber(t *testing.T) {
	beforeETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		t.Errorf("Error getting ETH balance")
	}
	beforeDAI, err := defiClient.BalanceOf(DAI)
	if err != nil {
		t.Errorf("Error getting DAI balance")
	}

	actions := new(Actions)

	actions.Add(
		defiClient.Kyberswap().SwapActions(big.NewInt(1e18), DAI, ETH),
	)

	err = defiClient.ExecuteActions(actions)

	afterETH, err := ethClient.BalanceAt(context.Background(), fromAddr, nil)
	afterDAI, err := defiClient.BalanceOf(DAI)

	if beforeETH.Cmp(afterETH) != 1 {
		t.Errorf("ETH balance not decreasing.")
	}
	if afterDAI.Cmp(beforeDAI) != 1 {
		t.Errorf("DAI hasn't increased!")
	}

}

func TestInteractWithFurucomboFlashLoanCompound(t *testing.T) {
	Approve(defiClient, DAI, common.HexToAddress(ProxyAddr), big.NewInt(3e18))
	beforecDAI, err := defiClient.BalanceOf(cDAI)
	if err != nil {
		t.Errorf("Error getting DAI balance")
	}

	actions := new(Actions)
	flashLoanActions := new(Actions)

	flashLoanActions.Add(
		defiClient.Compound().SupplyActions(big.NewInt(1e18), DAI),
		defiClient.SupplyFundActions(big.NewInt(2e18), DAI),
		defiClient.Compound().RedeemActions(big.NewInt(1), DAI),
	)

	actions.Add(
		// defiClient.SupplyFundActions(big.NewInt(1e18), DAI),
		defiClient.Aave().FlashLoanActions(
			big.NewInt(1e18),
			DAI,
			flashLoanActions,
		),
	)

	err = defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

	aftercDAI, err := defiClient.BalanceOf(cDAI)
	if beforecDAI.Cmp(aftercDAI) != -1 {
		t.Errorf("cdai balance not increasing.")
	}
}

// func TestInteractWithFurucomboFlashSwapCompound(t *testing.T) {
// 	Approve(defiClient, DAI, common.HexToAddress(FurucomboAddr), big.NewInt(2e18))
// 	beforecDAI, err := defiClient.BalanceOf(cDAI)
// 	if err != nil {
// 		t.Errorf("Error getting DAI balance")
// 	}

// 	actions := new(Actions)
// 	flashLoanActions := new(Actions)

// 	flashLoanActions.Add(
// 		defiClient.Compound().SupplyActions(big.NewInt(1e18), DAI),
// 		defiClient.SupplyFundActions(big.NewInt(2e18), DAI),
// 		defiClient.Compound().RedeemActions(big.NewInt(1), DAI),
// 	)

// 	actions.Add(
// 		defiClient.Uniswap().FlashSwapActions(
// 			big.NewInt(1e18),
// 			DAI,
// 			DAI,
// 			flashLoanActions,
// 		),
// 	)

// 	err = defiClient.ExecuteActions(actions)

// 	if err != nil {
// 		t.Errorf("Failed to interact with Furucombo: %v", err)
// 	}

// 	aftercDAI, err := defiClient.BalanceOf(cDAI)
// 	if beforecDAI.Cmp(aftercDAI) != -1 {
// 		t.Errorf("cdai balance not increasing.")
// 	}
// }

func TestInteractWithFurucomboCurve(t *testing.T) {
	Approve(defiClient, DAI, common.HexToAddress(ProxyAddr), big.NewInt(2e18))
	beforeUSDC, err := defiClient.BalanceOf(USDC)
	if err != nil {
		t.Errorf("Error getting DAI balance")
	}

	actions := new(Actions)

	actions.Add(
		defiClient.Curve().ExchangeActions(
			common.HexToAddress(c3Pool),
			CoinToAddressMap[DAI],
			CoinToAddressMap[USDC],
			big.NewInt(0),
			big.NewInt(1),
			big.NewInt(1e18),
			big.NewInt(1e5)),
	)

	err = defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

	afterUSDC, err := defiClient.BalanceOf(USDC)
	if beforeUSDC.Cmp(afterUSDC) != -1 {
		t.Errorf("USDC balance not increasing. %v %v", beforeUSDC, afterUSDC)
	}
}

// Supplying DAI to the Curve 3 pool
func TestInteractWithFurucomboCurveAddLiquidity(t *testing.T) {
	Approve(defiClient, DAI, common.HexToAddress(ProxyAddr), big.NewInt(2e18))
	beforeDAI, err := defiClient.BalanceOf(DAI)
	if err != nil {
		t.Errorf("Error getting DAI balance")
	}

	actions := new(Actions)

	actions.Add(
		defiClient.Curve().AddLiquidityActions(
			common.HexToAddress(c3Pool),
			common.HexToAddress(threePoolCrv),
			[]common.Address{CoinToAddressMap[DAI], CoinToAddressMap[USDC], CoinToAddressMap[USDT]},
			[]*big.Int{big.NewInt(1e18), big.NewInt(0), big.NewInt(0)},
			big.NewInt(0)),
	)

	err = defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

	afterDAI, err := defiClient.BalanceOf(DAI)
	if beforeDAI.Cmp(afterDAI) != 1 {
		t.Errorf("USDC balance not decreasing. %v %v", beforeDAI, afterDAI)
	}
}

func TestInteractWithFurucomboMaker(t *testing.T) {
	beforeDAI, err := defiClient.BalanceOf(DAI)
	if err != nil {
		t.Errorf("Error getting DAI balance")
	}

	actions := new(Actions)

	collateralAmount := big.NewInt(0)
	collateralAmount.SetString("2000000000000000000", 10)

	outputAmount := big.NewInt(0)
	outputAmount.SetString("511000000000000000000", 10)
	actions.Add(
		defiClient.Maker().GenerateDaiAction(collateralAmount, outputAmount, ETH),
	)

	err = defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

	afterDAI, err := defiClient.BalanceOf(DAI)
	if beforeDAI.Cmp(afterDAI) != -1 {
		t.Errorf("dai balance not increasing: %v, %v.", beforeDAI, afterDAI)
	}
}

func TestInteractWithFurucomboMakerUSDC(t *testing.T) {
	beforeDAI, err := defiClient.BalanceOf(DAI)

	if err != nil {
		t.Errorf("Error getting DAI balance")
	}

	Approve(defiClient, USDC, common.HexToAddress(ProxyAddr), big.NewInt(1e18))
	actions := new(Actions)

	collateralAmount := big.NewInt(0)
	collateralAmount.SetString("1000000000", 10)
	outputAmount := big.NewInt(0)
	outputAmount.SetString("520000000000000000000", 10)
	actions.Add(
		defiClient.Uniswap().SwapActions(big.NewInt(5e18), USDC, ETH),
		defiClient.Maker().GenerateDaiAction(collateralAmount, outputAmount, USDC),
	)

	err = defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

	afterDAI, err := defiClient.BalanceOf(DAI)
	if beforeDAI.Cmp(afterDAI) != -1 {
		t.Errorf("dai balance not increasing: %v, %v.", beforeDAI, afterDAI)
	}
}

func TestInteractWithFurucomboBalancer(t *testing.T) {
	beforeDAI, err := defiClient.BalanceOf(ETH)
	Approve(defiClient, DAI, common.HexToAddress(ProxyAddr), big.NewInt(6e18))

	if err != nil {
		t.Errorf("Error getting DAI balance")
	}

	actions := new(Actions)

	actions.Add(
		defiClient.Balancer().Swap(DAI, ETH, big.NewInt(6e18)),
	)

	err = defiClient.ExecuteActions(actions)

	if err != nil {
		t.Errorf("Failed to interact with Furucombo: %v", err)
	}

	afterDAI, err := defiClient.BalanceOf(ETH)
	if beforeDAI.Cmp(afterDAI) != -1 {
		t.Errorf("dai balance not increasing: %v, %v.", beforeDAI, afterDAI)
	}
}
