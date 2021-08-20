// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package router

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// IDiamondCutFacetCut is an auto generated low-level Go binding around an user-defined struct.
type IDiamondCutFacetCut struct {
	FacetAddress      common.Address
	Action            uint8
	FunctionSelectors [][4]byte
}

// IDiamondLoupeFacet is an auto generated low-level Go binding around an user-defined struct.
type IDiamondLoupeFacet struct {
	FacetAddress      common.Address
	FunctionSelectors [][4]byte
}

// WrappedTokenParams is an auto generated low-level Go binding around an user-defined struct.
type WrappedTokenParams struct {
	Name     string
	Symbol   string
	Decimals uint8
}

// RouterABI is the input ABI used to generate the binding from.
const RouterABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"receiver\",\"type\":\"bytes\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"member\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Claim\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"facetAddress\",\"type\":\"address\"},{\"internalType\":\"enumIDiamondCut.FacetCutAction\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"bytes4[]\",\"name\":\"functionSelectors\",\"type\":\"bytes4[]\"}],\"indexed\":false,\"internalType\":\"structIDiamondCut.FacetCut[]\",\"name\":\"_diamondCut\",\"type\":\"tuple[]\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_init\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"_calldata\",\"type\":\"bytes\"}],\"name\":\"DiamondCut\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"targetChain\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"receiver\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"serviceFee\",\"type\":\"uint256\"}],\"name\":\"Lock\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"member\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"MemberUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"serviceFee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"NativeTokenUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newServiceFee\",\"type\":\"uint256\"}],\"name\":\"ServiceFeeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"serviceFee\",\"type\":\"uint256\"}],\"name\":\"Unlock\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"sourceChain\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"nativeToken\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"}],\"name\":\"WrappedTokenDeployed\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"administrativeNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_wrappedToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_receiver\",\"type\":\"bytes\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_wrappedToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_receiver\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"burnWithPermit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_account\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"claimedRewardsPerAccount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_sourceChain\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_nativeToken\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint8\",\"name\":\"decimals\",\"type\":\"uint8\"}],\"internalType\":\"structWrappedTokenParams\",\"name\":\"_tokenParams\",\"type\":\"tuple\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"deployWrappedToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"facetAddress\",\"type\":\"address\"},{\"internalType\":\"enumIDiamondCut.FacetCutAction\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"bytes4[]\",\"name\":\"functionSelectors\",\"type\":\"bytes4[]\"}],\"internalType\":\"structIDiamondCut.FacetCut[]\",\"name\":\"_diamondCut\",\"type\":\"tuple[]\"},{\"internalType\":\"address\",\"name\":\"_init\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"_calldata\",\"type\":\"bytes\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"diamondCut\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"_functionSelector\",\"type\":\"bytes4\"}],\"name\":\"facetAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"facetAddress_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"facetAddresses\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"facetAddresses_\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_facet\",\"type\":\"address\"}],\"name\":\"facetFunctionSelectors\",\"outputs\":[{\"internalType\":\"bytes4[]\",\"name\":\"facetFunctionSelectors_\",\"type\":\"bytes4[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"facets\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"facetAddress\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"functionSelectors\",\"type\":\"bytes4[]\"}],\"internalType\":\"structIDiamondLoupe.Facet[]\",\"name\":\"facets_\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_chainId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_ethHash\",\"type\":\"bytes32\"}],\"name\":\"hashesUsed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_precision\",\"type\":\"uint256\"}],\"name\":\"initFeeCalculator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_members\",\"type\":\"address[]\"}],\"name\":\"initGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initRouter\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_member\",\"type\":\"address\"}],\"name\":\"isMember\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_targetChain\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_nativeToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_receiver\",\"type\":\"bytes\"}],\"name\":\"lock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_targetChain\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_nativeToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_receiver\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"lockWithPermit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"memberAt\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"membersCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_sourceChain\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_transactionId\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"_wrappedToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_receiver\",\"type\":\"address\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"nativeTokenAt\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nativeTokensCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_serviceFeePercentage\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"setServiceFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"tokenFeeData\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"serviceFeePercentage\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"feesAccrued\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"previousAccrued\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"accumulator\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_sourceChain\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_transactionId\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"_nativeToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_receiver\",\"type\":\"address\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"unlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_account\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"_status\",\"type\":\"bool\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"updateMember\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_nativeToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_serviceFee\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"_status\",\"type\":\"bool\"},{\"internalType\":\"bytes[]\",\"name\":\"_signatures\",\"type\":\"bytes[]\"}],\"name\":\"updateNativeToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Router is an auto generated Go binding around an Ethereum contract.
type Router struct {
	RouterCaller     // Read-only binding to the contract
	RouterTransactor // Write-only binding to the contract
	RouterFilterer   // Log filterer for contract events
}

// RouterCaller is an auto generated read-only Go binding around an Ethereum contract.
type RouterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RouterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RouterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RouterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RouterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RouterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RouterSession struct {
	Contract     *Router           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RouterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RouterCallerSession struct {
	Contract *RouterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// RouterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RouterTransactorSession struct {
	Contract     *RouterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RouterRaw is an auto generated low-level Go binding around an Ethereum contract.
type RouterRaw struct {
	Contract *Router // Generic contract binding to access the raw methods on
}

// RouterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RouterCallerRaw struct {
	Contract *RouterCaller // Generic read-only contract binding to access the raw methods on
}

// RouterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RouterTransactorRaw struct {
	Contract *RouterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRouter creates a new instance of Router, bound to a specific deployed contract.
func NewRouter(address common.Address, backend bind.ContractBackend) (*Router, error) {
	contract, err := bindRouter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Router{RouterCaller: RouterCaller{contract: contract}, RouterTransactor: RouterTransactor{contract: contract}, RouterFilterer: RouterFilterer{contract: contract}}, nil
}

// NewRouterCaller creates a new read-only instance of Router, bound to a specific deployed contract.
func NewRouterCaller(address common.Address, caller bind.ContractCaller) (*RouterCaller, error) {
	contract, err := bindRouter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RouterCaller{contract: contract}, nil
}

// NewRouterTransactor creates a new write-only instance of Router, bound to a specific deployed contract.
func NewRouterTransactor(address common.Address, transactor bind.ContractTransactor) (*RouterTransactor, error) {
	contract, err := bindRouter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RouterTransactor{contract: contract}, nil
}

// NewRouterFilterer creates a new log filterer instance of Router, bound to a specific deployed contract.
func NewRouterFilterer(address common.Address, filterer bind.ContractFilterer) (*RouterFilterer, error) {
	contract, err := bindRouter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RouterFilterer{contract: contract}, nil
}

// bindRouter binds a generic wrapper to an already deployed contract.
func bindRouter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RouterABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Router *RouterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Router.Contract.RouterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Router *RouterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Router.Contract.RouterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Router *RouterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Router.Contract.RouterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Router *RouterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Router.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Router *RouterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Router.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Router *RouterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Router.Contract.contract.Transact(opts, method, params...)
}

// AdministrativeNonce is a free data retrieval call binding the contract method 0x2567516a.
//
// Solidity: function administrativeNonce() view returns(uint256)
func (_Router *RouterCaller) AdministrativeNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "administrativeNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AdministrativeNonce is a free data retrieval call binding the contract method 0x2567516a.
//
// Solidity: function administrativeNonce() view returns(uint256)
func (_Router *RouterSession) AdministrativeNonce() (*big.Int, error) {
	return _Router.Contract.AdministrativeNonce(&_Router.CallOpts)
}

// AdministrativeNonce is a free data retrieval call binding the contract method 0x2567516a.
//
// Solidity: function administrativeNonce() view returns(uint256)
func (_Router *RouterCallerSession) AdministrativeNonce() (*big.Int, error) {
	return _Router.Contract.AdministrativeNonce(&_Router.CallOpts)
}

// ClaimedRewardsPerAccount is a free data retrieval call binding the contract method 0x1b3738d5.
//
// Solidity: function claimedRewardsPerAccount(address _account, address _token) view returns(uint256)
func (_Router *RouterCaller) ClaimedRewardsPerAccount(opts *bind.CallOpts, _account common.Address, _token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "claimedRewardsPerAccount", _account, _token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ClaimedRewardsPerAccount is a free data retrieval call binding the contract method 0x1b3738d5.
//
// Solidity: function claimedRewardsPerAccount(address _account, address _token) view returns(uint256)
func (_Router *RouterSession) ClaimedRewardsPerAccount(_account common.Address, _token common.Address) (*big.Int, error) {
	return _Router.Contract.ClaimedRewardsPerAccount(&_Router.CallOpts, _account, _token)
}

// ClaimedRewardsPerAccount is a free data retrieval call binding the contract method 0x1b3738d5.
//
// Solidity: function claimedRewardsPerAccount(address _account, address _token) view returns(uint256)
func (_Router *RouterCallerSession) ClaimedRewardsPerAccount(_account common.Address, _token common.Address) (*big.Int, error) {
	return _Router.Contract.ClaimedRewardsPerAccount(&_Router.CallOpts, _account, _token)
}

// FacetAddress is a free data retrieval call binding the contract method 0xcdffacc6.
//
// Solidity: function facetAddress(bytes4 _functionSelector) view returns(address facetAddress_)
func (_Router *RouterCaller) FacetAddress(opts *bind.CallOpts, _functionSelector [4]byte) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "facetAddress", _functionSelector)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FacetAddress is a free data retrieval call binding the contract method 0xcdffacc6.
//
// Solidity: function facetAddress(bytes4 _functionSelector) view returns(address facetAddress_)
func (_Router *RouterSession) FacetAddress(_functionSelector [4]byte) (common.Address, error) {
	return _Router.Contract.FacetAddress(&_Router.CallOpts, _functionSelector)
}

// FacetAddress is a free data retrieval call binding the contract method 0xcdffacc6.
//
// Solidity: function facetAddress(bytes4 _functionSelector) view returns(address facetAddress_)
func (_Router *RouterCallerSession) FacetAddress(_functionSelector [4]byte) (common.Address, error) {
	return _Router.Contract.FacetAddress(&_Router.CallOpts, _functionSelector)
}

// FacetAddresses is a free data retrieval call binding the contract method 0x52ef6b2c.
//
// Solidity: function facetAddresses() view returns(address[] facetAddresses_)
func (_Router *RouterCaller) FacetAddresses(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "facetAddresses")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// FacetAddresses is a free data retrieval call binding the contract method 0x52ef6b2c.
//
// Solidity: function facetAddresses() view returns(address[] facetAddresses_)
func (_Router *RouterSession) FacetAddresses() ([]common.Address, error) {
	return _Router.Contract.FacetAddresses(&_Router.CallOpts)
}

// FacetAddresses is a free data retrieval call binding the contract method 0x52ef6b2c.
//
// Solidity: function facetAddresses() view returns(address[] facetAddresses_)
func (_Router *RouterCallerSession) FacetAddresses() ([]common.Address, error) {
	return _Router.Contract.FacetAddresses(&_Router.CallOpts)
}

// FacetFunctionSelectors is a free data retrieval call binding the contract method 0xadfca15e.
//
// Solidity: function facetFunctionSelectors(address _facet) view returns(bytes4[] facetFunctionSelectors_)
func (_Router *RouterCaller) FacetFunctionSelectors(opts *bind.CallOpts, _facet common.Address) ([][4]byte, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "facetFunctionSelectors", _facet)

	if err != nil {
		return *new([][4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][4]byte)).(*[][4]byte)

	return out0, err

}

// FacetFunctionSelectors is a free data retrieval call binding the contract method 0xadfca15e.
//
// Solidity: function facetFunctionSelectors(address _facet) view returns(bytes4[] facetFunctionSelectors_)
func (_Router *RouterSession) FacetFunctionSelectors(_facet common.Address) ([][4]byte, error) {
	return _Router.Contract.FacetFunctionSelectors(&_Router.CallOpts, _facet)
}

// FacetFunctionSelectors is a free data retrieval call binding the contract method 0xadfca15e.
//
// Solidity: function facetFunctionSelectors(address _facet) view returns(bytes4[] facetFunctionSelectors_)
func (_Router *RouterCallerSession) FacetFunctionSelectors(_facet common.Address) ([][4]byte, error) {
	return _Router.Contract.FacetFunctionSelectors(&_Router.CallOpts, _facet)
}

// Facets is a free data retrieval call binding the contract method 0x7a0ed627.
//
// Solidity: function facets() view returns((address,bytes4[])[] facets_)
func (_Router *RouterCaller) Facets(opts *bind.CallOpts) ([]IDiamondLoupeFacet, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "facets")

	if err != nil {
		return *new([]IDiamondLoupeFacet), err
	}

	out0 := *abi.ConvertType(out[0], new([]IDiamondLoupeFacet)).(*[]IDiamondLoupeFacet)

	return out0, err

}

// Facets is a free data retrieval call binding the contract method 0x7a0ed627.
//
// Solidity: function facets() view returns((address,bytes4[])[] facets_)
func (_Router *RouterSession) Facets() ([]IDiamondLoupeFacet, error) {
	return _Router.Contract.Facets(&_Router.CallOpts)
}

// Facets is a free data retrieval call binding the contract method 0x7a0ed627.
//
// Solidity: function facets() view returns((address,bytes4[])[] facets_)
func (_Router *RouterCallerSession) Facets() ([]IDiamondLoupeFacet, error) {
	return _Router.Contract.Facets(&_Router.CallOpts)
}

// HashesUsed is a free data retrieval call binding the contract method 0xde430c0a.
//
// Solidity: function hashesUsed(uint256 _chainId, bytes32 _ethHash) view returns(bool)
func (_Router *RouterCaller) HashesUsed(opts *bind.CallOpts, _chainId *big.Int, _ethHash [32]byte) (bool, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "hashesUsed", _chainId, _ethHash)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HashesUsed is a free data retrieval call binding the contract method 0xde430c0a.
//
// Solidity: function hashesUsed(uint256 _chainId, bytes32 _ethHash) view returns(bool)
func (_Router *RouterSession) HashesUsed(_chainId *big.Int, _ethHash [32]byte) (bool, error) {
	return _Router.Contract.HashesUsed(&_Router.CallOpts, _chainId, _ethHash)
}

// HashesUsed is a free data retrieval call binding the contract method 0xde430c0a.
//
// Solidity: function hashesUsed(uint256 _chainId, bytes32 _ethHash) view returns(bool)
func (_Router *RouterCallerSession) HashesUsed(_chainId *big.Int, _ethHash [32]byte) (bool, error) {
	return _Router.Contract.HashesUsed(&_Router.CallOpts, _chainId, _ethHash)
}

// IsMember is a free data retrieval call binding the contract method 0xa230c524.
//
// Solidity: function isMember(address _member) view returns(bool)
func (_Router *RouterCaller) IsMember(opts *bind.CallOpts, _member common.Address) (bool, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "isMember", _member)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsMember is a free data retrieval call binding the contract method 0xa230c524.
//
// Solidity: function isMember(address _member) view returns(bool)
func (_Router *RouterSession) IsMember(_member common.Address) (bool, error) {
	return _Router.Contract.IsMember(&_Router.CallOpts, _member)
}

// IsMember is a free data retrieval call binding the contract method 0xa230c524.
//
// Solidity: function isMember(address _member) view returns(bool)
func (_Router *RouterCallerSession) IsMember(_member common.Address) (bool, error) {
	return _Router.Contract.IsMember(&_Router.CallOpts, _member)
}

// MemberAt is a free data retrieval call binding the contract method 0xac0250f7.
//
// Solidity: function memberAt(uint256 _index) view returns(address)
func (_Router *RouterCaller) MemberAt(opts *bind.CallOpts, _index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "memberAt", _index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MemberAt is a free data retrieval call binding the contract method 0xac0250f7.
//
// Solidity: function memberAt(uint256 _index) view returns(address)
func (_Router *RouterSession) MemberAt(_index *big.Int) (common.Address, error) {
	return _Router.Contract.MemberAt(&_Router.CallOpts, _index)
}

// MemberAt is a free data retrieval call binding the contract method 0xac0250f7.
//
// Solidity: function memberAt(uint256 _index) view returns(address)
func (_Router *RouterCallerSession) MemberAt(_index *big.Int) (common.Address, error) {
	return _Router.Contract.MemberAt(&_Router.CallOpts, _index)
}

// MembersCount is a free data retrieval call binding the contract method 0x297f9af0.
//
// Solidity: function membersCount() view returns(uint256)
func (_Router *RouterCaller) MembersCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "membersCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MembersCount is a free data retrieval call binding the contract method 0x297f9af0.
//
// Solidity: function membersCount() view returns(uint256)
func (_Router *RouterSession) MembersCount() (*big.Int, error) {
	return _Router.Contract.MembersCount(&_Router.CallOpts)
}

// MembersCount is a free data retrieval call binding the contract method 0x297f9af0.
//
// Solidity: function membersCount() view returns(uint256)
func (_Router *RouterCallerSession) MembersCount() (*big.Int, error) {
	return _Router.Contract.MembersCount(&_Router.CallOpts)
}

// NativeTokenAt is a free data retrieval call binding the contract method 0x352036ff.
//
// Solidity: function nativeTokenAt(uint256 _index) view returns(address)
func (_Router *RouterCaller) NativeTokenAt(opts *bind.CallOpts, _index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "nativeTokenAt", _index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NativeTokenAt is a free data retrieval call binding the contract method 0x352036ff.
//
// Solidity: function nativeTokenAt(uint256 _index) view returns(address)
func (_Router *RouterSession) NativeTokenAt(_index *big.Int) (common.Address, error) {
	return _Router.Contract.NativeTokenAt(&_Router.CallOpts, _index)
}

// NativeTokenAt is a free data retrieval call binding the contract method 0x352036ff.
//
// Solidity: function nativeTokenAt(uint256 _index) view returns(address)
func (_Router *RouterCallerSession) NativeTokenAt(_index *big.Int) (common.Address, error) {
	return _Router.Contract.NativeTokenAt(&_Router.CallOpts, _index)
}

// NativeTokensCount is a free data retrieval call binding the contract method 0x1b74f8fe.
//
// Solidity: function nativeTokensCount() view returns(uint256)
func (_Router *RouterCaller) NativeTokensCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "nativeTokensCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NativeTokensCount is a free data retrieval call binding the contract method 0x1b74f8fe.
//
// Solidity: function nativeTokensCount() view returns(uint256)
func (_Router *RouterSession) NativeTokensCount() (*big.Int, error) {
	return _Router.Contract.NativeTokensCount(&_Router.CallOpts)
}

// NativeTokensCount is a free data retrieval call binding the contract method 0x1b74f8fe.
//
// Solidity: function nativeTokensCount() view returns(uint256)
func (_Router *RouterCallerSession) NativeTokensCount() (*big.Int, error) {
	return _Router.Contract.NativeTokensCount(&_Router.CallOpts)
}

// TokenFeeData is a free data retrieval call binding the contract method 0xeb6fc3b1.
//
// Solidity: function tokenFeeData(address _token) view returns(uint256 serviceFeePercentage, uint256 feesAccrued, uint256 previousAccrued, uint256 accumulator)
func (_Router *RouterCaller) TokenFeeData(opts *bind.CallOpts, _token common.Address) (struct {
	ServiceFeePercentage *big.Int
	FeesAccrued          *big.Int
	PreviousAccrued      *big.Int
	Accumulator          *big.Int
}, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "tokenFeeData", _token)

	outstruct := new(struct {
		ServiceFeePercentage *big.Int
		FeesAccrued          *big.Int
		PreviousAccrued      *big.Int
		Accumulator          *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ServiceFeePercentage = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.FeesAccrued = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.PreviousAccrued = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Accumulator = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// TokenFeeData is a free data retrieval call binding the contract method 0xeb6fc3b1.
//
// Solidity: function tokenFeeData(address _token) view returns(uint256 serviceFeePercentage, uint256 feesAccrued, uint256 previousAccrued, uint256 accumulator)
func (_Router *RouterSession) TokenFeeData(_token common.Address) (struct {
	ServiceFeePercentage *big.Int
	FeesAccrued          *big.Int
	PreviousAccrued      *big.Int
	Accumulator          *big.Int
}, error) {
	return _Router.Contract.TokenFeeData(&_Router.CallOpts, _token)
}

// TokenFeeData is a free data retrieval call binding the contract method 0xeb6fc3b1.
//
// Solidity: function tokenFeeData(address _token) view returns(uint256 serviceFeePercentage, uint256 feesAccrued, uint256 previousAccrued, uint256 accumulator)
func (_Router *RouterCallerSession) TokenFeeData(_token common.Address) (struct {
	ServiceFeePercentage *big.Int
	FeesAccrued          *big.Int
	PreviousAccrued      *big.Int
	Accumulator          *big.Int
}, error) {
	return _Router.Contract.TokenFeeData(&_Router.CallOpts, _token)
}

// Burn is a paid mutator transaction binding the contract method 0x44d17187.
//
// Solidity: function burn(address _wrappedToken, uint256 _amount, bytes _receiver) returns()
func (_Router *RouterTransactor) Burn(opts *bind.TransactOpts, _wrappedToken common.Address, _amount *big.Int, _receiver []byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "burn", _wrappedToken, _amount, _receiver)
}

// Burn is a paid mutator transaction binding the contract method 0x44d17187.
//
// Solidity: function burn(address _wrappedToken, uint256 _amount, bytes _receiver) returns()
func (_Router *RouterSession) Burn(_wrappedToken common.Address, _amount *big.Int, _receiver []byte) (*types.Transaction, error) {
	return _Router.Contract.Burn(&_Router.TransactOpts, _wrappedToken, _amount, _receiver)
}

// Burn is a paid mutator transaction binding the contract method 0x44d17187.
//
// Solidity: function burn(address _wrappedToken, uint256 _amount, bytes _receiver) returns()
func (_Router *RouterTransactorSession) Burn(_wrappedToken common.Address, _amount *big.Int, _receiver []byte) (*types.Transaction, error) {
	return _Router.Contract.Burn(&_Router.TransactOpts, _wrappedToken, _amount, _receiver)
}

// BurnWithPermit is a paid mutator transaction binding the contract method 0xfb828b1d.
//
// Solidity: function burnWithPermit(address _wrappedToken, uint256 _amount, bytes _receiver, uint256 _deadline, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_Router *RouterTransactor) BurnWithPermit(opts *bind.TransactOpts, _wrappedToken common.Address, _amount *big.Int, _receiver []byte, _deadline *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "burnWithPermit", _wrappedToken, _amount, _receiver, _deadline, _v, _r, _s)
}

// BurnWithPermit is a paid mutator transaction binding the contract method 0xfb828b1d.
//
// Solidity: function burnWithPermit(address _wrappedToken, uint256 _amount, bytes _receiver, uint256 _deadline, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_Router *RouterSession) BurnWithPermit(_wrappedToken common.Address, _amount *big.Int, _receiver []byte, _deadline *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _Router.Contract.BurnWithPermit(&_Router.TransactOpts, _wrappedToken, _amount, _receiver, _deadline, _v, _r, _s)
}

// BurnWithPermit is a paid mutator transaction binding the contract method 0xfb828b1d.
//
// Solidity: function burnWithPermit(address _wrappedToken, uint256 _amount, bytes _receiver, uint256 _deadline, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_Router *RouterTransactorSession) BurnWithPermit(_wrappedToken common.Address, _amount *big.Int, _receiver []byte, _deadline *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _Router.Contract.BurnWithPermit(&_Router.TransactOpts, _wrappedToken, _amount, _receiver, _deadline, _v, _r, _s)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address _token) returns()
func (_Router *RouterTransactor) Claim(opts *bind.TransactOpts, _token common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "claim", _token)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address _token) returns()
func (_Router *RouterSession) Claim(_token common.Address) (*types.Transaction, error) {
	return _Router.Contract.Claim(&_Router.TransactOpts, _token)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address _token) returns()
func (_Router *RouterTransactorSession) Claim(_token common.Address) (*types.Transaction, error) {
	return _Router.Contract.Claim(&_Router.TransactOpts, _token)
}

// DeployWrappedToken is a paid mutator transaction binding the contract method 0x3b10603a.
//
// Solidity: function deployWrappedToken(uint256 _sourceChain, bytes _nativeToken, (string,string,uint8) _tokenParams, bytes[] _signatures) returns()
func (_Router *RouterTransactor) DeployWrappedToken(opts *bind.TransactOpts, _sourceChain *big.Int, _nativeToken []byte, _tokenParams WrappedTokenParams, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "deployWrappedToken", _sourceChain, _nativeToken, _tokenParams, _signatures)
}

// DeployWrappedToken is a paid mutator transaction binding the contract method 0x3b10603a.
//
// Solidity: function deployWrappedToken(uint256 _sourceChain, bytes _nativeToken, (string,string,uint8) _tokenParams, bytes[] _signatures) returns()
func (_Router *RouterSession) DeployWrappedToken(_sourceChain *big.Int, _nativeToken []byte, _tokenParams WrappedTokenParams, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.DeployWrappedToken(&_Router.TransactOpts, _sourceChain, _nativeToken, _tokenParams, _signatures)
}

// DeployWrappedToken is a paid mutator transaction binding the contract method 0x3b10603a.
//
// Solidity: function deployWrappedToken(uint256 _sourceChain, bytes _nativeToken, (string,string,uint8) _tokenParams, bytes[] _signatures) returns()
func (_Router *RouterTransactorSession) DeployWrappedToken(_sourceChain *big.Int, _nativeToken []byte, _tokenParams WrappedTokenParams, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.DeployWrappedToken(&_Router.TransactOpts, _sourceChain, _nativeToken, _tokenParams, _signatures)
}

// DiamondCut is a paid mutator transaction binding the contract method 0x41ea4fdb.
//
// Solidity: function diamondCut((address,uint8,bytes4[])[] _diamondCut, address _init, bytes _calldata, bytes[] _signatures) returns()
func (_Router *RouterTransactor) DiamondCut(opts *bind.TransactOpts, _diamondCut []IDiamondCutFacetCut, _init common.Address, _calldata []byte, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "diamondCut", _diamondCut, _init, _calldata, _signatures)
}

// DiamondCut is a paid mutator transaction binding the contract method 0x41ea4fdb.
//
// Solidity: function diamondCut((address,uint8,bytes4[])[] _diamondCut, address _init, bytes _calldata, bytes[] _signatures) returns()
func (_Router *RouterSession) DiamondCut(_diamondCut []IDiamondCutFacetCut, _init common.Address, _calldata []byte, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.DiamondCut(&_Router.TransactOpts, _diamondCut, _init, _calldata, _signatures)
}

// DiamondCut is a paid mutator transaction binding the contract method 0x41ea4fdb.
//
// Solidity: function diamondCut((address,uint8,bytes4[])[] _diamondCut, address _init, bytes _calldata, bytes[] _signatures) returns()
func (_Router *RouterTransactorSession) DiamondCut(_diamondCut []IDiamondCutFacetCut, _init common.Address, _calldata []byte, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.DiamondCut(&_Router.TransactOpts, _diamondCut, _init, _calldata, _signatures)
}

// InitFeeCalculator is a paid mutator transaction binding the contract method 0xe3c9d084.
//
// Solidity: function initFeeCalculator(uint256 _precision) returns()
func (_Router *RouterTransactor) InitFeeCalculator(opts *bind.TransactOpts, _precision *big.Int) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "initFeeCalculator", _precision)
}

// InitFeeCalculator is a paid mutator transaction binding the contract method 0xe3c9d084.
//
// Solidity: function initFeeCalculator(uint256 _precision) returns()
func (_Router *RouterSession) InitFeeCalculator(_precision *big.Int) (*types.Transaction, error) {
	return _Router.Contract.InitFeeCalculator(&_Router.TransactOpts, _precision)
}

// InitFeeCalculator is a paid mutator transaction binding the contract method 0xe3c9d084.
//
// Solidity: function initFeeCalculator(uint256 _precision) returns()
func (_Router *RouterTransactorSession) InitFeeCalculator(_precision *big.Int) (*types.Transaction, error) {
	return _Router.Contract.InitFeeCalculator(&_Router.TransactOpts, _precision)
}

// InitGovernance is a paid mutator transaction binding the contract method 0x8f98a45b.
//
// Solidity: function initGovernance(address[] _members) returns()
func (_Router *RouterTransactor) InitGovernance(opts *bind.TransactOpts, _members []common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "initGovernance", _members)
}

// InitGovernance is a paid mutator transaction binding the contract method 0x8f98a45b.
//
// Solidity: function initGovernance(address[] _members) returns()
func (_Router *RouterSession) InitGovernance(_members []common.Address) (*types.Transaction, error) {
	return _Router.Contract.InitGovernance(&_Router.TransactOpts, _members)
}

// InitGovernance is a paid mutator transaction binding the contract method 0x8f98a45b.
//
// Solidity: function initGovernance(address[] _members) returns()
func (_Router *RouterTransactorSession) InitGovernance(_members []common.Address) (*types.Transaction, error) {
	return _Router.Contract.InitGovernance(&_Router.TransactOpts, _members)
}

// InitRouter is a paid mutator transaction binding the contract method 0xe5026ea2.
//
// Solidity: function initRouter() returns()
func (_Router *RouterTransactor) InitRouter(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "initRouter")
}

// InitRouter is a paid mutator transaction binding the contract method 0xe5026ea2.
//
// Solidity: function initRouter() returns()
func (_Router *RouterSession) InitRouter() (*types.Transaction, error) {
	return _Router.Contract.InitRouter(&_Router.TransactOpts)
}

// InitRouter is a paid mutator transaction binding the contract method 0xe5026ea2.
//
// Solidity: function initRouter() returns()
func (_Router *RouterTransactorSession) InitRouter() (*types.Transaction, error) {
	return _Router.Contract.InitRouter(&_Router.TransactOpts)
}

// Lock is a paid mutator transaction binding the contract method 0xb258848a.
//
// Solidity: function lock(uint256 _targetChain, address _nativeToken, uint256 _amount, bytes _receiver) returns()
func (_Router *RouterTransactor) Lock(opts *bind.TransactOpts, _targetChain *big.Int, _nativeToken common.Address, _amount *big.Int, _receiver []byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "lock", _targetChain, _nativeToken, _amount, _receiver)
}

// Lock is a paid mutator transaction binding the contract method 0xb258848a.
//
// Solidity: function lock(uint256 _targetChain, address _nativeToken, uint256 _amount, bytes _receiver) returns()
func (_Router *RouterSession) Lock(_targetChain *big.Int, _nativeToken common.Address, _amount *big.Int, _receiver []byte) (*types.Transaction, error) {
	return _Router.Contract.Lock(&_Router.TransactOpts, _targetChain, _nativeToken, _amount, _receiver)
}

// Lock is a paid mutator transaction binding the contract method 0xb258848a.
//
// Solidity: function lock(uint256 _targetChain, address _nativeToken, uint256 _amount, bytes _receiver) returns()
func (_Router *RouterTransactorSession) Lock(_targetChain *big.Int, _nativeToken common.Address, _amount *big.Int, _receiver []byte) (*types.Transaction, error) {
	return _Router.Contract.Lock(&_Router.TransactOpts, _targetChain, _nativeToken, _amount, _receiver)
}

// LockWithPermit is a paid mutator transaction binding the contract method 0xe1bf71ea.
//
// Solidity: function lockWithPermit(uint256 _targetChain, address _nativeToken, uint256 _amount, bytes _receiver, uint256 _deadline, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_Router *RouterTransactor) LockWithPermit(opts *bind.TransactOpts, _targetChain *big.Int, _nativeToken common.Address, _amount *big.Int, _receiver []byte, _deadline *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "lockWithPermit", _targetChain, _nativeToken, _amount, _receiver, _deadline, _v, _r, _s)
}

// LockWithPermit is a paid mutator transaction binding the contract method 0xe1bf71ea.
//
// Solidity: function lockWithPermit(uint256 _targetChain, address _nativeToken, uint256 _amount, bytes _receiver, uint256 _deadline, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_Router *RouterSession) LockWithPermit(_targetChain *big.Int, _nativeToken common.Address, _amount *big.Int, _receiver []byte, _deadline *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _Router.Contract.LockWithPermit(&_Router.TransactOpts, _targetChain, _nativeToken, _amount, _receiver, _deadline, _v, _r, _s)
}

// LockWithPermit is a paid mutator transaction binding the contract method 0xe1bf71ea.
//
// Solidity: function lockWithPermit(uint256 _targetChain, address _nativeToken, uint256 _amount, bytes _receiver, uint256 _deadline, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_Router *RouterTransactorSession) LockWithPermit(_targetChain *big.Int, _nativeToken common.Address, _amount *big.Int, _receiver []byte, _deadline *big.Int, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _Router.Contract.LockWithPermit(&_Router.TransactOpts, _targetChain, _nativeToken, _amount, _receiver, _deadline, _v, _r, _s)
}

// Mint is a paid mutator transaction binding the contract method 0x914442d7.
//
// Solidity: function mint(uint256 _sourceChain, bytes _transactionId, address _wrappedToken, uint256 _amount, address _receiver, bytes[] _signatures) returns()
func (_Router *RouterTransactor) Mint(opts *bind.TransactOpts, _sourceChain *big.Int, _transactionId []byte, _wrappedToken common.Address, _amount *big.Int, _receiver common.Address, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "mint", _sourceChain, _transactionId, _wrappedToken, _amount, _receiver, _signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x914442d7.
//
// Solidity: function mint(uint256 _sourceChain, bytes _transactionId, address _wrappedToken, uint256 _amount, address _receiver, bytes[] _signatures) returns()
func (_Router *RouterSession) Mint(_sourceChain *big.Int, _transactionId []byte, _wrappedToken common.Address, _amount *big.Int, _receiver common.Address, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint(&_Router.TransactOpts, _sourceChain, _transactionId, _wrappedToken, _amount, _receiver, _signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x914442d7.
//
// Solidity: function mint(uint256 _sourceChain, bytes _transactionId, address _wrappedToken, uint256 _amount, address _receiver, bytes[] _signatures) returns()
func (_Router *RouterTransactorSession) Mint(_sourceChain *big.Int, _transactionId []byte, _wrappedToken common.Address, _amount *big.Int, _receiver common.Address, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint(&_Router.TransactOpts, _sourceChain, _transactionId, _wrappedToken, _amount, _receiver, _signatures)
}

// SetServiceFee is a paid mutator transaction binding the contract method 0x35cba138.
//
// Solidity: function setServiceFee(address _token, uint256 _serviceFeePercentage, bytes[] _signatures) returns()
func (_Router *RouterTransactor) SetServiceFee(opts *bind.TransactOpts, _token common.Address, _serviceFeePercentage *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "setServiceFee", _token, _serviceFeePercentage, _signatures)
}

// SetServiceFee is a paid mutator transaction binding the contract method 0x35cba138.
//
// Solidity: function setServiceFee(address _token, uint256 _serviceFeePercentage, bytes[] _signatures) returns()
func (_Router *RouterSession) SetServiceFee(_token common.Address, _serviceFeePercentage *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.SetServiceFee(&_Router.TransactOpts, _token, _serviceFeePercentage, _signatures)
}

// SetServiceFee is a paid mutator transaction binding the contract method 0x35cba138.
//
// Solidity: function setServiceFee(address _token, uint256 _serviceFeePercentage, bytes[] _signatures) returns()
func (_Router *RouterTransactorSession) SetServiceFee(_token common.Address, _serviceFeePercentage *big.Int, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.SetServiceFee(&_Router.TransactOpts, _token, _serviceFeePercentage, _signatures)
}

// Unlock is a paid mutator transaction binding the contract method 0x3b68d993.
//
// Solidity: function unlock(uint256 _sourceChain, bytes _transactionId, address _nativeToken, uint256 _amount, address _receiver, bytes[] _signatures) returns()
func (_Router *RouterTransactor) Unlock(opts *bind.TransactOpts, _sourceChain *big.Int, _transactionId []byte, _nativeToken common.Address, _amount *big.Int, _receiver common.Address, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "unlock", _sourceChain, _transactionId, _nativeToken, _amount, _receiver, _signatures)
}

// Unlock is a paid mutator transaction binding the contract method 0x3b68d993.
//
// Solidity: function unlock(uint256 _sourceChain, bytes _transactionId, address _nativeToken, uint256 _amount, address _receiver, bytes[] _signatures) returns()
func (_Router *RouterSession) Unlock(_sourceChain *big.Int, _transactionId []byte, _nativeToken common.Address, _amount *big.Int, _receiver common.Address, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Unlock(&_Router.TransactOpts, _sourceChain, _transactionId, _nativeToken, _amount, _receiver, _signatures)
}

// Unlock is a paid mutator transaction binding the contract method 0x3b68d993.
//
// Solidity: function unlock(uint256 _sourceChain, bytes _transactionId, address _nativeToken, uint256 _amount, address _receiver, bytes[] _signatures) returns()
func (_Router *RouterTransactorSession) Unlock(_sourceChain *big.Int, _transactionId []byte, _nativeToken common.Address, _amount *big.Int, _receiver common.Address, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Unlock(&_Router.TransactOpts, _sourceChain, _transactionId, _nativeToken, _amount, _receiver, _signatures)
}

// UpdateMember is a paid mutator transaction binding the contract method 0x9b6af648.
//
// Solidity: function updateMember(address _account, bool _status, bytes[] _signatures) returns()
func (_Router *RouterTransactor) UpdateMember(opts *bind.TransactOpts, _account common.Address, _status bool, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "updateMember", _account, _status, _signatures)
}

// UpdateMember is a paid mutator transaction binding the contract method 0x9b6af648.
//
// Solidity: function updateMember(address _account, bool _status, bytes[] _signatures) returns()
func (_Router *RouterSession) UpdateMember(_account common.Address, _status bool, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.UpdateMember(&_Router.TransactOpts, _account, _status, _signatures)
}

// UpdateMember is a paid mutator transaction binding the contract method 0x9b6af648.
//
// Solidity: function updateMember(address _account, bool _status, bytes[] _signatures) returns()
func (_Router *RouterTransactorSession) UpdateMember(_account common.Address, _status bool, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.UpdateMember(&_Router.TransactOpts, _account, _status, _signatures)
}

// UpdateNativeToken is a paid mutator transaction binding the contract method 0x61dba72f.
//
// Solidity: function updateNativeToken(address _nativeToken, uint256 _serviceFee, bool _status, bytes[] _signatures) returns()
func (_Router *RouterTransactor) UpdateNativeToken(opts *bind.TransactOpts, _nativeToken common.Address, _serviceFee *big.Int, _status bool, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "updateNativeToken", _nativeToken, _serviceFee, _status, _signatures)
}

// UpdateNativeToken is a paid mutator transaction binding the contract method 0x61dba72f.
//
// Solidity: function updateNativeToken(address _nativeToken, uint256 _serviceFee, bool _status, bytes[] _signatures) returns()
func (_Router *RouterSession) UpdateNativeToken(_nativeToken common.Address, _serviceFee *big.Int, _status bool, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.UpdateNativeToken(&_Router.TransactOpts, _nativeToken, _serviceFee, _status, _signatures)
}

// UpdateNativeToken is a paid mutator transaction binding the contract method 0x61dba72f.
//
// Solidity: function updateNativeToken(address _nativeToken, uint256 _serviceFee, bool _status, bytes[] _signatures) returns()
func (_Router *RouterTransactorSession) UpdateNativeToken(_nativeToken common.Address, _serviceFee *big.Int, _status bool, _signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.UpdateNativeToken(&_Router.TransactOpts, _nativeToken, _serviceFee, _status, _signatures)
}

// RouterBurnIterator is returned from FilterBurn and is used to iterate over the raw logs and unpacked data for Burn events raised by the Router contract.
type RouterBurnIterator struct {
	Event *RouterBurn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterBurnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterBurn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterBurn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterBurnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterBurnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterBurn represents a Burn event raised by the Router contract.
type RouterBurn struct {
	Token    common.Address
	Amount   *big.Int
	Receiver []byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0x8d38f5a0c1764ff1cca876ce8fe136163fddfce925659e6ad05437cfff6fd392.
//
// Solidity: event Burn(address token, uint256 amount, bytes receiver)
func (_Router *RouterFilterer) FilterBurn(opts *bind.FilterOpts) (*RouterBurnIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "Burn")
	if err != nil {
		return nil, err
	}
	return &RouterBurnIterator{contract: _Router.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0x8d38f5a0c1764ff1cca876ce8fe136163fddfce925659e6ad05437cfff6fd392.
//
// Solidity: event Burn(address token, uint256 amount, bytes receiver)
func (_Router *RouterFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *RouterBurn) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "Burn")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterBurn)
				if err := _Router.contract.UnpackLog(event, "Burn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBurn is a log parse operation binding the contract event 0x8d38f5a0c1764ff1cca876ce8fe136163fddfce925659e6ad05437cfff6fd392.
//
// Solidity: event Burn(address token, uint256 amount, bytes receiver)
func (_Router *RouterFilterer) ParseBurn(log types.Log) (*RouterBurn, error) {
	event := new(RouterBurn)
	if err := _Router.contract.UnpackLog(event, "Burn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterClaimIterator is returned from FilterClaim and is used to iterate over the raw logs and unpacked data for Claim events raised by the Router contract.
type RouterClaimIterator struct {
	Event *RouterClaim // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterClaimIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterClaim)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterClaim)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterClaimIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterClaimIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterClaim represents a Claim event raised by the Router contract.
type RouterClaim struct {
	Member common.Address
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterClaim is a free log retrieval operation binding the contract event 0x70eb43c4a8ae8c40502dcf22436c509c28d6ff421cf07c491be56984bd987068.
//
// Solidity: event Claim(address member, address token, uint256 amount)
func (_Router *RouterFilterer) FilterClaim(opts *bind.FilterOpts) (*RouterClaimIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "Claim")
	if err != nil {
		return nil, err
	}
	return &RouterClaimIterator{contract: _Router.contract, event: "Claim", logs: logs, sub: sub}, nil
}

// WatchClaim is a free log subscription operation binding the contract event 0x70eb43c4a8ae8c40502dcf22436c509c28d6ff421cf07c491be56984bd987068.
//
// Solidity: event Claim(address member, address token, uint256 amount)
func (_Router *RouterFilterer) WatchClaim(opts *bind.WatchOpts, sink chan<- *RouterClaim) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "Claim")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterClaim)
				if err := _Router.contract.UnpackLog(event, "Claim", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseClaim is a log parse operation binding the contract event 0x70eb43c4a8ae8c40502dcf22436c509c28d6ff421cf07c491be56984bd987068.
//
// Solidity: event Claim(address member, address token, uint256 amount)
func (_Router *RouterFilterer) ParseClaim(log types.Log) (*RouterClaim, error) {
	event := new(RouterClaim)
	if err := _Router.contract.UnpackLog(event, "Claim", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterDiamondCutIterator is returned from FilterDiamondCut and is used to iterate over the raw logs and unpacked data for DiamondCut events raised by the Router contract.
type RouterDiamondCutIterator struct {
	Event *RouterDiamondCut // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterDiamondCutIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterDiamondCut)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterDiamondCut)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterDiamondCutIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterDiamondCutIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterDiamondCut represents a DiamondCut event raised by the Router contract.
type RouterDiamondCut struct {
	DiamondCut []IDiamondCutFacetCut
	Init       common.Address
	Calldata   []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterDiamondCut is a free log retrieval operation binding the contract event 0x8faa70878671ccd212d20771b795c50af8fd3ff6cf27f4bde57e5d4de0aeb673.
//
// Solidity: event DiamondCut((address,uint8,bytes4[])[] _diamondCut, address _init, bytes _calldata)
func (_Router *RouterFilterer) FilterDiamondCut(opts *bind.FilterOpts) (*RouterDiamondCutIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "DiamondCut")
	if err != nil {
		return nil, err
	}
	return &RouterDiamondCutIterator{contract: _Router.contract, event: "DiamondCut", logs: logs, sub: sub}, nil
}

// WatchDiamondCut is a free log subscription operation binding the contract event 0x8faa70878671ccd212d20771b795c50af8fd3ff6cf27f4bde57e5d4de0aeb673.
//
// Solidity: event DiamondCut((address,uint8,bytes4[])[] _diamondCut, address _init, bytes _calldata)
func (_Router *RouterFilterer) WatchDiamondCut(opts *bind.WatchOpts, sink chan<- *RouterDiamondCut) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "DiamondCut")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterDiamondCut)
				if err := _Router.contract.UnpackLog(event, "DiamondCut", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDiamondCut is a log parse operation binding the contract event 0x8faa70878671ccd212d20771b795c50af8fd3ff6cf27f4bde57e5d4de0aeb673.
//
// Solidity: event DiamondCut((address,uint8,bytes4[])[] _diamondCut, address _init, bytes _calldata)
func (_Router *RouterFilterer) ParseDiamondCut(log types.Log) (*RouterDiamondCut, error) {
	event := new(RouterDiamondCut)
	if err := _Router.contract.UnpackLog(event, "DiamondCut", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterLockIterator is returned from FilterLock and is used to iterate over the raw logs and unpacked data for Lock events raised by the Router contract.
type RouterLockIterator struct {
	Event *RouterLock // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterLockIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterLock)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterLock)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterLockIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterLockIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterLock represents a Lock event raised by the Router contract.
type RouterLock struct {
	TargetChain *big.Int
	Token       common.Address
	Receiver    []byte
	Amount      *big.Int
	ServiceFee  *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterLock is a free log retrieval operation binding the contract event 0xaa3a3bc72b8c754ca6ee8425a5531bafec37569ec012d62d5f682ca909ae06f1.
//
// Solidity: event Lock(uint256 targetChain, address token, bytes receiver, uint256 amount, uint256 serviceFee)
func (_Router *RouterFilterer) FilterLock(opts *bind.FilterOpts) (*RouterLockIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "Lock")
	if err != nil {
		return nil, err
	}
	return &RouterLockIterator{contract: _Router.contract, event: "Lock", logs: logs, sub: sub}, nil
}

// WatchLock is a free log subscription operation binding the contract event 0xaa3a3bc72b8c754ca6ee8425a5531bafec37569ec012d62d5f682ca909ae06f1.
//
// Solidity: event Lock(uint256 targetChain, address token, bytes receiver, uint256 amount, uint256 serviceFee)
func (_Router *RouterFilterer) WatchLock(opts *bind.WatchOpts, sink chan<- *RouterLock) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "Lock")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterLock)
				if err := _Router.contract.UnpackLog(event, "Lock", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseLock is a log parse operation binding the contract event 0xaa3a3bc72b8c754ca6ee8425a5531bafec37569ec012d62d5f682ca909ae06f1.
//
// Solidity: event Lock(uint256 targetChain, address token, bytes receiver, uint256 amount, uint256 serviceFee)
func (_Router *RouterFilterer) ParseLock(log types.Log) (*RouterLock, error) {
	event := new(RouterLock)
	if err := _Router.contract.UnpackLog(event, "Lock", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterMemberUpdatedIterator is returned from FilterMemberUpdated and is used to iterate over the raw logs and unpacked data for MemberUpdated events raised by the Router contract.
type RouterMemberUpdatedIterator struct {
	Event *RouterMemberUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterMemberUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterMemberUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterMemberUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterMemberUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterMemberUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterMemberUpdated represents a MemberUpdated event raised by the Router contract.
type RouterMemberUpdated struct {
	Member common.Address
	Status bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMemberUpdated is a free log retrieval operation binding the contract event 0x30f1d11f11278ba2cc669fd4c95ee8d46ede2c82f6af0b74e4f427369b3522d3.
//
// Solidity: event MemberUpdated(address member, bool status)
func (_Router *RouterFilterer) FilterMemberUpdated(opts *bind.FilterOpts) (*RouterMemberUpdatedIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "MemberUpdated")
	if err != nil {
		return nil, err
	}
	return &RouterMemberUpdatedIterator{contract: _Router.contract, event: "MemberUpdated", logs: logs, sub: sub}, nil
}

// WatchMemberUpdated is a free log subscription operation binding the contract event 0x30f1d11f11278ba2cc669fd4c95ee8d46ede2c82f6af0b74e4f427369b3522d3.
//
// Solidity: event MemberUpdated(address member, bool status)
func (_Router *RouterFilterer) WatchMemberUpdated(opts *bind.WatchOpts, sink chan<- *RouterMemberUpdated) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "MemberUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterMemberUpdated)
				if err := _Router.contract.UnpackLog(event, "MemberUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMemberUpdated is a log parse operation binding the contract event 0x30f1d11f11278ba2cc669fd4c95ee8d46ede2c82f6af0b74e4f427369b3522d3.
//
// Solidity: event MemberUpdated(address member, bool status)
func (_Router *RouterFilterer) ParseMemberUpdated(log types.Log) (*RouterMemberUpdated, error) {
	event := new(RouterMemberUpdated)
	if err := _Router.contract.UnpackLog(event, "MemberUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterMintIterator is returned from FilterMint and is used to iterate over the raw logs and unpacked data for Mint events raised by the Router contract.
type RouterMintIterator struct {
	Event *RouterMint // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterMintIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterMint)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterMint)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterMintIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterMintIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterMint represents a Mint event raised by the Router contract.
type RouterMint struct {
	Token    common.Address
	Amount   *big.Int
	Receiver common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0xbcad3d7d3dfccb90d49c6063bf70f828901fefc88937d90af74e58e6e55bc39d.
//
// Solidity: event Mint(address token, uint256 amount, address receiver)
func (_Router *RouterFilterer) FilterMint(opts *bind.FilterOpts) (*RouterMintIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "Mint")
	if err != nil {
		return nil, err
	}
	return &RouterMintIterator{contract: _Router.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0xbcad3d7d3dfccb90d49c6063bf70f828901fefc88937d90af74e58e6e55bc39d.
//
// Solidity: event Mint(address token, uint256 amount, address receiver)
func (_Router *RouterFilterer) WatchMint(opts *bind.WatchOpts, sink chan<- *RouterMint) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "Mint")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterMint)
				if err := _Router.contract.UnpackLog(event, "Mint", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMint is a log parse operation binding the contract event 0xbcad3d7d3dfccb90d49c6063bf70f828901fefc88937d90af74e58e6e55bc39d.
//
// Solidity: event Mint(address token, uint256 amount, address receiver)
func (_Router *RouterFilterer) ParseMint(log types.Log) (*RouterMint, error) {
	event := new(RouterMint)
	if err := _Router.contract.UnpackLog(event, "Mint", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterNativeTokenUpdatedIterator is returned from FilterNativeTokenUpdated and is used to iterate over the raw logs and unpacked data for NativeTokenUpdated events raised by the Router contract.
type RouterNativeTokenUpdatedIterator struct {
	Event *RouterNativeTokenUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterNativeTokenUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterNativeTokenUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterNativeTokenUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterNativeTokenUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterNativeTokenUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterNativeTokenUpdated represents a NativeTokenUpdated event raised by the Router contract.
type RouterNativeTokenUpdated struct {
	Token      common.Address
	ServiceFee *big.Int
	Status     bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterNativeTokenUpdated is a free log retrieval operation binding the contract event 0x62f51bef49e8a6a5d65e8aef0916ba65fc03e95b3c5c828b6c065f357a24dd34.
//
// Solidity: event NativeTokenUpdated(address token, uint256 serviceFee, bool status)
func (_Router *RouterFilterer) FilterNativeTokenUpdated(opts *bind.FilterOpts) (*RouterNativeTokenUpdatedIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "NativeTokenUpdated")
	if err != nil {
		return nil, err
	}
	return &RouterNativeTokenUpdatedIterator{contract: _Router.contract, event: "NativeTokenUpdated", logs: logs, sub: sub}, nil
}

// WatchNativeTokenUpdated is a free log subscription operation binding the contract event 0x62f51bef49e8a6a5d65e8aef0916ba65fc03e95b3c5c828b6c065f357a24dd34.
//
// Solidity: event NativeTokenUpdated(address token, uint256 serviceFee, bool status)
func (_Router *RouterFilterer) WatchNativeTokenUpdated(opts *bind.WatchOpts, sink chan<- *RouterNativeTokenUpdated) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "NativeTokenUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterNativeTokenUpdated)
				if err := _Router.contract.UnpackLog(event, "NativeTokenUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNativeTokenUpdated is a log parse operation binding the contract event 0x62f51bef49e8a6a5d65e8aef0916ba65fc03e95b3c5c828b6c065f357a24dd34.
//
// Solidity: event NativeTokenUpdated(address token, uint256 serviceFee, bool status)
func (_Router *RouterFilterer) ParseNativeTokenUpdated(log types.Log) (*RouterNativeTokenUpdated, error) {
	event := new(RouterNativeTokenUpdated)
	if err := _Router.contract.UnpackLog(event, "NativeTokenUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterServiceFeeSetIterator is returned from FilterServiceFeeSet and is used to iterate over the raw logs and unpacked data for ServiceFeeSet events raised by the Router contract.
type RouterServiceFeeSetIterator struct {
	Event *RouterServiceFeeSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterServiceFeeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterServiceFeeSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterServiceFeeSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterServiceFeeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterServiceFeeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterServiceFeeSet represents a ServiceFeeSet event raised by the Router contract.
type RouterServiceFeeSet struct {
	Account       common.Address
	Token         common.Address
	NewServiceFee *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterServiceFeeSet is a free log retrieval operation binding the contract event 0xdf569f3a847aa48bacde580cf8f9884aee143cebb7d535609b1ba812fdf65e96.
//
// Solidity: event ServiceFeeSet(address account, address token, uint256 newServiceFee)
func (_Router *RouterFilterer) FilterServiceFeeSet(opts *bind.FilterOpts) (*RouterServiceFeeSetIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "ServiceFeeSet")
	if err != nil {
		return nil, err
	}
	return &RouterServiceFeeSetIterator{contract: _Router.contract, event: "ServiceFeeSet", logs: logs, sub: sub}, nil
}

// WatchServiceFeeSet is a free log subscription operation binding the contract event 0xdf569f3a847aa48bacde580cf8f9884aee143cebb7d535609b1ba812fdf65e96.
//
// Solidity: event ServiceFeeSet(address account, address token, uint256 newServiceFee)
func (_Router *RouterFilterer) WatchServiceFeeSet(opts *bind.WatchOpts, sink chan<- *RouterServiceFeeSet) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "ServiceFeeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterServiceFeeSet)
				if err := _Router.contract.UnpackLog(event, "ServiceFeeSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseServiceFeeSet is a log parse operation binding the contract event 0xdf569f3a847aa48bacde580cf8f9884aee143cebb7d535609b1ba812fdf65e96.
//
// Solidity: event ServiceFeeSet(address account, address token, uint256 newServiceFee)
func (_Router *RouterFilterer) ParseServiceFeeSet(log types.Log) (*RouterServiceFeeSet, error) {
	event := new(RouterServiceFeeSet)
	if err := _Router.contract.UnpackLog(event, "ServiceFeeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterUnlockIterator is returned from FilterUnlock and is used to iterate over the raw logs and unpacked data for Unlock events raised by the Router contract.
type RouterUnlockIterator struct {
	Event *RouterUnlock // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterUnlockIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterUnlock)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterUnlock)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterUnlockIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterUnlockIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterUnlock represents a Unlock event raised by the Router contract.
type RouterUnlock struct {
	Token      common.Address
	Amount     *big.Int
	Receiver   common.Address
	ServiceFee *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterUnlock is a free log retrieval operation binding the contract event 0xe9dd21edb9d48d83b9a32dee73f3307e8a8bcf2b8bc87bb7974dbd0510e53caa.
//
// Solidity: event Unlock(address token, uint256 amount, address receiver, uint256 serviceFee)
func (_Router *RouterFilterer) FilterUnlock(opts *bind.FilterOpts) (*RouterUnlockIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "Unlock")
	if err != nil {
		return nil, err
	}
	return &RouterUnlockIterator{contract: _Router.contract, event: "Unlock", logs: logs, sub: sub}, nil
}

// WatchUnlock is a free log subscription operation binding the contract event 0xe9dd21edb9d48d83b9a32dee73f3307e8a8bcf2b8bc87bb7974dbd0510e53caa.
//
// Solidity: event Unlock(address token, uint256 amount, address receiver, uint256 serviceFee)
func (_Router *RouterFilterer) WatchUnlock(opts *bind.WatchOpts, sink chan<- *RouterUnlock) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "Unlock")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterUnlock)
				if err := _Router.contract.UnpackLog(event, "Unlock", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUnlock is a log parse operation binding the contract event 0xe9dd21edb9d48d83b9a32dee73f3307e8a8bcf2b8bc87bb7974dbd0510e53caa.
//
// Solidity: event Unlock(address token, uint256 amount, address receiver, uint256 serviceFee)
func (_Router *RouterFilterer) ParseUnlock(log types.Log) (*RouterUnlock, error) {
	event := new(RouterUnlock)
	if err := _Router.contract.UnpackLog(event, "Unlock", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterWrappedTokenDeployedIterator is returned from FilterWrappedTokenDeployed and is used to iterate over the raw logs and unpacked data for WrappedTokenDeployed events raised by the Router contract.
type RouterWrappedTokenDeployedIterator struct {
	Event *RouterWrappedTokenDeployed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *RouterWrappedTokenDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterWrappedTokenDeployed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(RouterWrappedTokenDeployed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *RouterWrappedTokenDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterWrappedTokenDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterWrappedTokenDeployed represents a WrappedTokenDeployed event raised by the Router contract.
type RouterWrappedTokenDeployed struct {
	SourceChain  *big.Int
	NativeToken  []byte
	WrappedToken common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterWrappedTokenDeployed is a free log retrieval operation binding the contract event 0x1ac5f8e99c47a193ef0460626b39362bb79296795b2bbac054bf053222eeab34.
//
// Solidity: event WrappedTokenDeployed(uint256 sourceChain, bytes nativeToken, address wrappedToken)
func (_Router *RouterFilterer) FilterWrappedTokenDeployed(opts *bind.FilterOpts) (*RouterWrappedTokenDeployedIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "WrappedTokenDeployed")
	if err != nil {
		return nil, err
	}
	return &RouterWrappedTokenDeployedIterator{contract: _Router.contract, event: "WrappedTokenDeployed", logs: logs, sub: sub}, nil
}

// WatchWrappedTokenDeployed is a free log subscription operation binding the contract event 0x1ac5f8e99c47a193ef0460626b39362bb79296795b2bbac054bf053222eeab34.
//
// Solidity: event WrappedTokenDeployed(uint256 sourceChain, bytes nativeToken, address wrappedToken)
func (_Router *RouterFilterer) WatchWrappedTokenDeployed(opts *bind.WatchOpts, sink chan<- *RouterWrappedTokenDeployed) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "WrappedTokenDeployed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterWrappedTokenDeployed)
				if err := _Router.contract.UnpackLog(event, "WrappedTokenDeployed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWrappedTokenDeployed is a log parse operation binding the contract event 0x1ac5f8e99c47a193ef0460626b39362bb79296795b2bbac054bf053222eeab34.
//
// Solidity: event WrappedTokenDeployed(uint256 sourceChain, bytes nativeToken, address wrappedToken)
func (_Router *RouterFilterer) ParseWrappedTokenDeployed(log types.Log) (*RouterWrappedTokenDeployed, error) {
	event := new(RouterWrappedTokenDeployed)
	if err := _Router.contract.UnpackLog(event, "WrappedTokenDeployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}