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

// RouterABI is the input ABI used to generate the binding from.
const RouterABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAsset\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"}],\"name\":\"AssetContractSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"serviceFee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"receiver\",\"type\":\"bytes\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Claim\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deprecate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"member\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"MemberUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"serviceFeeInWTokens\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"txCost\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newServiceFee\",\"type\":\"uint256\"}],\"name\":\"ServiceFeeSet\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"assetToTokenId\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"assetsData\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"feesAccrued\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"previousAccrued\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"accumulator\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_member\",\"type\":\"address\"}],\"name\":\"isMember\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"memberAt\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"membersCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"mintTransfers\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"isExecuted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"serviceFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_serviceFee\",\"type\":\"uint256\"}],\"name\":\"setServiceFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"tokenIdToAsset\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"txCost\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"receiver\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"isMember\",\"type\":\"bool\"}],\"name\":\"updateMember\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAsset\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"tokenID\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"}],\"name\":\"updateAsset\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"}],\"name\":\"isAsset\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"assetsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"assetAt\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

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

// AssetAt is a free data retrieval call binding the contract method 0xaa9239f5.
//
// Solidity: function assetAt(uint256 index) view returns(address)
func (_Router *RouterCaller) AssetAt(opts *bind.CallOpts, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "assetAt", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AssetAt is a free data retrieval call binding the contract method 0xaa9239f5.
//
// Solidity: function assetAt(uint256 index) view returns(address)
func (_Router *RouterSession) AssetAt(index *big.Int) (common.Address, error) {
	return _Router.Contract.AssetAt(&_Router.CallOpts, index)
}

// AssetAt is a free data retrieval call binding the contract method 0xaa9239f5.
//
// Solidity: function assetAt(uint256 index) view returns(address)
func (_Router *RouterCallerSession) AssetAt(index *big.Int) (common.Address, error) {
	return _Router.Contract.AssetAt(&_Router.CallOpts, index)
}

// AssetToTokenId is a free data retrieval call binding the contract method 0xecec967c.
//
// Solidity: function assetToTokenId(address ) view returns(bytes)
func (_Router *RouterCaller) AssetToTokenId(opts *bind.CallOpts, arg0 common.Address) ([]byte, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "assetToTokenId", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// AssetToTokenId is a free data retrieval call binding the contract method 0xecec967c.
//
// Solidity: function assetToTokenId(address ) view returns(bytes)
func (_Router *RouterSession) AssetToTokenId(arg0 common.Address) ([]byte, error) {
	return _Router.Contract.AssetToTokenId(&_Router.CallOpts, arg0)
}

// AssetToTokenId is a free data retrieval call binding the contract method 0xecec967c.
//
// Solidity: function assetToTokenId(address ) view returns(bytes)
func (_Router *RouterCallerSession) AssetToTokenId(arg0 common.Address) ([]byte, error) {
	return _Router.Contract.AssetToTokenId(&_Router.CallOpts, arg0)
}

// AssetsCount is a free data retrieval call binding the contract method 0xcd9df190.
//
// Solidity: function assetsCount() view returns(uint256)
func (_Router *RouterCaller) AssetsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "assetsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AssetsCount is a free data retrieval call binding the contract method 0xcd9df190.
//
// Solidity: function assetsCount() view returns(uint256)
func (_Router *RouterSession) AssetsCount() (*big.Int, error) {
	return _Router.Contract.AssetsCount(&_Router.CallOpts)
}

// AssetsCount is a free data retrieval call binding the contract method 0xcd9df190.
//
// Solidity: function assetsCount() view returns(uint256)
func (_Router *RouterCallerSession) AssetsCount() (*big.Int, error) {
	return _Router.Contract.AssetsCount(&_Router.CallOpts)
}

// AssetsData is a free data retrieval call binding the contract method 0x35558e9b.
//
// Solidity: function assetsData(address ) view returns(uint256 feesAccrued, uint256 previousAccrued, uint256 accumulator)
func (_Router *RouterCaller) AssetsData(opts *bind.CallOpts, arg0 common.Address) (struct {
	FeesAccrued     *big.Int
	PreviousAccrued *big.Int
	Accumulator     *big.Int
}, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "assetsData", arg0)

	outstruct := new(struct {
		FeesAccrued     *big.Int
		PreviousAccrued *big.Int
		Accumulator     *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.FeesAccrued = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.PreviousAccrued = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Accumulator = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// AssetsData is a free data retrieval call binding the contract method 0x35558e9b.
//
// Solidity: function assetsData(address ) view returns(uint256 feesAccrued, uint256 previousAccrued, uint256 accumulator)
func (_Router *RouterSession) AssetsData(arg0 common.Address) (struct {
	FeesAccrued     *big.Int
	PreviousAccrued *big.Int
	Accumulator     *big.Int
}, error) {
	return _Router.Contract.AssetsData(&_Router.CallOpts, arg0)
}

// AssetsData is a free data retrieval call binding the contract method 0x35558e9b.
//
// Solidity: function assetsData(address ) view returns(uint256 feesAccrued, uint256 previousAccrued, uint256 accumulator)
func (_Router *RouterCallerSession) AssetsData(arg0 common.Address) (struct {
	FeesAccrued     *big.Int
	PreviousAccrued *big.Int
	Accumulator     *big.Int
}, error) {
	return _Router.Contract.AssetsData(&_Router.CallOpts, arg0)
}

// IsAsset is a free data retrieval call binding the contract method 0xc87fa42a.
//
// Solidity: function isAsset(address asset) view returns(bool)
func (_Router *RouterCaller) IsAsset(opts *bind.CallOpts, asset common.Address) (bool, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "isAsset", asset)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAsset is a free data retrieval call binding the contract method 0xc87fa42a.
//
// Solidity: function isAsset(address asset) view returns(bool)
func (_Router *RouterSession) IsAsset(asset common.Address) (bool, error) {
	return _Router.Contract.IsAsset(&_Router.CallOpts, asset)
}

// IsAsset is a free data retrieval call binding the contract method 0xc87fa42a.
//
// Solidity: function isAsset(address asset) view returns(bool)
func (_Router *RouterCallerSession) IsAsset(asset common.Address) (bool, error) {
	return _Router.Contract.IsAsset(&_Router.CallOpts, asset)
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
// Solidity: function memberAt(uint256 index) view returns(address)
func (_Router *RouterCaller) MemberAt(opts *bind.CallOpts, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "memberAt", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MemberAt is a free data retrieval call binding the contract method 0xac0250f7.
//
// Solidity: function memberAt(uint256 index) view returns(address)
func (_Router *RouterSession) MemberAt(index *big.Int) (common.Address, error) {
	return _Router.Contract.MemberAt(&_Router.CallOpts, index)
}

// MemberAt is a free data retrieval call binding the contract method 0xac0250f7.
//
// Solidity: function memberAt(uint256 index) view returns(address)
func (_Router *RouterCallerSession) MemberAt(index *big.Int) (common.Address, error) {
	return _Router.Contract.MemberAt(&_Router.CallOpts, index)
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

// MintTransfers is a free data retrieval call binding the contract method 0xed1018ad.
//
// Solidity: function mintTransfers(bytes ) view returns(bool isExecuted)
func (_Router *RouterCaller) MintTransfers(opts *bind.CallOpts, arg0 []byte) (bool, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "mintTransfers", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MintTransfers is a free data retrieval call binding the contract method 0xed1018ad.
//
// Solidity: function mintTransfers(bytes ) view returns(bool isExecuted)
func (_Router *RouterSession) MintTransfers(arg0 []byte) (bool, error) {
	return _Router.Contract.MintTransfers(&_Router.CallOpts, arg0)
}

// MintTransfers is a free data retrieval call binding the contract method 0xed1018ad.
//
// Solidity: function mintTransfers(bytes ) view returns(bool isExecuted)
func (_Router *RouterCallerSession) MintTransfers(arg0 []byte) (bool, error) {
	return _Router.Contract.MintTransfers(&_Router.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Router *RouterCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Router *RouterSession) Owner() (common.Address, error) {
	return _Router.Contract.Owner(&_Router.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Router *RouterCallerSession) Owner() (common.Address, error) {
	return _Router.Contract.Owner(&_Router.CallOpts)
}

// ServiceFee is a free data retrieval call binding the contract method 0x8abdf5aa.
//
// Solidity: function serviceFee() view returns(uint256)
func (_Router *RouterCaller) ServiceFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "serviceFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ServiceFee is a free data retrieval call binding the contract method 0x8abdf5aa.
//
// Solidity: function serviceFee() view returns(uint256)
func (_Router *RouterSession) ServiceFee() (*big.Int, error) {
	return _Router.Contract.ServiceFee(&_Router.CallOpts)
}

// ServiceFee is a free data retrieval call binding the contract method 0x8abdf5aa.
//
// Solidity: function serviceFee() view returns(uint256)
func (_Router *RouterCallerSession) ServiceFee() (*big.Int, error) {
	return _Router.Contract.ServiceFee(&_Router.CallOpts)
}

// TokenIdToAsset is a free data retrieval call binding the contract method 0xae5e2696.
//
// Solidity: function tokenIdToAsset(bytes ) view returns(address)
func (_Router *RouterCaller) TokenIdToAsset(opts *bind.CallOpts, arg0 []byte) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "tokenIdToAsset", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenIdToAsset is a free data retrieval call binding the contract method 0xae5e2696.
//
// Solidity: function tokenIdToAsset(bytes ) view returns(address)
func (_Router *RouterSession) TokenIdToAsset(arg0 []byte) (common.Address, error) {
	return _Router.Contract.TokenIdToAsset(&_Router.CallOpts, arg0)
}

// TokenIdToAsset is a free data retrieval call binding the contract method 0xae5e2696.
//
// Solidity: function tokenIdToAsset(bytes ) view returns(address)
func (_Router *RouterCallerSession) TokenIdToAsset(arg0 []byte) (common.Address, error) {
	return _Router.Contract.TokenIdToAsset(&_Router.CallOpts, arg0)
}

// Burn is a paid mutator transaction binding the contract method 0x8c99319f.
//
// Solidity: function burn(uint256 amount, bytes receiver, address asset) returns()
func (_Router *RouterTransactor) Burn(opts *bind.TransactOpts, amount *big.Int, receiver []byte, asset common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "burn", amount, receiver, asset)
}

// Burn is a paid mutator transaction binding the contract method 0x8c99319f.
//
// Solidity: function burn(uint256 amount, bytes receiver, address asset) returns()
func (_Router *RouterSession) Burn(amount *big.Int, receiver []byte, asset common.Address) (*types.Transaction, error) {
	return _Router.Contract.Burn(&_Router.TransactOpts, amount, receiver, asset)
}

// Burn is a paid mutator transaction binding the contract method 0x8c99319f.
//
// Solidity: function burn(uint256 amount, bytes receiver, address asset) returns()
func (_Router *RouterTransactorSession) Burn(amount *big.Int, receiver []byte, asset common.Address) (*types.Transaction, error) {
	return _Router.Contract.Burn(&_Router.TransactOpts, amount, receiver, asset)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address asset) returns()
func (_Router *RouterTransactor) Claim(opts *bind.TransactOpts, asset common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "claim", asset)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address asset) returns()
func (_Router *RouterSession) Claim(asset common.Address) (*types.Transaction, error) {
	return _Router.Contract.Claim(&_Router.TransactOpts, asset)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address asset) returns()
func (_Router *RouterTransactorSession) Claim(asset common.Address) (*types.Transaction, error) {
	return _Router.Contract.Claim(&_Router.TransactOpts, asset)
}

// Mint is a paid mutator transaction binding the contract method 0x724ae0eb.
//
// Solidity: function mint(bytes transactionId, address asset, address receiver, uint256 amount, uint256 txCost, bytes[] signatures) returns()
func (_Router *RouterTransactor) Mint(opts *bind.TransactOpts, transactionId []byte, asset common.Address, receiver common.Address, amount *big.Int, txCost *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "mint", transactionId, asset, receiver, amount, txCost, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x724ae0eb.
//
// Solidity: function mint(bytes transactionId, address asset, address receiver, uint256 amount, uint256 txCost, bytes[] signatures) returns()
func (_Router *RouterSession) Mint(transactionId []byte, asset common.Address, receiver common.Address, amount *big.Int, txCost *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint(&_Router.TransactOpts, transactionId, asset, receiver, amount, txCost, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x724ae0eb.
//
// Solidity: function mint(bytes transactionId, address asset, address receiver, uint256 amount, uint256 txCost, bytes[] signatures) returns()
func (_Router *RouterTransactorSession) Mint(transactionId []byte, asset common.Address, receiver common.Address, amount *big.Int, txCost *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint(&_Router.TransactOpts, transactionId, asset, receiver, amount, txCost, signatures)
}

// Mint0 is a paid mutator transaction binding the contract method 0x9949e912.
//
// Solidity: function mint(bytes transactionId, address asset, address receiver, uint256 amount, bytes[] signatures) returns()
func (_Router *RouterTransactor) Mint0(opts *bind.TransactOpts, transactionId []byte, asset common.Address, receiver common.Address, amount *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "mint0", transactionId, asset, receiver, amount, signatures)
}

// Mint0 is a paid mutator transaction binding the contract method 0x9949e912.
//
// Solidity: function mint(bytes transactionId, address asset, address receiver, uint256 amount, bytes[] signatures) returns()
func (_Router *RouterSession) Mint0(transactionId []byte, asset common.Address, receiver common.Address, amount *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint0(&_Router.TransactOpts, transactionId, asset, receiver, amount, signatures)
}

// Mint0 is a paid mutator transaction binding the contract method 0x9949e912.
//
// Solidity: function mint(bytes transactionId, address asset, address receiver, uint256 amount, bytes[] signatures) returns()
func (_Router *RouterTransactorSession) Mint0(transactionId []byte, asset common.Address, receiver common.Address, amount *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint0(&_Router.TransactOpts, transactionId, asset, receiver, amount, signatures)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Router *RouterTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Router *RouterSession) RenounceOwnership() (*types.Transaction, error) {
	return _Router.Contract.RenounceOwnership(&_Router.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Router *RouterTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Router.Contract.RenounceOwnership(&_Router.TransactOpts)
}

// SetServiceFee is a paid mutator transaction binding the contract method 0x5cdf76f8.
//
// Solidity: function setServiceFee(uint256 _serviceFee) returns()
func (_Router *RouterTransactor) SetServiceFee(opts *bind.TransactOpts, _serviceFee *big.Int) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "setServiceFee", _serviceFee)
}

// SetServiceFee is a paid mutator transaction binding the contract method 0x5cdf76f8.
//
// Solidity: function setServiceFee(uint256 _serviceFee) returns()
func (_Router *RouterSession) SetServiceFee(_serviceFee *big.Int) (*types.Transaction, error) {
	return _Router.Contract.SetServiceFee(&_Router.TransactOpts, _serviceFee)
}

// SetServiceFee is a paid mutator transaction binding the contract method 0x5cdf76f8.
//
// Solidity: function setServiceFee(uint256 _serviceFee) returns()
func (_Router *RouterTransactorSession) SetServiceFee(_serviceFee *big.Int) (*types.Transaction, error) {
	return _Router.Contract.SetServiceFee(&_Router.TransactOpts, _serviceFee)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Router *RouterTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Router *RouterSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Router.Contract.TransferOwnership(&_Router.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Router *RouterTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Router.Contract.TransferOwnership(&_Router.TransactOpts, newOwner)
}

// UpdateAsset is a paid mutator transaction binding the contract method 0x86545fa4.
//
// Solidity: function updateAsset(address newAsset, bytes tokenID, bool isActive) returns()
func (_Router *RouterTransactor) UpdateAsset(opts *bind.TransactOpts, newAsset common.Address, tokenID []byte, isActive bool) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "updateAsset", newAsset, tokenID, isActive)
}

// UpdateAsset is a paid mutator transaction binding the contract method 0x86545fa4.
//
// Solidity: function updateAsset(address newAsset, bytes tokenID, bool isActive) returns()
func (_Router *RouterSession) UpdateAsset(newAsset common.Address, tokenID []byte, isActive bool) (*types.Transaction, error) {
	return _Router.Contract.UpdateAsset(&_Router.TransactOpts, newAsset, tokenID, isActive)
}

// UpdateAsset is a paid mutator transaction binding the contract method 0x86545fa4.
//
// Solidity: function updateAsset(address newAsset, bytes tokenID, bool isActive) returns()
func (_Router *RouterTransactorSession) UpdateAsset(newAsset common.Address, tokenID []byte, isActive bool) (*types.Transaction, error) {
	return _Router.Contract.UpdateAsset(&_Router.TransactOpts, newAsset, tokenID, isActive)
}

// UpdateMember is a paid mutator transaction binding the contract method 0x05566101.
//
// Solidity: function updateMember(address account, bool isMember) returns()
func (_Router *RouterTransactor) UpdateMember(opts *bind.TransactOpts, account common.Address, isMember bool) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "updateMember", account, isMember)
}

// UpdateMember is a paid mutator transaction binding the contract method 0x05566101.
//
// Solidity: function updateMember(address account, bool isMember) returns()
func (_Router *RouterSession) UpdateMember(account common.Address, isMember bool) (*types.Transaction, error) {
	return _Router.Contract.UpdateMember(&_Router.TransactOpts, account, isMember)
}

// UpdateMember is a paid mutator transaction binding the contract method 0x05566101.
//
// Solidity: function updateMember(address account, bool isMember) returns()
func (_Router *RouterTransactorSession) UpdateMember(account common.Address, isMember bool) (*types.Transaction, error) {
	return _Router.Contract.UpdateMember(&_Router.TransactOpts, account, isMember)
}

// RouterAssetContractSetIterator is returned from FilterAssetContractSet and is used to iterate over the raw logs and unpacked data for AssetContractSet events raised by the Router contract.
type RouterAssetContractSetIterator struct {
	Event *RouterAssetContractSet // Event containing the contract specifics and raw log

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
func (it *RouterAssetContractSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterAssetContractSet)
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
		it.Event = new(RouterAssetContractSet)
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
func (it *RouterAssetContractSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterAssetContractSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterAssetContractSet represents a AssetContractSet event raised by the Router contract.
type RouterAssetContractSet struct {
	NewAsset common.Address
	IsActive bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterAssetContractSet is a free log retrieval operation binding the contract event 0x46b131307a2dfcf4a7c0e56a0f8e948d43e7e652c64cf5c227c255d3fb91a0fc.
//
// Solidity: event AssetContractSet(address newAsset, bool isActive)
func (_Router *RouterFilterer) FilterAssetContractSet(opts *bind.FilterOpts) (*RouterAssetContractSetIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "AssetContractSet")
	if err != nil {
		return nil, err
	}
	return &RouterAssetContractSetIterator{contract: _Router.contract, event: "AssetContractSet", logs: logs, sub: sub}, nil
}

// WatchAssetContractSet is a free log subscription operation binding the contract event 0x46b131307a2dfcf4a7c0e56a0f8e948d43e7e652c64cf5c227c255d3fb91a0fc.
//
// Solidity: event AssetContractSet(address newAsset, bool isActive)
func (_Router *RouterFilterer) WatchAssetContractSet(opts *bind.WatchOpts, sink chan<- *RouterAssetContractSet) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "AssetContractSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterAssetContractSet)
				if err := _Router.contract.UnpackLog(event, "AssetContractSet", log); err != nil {
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

// ParseAssetContractSet is a log parse operation binding the contract event 0x46b131307a2dfcf4a7c0e56a0f8e948d43e7e652c64cf5c227c255d3fb91a0fc.
//
// Solidity: event AssetContractSet(address newAsset, bool isActive)
func (_Router *RouterFilterer) ParseAssetContractSet(log types.Log) (*RouterAssetContractSet, error) {
	event := new(RouterAssetContractSet)
	if err := _Router.contract.UnpackLog(event, "AssetContractSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
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
	Account    common.Address
	Amount     *big.Int
	ServiceFee *big.Int
	Receiver   []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0x8da2fc26da2245514483a393963ce93cac8be27cf30bbbc78569ff2ffe3eda16.
//
// Solidity: event Burn(address indexed account, uint256 amount, uint256 serviceFee, bytes receiver)
func (_Router *RouterFilterer) FilterBurn(opts *bind.FilterOpts, account []common.Address) (*RouterBurnIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Router.contract.FilterLogs(opts, "Burn", accountRule)
	if err != nil {
		return nil, err
	}
	return &RouterBurnIterator{contract: _Router.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0x8da2fc26da2245514483a393963ce93cac8be27cf30bbbc78569ff2ffe3eda16.
//
// Solidity: event Burn(address indexed account, uint256 amount, uint256 serviceFee, bytes receiver)
func (_Router *RouterFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *RouterBurn, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Router.contract.WatchLogs(opts, "Burn", accountRule)
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

// ParseBurn is a log parse operation binding the contract event 0x8da2fc26da2245514483a393963ce93cac8be27cf30bbbc78569ff2ffe3eda16.
//
// Solidity: event Burn(address indexed account, uint256 amount, uint256 serviceFee, bytes receiver)
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
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterClaim is a free log retrieval operation binding the contract event 0x47cee97cb7acd717b3c0aa1435d004cd5b3c8c57d70dbceb4e4458bbd60e39d4.
//
// Solidity: event Claim(address indexed account, uint256 amount)
func (_Router *RouterFilterer) FilterClaim(opts *bind.FilterOpts, account []common.Address) (*RouterClaimIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Router.contract.FilterLogs(opts, "Claim", accountRule)
	if err != nil {
		return nil, err
	}
	return &RouterClaimIterator{contract: _Router.contract, event: "Claim", logs: logs, sub: sub}, nil
}

// WatchClaim is a free log subscription operation binding the contract event 0x47cee97cb7acd717b3c0aa1435d004cd5b3c8c57d70dbceb4e4458bbd60e39d4.
//
// Solidity: event Claim(address indexed account, uint256 amount)
func (_Router *RouterFilterer) WatchClaim(opts *bind.WatchOpts, sink chan<- *RouterClaim, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Router.contract.WatchLogs(opts, "Claim", accountRule)
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

// ParseClaim is a log parse operation binding the contract event 0x47cee97cb7acd717b3c0aa1435d004cd5b3c8c57d70dbceb4e4458bbd60e39d4.
//
// Solidity: event Claim(address indexed account, uint256 amount)
func (_Router *RouterFilterer) ParseClaim(log types.Log) (*RouterClaim, error) {
	event := new(RouterClaim)
	if err := _Router.contract.UnpackLog(event, "Claim", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterDeprecateIterator is returned from FilterDeprecate and is used to iterate over the raw logs and unpacked data for Deprecate events raised by the Router contract.
type RouterDeprecateIterator struct {
	Event *RouterDeprecate // Event containing the contract specifics and raw log

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
func (it *RouterDeprecateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterDeprecate)
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
		it.Event = new(RouterDeprecate)
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
func (it *RouterDeprecateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterDeprecateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterDeprecate represents a Deprecate event raised by the Router contract.
type RouterDeprecate struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDeprecate is a free log retrieval operation binding the contract event 0x33a356c68d152c1cdfdce73f7e4644d382dcbe94098932def08c281a37cdf6bd.
//
// Solidity: event Deprecate(address account, uint256 amount)
func (_Router *RouterFilterer) FilterDeprecate(opts *bind.FilterOpts) (*RouterDeprecateIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "Deprecate")
	if err != nil {
		return nil, err
	}
	return &RouterDeprecateIterator{contract: _Router.contract, event: "Deprecate", logs: logs, sub: sub}, nil
}

// WatchDeprecate is a free log subscription operation binding the contract event 0x33a356c68d152c1cdfdce73f7e4644d382dcbe94098932def08c281a37cdf6bd.
//
// Solidity: event Deprecate(address account, uint256 amount)
func (_Router *RouterFilterer) WatchDeprecate(opts *bind.WatchOpts, sink chan<- *RouterDeprecate) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "Deprecate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterDeprecate)
				if err := _Router.contract.UnpackLog(event, "Deprecate", log); err != nil {
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

// ParseDeprecate is a log parse operation binding the contract event 0x33a356c68d152c1cdfdce73f7e4644d382dcbe94098932def08c281a37cdf6bd.
//
// Solidity: event Deprecate(address account, uint256 amount)
func (_Router *RouterFilterer) ParseDeprecate(log types.Log) (*RouterDeprecate, error) {
	event := new(RouterDeprecate)
	if err := _Router.contract.UnpackLog(event, "Deprecate", log); err != nil {
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
	Account             common.Address
	Amount              *big.Int
	ServiceFeeInWTokens *big.Int
	TxCost              *big.Int
	TransactionId       common.Hash
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0xab5e0e0fc4eb29aaba31da7756245e8580bb49834b10faff58b98807a565808f.
//
// Solidity: event Mint(address indexed account, uint256 amount, uint256 serviceFeeInWTokens, uint256 txCost, bytes indexed transactionId)
func (_Router *RouterFilterer) FilterMint(opts *bind.FilterOpts, account []common.Address, transactionId [][]byte) (*RouterMintIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _Router.contract.FilterLogs(opts, "Mint", accountRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &RouterMintIterator{contract: _Router.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0xab5e0e0fc4eb29aaba31da7756245e8580bb49834b10faff58b98807a565808f.
//
// Solidity: event Mint(address indexed account, uint256 amount, uint256 serviceFeeInWTokens, uint256 txCost, bytes indexed transactionId)
func (_Router *RouterFilterer) WatchMint(opts *bind.WatchOpts, sink chan<- *RouterMint, account []common.Address, transactionId [][]byte) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _Router.contract.WatchLogs(opts, "Mint", accountRule, transactionIdRule)
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

// ParseMint is a log parse operation binding the contract event 0xab5e0e0fc4eb29aaba31da7756245e8580bb49834b10faff58b98807a565808f.
//
// Solidity: event Mint(address indexed account, uint256 amount, uint256 serviceFeeInWTokens, uint256 txCost, bytes indexed transactionId)
func (_Router *RouterFilterer) ParseMint(log types.Log) (*RouterMint, error) {
	event := new(RouterMint)
	if err := _Router.contract.UnpackLog(event, "Mint", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Router contract.
type RouterOwnershipTransferredIterator struct {
	Event *RouterOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *RouterOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterOwnershipTransferred)
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
		it.Event = new(RouterOwnershipTransferred)
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
func (it *RouterOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterOwnershipTransferred represents a OwnershipTransferred event raised by the Router contract.
type RouterOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Router *RouterFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*RouterOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Router.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &RouterOwnershipTransferredIterator{contract: _Router.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Router *RouterFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *RouterOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Router.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterOwnershipTransferred)
				if err := _Router.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Router *RouterFilterer) ParseOwnershipTransferred(log types.Log) (*RouterOwnershipTransferred, error) {
	event := new(RouterOwnershipTransferred)
	if err := _Router.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
	NewServiceFee *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterServiceFeeSet is a free log retrieval operation binding the contract event 0x1a1f706a54e5071e5d16465e3d926d20df26a148f666efe1bdbfbc65d4f41b5b.
//
// Solidity: event ServiceFeeSet(address account, uint256 newServiceFee)
func (_Router *RouterFilterer) FilterServiceFeeSet(opts *bind.FilterOpts) (*RouterServiceFeeSetIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "ServiceFeeSet")
	if err != nil {
		return nil, err
	}
	return &RouterServiceFeeSetIterator{contract: _Router.contract, event: "ServiceFeeSet", logs: logs, sub: sub}, nil
}

// WatchServiceFeeSet is a free log subscription operation binding the contract event 0x1a1f706a54e5071e5d16465e3d926d20df26a148f666efe1bdbfbc65d4f41b5b.
//
// Solidity: event ServiceFeeSet(address account, uint256 newServiceFee)
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

// ParseServiceFeeSet is a log parse operation binding the contract event 0x1a1f706a54e5071e5d16465e3d926d20df26a148f666efe1bdbfbc65d4f41b5b.
//
// Solidity: event ServiceFeeSet(address account, uint256 newServiceFee)
func (_Router *RouterFilterer) ParseServiceFeeSet(log types.Log) (*RouterServiceFeeSet, error) {
	event := new(RouterServiceFeeSet)
	if err := _Router.contract.UnpackLog(event, "ServiceFeeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
