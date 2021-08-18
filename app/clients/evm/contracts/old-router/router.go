// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package old_router

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
const RouterABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_controller\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wrappedAsset\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"receiver\",\"type\":\"bytes\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"member\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"MemberUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wrappedAsset\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"native\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"wrapped\",\"type\":\"address\"}],\"name\":\"PairAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"native\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"wrapped\",\"type\":\"address\"}],\"name\":\"PairRemoved\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"native\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"wrapped\",\"type\":\"address\"}],\"name\":\"addPair\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"receiver\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"wrappedAsset\",\"type\":\"address\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"wrappedAsset\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"receiver\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"burnWithPermit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"controller\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"executedTransactions\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_member\",\"type\":\"address\"}],\"name\":\"isMember\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"wrappedAsset\",\"type\":\"address\"}],\"name\":\"isSupportedAsset\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"memberAt\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"membersCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"wrappedAsset\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"nativeToWrapped\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"native\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"wrapped\",\"type\":\"address\"}],\"name\":\"removePair\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"isMember\",\"type\":\"bool\"}],\"name\":\"updateMember\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"wrappedAssetAt\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"wrappedAssetsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"wrappedToNative\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

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

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_Router *RouterCaller) Controller(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "controller")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_Router *RouterSession) Controller() (common.Address, error) {
	return _Router.Contract.Controller(&_Router.CallOpts)
}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_Router *RouterCallerSession) Controller() (common.Address, error) {
	return _Router.Contract.Controller(&_Router.CallOpts)
}

// ExecutedTransactions is a free data retrieval call binding the contract method 0x4c2c7559.
//
// Solidity: function executedTransactions(bytes ) view returns(bool)
func (_Router *RouterCaller) ExecutedTransactions(opts *bind.CallOpts, arg0 []byte) (bool, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "executedTransactions", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ExecutedTransactions is a free data retrieval call binding the contract method 0x4c2c7559.
//
// Solidity: function executedTransactions(bytes ) view returns(bool)
func (_Router *RouterSession) ExecutedTransactions(arg0 []byte) (bool, error) {
	return _Router.Contract.ExecutedTransactions(&_Router.CallOpts, arg0)
}

// ExecutedTransactions is a free data retrieval call binding the contract method 0x4c2c7559.
//
// Solidity: function executedTransactions(bytes ) view returns(bool)
func (_Router *RouterCallerSession) ExecutedTransactions(arg0 []byte) (bool, error) {
	return _Router.Contract.ExecutedTransactions(&_Router.CallOpts, arg0)
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

// IsSupportedAsset is a free data retrieval call binding the contract method 0x9be918e6.
//
// Solidity: function isSupportedAsset(address wrappedAsset) view returns(bool)
func (_Router *RouterCaller) IsSupportedAsset(opts *bind.CallOpts, wrappedAsset common.Address) (bool, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "isSupportedAsset", wrappedAsset)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsSupportedAsset is a free data retrieval call binding the contract method 0x9be918e6.
//
// Solidity: function isSupportedAsset(address wrappedAsset) view returns(bool)
func (_Router *RouterSession) IsSupportedAsset(wrappedAsset common.Address) (bool, error) {
	return _Router.Contract.IsSupportedAsset(&_Router.CallOpts, wrappedAsset)
}

// IsSupportedAsset is a free data retrieval call binding the contract method 0x9be918e6.
//
// Solidity: function isSupportedAsset(address wrappedAsset) view returns(bool)
func (_Router *RouterCallerSession) IsSupportedAsset(wrappedAsset common.Address) (bool, error) {
	return _Router.Contract.IsSupportedAsset(&_Router.CallOpts, wrappedAsset)
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

// NativeToWrapped is a free data retrieval call binding the contract method 0x1104296b.
//
// Solidity: function nativeToWrapped(bytes ) view returns(address)
func (_Router *RouterCaller) NativeToWrapped(opts *bind.CallOpts, arg0 []byte) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "nativeToWrapped", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NativeToWrapped is a free data retrieval call binding the contract method 0x1104296b.
//
// Solidity: function nativeToWrapped(bytes ) view returns(address)
func (_Router *RouterSession) NativeToWrapped(arg0 []byte) (common.Address, error) {
	return _Router.Contract.NativeToWrapped(&_Router.CallOpts, arg0)
}

// NativeToWrapped is a free data retrieval call binding the contract method 0x1104296b.
//
// Solidity: function nativeToWrapped(bytes ) view returns(address)
func (_Router *RouterCallerSession) NativeToWrapped(arg0 []byte) (common.Address, error) {
	return _Router.Contract.NativeToWrapped(&_Router.CallOpts, arg0)
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

// WrappedAssetAt is a free data retrieval call binding the contract method 0xa1dce12c.
//
// Solidity: function wrappedAssetAt(uint256 index) view returns(address)
func (_Router *RouterCaller) WrappedAssetAt(opts *bind.CallOpts, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "wrappedAssetAt", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WrappedAssetAt is a free data retrieval call binding the contract method 0xa1dce12c.
//
// Solidity: function wrappedAssetAt(uint256 index) view returns(address)
func (_Router *RouterSession) WrappedAssetAt(index *big.Int) (common.Address, error) {
	return _Router.Contract.WrappedAssetAt(&_Router.CallOpts, index)
}

// WrappedAssetAt is a free data retrieval call binding the contract method 0xa1dce12c.
//
// Solidity: function wrappedAssetAt(uint256 index) view returns(address)
func (_Router *RouterCallerSession) WrappedAssetAt(index *big.Int) (common.Address, error) {
	return _Router.Contract.WrappedAssetAt(&_Router.CallOpts, index)
}

// WrappedAssetsCount is a free data retrieval call binding the contract method 0x868a503c.
//
// Solidity: function wrappedAssetsCount() view returns(uint256)
func (_Router *RouterCaller) WrappedAssetsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "wrappedAssetsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WrappedAssetsCount is a free data retrieval call binding the contract method 0x868a503c.
//
// Solidity: function wrappedAssetsCount() view returns(uint256)
func (_Router *RouterSession) WrappedAssetsCount() (*big.Int, error) {
	return _Router.Contract.WrappedAssetsCount(&_Router.CallOpts)
}

// WrappedAssetsCount is a free data retrieval call binding the contract method 0x868a503c.
//
// Solidity: function wrappedAssetsCount() view returns(uint256)
func (_Router *RouterCallerSession) WrappedAssetsCount() (*big.Int, error) {
	return _Router.Contract.WrappedAssetsCount(&_Router.CallOpts)
}

// WrappedToNative is a free data retrieval call binding the contract method 0xa5e83bcb.
//
// Solidity: function wrappedToNative(address ) view returns(bytes)
func (_Router *RouterCaller) WrappedToNative(opts *bind.CallOpts, arg0 common.Address) ([]byte, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "wrappedToNative", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// WrappedToNative is a free data retrieval call binding the contract method 0xa5e83bcb.
//
// Solidity: function wrappedToNative(address ) view returns(bytes)
func (_Router *RouterSession) WrappedToNative(arg0 common.Address) ([]byte, error) {
	return _Router.Contract.WrappedToNative(&_Router.CallOpts, arg0)
}

// WrappedToNative is a free data retrieval call binding the contract method 0xa5e83bcb.
//
// Solidity: function wrappedToNative(address ) view returns(bytes)
func (_Router *RouterCallerSession) WrappedToNative(arg0 common.Address) ([]byte, error) {
	return _Router.Contract.WrappedToNative(&_Router.CallOpts, arg0)
}

// AddPair is a paid mutator transaction binding the contract method 0x746c26fb.
//
// Solidity: function addPair(bytes native, address wrapped) returns()
func (_Router *RouterTransactor) AddPair(opts *bind.TransactOpts, native []byte, wrapped common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "addPair", native, wrapped)
}

// AddPair is a paid mutator transaction binding the contract method 0x746c26fb.
//
// Solidity: function addPair(bytes native, address wrapped) returns()
func (_Router *RouterSession) AddPair(native []byte, wrapped common.Address) (*types.Transaction, error) {
	return _Router.Contract.AddPair(&_Router.TransactOpts, native, wrapped)
}

// AddPair is a paid mutator transaction binding the contract method 0x746c26fb.
//
// Solidity: function addPair(bytes native, address wrapped) returns()
func (_Router *RouterTransactorSession) AddPair(native []byte, wrapped common.Address) (*types.Transaction, error) {
	return _Router.Contract.AddPair(&_Router.TransactOpts, native, wrapped)
}

// Burn is a paid mutator transaction binding the contract method 0x8c99319f.
//
// Solidity: function burn(uint256 amount, bytes receiver, address wrappedAsset) returns()
func (_Router *RouterTransactor) Burn(opts *bind.TransactOpts, amount *big.Int, receiver []byte, wrappedAsset common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "burn", amount, receiver, wrappedAsset)
}

// Burn is a paid mutator transaction binding the contract method 0x8c99319f.
//
// Solidity: function burn(uint256 amount, bytes receiver, address wrappedAsset) returns()
func (_Router *RouterSession) Burn(amount *big.Int, receiver []byte, wrappedAsset common.Address) (*types.Transaction, error) {
	return _Router.Contract.Burn(&_Router.TransactOpts, amount, receiver, wrappedAsset)
}

// Burn is a paid mutator transaction binding the contract method 0x8c99319f.
//
// Solidity: function burn(uint256 amount, bytes receiver, address wrappedAsset) returns()
func (_Router *RouterTransactorSession) Burn(amount *big.Int, receiver []byte, wrappedAsset common.Address) (*types.Transaction, error) {
	return _Router.Contract.Burn(&_Router.TransactOpts, amount, receiver, wrappedAsset)
}

// BurnWithPermit is a paid mutator transaction binding the contract method 0x63f4c78a.
//
// Solidity: function burnWithPermit(address wrappedAsset, bytes receiver, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_Router *RouterTransactor) BurnWithPermit(opts *bind.TransactOpts, wrappedAsset common.Address, receiver []byte, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "burnWithPermit", wrappedAsset, receiver, amount, deadline, v, r, s)
}

// BurnWithPermit is a paid mutator transaction binding the contract method 0x63f4c78a.
//
// Solidity: function burnWithPermit(address wrappedAsset, bytes receiver, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_Router *RouterSession) BurnWithPermit(wrappedAsset common.Address, receiver []byte, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Router.Contract.BurnWithPermit(&_Router.TransactOpts, wrappedAsset, receiver, amount, deadline, v, r, s)
}

// BurnWithPermit is a paid mutator transaction binding the contract method 0x63f4c78a.
//
// Solidity: function burnWithPermit(address wrappedAsset, bytes receiver, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_Router *RouterTransactorSession) BurnWithPermit(wrappedAsset common.Address, receiver []byte, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Router.Contract.BurnWithPermit(&_Router.TransactOpts, wrappedAsset, receiver, amount, deadline, v, r, s)
}

// Mint is a paid mutator transaction binding the contract method 0x9949e912.
//
// Solidity: function mint(bytes transactionId, address wrappedAsset, address receiver, uint256 amount, bytes[] signatures) returns()
func (_Router *RouterTransactor) Mint(opts *bind.TransactOpts, transactionId []byte, wrappedAsset common.Address, receiver common.Address, amount *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "mint", transactionId, wrappedAsset, receiver, amount, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x9949e912.
//
// Solidity: function mint(bytes transactionId, address wrappedAsset, address receiver, uint256 amount, bytes[] signatures) returns()
func (_Router *RouterSession) Mint(transactionId []byte, wrappedAsset common.Address, receiver common.Address, amount *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint(&_Router.TransactOpts, transactionId, wrappedAsset, receiver, amount, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x9949e912.
//
// Solidity: function mint(bytes transactionId, address wrappedAsset, address receiver, uint256 amount, bytes[] signatures) returns()
func (_Router *RouterTransactorSession) Mint(transactionId []byte, wrappedAsset common.Address, receiver common.Address, amount *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint(&_Router.TransactOpts, transactionId, wrappedAsset, receiver, amount, signatures)
}

// RemovePair is a paid mutator transaction binding the contract method 0xef91259d.
//
// Solidity: function removePair(bytes native, address wrapped) returns()
func (_Router *RouterTransactor) RemovePair(opts *bind.TransactOpts, native []byte, wrapped common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "removePair", native, wrapped)
}

// RemovePair is a paid mutator transaction binding the contract method 0xef91259d.
//
// Solidity: function removePair(bytes native, address wrapped) returns()
func (_Router *RouterSession) RemovePair(native []byte, wrapped common.Address) (*types.Transaction, error) {
	return _Router.Contract.RemovePair(&_Router.TransactOpts, native, wrapped)
}

// RemovePair is a paid mutator transaction binding the contract method 0xef91259d.
//
// Solidity: function removePair(bytes native, address wrapped) returns()
func (_Router *RouterTransactorSession) RemovePair(native []byte, wrapped common.Address) (*types.Transaction, error) {
	return _Router.Contract.RemovePair(&_Router.TransactOpts, native, wrapped)
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
	Account      common.Address
	WrappedAsset common.Address
	Amount       *big.Int
	Receiver     []byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0x21e3428dd0c1de86ad5cb0d4df91049f26e9711489b51d1e9b5d5a53ebae1f63.
//
// Solidity: event Burn(address indexed account, address indexed wrappedAsset, uint256 amount, bytes receiver)
func (_Router *RouterFilterer) FilterBurn(opts *bind.FilterOpts, account []common.Address, wrappedAsset []common.Address) (*RouterBurnIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var wrappedAssetRule []interface{}
	for _, wrappedAssetItem := range wrappedAsset {
		wrappedAssetRule = append(wrappedAssetRule, wrappedAssetItem)
	}

	logs, sub, err := _Router.contract.FilterLogs(opts, "Burn", accountRule, wrappedAssetRule)
	if err != nil {
		return nil, err
	}
	return &RouterBurnIterator{contract: _Router.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0x21e3428dd0c1de86ad5cb0d4df91049f26e9711489b51d1e9b5d5a53ebae1f63.
//
// Solidity: event Burn(address indexed account, address indexed wrappedAsset, uint256 amount, bytes receiver)
func (_Router *RouterFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *RouterBurn, account []common.Address, wrappedAsset []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var wrappedAssetRule []interface{}
	for _, wrappedAssetItem := range wrappedAsset {
		wrappedAssetRule = append(wrappedAssetRule, wrappedAssetItem)
	}

	logs, sub, err := _Router.contract.WatchLogs(opts, "Burn", accountRule, wrappedAssetRule)
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

// ParseBurn is a log parse operation binding the contract event 0x21e3428dd0c1de86ad5cb0d4df91049f26e9711489b51d1e9b5d5a53ebae1f63.
//
// Solidity: event Burn(address indexed account, address indexed wrappedAsset, uint256 amount, bytes receiver)
func (_Router *RouterFilterer) ParseBurn(log types.Log) (*RouterBurn, error) {
	event := new(RouterBurn)
	if err := _Router.contract.UnpackLog(event, "Burn", log); err != nil {
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
	Account       common.Address
	WrappedAsset  common.Address
	Amount        *big.Int
	TransactionId common.Hash
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0x30c99b8e9449992bd7616d2645b02ee0a2b6f229a45d6b56b686963ff0c53497.
//
// Solidity: event Mint(address indexed account, address indexed wrappedAsset, uint256 amount, bytes indexed transactionId)
func (_Router *RouterFilterer) FilterMint(opts *bind.FilterOpts, account []common.Address, wrappedAsset []common.Address, transactionId [][]byte) (*RouterMintIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var wrappedAssetRule []interface{}
	for _, wrappedAssetItem := range wrappedAsset {
		wrappedAssetRule = append(wrappedAssetRule, wrappedAssetItem)
	}

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _Router.contract.FilterLogs(opts, "Mint", accountRule, wrappedAssetRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &RouterMintIterator{contract: _Router.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0x30c99b8e9449992bd7616d2645b02ee0a2b6f229a45d6b56b686963ff0c53497.
//
// Solidity: event Mint(address indexed account, address indexed wrappedAsset, uint256 amount, bytes indexed transactionId)
func (_Router *RouterFilterer) WatchMint(opts *bind.WatchOpts, sink chan<- *RouterMint, account []common.Address, wrappedAsset []common.Address, transactionId [][]byte) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var wrappedAssetRule []interface{}
	for _, wrappedAssetItem := range wrappedAsset {
		wrappedAssetRule = append(wrappedAssetRule, wrappedAssetItem)
	}

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _Router.contract.WatchLogs(opts, "Mint", accountRule, wrappedAssetRule, transactionIdRule)
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

// ParseMint is a log parse operation binding the contract event 0x30c99b8e9449992bd7616d2645b02ee0a2b6f229a45d6b56b686963ff0c53497.
//
// Solidity: event Mint(address indexed account, address indexed wrappedAsset, uint256 amount, bytes indexed transactionId)
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

// RouterPairAddedIterator is returned from FilterPairAdded and is used to iterate over the raw logs and unpacked data for PairAdded events raised by the Router contract.
type RouterPairAddedIterator struct {
	Event *RouterPairAdded // Event containing the contract specifics and raw log

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
func (it *RouterPairAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterPairAdded)
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
		it.Event = new(RouterPairAdded)
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
func (it *RouterPairAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterPairAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterPairAdded represents a PairAdded event raised by the Router contract.
type RouterPairAdded struct {
	Native  []byte
	Wrapped common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPairAdded is a free log retrieval operation binding the contract event 0xb5b48dbe7e7da4cacb0e117526f73468db287d8677e56aa66ba5cbf0ae0268d2.
//
// Solidity: event PairAdded(bytes native, address wrapped)
func (_Router *RouterFilterer) FilterPairAdded(opts *bind.FilterOpts) (*RouterPairAddedIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "PairAdded")
	if err != nil {
		return nil, err
	}
	return &RouterPairAddedIterator{contract: _Router.contract, event: "PairAdded", logs: logs, sub: sub}, nil
}

// WatchPairAdded is a free log subscription operation binding the contract event 0xb5b48dbe7e7da4cacb0e117526f73468db287d8677e56aa66ba5cbf0ae0268d2.
//
// Solidity: event PairAdded(bytes native, address wrapped)
func (_Router *RouterFilterer) WatchPairAdded(opts *bind.WatchOpts, sink chan<- *RouterPairAdded) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "PairAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterPairAdded)
				if err := _Router.contract.UnpackLog(event, "PairAdded", log); err != nil {
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

// ParsePairAdded is a log parse operation binding the contract event 0xb5b48dbe7e7da4cacb0e117526f73468db287d8677e56aa66ba5cbf0ae0268d2.
//
// Solidity: event PairAdded(bytes native, address wrapped)
func (_Router *RouterFilterer) ParsePairAdded(log types.Log) (*RouterPairAdded, error) {
	event := new(RouterPairAdded)
	if err := _Router.contract.UnpackLog(event, "PairAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RouterPairRemovedIterator is returned from FilterPairRemoved and is used to iterate over the raw logs and unpacked data for PairRemoved events raised by the Router contract.
type RouterPairRemovedIterator struct {
	Event *RouterPairRemoved // Event containing the contract specifics and raw log

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
func (it *RouterPairRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterPairRemoved)
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
		it.Event = new(RouterPairRemoved)
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
func (it *RouterPairRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterPairRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterPairRemoved represents a PairRemoved event raised by the Router contract.
type RouterPairRemoved struct {
	Native  []byte
	Wrapped common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPairRemoved is a free log retrieval operation binding the contract event 0x3812e2f41e41b2813c984b5c8e00dde55d97583c8440466f9c597ff232d0e381.
//
// Solidity: event PairRemoved(bytes native, address wrapped)
func (_Router *RouterFilterer) FilterPairRemoved(opts *bind.FilterOpts) (*RouterPairRemovedIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "PairRemoved")
	if err != nil {
		return nil, err
	}
	return &RouterPairRemovedIterator{contract: _Router.contract, event: "PairRemoved", logs: logs, sub: sub}, nil
}

// WatchPairRemoved is a free log subscription operation binding the contract event 0x3812e2f41e41b2813c984b5c8e00dde55d97583c8440466f9c597ff232d0e381.
//
// Solidity: event PairRemoved(bytes native, address wrapped)
func (_Router *RouterFilterer) WatchPairRemoved(opts *bind.WatchOpts, sink chan<- *RouterPairRemoved) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "PairRemoved")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterPairRemoved)
				if err := _Router.contract.UnpackLog(event, "PairRemoved", log); err != nil {
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

// ParsePairRemoved is a log parse operation binding the contract event 0x3812e2f41e41b2813c984b5c8e00dde55d97583c8440466f9c597ff232d0e381.
//
// Solidity: event PairRemoved(bytes native, address wrapped)
func (_Router *RouterFilterer) ParsePairRemoved(log types.Log) (*RouterPairRemoved, error) {
	event := new(RouterPairRemoved)
	if err := _Router.contract.UnpackLog(event, "PairRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
