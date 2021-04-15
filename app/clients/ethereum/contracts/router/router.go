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
const RouterABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_serviceFee\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_controllerAddress\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"serviceFee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"receiver\",\"type\":\"bytes\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Claim\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"member\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"MemberUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"serviceFeeInWTokens\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"txCost\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newServiceFee\",\"type\":\"uint256\"}],\"name\":\"ServiceFeeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"nativeToken\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"}],\"name\":\"TokenUpdate\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"controllerAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"member\",\"type\":\"address\"}],\"name\":\"getTxCostsPerMember\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_member\",\"type\":\"address\"}],\"name\":\"isMember\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"memberAt\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"membersCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"mintTransfers\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"isExecuted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"nativeToWrappedToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"serviceFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_serviceFee\",\"type\":\"uint256\"}],\"name\":\"setServiceFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"wrappedToNativeToken\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"wrappedTokensData\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"feesAccrued\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"previousAccrued\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"accumulator\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"txCost\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"mintWithReimbursement\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"receiver\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"isMember\",\"type\":\"bool\"}],\"name\":\"updateMember\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newWrappedToken\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"tokenID\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"}],\"name\":\"updateWrappedToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"wrappedToken\",\"type\":\"address\"}],\"name\":\"isSupportedToken\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"wrappedTokensCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"wrappedTokenAt\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

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

// ControllerAddress is a free data retrieval call binding the contract method 0x4b24ea47.
//
// Solidity: function controllerAddress() view returns(address)
func (_Router *RouterCaller) ControllerAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "controllerAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ControllerAddress is a free data retrieval call binding the contract method 0x4b24ea47.
//
// Solidity: function controllerAddress() view returns(address)
func (_Router *RouterSession) ControllerAddress() (common.Address, error) {
	return _Router.Contract.ControllerAddress(&_Router.CallOpts)
}

// ControllerAddress is a free data retrieval call binding the contract method 0x4b24ea47.
//
// Solidity: function controllerAddress() view returns(address)
func (_Router *RouterCallerSession) ControllerAddress() (common.Address, error) {
	return _Router.Contract.ControllerAddress(&_Router.CallOpts)
}

// GetTxCostsPerMember is a free data retrieval call binding the contract method 0x20990243.
//
// Solidity: function getTxCostsPerMember(address wrappedToken, address member) view returns(uint256)
func (_Router *RouterCaller) GetTxCostsPerMember(opts *bind.CallOpts, wrappedToken common.Address, member common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "getTxCostsPerMember", wrappedToken, member)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTxCostsPerMember is a free data retrieval call binding the contract method 0x20990243.
//
// Solidity: function getTxCostsPerMember(address wrappedToken, address member) view returns(uint256)
func (_Router *RouterSession) GetTxCostsPerMember(wrappedToken common.Address, member common.Address) (*big.Int, error) {
	return _Router.Contract.GetTxCostsPerMember(&_Router.CallOpts, wrappedToken, member)
}

// GetTxCostsPerMember is a free data retrieval call binding the contract method 0x20990243.
//
// Solidity: function getTxCostsPerMember(address wrappedToken, address member) view returns(uint256)
func (_Router *RouterCallerSession) GetTxCostsPerMember(wrappedToken common.Address, member common.Address) (*big.Int, error) {
	return _Router.Contract.GetTxCostsPerMember(&_Router.CallOpts, wrappedToken, member)
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

// IsSupportedToken is a free data retrieval call binding the contract method 0x240028e8.
//
// Solidity: function isSupportedToken(address wrappedToken) view returns(bool)
func (_Router *RouterCaller) IsSupportedToken(opts *bind.CallOpts, wrappedToken common.Address) (bool, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "isSupportedToken", wrappedToken)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsSupportedToken is a free data retrieval call binding the contract method 0x240028e8.
//
// Solidity: function isSupportedToken(address wrappedToken) view returns(bool)
func (_Router *RouterSession) IsSupportedToken(wrappedToken common.Address) (bool, error) {
	return _Router.Contract.IsSupportedToken(&_Router.CallOpts, wrappedToken)
}

// IsSupportedToken is a free data retrieval call binding the contract method 0x240028e8.
//
// Solidity: function isSupportedToken(address wrappedToken) view returns(bool)
func (_Router *RouterCallerSession) IsSupportedToken(wrappedToken common.Address) (bool, error) {
	return _Router.Contract.IsSupportedToken(&_Router.CallOpts, wrappedToken)
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

// NativeToWrappedToken is a free data retrieval call binding the contract method 0x2b793cc8.
//
// Solidity: function nativeToWrappedToken(bytes ) view returns(address)
func (_Router *RouterCaller) NativeToWrappedToken(opts *bind.CallOpts, arg0 []byte) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "nativeToWrappedToken", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NativeToWrappedToken is a free data retrieval call binding the contract method 0x2b793cc8.
//
// Solidity: function nativeToWrappedToken(bytes ) view returns(address)
func (_Router *RouterSession) NativeToWrappedToken(arg0 []byte) (common.Address, error) {
	return _Router.Contract.NativeToWrappedToken(&_Router.CallOpts, arg0)
}

// NativeToWrappedToken is a free data retrieval call binding the contract method 0x2b793cc8.
//
// Solidity: function nativeToWrappedToken(bytes ) view returns(address)
func (_Router *RouterCallerSession) NativeToWrappedToken(arg0 []byte) (common.Address, error) {
	return _Router.Contract.NativeToWrappedToken(&_Router.CallOpts, arg0)
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

// WrappedToNativeToken is a free data retrieval call binding the contract method 0x0aa3824d.
//
// Solidity: function wrappedToNativeToken(address ) view returns(bytes)
func (_Router *RouterCaller) WrappedToNativeToken(opts *bind.CallOpts, arg0 common.Address) ([]byte, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "wrappedToNativeToken", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// WrappedToNativeToken is a free data retrieval call binding the contract method 0x0aa3824d.
//
// Solidity: function wrappedToNativeToken(address ) view returns(bytes)
func (_Router *RouterSession) WrappedToNativeToken(arg0 common.Address) ([]byte, error) {
	return _Router.Contract.WrappedToNativeToken(&_Router.CallOpts, arg0)
}

// WrappedToNativeToken is a free data retrieval call binding the contract method 0x0aa3824d.
//
// Solidity: function wrappedToNativeToken(address ) view returns(bytes)
func (_Router *RouterCallerSession) WrappedToNativeToken(arg0 common.Address) ([]byte, error) {
	return _Router.Contract.WrappedToNativeToken(&_Router.CallOpts, arg0)
}

// WrappedTokenAt is a free data retrieval call binding the contract method 0xae963d45.
//
// Solidity: function wrappedTokenAt(uint256 index) view returns(address)
func (_Router *RouterCaller) WrappedTokenAt(opts *bind.CallOpts, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "wrappedTokenAt", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WrappedTokenAt is a free data retrieval call binding the contract method 0xae963d45.
//
// Solidity: function wrappedTokenAt(uint256 index) view returns(address)
func (_Router *RouterSession) WrappedTokenAt(index *big.Int) (common.Address, error) {
	return _Router.Contract.WrappedTokenAt(&_Router.CallOpts, index)
}

// WrappedTokenAt is a free data retrieval call binding the contract method 0xae963d45.
//
// Solidity: function wrappedTokenAt(uint256 index) view returns(address)
func (_Router *RouterCallerSession) WrappedTokenAt(index *big.Int) (common.Address, error) {
	return _Router.Contract.WrappedTokenAt(&_Router.CallOpts, index)
}

// WrappedTokensCount is a free data retrieval call binding the contract method 0xb753cf46.
//
// Solidity: function wrappedTokensCount() view returns(uint256)
func (_Router *RouterCaller) WrappedTokensCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "wrappedTokensCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WrappedTokensCount is a free data retrieval call binding the contract method 0xb753cf46.
//
// Solidity: function wrappedTokensCount() view returns(uint256)
func (_Router *RouterSession) WrappedTokensCount() (*big.Int, error) {
	return _Router.Contract.WrappedTokensCount(&_Router.CallOpts)
}

// WrappedTokensCount is a free data retrieval call binding the contract method 0xb753cf46.
//
// Solidity: function wrappedTokensCount() view returns(uint256)
func (_Router *RouterCallerSession) WrappedTokensCount() (*big.Int, error) {
	return _Router.Contract.WrappedTokensCount(&_Router.CallOpts)
}

// WrappedTokensData is a free data retrieval call binding the contract method 0xf804b7a2.
//
// Solidity: function wrappedTokensData(address ) view returns(uint256 feesAccrued, uint256 previousAccrued, uint256 accumulator)
func (_Router *RouterCaller) WrappedTokensData(opts *bind.CallOpts, arg0 common.Address) (struct {
	FeesAccrued     *big.Int
	PreviousAccrued *big.Int
	Accumulator     *big.Int
}, error) {
	var out []interface{}
	err := _Router.contract.Call(opts, &out, "wrappedTokensData", arg0)

	outstruct := new(struct {
		FeesAccrued     *big.Int
		PreviousAccrued *big.Int
		Accumulator     *big.Int
	})

	outstruct.FeesAccrued = out[0].(*big.Int)
	outstruct.PreviousAccrued = out[1].(*big.Int)
	outstruct.Accumulator = out[2].(*big.Int)

	return *outstruct, err

}

// WrappedTokensData is a free data retrieval call binding the contract method 0xf804b7a2.
//
// Solidity: function wrappedTokensData(address ) view returns(uint256 feesAccrued, uint256 previousAccrued, uint256 accumulator)
func (_Router *RouterSession) WrappedTokensData(arg0 common.Address) (struct {
	FeesAccrued     *big.Int
	PreviousAccrued *big.Int
	Accumulator     *big.Int
}, error) {
	return _Router.Contract.WrappedTokensData(&_Router.CallOpts, arg0)
}

// WrappedTokensData is a free data retrieval call binding the contract method 0xf804b7a2.
//
// Solidity: function wrappedTokensData(address ) view returns(uint256 feesAccrued, uint256 previousAccrued, uint256 accumulator)
func (_Router *RouterCallerSession) WrappedTokensData(arg0 common.Address) (struct {
	FeesAccrued     *big.Int
	PreviousAccrued *big.Int
	Accumulator     *big.Int
}, error) {
	return _Router.Contract.WrappedTokensData(&_Router.CallOpts, arg0)
}

// Burn is a paid mutator transaction binding the contract method 0x8c99319f.
//
// Solidity: function burn(uint256 amount, bytes receiver, address wrappedToken) returns()
func (_Router *RouterTransactor) Burn(opts *bind.TransactOpts, amount *big.Int, receiver []byte, wrappedToken common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "burn", amount, receiver, wrappedToken)
}

// Burn is a paid mutator transaction binding the contract method 0x8c99319f.
//
// Solidity: function burn(uint256 amount, bytes receiver, address wrappedToken) returns()
func (_Router *RouterSession) Burn(amount *big.Int, receiver []byte, wrappedToken common.Address) (*types.Transaction, error) {
	return _Router.Contract.Burn(&_Router.TransactOpts, amount, receiver, wrappedToken)
}

// Burn is a paid mutator transaction binding the contract method 0x8c99319f.
//
// Solidity: function burn(uint256 amount, bytes receiver, address wrappedToken) returns()
func (_Router *RouterTransactorSession) Burn(amount *big.Int, receiver []byte, wrappedToken common.Address) (*types.Transaction, error) {
	return _Router.Contract.Burn(&_Router.TransactOpts, amount, receiver, wrappedToken)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address wrappedToken) returns()
func (_Router *RouterTransactor) Claim(opts *bind.TransactOpts, wrappedToken common.Address) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "claim", wrappedToken)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address wrappedToken) returns()
func (_Router *RouterSession) Claim(wrappedToken common.Address) (*types.Transaction, error) {
	return _Router.Contract.Claim(&_Router.TransactOpts, wrappedToken)
}

// Claim is a paid mutator transaction binding the contract method 0x1e83409a.
//
// Solidity: function claim(address wrappedToken) returns()
func (_Router *RouterTransactorSession) Claim(wrappedToken common.Address) (*types.Transaction, error) {
	return _Router.Contract.Claim(&_Router.TransactOpts, wrappedToken)
}

// Mint is a paid mutator transaction binding the contract method 0x9949e912.
//
// Solidity: function mint(bytes transactionId, address wrappedToken, address receiver, uint256 amount, bytes[] signatures) returns()
func (_Router *RouterTransactor) Mint(opts *bind.TransactOpts, transactionId []byte, wrappedToken common.Address, receiver common.Address, amount *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "mint", transactionId, wrappedToken, receiver, amount, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x9949e912.
//
// Solidity: function mint(bytes transactionId, address wrappedToken, address receiver, uint256 amount, bytes[] signatures) returns()
func (_Router *RouterSession) Mint(transactionId []byte, wrappedToken common.Address, receiver common.Address, amount *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint(&_Router.TransactOpts, transactionId, wrappedToken, receiver, amount, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0x9949e912.
//
// Solidity: function mint(bytes transactionId, address wrappedToken, address receiver, uint256 amount, bytes[] signatures) returns()
func (_Router *RouterTransactorSession) Mint(transactionId []byte, wrappedToken common.Address, receiver common.Address, amount *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.Mint(&_Router.TransactOpts, transactionId, wrappedToken, receiver, amount, signatures)
}

// MintWithReimbursement is a paid mutator transaction binding the contract method 0x9787221b.
//
// Solidity: function mintWithReimbursement(bytes transactionId, address wrappedToken, address receiver, uint256 amount, uint256 txCost, bytes[] signatures) returns()
func (_Router *RouterTransactor) MintWithReimbursement(opts *bind.TransactOpts, transactionId []byte, wrappedToken common.Address, receiver common.Address, amount *big.Int, txCost *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "mintWithReimbursement", transactionId, wrappedToken, receiver, amount, txCost, signatures)
}

// MintWithReimbursement is a paid mutator transaction binding the contract method 0x9787221b.
//
// Solidity: function mintWithReimbursement(bytes transactionId, address wrappedToken, address receiver, uint256 amount, uint256 txCost, bytes[] signatures) returns()
func (_Router *RouterSession) MintWithReimbursement(transactionId []byte, wrappedToken common.Address, receiver common.Address, amount *big.Int, txCost *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.MintWithReimbursement(&_Router.TransactOpts, transactionId, wrappedToken, receiver, amount, txCost, signatures)
}

// MintWithReimbursement is a paid mutator transaction binding the contract method 0x9787221b.
//
// Solidity: function mintWithReimbursement(bytes transactionId, address wrappedToken, address receiver, uint256 amount, uint256 txCost, bytes[] signatures) returns()
func (_Router *RouterTransactorSession) MintWithReimbursement(transactionId []byte, wrappedToken common.Address, receiver common.Address, amount *big.Int, txCost *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Router.Contract.MintWithReimbursement(&_Router.TransactOpts, transactionId, wrappedToken, receiver, amount, txCost, signatures)
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

// UpdateWrappedToken is a paid mutator transaction binding the contract method 0xb9080c46.
//
// Solidity: function updateWrappedToken(address newWrappedToken, bytes tokenID, bool isActive) returns()
func (_Router *RouterTransactor) UpdateWrappedToken(opts *bind.TransactOpts, newWrappedToken common.Address, tokenID []byte, isActive bool) (*types.Transaction, error) {
	return _Router.contract.Transact(opts, "updateWrappedToken", newWrappedToken, tokenID, isActive)
}

// UpdateWrappedToken is a paid mutator transaction binding the contract method 0xb9080c46.
//
// Solidity: function updateWrappedToken(address newWrappedToken, bytes tokenID, bool isActive) returns()
func (_Router *RouterSession) UpdateWrappedToken(newWrappedToken common.Address, tokenID []byte, isActive bool) (*types.Transaction, error) {
	return _Router.Contract.UpdateWrappedToken(&_Router.TransactOpts, newWrappedToken, tokenID, isActive)
}

// UpdateWrappedToken is a paid mutator transaction binding the contract method 0xb9080c46.
//
// Solidity: function updateWrappedToken(address newWrappedToken, bytes tokenID, bool isActive) returns()
func (_Router *RouterTransactorSession) UpdateWrappedToken(newWrappedToken common.Address, tokenID []byte, isActive bool) (*types.Transaction, error) {
	return _Router.Contract.UpdateWrappedToken(&_Router.TransactOpts, newWrappedToken, tokenID, isActive)
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
	WrappedToken common.Address
	Amount       *big.Int
	ServiceFee   *big.Int
	Receiver     []byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0xdc10f4e7dc07ecbfb7161b29d366abee4dd6e2413104fcb7c324aeccedb2f15f.
//
// Solidity: event Burn(address indexed account, address indexed wrappedToken, uint256 amount, uint256 serviceFee, bytes receiver)
func (_Router *RouterFilterer) FilterBurn(opts *bind.FilterOpts, account []common.Address, wrappedToken []common.Address) (*RouterBurnIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var wrappedTokenRule []interface{}
	for _, wrappedTokenItem := range wrappedToken {
		wrappedTokenRule = append(wrappedTokenRule, wrappedTokenItem)
	}

	logs, sub, err := _Router.contract.FilterLogs(opts, "Burn", accountRule, wrappedTokenRule)
	if err != nil {
		return nil, err
	}
	return &RouterBurnIterator{contract: _Router.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0xdc10f4e7dc07ecbfb7161b29d366abee4dd6e2413104fcb7c324aeccedb2f15f.
//
// Solidity: event Burn(address indexed account, address indexed wrappedToken, uint256 amount, uint256 serviceFee, bytes receiver)
func (_Router *RouterFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *RouterBurn, account []common.Address, wrappedToken []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var wrappedTokenRule []interface{}
	for _, wrappedTokenItem := range wrappedToken {
		wrappedTokenRule = append(wrappedTokenRule, wrappedTokenItem)
	}

	logs, sub, err := _Router.contract.WatchLogs(opts, "Burn", accountRule, wrappedTokenRule)
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

// ParseBurn is a log parse operation binding the contract event 0xdc10f4e7dc07ecbfb7161b29d366abee4dd6e2413104fcb7c324aeccedb2f15f.
//
// Solidity: event Burn(address indexed account, address indexed wrappedToken, uint256 amount, uint256 serviceFee, bytes receiver)
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
	WrappedToken        common.Address
	Amount              *big.Int
	ServiceFeeInWTokens *big.Int
	TxCost              *big.Int
	TransactionId       common.Hash
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0x43921b73dbf816ea45b5fb03d9fce2201c7d8f5a8ed24da1726402df8c718551.
//
// Solidity: event Mint(address indexed account, address indexed wrappedToken, uint256 amount, uint256 serviceFeeInWTokens, uint256 txCost, bytes indexed transactionId)
func (_Router *RouterFilterer) FilterMint(opts *bind.FilterOpts, account []common.Address, wrappedToken []common.Address, transactionId [][]byte) (*RouterMintIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var wrappedTokenRule []interface{}
	for _, wrappedTokenItem := range wrappedToken {
		wrappedTokenRule = append(wrappedTokenRule, wrappedTokenItem)
	}

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _Router.contract.FilterLogs(opts, "Mint", accountRule, wrappedTokenRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &RouterMintIterator{contract: _Router.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0x43921b73dbf816ea45b5fb03d9fce2201c7d8f5a8ed24da1726402df8c718551.
//
// Solidity: event Mint(address indexed account, address indexed wrappedToken, uint256 amount, uint256 serviceFeeInWTokens, uint256 txCost, bytes indexed transactionId)
func (_Router *RouterFilterer) WatchMint(opts *bind.WatchOpts, sink chan<- *RouterMint, account []common.Address, wrappedToken []common.Address, transactionId [][]byte) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var wrappedTokenRule []interface{}
	for _, wrappedTokenItem := range wrappedToken {
		wrappedTokenRule = append(wrappedTokenRule, wrappedTokenItem)
	}

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _Router.contract.WatchLogs(opts, "Mint", accountRule, wrappedTokenRule, transactionIdRule)
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

// ParseMint is a log parse operation binding the contract event 0x43921b73dbf816ea45b5fb03d9fce2201c7d8f5a8ed24da1726402df8c718551.
//
// Solidity: event Mint(address indexed account, address indexed wrappedToken, uint256 amount, uint256 serviceFeeInWTokens, uint256 txCost, bytes indexed transactionId)
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

// RouterTokenUpdateIterator is returned from FilterTokenUpdate and is used to iterate over the raw logs and unpacked data for TokenUpdate events raised by the Router contract.
type RouterTokenUpdateIterator struct {
	Event *RouterTokenUpdate // Event containing the contract specifics and raw log

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
func (it *RouterTokenUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RouterTokenUpdate)
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
		it.Event = new(RouterTokenUpdate)
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
func (it *RouterTokenUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RouterTokenUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RouterTokenUpdate represents a TokenUpdate event raised by the Router contract.
type RouterTokenUpdate struct {
	WrappedToken common.Address
	NativeToken  []byte
	IsActive     bool
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterTokenUpdate is a free log retrieval operation binding the contract event 0x1fdba7efccb3b2b0ed68be70ff537de605c6a8d279fc25899e5426cb4ddd3d8b.
//
// Solidity: event TokenUpdate(address wrappedToken, bytes nativeToken, bool isActive)
func (_Router *RouterFilterer) FilterTokenUpdate(opts *bind.FilterOpts) (*RouterTokenUpdateIterator, error) {

	logs, sub, err := _Router.contract.FilterLogs(opts, "TokenUpdate")
	if err != nil {
		return nil, err
	}
	return &RouterTokenUpdateIterator{contract: _Router.contract, event: "TokenUpdate", logs: logs, sub: sub}, nil
}

// WatchTokenUpdate is a free log subscription operation binding the contract event 0x1fdba7efccb3b2b0ed68be70ff537de605c6a8d279fc25899e5426cb4ddd3d8b.
//
// Solidity: event TokenUpdate(address wrappedToken, bytes nativeToken, bool isActive)
func (_Router *RouterFilterer) WatchTokenUpdate(opts *bind.WatchOpts, sink chan<- *RouterTokenUpdate) (event.Subscription, error) {

	logs, sub, err := _Router.contract.WatchLogs(opts, "TokenUpdate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RouterTokenUpdate)
				if err := _Router.contract.UnpackLog(event, "TokenUpdate", log); err != nil {
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

// ParseTokenUpdate is a log parse operation binding the contract event 0x1fdba7efccb3b2b0ed68be70ff537de605c6a8d279fc25899e5426cb4ddd3d8b.
//
// Solidity: event TokenUpdate(address wrappedToken, bytes nativeToken, bool isActive)
func (_Router *RouterFilterer) ParseTokenUpdate(log types.Log) (*RouterTokenUpdate, error) {
	event := new(RouterTokenUpdate)
	if err := _Router.contract.UnpackLog(event, "TokenUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
