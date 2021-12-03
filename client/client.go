package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/rafaelescrich/go-defi-1/binding/haave"
	"github.com/rafaelescrich/go-defi-1/binding/hbalancer_exchange"
	"github.com/rafaelescrich/go-defi-1/binding/hcether"
	"github.com/rafaelescrich/go-defi-1/binding/hctoken"
	"github.com/rafaelescrich/go-defi-1/binding/hcurve"
	"github.com/rafaelescrich/go-defi-1/binding/hkyber"
	"github.com/rafaelescrich/go-defi-1/binding/hmaker"
	"github.com/rafaelescrich/go-defi-1/binding/huniswap"
	"github.com/rafaelescrich/go-defi-1/binding/hyearn"

	"github.com/rafaelescrich/go-defi-1/binding/herc20tokenin"
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/rafaelescrich/go-defi-1/binding/aave/lendingpool"
	ceth_binding "github.com/rafaelescrich/go-defi-1/binding/compound/cETH"
	"github.com/rafaelescrich/go-defi-1/binding/compound/cToken"
	"github.com/rafaelescrich/go-defi-1/binding/erc20"
	"github.com/rafaelescrich/go-defi-1/binding/furucombo"
	"github.com/rafaelescrich/go-defi-1/binding/swapper"
	"github.com/rafaelescrich/go-defi-1/binding/uniswap"
	"github.com/rafaelescrich/go-defi-1/binding/yearn/yregistry"
	"github.com/rafaelescrich/go-defi-1/binding/yearn/yvault"
	"github.com/rafaelescrich/go-defi-1/binding/yearn/yweth"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type rateModel int64

const (
	// StableRate https://medium.com/aave/aave-borrowing-rates-upgraded-f6c8b27973a7
	StableRate rateModel = 1
	// VariableRate https://medium.com/aave/aave-borrowing-rates-upgraded-f6c8b27973a7
	VariableRate rateModel = 2
)

type coinType int

const (
	// ETH is ether.
	ETH coinType = iota
	// BAT is basic attention token.
	BAT coinType = iota
	// COMP is the governance token for Compound.
	COMP coinType = iota
	// DAI is the stable coin.
	DAI coinType = iota
	// REP is Augur reputation token.
	REP coinType = iota
	// SAI is Single Collateral DAI.
	SAI coinType = iota
	// UNI is the governance token for Uniswap.
	UNI coinType = iota
	// USDC is the stable coin by Circle.
	USDC coinType = iota
	// USDT is the stable coin.
	USDT coinType = iota
	// WBTC is wrapped BTC.
	WBTC coinType = iota
	// ZRX is the utility token for 0x.
	ZRX coinType = iota
	// BUSD is the Binance USD token.
	BUSD coinType = iota
	// YFI is the yearn governance token.
	YFI coinType = iota
	// AAVE is the Aave governance token.
	AAVE coinType = iota

	// cToken is the token that user receive after deposit into Yearn
	cETH = iota

	cDAI = iota

	cUSDC = iota

	yWETH = iota
)

const (
	// uniswapAddr is UniswapV2Router, see here: https://uniswap.org/docs/v2/smart-contracts/router02/#address
	uniswapAddr             string = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
	yRegistryAddr           string = "0x3eE41C098f9666ed2eA246f4D2558010e59d63A0"
	yETHVaultAddr           string = "0xe1237aA7f535b0CC33Fd973D66cBf830354D16c7"
	aaveLendingPoolAddr     string = "0x398eC7346DcD622eDc5ae82352F02bE94C62d119"
	aaveLendingPoolCoreAddr string = "0x3dfd23A6c5E8BbcFc9581d2E864a68feb6a076d3"
	FurucomboAddr           string = "0xfFffFffF2ba8F66D4e51811C5190992176930278"
	// Proxy and Handler related addresses

	// ProxyAddr is the address of the proxy contract.
	ProxyAddr             string = "0x57805e5a227937bac2b0fdacaa30413ddac6b8e1"
	hCEtherAddr           string = "0x9A1049f7f87Dbb0468C745d9B3952e23d5d6CE5e"
	hErcInAddr            string = "0x914490a362f4507058403a99e28bdf685c5c767f"
	hCTokenAddr           string = "0x8973D623d883c5641Dd3906625Aac31cdC8790c5"
	hMakerDaoAddr         string = "0x294fbca49c8a855e04d7d82b28256b086d39afea"
	hUniswapAddr          string = "0x58a21cfcee675d65d577b251668f7dc46ea9c3a0"
	hCurveAddr            string = "0xa36dfb057010c419c5917f3d68b4520db3671cdb"
	hYearnAddr            string = "0xC50C8F34c9955217a6b3e385a069184DCE17fD2A"
	hAaveAddr             string = "0xf579b009748a62b1978639d6b54259f8dc915229"
	hOneInch              string = "0x783f5c56e3c8b23d90e4a271d7acbe914bfcd319"
	hFunds                string = "0xf9b03e9ea64b2311b0221b2854edd6df97669c09"
	hKyberAddr            string = "0xe2a3431508cd8e72d53a0e4b57c24af2899322a0"
	hBalancerExchangeAddr string = "0x892dD6ebd2e3E1c0D6592309bA82a0095830D6d6"

	// TODO: The following is not on mainnet yet
	hSushiswapAddr string = "0xB6F469a8930dd5111c0EA76571c7E86298A171f7"
	hSwapper       string = "0x017F3f2EB0c55DDF49B95ad38Cd2737ACf64AB4d"

	// Curve pool addresses
	cCompound string = "0xA2B47E3D5c44877cca798226B7B8118F9BFb7A56"
	cUsdt     string = "0x52EA46506B9CC5Ef470C5bf89f17Dc28bB35D85C"
	cY        string = "0x45F783CCE6B7FF23B2ab2D70e416cdb7D6055f51"
	cBusd     string = "0x79a8C46DeA5aDa233ABaFFD40F3A0A2B1e5A4F27"
	cSusd     string = "0xA5407eAE9Ba41422680e2e00537571bcC53efBfD"
	cRen      string = "0x93054188d876f558f4a66B2EF1d97d16eDf0895B"
	cSbtc     string = "0x7fC77b5c7614E1533320Ea6DDc2Eb61fa00A9714"
	cHbtc     string = "0x4ca9b3063ec5866a4b82e437059d2c43d1be596f"
	c3Pool    string = "0xbebc44782c7db0a1a60cb6fe97d0b483032ff1c7"
	cGusd     string = "0x4f062658eaaf2c1ccf8c8e36d6824cdf41167956"
	cHusd     string = "0x3eF6A01A0f81D6046290f3e2A8c5b843e738E604"
	cUsdk     string = "0x3e01dd8a5e1fb3481f0f589056b428fc308af0fb"
	cUsdn     string = "0x0f9cb53Ebe405d49A0bbdBD291A65Ff571bC83e1"

	// Curve token addresses
	compCrv      string = "0x845838DF265Dcd2c412A1Dc9e959c7d08537f8a2"
	usdtCrv      string = "0x9fC689CCaDa600B6DF723D9E47D84d76664a1F23"
	yCrv         string = "0xdF5e0e81Dff6FAF3A7e52BA697820c5e32D806A8"
	busdCrv      string = "0x3B3Ac5386837Dc563660FB6a0937DFAa5924333B"
	susdCrv      string = "0xC25a3A3b969415c80451098fa907EC722572917F"
	renCrv       string = "0x49849C98ae39Fff122806C06791Fa73784FB3675"
	sbtcCrv      string = "0x075b1bb99792c9E1041bA13afEf80C91a1e70fB3"
	hbtcCrv      string = "0xb19059ebb43466C323583928285a49f558E572Fd"
	threePoolCrv string = "0x6c3F90f043a72FA612cbac8115EE7e52BDe6E490"
	gusdCrv      string = "0xD2967f45c4f384DEEa880F807Be904762a3DeA07"
	husdCrv      string = "0x5B5CFE992AdAC0C9D48E05854B2d91C73a003858"
	usdkCrv      string = "0x97E2768e8E73511cA874545DC5Ff8067eB19B787"
	usdnCrv      string = "0x4f3E8F405CF5aFC05D68142F3783bDfE13811522"
)

// CoinToAddressMap returns a mapping from coin to address
var CoinToAddressMap = map[coinType]common.Address{
	ETH:   common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
	BAT:   common.HexToAddress("0x0d8775f648430679a709e98d2b0cb6250d2887ef"),
	COMP:  common.HexToAddress("0xc00e94cb662c3520282e6f5717214004a7f26888"),
	DAI:   common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
	USDC:  common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"),
	USDT:  common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
	cETH:  common.HexToAddress("0x4ddc2d193948926d02f9b1fe9e1daa0718270ed5"),
	cDAI:  common.HexToAddress("0x5d3a536e4d6dbd6114cc1ead35777bab948e3643"),
	cUSDC: common.HexToAddress("0x39aa39c021dfbae8fac545936693ac917d5e7563"),
	BUSD:  common.HexToAddress("0x4Fabb145d64652a948d72533023f6E7A623C7C53"),
	yWETH: common.HexToAddress("0xe1237aA7f535b0CC33Fd973D66cBf830354D16c7"),
}

// CoinToCompoundMap returns a mapping from coin to compound address
var CoinToCompoundMap = map[coinType]common.Address{
	ETH:  common.HexToAddress("0x4ddc2d193948926d02f9b1fe9e1daa0718270ed5"),
	DAI:  common.HexToAddress("0x5d3a536e4d6dbd6114cc1ead35777bab948e3643"),
	USDC: common.HexToAddress("0x39aa39c021dfbae8fac545936693ac917d5e7563"),
}

// CoinToJoinMap maps the coin type to its corresponding Join which is a MakerDao terminology meaning an adapter to
// deposit and withdraw unlocked collateral.
var CoinToJoinMap = map[coinType]common.Address{
	DAI:  common.HexToAddress("0x9759A6Ac90977b93B58547b4A71c78317f391A28"),
	ETH:  common.HexToAddress("0x2F0b23f53734252Bda2277357e97e1517d6B042A"),
	USDC: common.HexToAddress("0x2600004fd1585f7270756DDc88aD9cfA10dD0428"),
	YFI:  common.HexToAddress("0x3ff33d9162aD47660083D7DC4bC02Fb231c81677"),
	USDT: common.HexToAddress("0x0Ac6A1D74E84C2dF9063bDDc31699FF2a2BB22A2"),
	UNI:  common.HexToAddress("0x2502F65D77cA13f183850b5f9272270454094A08"),
	AAVE: common.HexToAddress("0x24e459F61cEAa7b1cE70Dbaea938940A7c5aD46e"),
}

// CoinToIlkMap maps the coin type to the corresponding Ilk as found in here:
// https://etherscan.io/address/0x8b4ce5DCbb01e0e1f0521cd8dCfb31B308E52c24
// Ilk is a MakerDao collateral type, each Ilk correspond to a type of collateral and
// user can query it's name, symbol, dec, gem, pip, join and flip.
var CoinToIlkMap = map[coinType][32]byte{
	ETH:  byte32PutString("4554482d41000000000000000000000000000000000000000000000000000000"),
	YFI:  byte32PutString("5946492d41000000000000000000000000000000000000000000000000000000"),
	USDC: byte32PutString("555344432d420000000000000000000000000000000000000000000000000000"),
	USDT: byte32PutString("555344542d410000000000000000000000000000000000000000000000000000"),
	UNI:  byte32PutString("554e4956324441494554482d4100000000000000000000000000000000000000"),
	AAVE: byte32PutString("414156452d410000000000000000000000000000000000000000000000000000"),
}

// Client is the new interface
type Client interface {
	Uniswap() UniswapClient
}

// NewClient Create a new client
// opts can be created using your private key
// ethclient can be created when you dial an ETH end point
func NewClient(opts *bind.TransactOpts, ethClient *ethclient.Client) *DefiClient {
	c := new(DefiClient)
	c.conn = ethClient
	c.opts = opts
	return c
}

// DefiClient is the struct that stores the information.
type DefiClient struct {
	opts *bind.TransactOpts
	conn *ethclient.Client
}

// BalanceOf returns the balance of a given coin.
func (c *DefiClient) BalanceOf(coin coinType) (*big.Int, error) {
	return c.balanceOf(CoinToAddressMap[coin])
}

func (c *DefiClient) balanceOf(addr common.Address) (*big.Int, error) {
	erc20, err := erc20.NewErc20(addr, c.conn)
	if err != nil {
		return nil, err
	}
	balance, err := erc20.BalanceOf(nil, c.opts.From)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// ExecuteActions sends one transaction for all the Defi interactions.
func (c *DefiClient) ExecuteActions(actions *Actions) error {
	gasPrice, err := c.SuggestGasPrice(nil)
	if err != nil {
		return err
	}

	return c.ExecuteActionsWithGasPrice(actions, gasPrice)
}

// SuggestGasPrice provides an estimation of the gas price based on the `blockNum`.
// If the blockNum is `nil`, it will automatically use the latest block data.
// The user can also specify a specific `blockNum` so that block will be used for the prediction.
func (c *DefiClient) SuggestGasPrice(blockNum *big.Int) (*big.Int, error) {
	if blockNum == nil {
		header, err := c.conn.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return nil, err
		}
		blockNum = header.Number
	}

	block, err := c.conn.BlockByNumber(context.Background(), blockNum)
	if err != nil {
		return nil, err
	}

	// If there is no transaction in the current block we fail back to the previous block.
	if block.Transactions().Len() == 0 {
		prvBlock := big.NewInt(0)
		prvBlock.Sub(blockNum, big.NewInt(1))
		return c.SuggestGasPrice(prvBlock)
	}

	sum := big.NewInt(0)
	for _, transaction := range block.Transactions() {
		sum.Add(sum, transaction.GasPrice())
	}

	size := big.NewInt(int64(len(block.Transactions())))
	average := big.NewInt(0)

	average.Div(sum, size)
	return average, nil
}

// ExecuteActionsWithGasPrice sends one transaction for all the Defi interactions with given gasPrice.
func (c *DefiClient) ExecuteActionsWithGasPrice(actions *Actions, gasPrice *big.Int) error {
	handlers, datas, totalEthers, err := c.CombineActions(actions)

	if err != nil {
		return err
	}

	proxy, err := furucombo.NewFurucombo(common.HexToAddress(ProxyAddr), c.conn)
	if err != nil {
		return nil
	}

	opts := &bind.TransactOpts{
		Value:    totalEthers,
		Signer:   c.opts.Signer,
		From:     c.opts.From,
		GasLimit: 5000000,
		GasPrice: gasPrice,
	}
	tx, err := proxy.BatchExec(opts, handlers, datas)
	if err != nil {
		return nil
	}
	receipt, err := bind.WaitMined(context.Background(), c.conn, tx)
	if err != nil {
		return err
	}
	if receipt.Status != 1 {
		return fmt.Errorf("tx receipt status is not 1, indicating a failure occurred")
	}
	return nil
}

// CombineActions takes in an `Actions` and returns a slice of handler address and a slice of call data
// if the combine is not successful, it will return the error.
func (c *DefiClient) CombineActions(actions *Actions) ([]common.Address, [][]byte, *big.Int, error) {
	handlers := []common.Address{}
	datas := make([][]byte, 0)
	totalEthers := big.NewInt(0)
	approvalTokens := make([]common.Address, 0)
	approvalAmounts := make([]*big.Int, 0)

	for i := 0; i < len(actions.Actions); i++ {
		handlers = append(handlers, actions.Actions[i].handlerAddr)
		datas = append(datas, actions.Actions[i].data)
		totalEthers.Add(totalEthers, actions.Actions[i].ethersNeeded)
		if len(actions.Actions[i].approvalTokens) > 0 {
			for j := 0; j < len(actions.Actions[i].approvalTokens); j++ {
				tokenAddr := actions.Actions[i].approvalTokens[j]
				tokenAmount := actions.Actions[i].approvalTokenAmounts[j]
				approvalTokens = append(approvalTokens, tokenAddr)
				balance, err := c.balanceOf(tokenAddr)
				if err != nil {
					return nil, nil, nil, err
				}
				if balance.Cmp(tokenAmount) == 1 {
					approvalAmounts = append(approvalAmounts, tokenAmount)
				} else {
					approvalAmounts = append(approvalAmounts, balance)
				}
			}

		}
	}

	if len(approvalTokens) > 0 {
		parsed, err := abi.JSON(strings.NewReader(herc20tokenin.Herc20tokeninABI))
		if err != nil {
			return nil, nil, nil, err
		}
		injectData, err := parsed.Pack("inject", approvalTokens, approvalAmounts)

		if err != nil {
			return nil, nil, nil, err
		}

		handlers = append([]common.Address{common.HexToAddress(hFunds)}, handlers...)
		datas = append([][]byte{injectData}, datas...)
	}

	return handlers, datas, totalEthers, nil
}

// SupplyFundActions transfer a certain amount of fund to the proxy
func (c *DefiClient) SupplyFundActions(size *big.Int, coin coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(herc20tokenin.Herc20tokeninABI))
	if err != nil {
		return nil
	}
	injectData, err := parsed.Pack(
		"inject", []common.Address{CoinToAddressMap[coin]}, []*big.Int{size})

	if err != nil {
		return nil
	}

	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hFunds),
				data:         injectData,
				ethersNeeded: big.NewInt(0),
			},
		},
	}
}

// action represents one action, e.g. supply to Compound, swap on Uniswap
type action struct {
	handlerAddr  common.Address
	data         []byte
	ethersNeeded *big.Int
	// There could be multiple tokens that we need to approve in the case of say Curve add liquidity or Flash loan
	approvalTokens       []common.Address
	approvalTokenAmounts []*big.Int
}

// Actions represents a list of Action.
type Actions struct {
	Actions []action
}

// Add adds actions together
// This is a variadic function so user can pass in any number of actions.
func (actions *Actions) Add(newActionss ...*Actions) error {
	if newActionss == nil {
		return fmt.Errorf("new action is nil")
	}
	for _, newActions := range newActionss {
		actions.Actions = append(actions.Actions, newActions.Actions...)
	}
	return nil
}

// Uniswap---------------------------------------------------------------------

// UniswapClient struct
type UniswapClient struct {
	client  *DefiClient
	uniswap *uniswap.Uniswap
}

// Uniswap returns a uniswap client.
func (c *DefiClient) Uniswap() *UniswapClient {
	uniClient := new(UniswapClient)
	uniClient.client = c
	uniswap, err := uniswap.NewUniswap(common.HexToAddress(uniswapAddr), c.conn)

	if err != nil {
		return nil
	}

	uniClient.uniswap = uniswap
	return uniClient
}

// TxHash represents a transaction hash.
type TxHash string

// Swap in the Uniswap Exchange.
func (c *UniswapClient) Swap(size int64, baseCurrency coinType, quoteCurrency coinType, receipient common.Address) error {
	if quoteCurrency == ETH {
		return c.swapETHToToken(size, baseCurrency, receipient)
	} else {
		err := Approve(c.client, quoteCurrency, common.HexToAddress(uniswapAddr), big.NewInt(size))
		if err != nil {
			return err
		}
		if baseCurrency == ETH {
			return c.swapTokenToETH(size, quoteCurrency, receipient)
		} else {
			return c.swapTokenToToken(size, baseCurrency, quoteCurrency, receipient)
		}
	}
}

func (c *UniswapClient) swapETHToToken(size int64, baseCurrency coinType, receipient common.Address) error {
	path := []common.Address{CoinToAddressMap[ETH], CoinToAddressMap[baseCurrency]}
	tx, err := c.uniswap.SwapExactETHForTokens(
		// TODO: there is basically no minimum output amount set, so this could cause huge slippage, need to fix.
		// Also the time stamp is set to 2038 January 1, it's better to set it dynamically.
		&bind.TransactOpts{
			Value:    big.NewInt(size),
			Signer:   c.client.opts.Signer,
			From:     c.client.opts.From,
			GasLimit: 500000,
			GasPrice: big.NewInt(20000000000),
		},
		big.NewInt(0),
		path, receipient,
		big.NewInt(2145916800),
	)
	if err != nil {
		return err
	}
	bind.WaitMined(context.Background(), c.client.conn, tx)
	return nil
}

func (c *UniswapClient) swapTokenToToken(size int64, baseCurrency coinType, quoteCurrency coinType, receipient common.Address) error {
	path := []common.Address{CoinToAddressMap[quoteCurrency], CoinToAddressMap[ETH], CoinToAddressMap[baseCurrency]}

	tx, err := c.uniswap.SwapExactTokensForTokens(
		// TODO: there is basically no minimum output amount set, so this could cause huge slippage, need to fix.
		// Also the time stamp is set to 2038 January 1, it's better to set it dynamically.
		c.client.opts, big.NewInt(size), big.NewInt(0), path, receipient, big.NewInt(2145916800))
	if err != nil {
		return err
	}
	bind.WaitMined(context.Background(), c.client.conn, tx)
	return nil
}

func (c *UniswapClient) swapTokenToETH(size int64, quoteCurrency coinType, receipient common.Address) error {
	path := []common.Address{CoinToAddressMap[quoteCurrency], CoinToAddressMap[ETH]}

	tx, err := c.uniswap.SwapExactTokensForETH(
		// TODO: there is basically no minimum output amount set, so this could cause huge slippage, need to fix.
		// Also the time stamp is set to 2038 January 1, it's better to set it dynamically.
		c.client.opts, big.NewInt(size), big.NewInt(0), path, receipient, big.NewInt(2145916800))
	if err != nil {
		return err
	}
	bind.WaitMined(context.Background(), c.client.conn, tx)
	return nil
}

// SwapActions create a new swap action.
func (c *UniswapClient) SwapActions(size *big.Int, baseCurrency coinType, quoteCurrency coinType) *Actions {
	var callData []byte
	var ethersNeeded = big.NewInt(0)
	if quoteCurrency == ETH {
		ethersNeeded = size
		callData = swapETHToTokenData(size, baseCurrency)
	} else {
		if baseCurrency == ETH {
			callData = swapTokenToETHData(size, quoteCurrency)
		} else {
			callData = swapTokenToTokenData(size, baseCurrency, quoteCurrency)
		}
	}

	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hUniswapAddr),
				data:         callData,
				ethersNeeded: ethersNeeded,
			},
		},
	}
}

func swapETHToTokenData(size *big.Int, baseCurrency coinType) []byte {
	parsed, err := abi.JSON(strings.NewReader(huniswap.HuniswapABI))
	if err != nil {
		return nil
	}
	data, err := parsed.Pack(
		"swapExactETHForTokens", size, big.NewInt(0), []common.Address{CoinToAddressMap[ETH], CoinToAddressMap[baseCurrency]})
	if err != nil {
		return nil
	}
	return data
}

func swapTokenToETHData(size *big.Int, quoteCurrency coinType) []byte {
	parsed, err := abi.JSON(strings.NewReader(huniswap.HuniswapABI))
	if err != nil {
		return nil
	}
	data, err := parsed.Pack(
		"swapExactTokensForETH", size, big.NewInt(0), []common.Address{CoinToAddressMap[quoteCurrency], CoinToAddressMap[ETH]})
	if err != nil {
		return nil
	}
	return data
}

func swapTokenToTokenData(size *big.Int, baseCurrency coinType, quoteCurrency coinType) []byte {
	parsed, err := abi.JSON(strings.NewReader(huniswap.HuniswapABI))
	if err != nil {
		return nil
	}
	data, err := parsed.Pack(
		"swapExactTokensForTokens", size, big.NewInt(0), []common.Address{
			CoinToAddressMap[quoteCurrency], CoinToAddressMap[ETH], CoinToAddressMap[baseCurrency]})
	if err != nil {
		return nil
	}
	return data
}

// FlashSwapActions create an action to perform flash swap on Uniswap.
func (c *UniswapClient) FlashSwapActions(size *big.Int, coinBorrow coinType, coinRepay coinType, actions *Actions) *Actions {
	handlers := []common.Address{}
	datas := make([][]byte, 0)
	totalEthers := big.NewInt(0)
	for i := 0; i < len(actions.Actions); i++ {
		handlers = append(handlers, actions.Actions[i].handlerAddr)
		datas = append(datas, actions.Actions[i].data)
		totalEthers.Add(totalEthers, actions.Actions[i].ethersNeeded)
	}

	proxy, err := abi.JSON(strings.NewReader(furucombo.FurucomboABI))
	if err != nil {
		return nil
	}
	payloadData, err := proxy.Pack("execs", handlers, datas)
	if err != nil {
		return nil
	}
	swapperAbi, err := abi.JSON(strings.NewReader(swapper.SwapperABI))
	if err != nil {
		return nil
	}
	// skip the first 4 bytes to omit the function selector
	flashSwapData, err := swapperAbi.Pack("startSwap", CoinToAddressMap[coinBorrow], size, CoinToAddressMap[coinRepay], payloadData[4:])
	if err != nil {
		return nil
	}

	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hSwapper),
				data:         flashSwapData,
				ethersNeeded: totalEthers,
			},
		},
	}
}

// Compound---------------------------------------------------------------------

// CompoundClient is an instance of Compound protocol.
type CompoundClient struct {
	client *DefiClient
}

// Compound returns a compound client.
func (c *DefiClient) Compound() *CompoundClient {
	compoundClient := new(CompoundClient)
	compoundClient.client = c

	return compoundClient
}

// Supply supplies token to compound.
func (c *CompoundClient) Supply(amount int64, coin coinType) error {
	var (
		tx  *types.Transaction
		err error
	)

	cTokenAddr, err := c.getPoolAddrFromCoin(coin)
	opts := &bind.TransactOpts{
		From:     c.client.opts.From,
		Signer:   c.client.opts.Signer,
		GasLimit: 500000,
		GasPrice: big.NewInt(20000000000),
	}

	switch coin {
	case ETH:
		opts.Value = big.NewInt(amount)
		cETHContract, err := ceth_binding.NewCETH(cTokenAddr, c.client.conn)
		if err != nil {
			return err
		}

		tx, err = cETHContract.Mint(opts)
	case BAT, COMP, DAI, REP, SAI, UNI, USDC, USDT, WBTC, ZRX:
		err = Approve(c.client, coin, cTokenAddr, big.NewInt(amount))
		if err != nil {
			return err
		}
		cTokenContract, err := cToken.NewCToken(cTokenAddr, c.client.conn)
		if err != nil {
			return err
		}
		tx, err = cTokenContract.Mint(opts, big.NewInt(amount))
	default:
		return fmt.Errorf("Not supported")
	}

	if err != nil {
		fmt.Printf("Error mint ctoken: %v", err)
		return err
	}

	bind.WaitMined(context.Background(), c.client.conn, tx)

	return nil
}

// Redeem supplies token to compound.
func (c *CompoundClient) Redeem(amount int64, coin coinType) error {
	var (
		tx  *types.Transaction
		err error
	)

	cTokenAddr, err := c.getPoolAddrFromCoin(coin)
	if err != nil {
		return err
	}

	opts := &bind.TransactOpts{
		From:     c.client.opts.From,
		Signer:   c.client.opts.Signer,
		GasLimit: 500000,
		GasPrice: big.NewInt(20000000000),
	}

	switch coin {
	case ETH:
		cETHContract, err := ceth_binding.NewCETH(cTokenAddr, c.client.conn)
		if err != nil {
			return fmt.Errorf("Error getting cETH contract: %v", err)
		}

		tx, err = cETHContract.Redeem(opts, big.NewInt(amount))
	case BAT, COMP, DAI, REP, SAI, UNI, USDC, USDT, WBTC, ZRX:
		cTokenContract, err := cToken.NewCToken(cTokenAddr, c.client.conn)
		if err != nil {
			return fmt.Errorf("Error getting cToken contract: %v", err)
		}

		tx, err = cTokenContract.Redeem(opts, big.NewInt(amount))
	}

	if err != nil {
		return err
	}

	bind.WaitMined(context.Background(), c.client.conn, tx)

	return nil
}

// BalanceOf return the balance of given cToken.
func (c *CompoundClient) BalanceOf(coin coinType) (*big.Int, error) {
	var (
		val *big.Int
		err error
	)

	cTokenAddr, err := c.getPoolAddrFromCoin(coin)
	if err != nil {
		return nil, err
	}

	switch coin {
	case ETH:
		cETHContract, err := ceth_binding.NewCETH(cTokenAddr, c.client.conn)
		if err != nil {
			return nil, fmt.Errorf("Error getting cETH contract")
		}

		val, err = cETHContract.BalanceOf(nil, c.client.opts.From)
	case BAT, COMP, DAI, REP, SAI, UNI, USDC, USDT, WBTC, ZRX:
		cTokenContract, err := cToken.NewCToken(cTokenAddr, c.client.conn)
		if err != nil {
			return nil, fmt.Errorf("Error getting cDai contract")
		}

		val, err = cTokenContract.BalanceOf(nil, c.client.opts.From)
	default:
		return nil, fmt.Errorf("Not support token in balanceOf: %v", coin)
	}

	if err != nil {
		return big.NewInt(0), fmt.Errorf("Error getting balance of cToken: %v", err)
	}
	return val, nil
}

// BalanceOfUnderlying return the balance of given cToken
func (c *CompoundClient) BalanceOfUnderlying(coin coinType) (*types.Transaction, error) {
	var (
		tx  *types.Transaction
		err error
	)

	cTokenAddr, err := c.getPoolAddrFromCoin(coin)
	if err != nil {
		return nil, err
	}

	switch coin {
	case ETH:
		cETHContract, err := ceth_binding.NewCETH(cTokenAddr, c.client.conn)
		if err != nil {
			fmt.Printf("Error getting cETH contract")
		}

		tx, err = cETHContract.BalanceOfUnderlying(nil, c.client.opts.From)
	case BAT, COMP, DAI, REP, SAI, UNI, USDC, USDT, WBTC, ZRX:
		cTokenContract, err := cToken.NewCToken(cTokenAddr, c.client.conn)
		if err != nil {
			return nil, fmt.Errorf("Error getting cDai contract")
		}

		tx, err = cTokenContract.BalanceOfUnderlying(nil, c.client.opts.From)
	}

	if err != nil {
		fmt.Printf("Error getting balance of cToken: %v", err)
		return nil, err
	}
	return tx, nil
}

// SupplyActions create a supply action to supply asset to Compound.
func (c *CompoundClient) SupplyActions(size *big.Int, coin coinType) *Actions {
	if coin == ETH {
		return c.supplyActionsETH(size, coin)
	} else {
		return c.supplyActionsERC20(size, coin)
	}
}

func (c *CompoundClient) supplyActionsETH(size *big.Int, coin coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hcether.HcetherABI))
	if err != nil {
		return nil
	}
	data, err := parsed.Pack("mint", size)
	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hCEtherAddr),
				data:         data,
				ethersNeeded: size,
			},
		},
	}
}

func (c *CompoundClient) supplyActionsERC20(size *big.Int, coin coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hctoken.HctokenABI))
	if err != nil {
		return nil
	}
	mintData, err := parsed.Pack("mint", CoinToCompoundMap[DAI], size)
	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hCTokenAddr),
				data:                 mintData,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       []common.Address{CoinToAddressMap[coin]},
				approvalTokenAmounts: []*big.Int{size},
			},
		},
	}
}

// RedeemActions create a Compound redeem action to be executed.
func (c *CompoundClient) RedeemActions(size *big.Int, coin coinType) *Actions {
	if coin == ETH {
		return c.redeemActionsETH(size, coin)
	} else {
		return c.redeemActionsERC20(size, coin)
	}
}

func (c *CompoundClient) redeemActionsETH(size *big.Int, coin coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hcether.HcetherABI))
	if err != nil {
		return nil
	}
	data, err := parsed.Pack("redeem", size)
	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hCEtherAddr),
				data:         data,
				ethersNeeded: size,
			},
		},
	}
}

func (c *CompoundClient) redeemActionsERC20(size *big.Int, coin coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hctoken.HctokenABI))
	if err != nil {
		return nil
	}
	redeemData, err := parsed.Pack("redeem", CoinToCompoundMap[coin], size)
	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hCTokenAddr),
				data:                 redeemData,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       []common.Address{CoinToCompoundMap[coin]},
				approvalTokenAmounts: []*big.Int{size},
			},
		},
	}
}

// FlashLoanActions create an action to perform Uniswap flashloan.
func (c *AaveClient) FlashLoanActions(size *big.Int, coin coinType, actions *Actions) *Actions {
	handlers := []common.Address{}
	datas := make([][]byte, 0)
	totalEthers := big.NewInt(0)
	for i := 0; i < len(actions.Actions); i++ {
		handlers = append(handlers, actions.Actions[i].handlerAddr)
		datas = append(datas, actions.Actions[i].data)
		totalEthers.Add(totalEthers, actions.Actions[i].ethersNeeded)
	}

	proxy, err := abi.JSON(strings.NewReader(furucombo.FurucomboABI))
	if err != nil {
		return nil
	}
	payloadData, err := proxy.Pack("execs", handlers, datas)
	if err != nil {
		return nil
	}
	haave, err := abi.JSON(strings.NewReader(haave.HaaveABI))
	if err != nil {
		return nil
	}
	// skip the first 4 bytes to omit the function selector
	flashLoanData, err := haave.Pack("flashLoan", CoinToAddressMap[coin], size, payloadData[4:])
	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hAaveAddr),
				data:         flashLoanData,
				ethersNeeded: totalEthers,
			},
		},
	}
}

func (c *CompoundClient) getPoolAddrFromCoin(coin coinType) (common.Address, error) {
	if val, ok := CoinToCompoundMap[coin]; ok {
		return val, nil
	}
	return common.Address{}, fmt.Errorf("No corresponding compound pool for token: %v", coin)
}

// yearn-----------------------------------------------------------------------------------------------------

// YearnClient is an instance of Compound protocol.
type YearnClient struct {
	client       *DefiClient
	tokenToVault map[common.Address]common.Address
}

// Yearn returns a Yearn client.
func (c *DefiClient) Yearn() *YearnClient {
	yearnClient := new(YearnClient)
	yearnClient.client = c

	yregistry, err := yregistry.NewYregistry(common.HexToAddress(yRegistryAddr), c.conn)
	if err != nil {
		return nil
	}

	vaults, err := yregistry.GetVaults(nil)
	if err != nil {
		return nil
	}

	vaultInfos, err := yregistry.GetVaultsInfo(nil)
	if err != nil {
		return nil
	}

	yearnClient.tokenToVault = make(map[common.Address]common.Address)
	for i := 0; i < len(vaults); i++ {
		yearnClient.tokenToVault[vaultInfos.TokenArray[i]] = vaults[i]
	}

	return yearnClient
}

func (c *YearnClient) addLiquidity(size *big.Int, coin coinType) error {
	var (
		tx  *types.Transaction
		err error
	)
	opts := &bind.TransactOpts{
		From:     c.client.opts.From,
		Signer:   c.client.opts.Signer,
		GasLimit: 500000,
		GasPrice: big.NewInt(20000000000),
	}

	if coin == ETH {
		weth, err := yweth.NewYweth(common.HexToAddress(yETHVaultAddr), c.client.conn)
		if err != nil {
			return fmt.Errorf("Error getting weth contract")
		}
		opts.Value = size
		tx, err = weth.DepositETH(opts)
	} else if coin != ETH {
		tokenAddr := CoinToAddressMap[coin]
		vaultAddr, ok := c.tokenToVault[tokenAddr]
		if !ok {
			return fmt.Errorf("No corresponding vault found for: %v ", coin)
		}
		err = Approve(c.client, coin, vaultAddr, size)
		yvault, err := yvault.NewYvault(vaultAddr, c.client.conn)
		if err != nil {
			return fmt.Errorf("Error getting weth contract")
		}
		opts.Value = size
		tx, err = yvault.Deposit(opts, size)
	}

	if err != nil {
		fmt.Printf("Error deposit into vault: %v", err)
		return err
	}

	bind.WaitMined(context.Background(), c.client.conn, tx)

	return nil
}

func (c *YearnClient) removeLiquidity(size *big.Int, coin coinType) error {
	var (
		tx  *types.Transaction
		err error
	)
	opts := &bind.TransactOpts{
		From:     c.client.opts.From,
		Signer:   c.client.opts.Signer,
		GasLimit: 500000,
		GasPrice: big.NewInt(20000000000),
	}

	if coin == ETH {
		weth, err := yweth.NewYweth(common.HexToAddress(yETHVaultAddr), c.client.conn)
		if err != nil {
			return fmt.Errorf("Error getting weth contract")
		}
		tx, err = weth.WithdrawETH(opts, size)
	} else if coin != ETH {
		tokenAddr := CoinToAddressMap[coin]
		vaultAddr, ok := c.tokenToVault[tokenAddr]
		if !ok {
			return fmt.Errorf("No corresponding vault found for: %v ", coin)
		}
		yvault, err := yvault.NewYvault(vaultAddr, c.client.conn)
		if err != nil {
			return fmt.Errorf("Error getting weth contract")
		}
		opts.Value = size
		tx, err = yvault.Withdraw(opts, size)
	}

	if err != nil {
		fmt.Printf("Error withdraw from vault: %v", err)
		return err
	}

	bind.WaitMined(context.Background(), c.client.conn, tx)

	return nil
}

// AddLiquidityActions creates an add liquidity action to Yearn.
func (c *YearnClient) AddLiquidityActions(size *big.Int, coin coinType) *Actions {
	if coin == ETH {
		return c.addLiquidityActionsETH(size, coin)
	} else {
		return c.addLiquidityActionsERC20(size, coin)
	}
}

func (c *YearnClient) addLiquidityActionsETH(size *big.Int, coin coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hyearn.HyearnABI))
	if err != nil {
		return nil
	}
	data, err := parsed.Pack("depositETH", size, common.HexToAddress(yETHVaultAddr))
	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hYearnAddr),
				data:         data,
				ethersNeeded: size,
			},
		},
	}
}

func (c *YearnClient) addLiquidityActionsERC20(size *big.Int, coin coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hyearn.HyearnABI))
	if err != nil {
		return nil
	}
	tokenAddr := CoinToAddressMap[coin]
	vaultAddr, ok := c.tokenToVault[tokenAddr]
	if !ok {
		return nil
	}
	data, err := parsed.Pack("deposit", vaultAddr, size)
	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hYearnAddr),
				data:                 data,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       []common.Address{CoinToAddressMap[coin]},
				approvalTokenAmounts: []*big.Int{size},
			},
		},
	}
}

// RemoveLiquidityActions creates a remove liquidity action to Yearn.
func (c *YearnClient) RemoveLiquidityActions(size *big.Int, coin coinType) *Actions {
	if coin == ETH {
		return c.removeLiquidityActionsETH(size, coin)
	} else {
		return c.removeLiquidityActionsERC20(size, coin)
	}
}

func (c *YearnClient) removeLiquidityActionsETH(size *big.Int, coin coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hyearn.HyearnABI))
	if err != nil {
		return nil
	}
	data, err := parsed.Pack("withdrawETH", common.HexToAddress(yETHVaultAddr), size)
	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hYearnAddr),
				data:                 data,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       []common.Address{common.HexToAddress(yETHVaultAddr)},
				approvalTokenAmounts: []*big.Int{size},
			},
		},
	}
}

func (c *YearnClient) removeLiquidityActionsERC20(size *big.Int, coin coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hyearn.HyearnABI))
	if err != nil {
		return nil
	}
	data, err := parsed.Pack("withdraw", common.HexToAddress(yETHVaultAddr), size)
	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hYearnAddr),
				data:         data,
				ethersNeeded: big.NewInt(0),
			},
		},
	}
}

// Aave----------------------------------------------------------------------------

// AaveClient is an instance of Aave protocol.
type AaveClient struct {
	client      *DefiClient
	lendingPool *lendingpool.Lendingpool
}

// Aave returns an Aave client which contains functions that you can use to interact with Aave.
func (c *DefiClient) Aave() *AaveClient {
	aaveClient := new(AaveClient)
	aaveClient.client = c

	lendingpool, err := lendingpool.NewLendingpool(common.HexToAddress(aaveLendingPoolAddr), c.conn)
	if err != nil {
		return nil
	}
	aaveClient.lendingPool = lendingpool
	return aaveClient
}

// Lend lend to the Aave lending pool.
func (c *AaveClient) Lend(size *big.Int, coin coinType) error {
	opts := &bind.TransactOpts{
		From:     c.client.opts.From,
		Signer:   c.client.opts.Signer,
		GasLimit: 500000,
		GasPrice: big.NewInt(20000000000),
	}

	if coin != ETH {
		Approve(c.client, coin, common.HexToAddress(aaveLendingPoolCoreAddr), size)
	}

	tx, err := c.lendingPool.Deposit(opts, CoinToAddressMap[coin], size, 0)
	if err != nil {
		return err
	}
	bind.WaitMined(context.Background(), c.client.conn, tx)
	return nil
}

// Borrow borrow money from lending pool.
func (c *AaveClient) Borrow(size *big.Int, coin coinType, interestRate rateModel) error {
	return nil
}

// ReserveData is a struct described the status of Aave lending pool.
type ReserveData struct {
	CurrentATokenBalance     *big.Int
	CurrentBorrowBalance     *big.Int
	PrincipalBorrowBalance   *big.Int
	BorrowRateMode           *big.Int
	BorrowRate               *big.Int
	LiquidityRate            *big.Int
	OriginationFee           *big.Int
	VariableBorrowIndex      *big.Int
	LastUpdateTimestamp      *big.Int
	UsageAsCollateralEnabled bool
}

// GetUserReserveData get the reserve data.
func (c *AaveClient) GetUserReserveData(addr common.Address, user common.Address) (ReserveData, error) {
	data, err := c.lendingPool.GetUserReserveData(nil, addr, user)
	if err != nil {
		return ReserveData{}, err
	}
	return data, nil
}

// Kyberswap----------------------------------------------------------------------

// KyberswapClient struct
type KyberswapClient struct {
	client *DefiClient
}

// Kyberswap returns a Kyberswap client.
func (c *DefiClient) Kyberswap() *KyberswapClient {
	kyberClient := new(KyberswapClient)
	kyberClient.client = c
	return kyberClient
}

// SwapActions creates a swap action.
func (c *KyberswapClient) SwapActions(size *big.Int, baseCurrency coinType, quoteCurrency coinType) *Actions {
	var (
		data         []byte
		err          error
		ethersNeeded *big.Int = big.NewInt(0)
	)

	parsed, err := abi.JSON(strings.NewReader(hkyber.HkyberABI))
	if err != nil {
		return nil
	}

	if quoteCurrency == ETH {
		ethersNeeded = size
		data, err = parsed.Pack("swapEtherToToken", size, CoinToAddressMap[baseCurrency], big.NewInt(0))
	} else {
		if baseCurrency == ETH {
			data, err = parsed.Pack("swapTokenToEther", CoinToAddressMap[baseCurrency], size, big.NewInt(0))
		} else {
			data, err = parsed.Pack("swapTokenToToken", CoinToAddressMap[baseCurrency], size, CoinToAddressMap[quoteCurrency], big.NewInt(0))
		}
	}

	if err != nil {
		return nil
	}

	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hKyberAddr),
				data:         data,
				ethersNeeded: ethersNeeded,
			},
		},
	}

}

// Sushiswap----------------------------------------------------------------------

// SushiswapClient struct
type SushiswapClient struct {
	client *DefiClient
}

// Sushiswap returns a Sushiswap client.
func (c *DefiClient) Sushiswap() *SushiswapClient {
	sushiswapClient := new(SushiswapClient)
	sushiswapClient.client = c
	return sushiswapClient
}

// SwapActions create a new swap action.
func (c *SushiswapClient) SwapActions(size *big.Int, baseCurrency coinType, quoteCurrency coinType) *Actions {
	var callData []byte
	var ethersNeeded = big.NewInt(0)
	var approvalTokens []common.Address = nil
	var approvalTokenAmounts []*big.Int = nil

	if quoteCurrency == ETH {
		ethersNeeded = size
		callData = swapETHToTokenData(size, baseCurrency)
	} else {
		if baseCurrency == ETH {
			approvalTokens = []common.Address{CoinToAddressMap[quoteCurrency]}
			approvalTokenAmounts = []*big.Int{size}
			callData = swapTokenToETHData(size, quoteCurrency)
		} else {
			approvalTokens = []common.Address{CoinToAddressMap[quoteCurrency]}
			approvalTokenAmounts = []*big.Int{size}
			callData = swapTokenToTokenData(size, baseCurrency, quoteCurrency)
		}
	}

	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hSushiswapAddr),
				data:                 callData,
				ethersNeeded:         ethersNeeded,
				approvalTokens:       approvalTokens,
				approvalTokenAmounts: approvalTokenAmounts,
			},
		},
	}
}

// Curve-------------------------------------------------------------------------

// CurveClient struct
type CurveClient struct {
	client *DefiClient
}

// Curve returns a Curve client.
func (c *DefiClient) Curve() *CurveClient {
	curveClient := new(CurveClient)
	curveClient.client = c
	return curveClient
}

// ExchangeActions creates a Curve exchange action to swap from one stable coin to another.
func (c *CurveClient) ExchangeActions(
	handler common.Address, token1Addr common.Address, token2Addr common.Address,
	i *big.Int, j *big.Int, dx *big.Int, minDy *big.Int) *Actions {

	parsed, err := abi.JSON(strings.NewReader(hcurve.HcurveABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("exchange", handler, token1Addr, token2Addr, i, j, dx, minDy)

	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hCurveAddr),
				data:                 data,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       []common.Address{token1Addr},
				approvalTokenAmounts: []*big.Int{dx},
			},
		},
	}
}

// ExchangeUnderlyingActions creates a Curve exchangeUnderlying action.
// `handler` is the address of the Curve pool.
// `token1Addr` is the address of the input token.
// `token2Addr` is the address of the output token.
// `i` is the index of the input token in the pool.
// `j` is the index of the output token in the pool.
// `dx` is the amount of the input token that you want to swap
// `minDy` is the minimum amount of the output token that you want to receive.
func (c *CurveClient) ExchangeUnderlyingActions(handler common.Address, token1Addr common.Address, token2Addr common.Address, i *big.Int, j *big.Int, dx *big.Int, minDy *big.Int) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hcurve.HcurveABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("exchangeUnderlying", handler, token1Addr, token2Addr, i, j, dx, minDy)
	if err != nil {
		return nil
	}

	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hCurveAddr),
				data:                 data,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       []common.Address{token1Addr},
				approvalTokenAmounts: []*big.Int{dx},
			},
		},
	}
}

// AddLiquidityActions adds liqudity to the given pool.
// `handler` is the address of the Curve pool.
// `pool` is the address of the pool token, e.g. bCRV token or 3CRV token.
// `tokens` is the addresses of the tokens that is in the pool.
// `amounts` is how much amount of each tokens you want to deposit.
// `minAmount` is the minimum amount of pool token that you want to get back as a result.
func (c *CurveClient) AddLiquidityActions(
	handler common.Address, pool common.Address, tokens []common.Address,
	amounts []*big.Int, minAmount *big.Int) *Actions {

	parsed, err := abi.JSON(strings.NewReader(hcurve.HcurveABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("addLiquidity", handler, pool, tokens, amounts, minAmount)

	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hCurveAddr),
				data:                 data,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       tokens,
				approvalTokenAmounts: amounts,
			},
		},
	}
}

// RemoveLiquidityActions creates remove liquidity action on Curve.
// `handler` is the address of the Curve pool.
// `pool` is the address of the pool token, e.g. bCRV token or 3CRV token.
// `tokenI` is the addresse of the tokens that you want to remove.
// `tokenAmount` is how much amount of token you want to deposit.
// `i` is the index of the token in the given pool.
// `minAmount` is the minimum amount of the underlying token that you want to get back as a result.
func (c *CurveClient) RemoveLiquidityActions(
	handler common.Address, pool common.Address, tokenI common.Address, tokenAmount *big.Int, i *big.Int, minAmount *big.Int,
) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hcurve.HcurveABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("removeLiquidityOneCoin", handler, pool, tokenI, tokenAmount, i, minAmount)

	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hCurveAddr),
				data:                 data,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       []common.Address{pool},
				approvalTokenAmounts: []*big.Int{tokenAmount},
			},
		},
	}
}

// Maker------------------------------------------------------------------------

// MakerClient is an instance of Maker protocol.
type MakerClient struct {
	client *DefiClient
}

// Maker creates a new instance of MakerClient
func (c *DefiClient) Maker() *MakerClient {
	makerClient := new(MakerClient)
	makerClient.client = c
	return makerClient
}

// GenerateDaiAction generate an action to create a vault and get some DAI
func (c *MakerClient) GenerateDaiAction(collateralAmount *big.Int, daiAmount *big.Int, collateralType coinType) *Actions {
	if collateralType == ETH {
		return c.generateDaiActionETH(collateralAmount, daiAmount)
	} else {
		return c.generateDaiActionErc20(collateralAmount, daiAmount, collateralType)
	}
}

func (c *MakerClient) generateDaiActionETH(collateralAmount *big.Int, daiAmount *big.Int) *Actions {

	parsed, err := abi.JSON(strings.NewReader(hmaker.HmakerABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("openLockETHAndDraw", collateralAmount, CoinToJoinMap[ETH], CoinToJoinMap[DAI], CoinToIlkMap[ETH], daiAmount)

	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hMakerDaoAddr),
				data:         data,
				ethersNeeded: collateralAmount,
			},
		},
	}
}

func (c *MakerClient) generateDaiActionErc20(collateralAmount *big.Int, daiAmount *big.Int, collateralType coinType) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hmaker.HmakerABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("openLockGemAndDraw", CoinToJoinMap[collateralType], CoinToJoinMap[DAI], CoinToIlkMap[collateralType], collateralAmount, daiAmount)

	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hMakerDaoAddr),
				data:                 data,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       []common.Address{CoinToAddressMap[collateralType]},
				approvalTokenAmounts: []*big.Int{collateralAmount},
			},
		},
	}
}

// DepositCollateralActions deposits additional collateral to the given vault.
func (c *MakerClient) DepositCollateralActions(collateralAmount *big.Int, collateralType coinType, cdp *big.Int) *Actions {
	if collateralType == ETH {
		return c.depositETHActions(collateralAmount, collateralType, cdp)
	} else {
		return c.depositERC20Actions(collateralAmount, collateralType, cdp)
	}
}

func (c *MakerClient) depositETHActions(collateralAmount *big.Int, collateralType coinType, cdp *big.Int) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hmaker.HmakerABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("safeLockETH", collateralAmount, CoinToJoinMap[ETH], cdp)

	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hMakerDaoAddr),
				data:         data,
				ethersNeeded: collateralAmount,
			},
		},
	}
}

func (c *MakerClient) depositERC20Actions(collateralAmount *big.Int, collateralType coinType, cdp *big.Int) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hmaker.HmakerABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("safeLockGem", CoinToJoinMap[collateralType], cdp, collateralAmount)

	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:          common.HexToAddress(hMakerDaoAddr),
				data:                 data,
				ethersNeeded:         big.NewInt(0),
				approvalTokens:       []common.Address{CoinToAddressMap[collateralType]},
				approvalTokenAmounts: []*big.Int{collateralAmount},
			},
		},
	}
}

// WipeAction creates a wipe action to decrease debt for th given cdp/vault.
func (c *MakerClient) WipeAction(daiAmount *big.Int, cdp *big.Int) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hmaker.HmakerABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("wipe", CoinToJoinMap[DAI], cdp, daiAmount)

	if err != nil {
		return nil
	}
	return &Actions{
		Actions: []action{
			{
				handlerAddr:  common.HexToAddress(hMakerDaoAddr),
				data:         data,
				ethersNeeded: big.NewInt(0),
			},
		},
	}
}

// Balancer-----------------------------------------------------------

// BalancerClient is an instance of Balancer protocol.
type BalancerClient struct {
	client *DefiClient
}

// Balancer creates a new instance of BalancerClient
func (c *DefiClient) Balancer() *BalancerClient {
	balancerClient := new(BalancerClient)
	balancerClient.client = c
	return balancerClient
}

// Swap swaps on Balancer Exchange
func (c *BalancerClient) Swap(inputCoin coinType, outputCoin coinType, inputAmount *big.Int) *Actions {
	parsed, err := abi.JSON(strings.NewReader(hbalancer_exchange.HbalancerExchangeABI))
	if err != nil {
		return nil
	}

	data, err := parsed.Pack("smartSwapExactIn", CoinToAddressMap[inputCoin], CoinToAddressMap[outputCoin], inputAmount, big.NewInt(0), big.NewInt(10))

	if err != nil {
		return nil
	}

	if inputCoin == ETH {
		return &Actions{
			Actions: []action{
				{
					handlerAddr:  common.HexToAddress(hBalancerExchangeAddr),
					data:         data,
					ethersNeeded: inputAmount,
				},
			},
		}
	} else {
		return &Actions{
			Actions: []action{
				{
					handlerAddr:          common.HexToAddress(hBalancerExchangeAddr),
					data:                 data,
					ethersNeeded:         big.NewInt(0),
					approvalTokens:       []common.Address{CoinToAddressMap[inputCoin]},
					approvalTokenAmounts: []*big.Int{inputAmount},
				},
			},
		}
	}

}

// utility------------------------------------------------------------------------

// Approve approves ERC-20 token transfer.
func Approve(client *DefiClient, coin coinType, addr common.Address, size *big.Int) error {
	erc20Contract, err := erc20.NewErc20(CoinToAddressMap[coin], client.conn)
	if err != nil {
		return err
	}
	opts := &bind.TransactOpts{
		Signer:   client.opts.Signer,
		From:     client.opts.From,
		GasLimit: 500000,
		GasPrice: big.NewInt(20000000000),
	}
	tx, err := erc20Contract.Approve(opts, addr, size)
	bind.WaitMined(context.Background(), client.conn, tx)
	return nil
}

// Convert string to a fixed length 32 byte.
func byte32PutString(s string) [32]byte {
	var res [32]byte
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return res
	}
	if len(s) > 32 {
		copy(res[:], decoded)
	} else {
		copy(res[32-len(s):], decoded)
	}
	return res
}
