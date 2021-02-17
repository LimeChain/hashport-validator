// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bridge

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

// BridgeABI is the input ABI used to generate the binding from.
const BridgeABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_whbarToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_serviceFee\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"receiverAddress\",\"type\":\"bytes\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"CustodianSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deprecate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"txCost\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newServiceFee\",\"type\":\"uint256\"}],\"name\":\"ServiceFeeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Unpaused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"custodianAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"custodianCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"feesAccrued\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_custodian\",\"type\":\"address\"}],\"name\":\"isCustodian\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"mintTransfers\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"isExecuted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"serviceFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"isOperator\",\"type\":\"bool\"}],\"name\":\"setCustodian\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalFeesAccrued\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"whbarToken\",\"outputs\":[{\"internalType\":\"contractWHBAR\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"txCost\",\"type\":\"uint256\"},{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"receiverAddress\",\"type\":\"bytes\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"transactionId\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"txCost\",\"type\":\"uint256\"}],\"name\":\"computeMessage\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_serviceFee\",\"type\":\"uint256\"}],\"name\":\"setServiceFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"deprecate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Bridge is an auto generated Go binding around an Ethereum contract.
type Bridge struct {
	BridgeCaller     // Read-only binding to the contract
	BridgeTransactor // Write-only binding to the contract
	BridgeFilterer   // Log filterer for contract events
}

// BridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeSession struct {
	Contract     *Bridge           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeCallerSession struct {
	Contract *BridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeTransactorSession struct {
	Contract     *BridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeRaw struct {
	Contract *Bridge // Generic contract binding to access the raw methods on
}

// BridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeCallerRaw struct {
	Contract *BridgeCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeTransactorRaw struct {
	Contract *BridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridge creates a new instance of Bridge, bound to a specific deployed contract.
func NewBridge(address common.Address, backend bind.ContractBackend) (*Bridge, error) {
	contract, err := bindBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

// NewBridgeCaller creates a new read-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeCaller(address common.Address, caller bind.ContractCaller) (*BridgeCaller, error) {
	contract, err := bindBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeCaller{contract: contract}, nil
}

// NewBridgeTransactor creates a new write-only instance of Bridge, bound to a specific deployed contract.
func NewBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeTransactor, error) {
	contract, err := bindBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeTransactor{contract: contract}, nil
}

// NewBridgeFilterer creates a new log filterer instance of Bridge, bound to a specific deployed contract.
func NewBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeFilterer, error) {
	contract, err := bindBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeFilterer{contract: contract}, nil
}

// bindBridge binds a generic wrapper to an already deployed contract.
func bindBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.BridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.BridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bridge *BridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bridge *BridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bridge *BridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bridge.Contract.contract.Transact(opts, method, params...)
}

// ComputeMessage is a free data retrieval call binding the contract method 0x72b78e8a.
//
// Solidity: function computeMessage(bytes transactionId, address receiver, uint256 amount, uint256 txCost) pure returns(bytes32)
func (_Bridge *BridgeCaller) ComputeMessage(opts *bind.CallOpts, transactionId []byte, receiver common.Address, amount *big.Int, txCost *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "computeMessage", transactionId, receiver, amount, txCost)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ComputeMessage is a free data retrieval call binding the contract method 0x72b78e8a.
//
// Solidity: function computeMessage(bytes transactionId, address receiver, uint256 amount, uint256 txCost) pure returns(bytes32)
func (_Bridge *BridgeSession) ComputeMessage(transactionId []byte, receiver common.Address, amount *big.Int, txCost *big.Int) ([32]byte, error) {
	return _Bridge.Contract.ComputeMessage(&_Bridge.CallOpts, transactionId, receiver, amount, txCost)
}

// ComputeMessage is a free data retrieval call binding the contract method 0x72b78e8a.
//
// Solidity: function computeMessage(bytes transactionId, address receiver, uint256 amount, uint256 txCost) pure returns(bytes32)
func (_Bridge *BridgeCallerSession) ComputeMessage(transactionId []byte, receiver common.Address, amount *big.Int, txCost *big.Int) ([32]byte, error) {
	return _Bridge.Contract.ComputeMessage(&_Bridge.CallOpts, transactionId, receiver, amount, txCost)
}

// CustodianAddress is a free data retrieval call binding the contract method 0x6396eff7.
//
// Solidity: function custodianAddress(uint256 index) view returns(address)
func (_Bridge *BridgeCaller) CustodianAddress(opts *bind.CallOpts, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "custodianAddress", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CustodianAddress is a free data retrieval call binding the contract method 0x6396eff7.
//
// Solidity: function custodianAddress(uint256 index) view returns(address)
func (_Bridge *BridgeSession) CustodianAddress(index *big.Int) (common.Address, error) {
	return _Bridge.Contract.CustodianAddress(&_Bridge.CallOpts, index)
}

// CustodianAddress is a free data retrieval call binding the contract method 0x6396eff7.
//
// Solidity: function custodianAddress(uint256 index) view returns(address)
func (_Bridge *BridgeCallerSession) CustodianAddress(index *big.Int) (common.Address, error) {
	return _Bridge.Contract.CustodianAddress(&_Bridge.CallOpts, index)
}

// CustodianCount is a free data retrieval call binding the contract method 0x6f7949f7.
//
// Solidity: function custodianCount() view returns(uint256)
func (_Bridge *BridgeCaller) CustodianCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "custodianCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CustodianCount is a free data retrieval call binding the contract method 0x6f7949f7.
//
// Solidity: function custodianCount() view returns(uint256)
func (_Bridge *BridgeSession) CustodianCount() (*big.Int, error) {
	return _Bridge.Contract.CustodianCount(&_Bridge.CallOpts)
}

// CustodianCount is a free data retrieval call binding the contract method 0x6f7949f7.
//
// Solidity: function custodianCount() view returns(uint256)
func (_Bridge *BridgeCallerSession) CustodianCount() (*big.Int, error) {
	return _Bridge.Contract.CustodianCount(&_Bridge.CallOpts)
}

// FeesAccrued is a free data retrieval call binding the contract method 0x9cca9238.
//
// Solidity: function feesAccrued(address ) view returns(uint256)
func (_Bridge *BridgeCaller) FeesAccrued(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "feesAccrued", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeesAccrued is a free data retrieval call binding the contract method 0x9cca9238.
//
// Solidity: function feesAccrued(address ) view returns(uint256)
func (_Bridge *BridgeSession) FeesAccrued(arg0 common.Address) (*big.Int, error) {
	return _Bridge.Contract.FeesAccrued(&_Bridge.CallOpts, arg0)
}

// FeesAccrued is a free data retrieval call binding the contract method 0x9cca9238.
//
// Solidity: function feesAccrued(address ) view returns(uint256)
func (_Bridge *BridgeCallerSession) FeesAccrued(arg0 common.Address) (*big.Int, error) {
	return _Bridge.Contract.FeesAccrued(&_Bridge.CallOpts, arg0)
}

// IsCustodian is a free data retrieval call binding the contract method 0x35c80c8c.
//
// Solidity: function isCustodian(address _custodian) view returns(bool)
func (_Bridge *BridgeCaller) IsCustodian(opts *bind.CallOpts, _custodian common.Address) (bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "isCustodian", _custodian)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsCustodian is a free data retrieval call binding the contract method 0x35c80c8c.
//
// Solidity: function isCustodian(address _custodian) view returns(bool)
func (_Bridge *BridgeSession) IsCustodian(_custodian common.Address) (bool, error) {
	return _Bridge.Contract.IsCustodian(&_Bridge.CallOpts, _custodian)
}

// IsCustodian is a free data retrieval call binding the contract method 0x35c80c8c.
//
// Solidity: function isCustodian(address _custodian) view returns(bool)
func (_Bridge *BridgeCallerSession) IsCustodian(_custodian common.Address) (bool, error) {
	return _Bridge.Contract.IsCustodian(&_Bridge.CallOpts, _custodian)
}

// MintTransfers is a free data retrieval call binding the contract method 0xed1018ad.
//
// Solidity: function mintTransfers(bytes ) view returns(bool isExecuted)
func (_Bridge *BridgeCaller) MintTransfers(opts *bind.CallOpts, arg0 []byte) (bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "mintTransfers", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MintTransfers is a free data retrieval call binding the contract method 0xed1018ad.
//
// Solidity: function mintTransfers(bytes ) view returns(bool isExecuted)
func (_Bridge *BridgeSession) MintTransfers(arg0 []byte) (bool, error) {
	return _Bridge.Contract.MintTransfers(&_Bridge.CallOpts, arg0)
}

// MintTransfers is a free data retrieval call binding the contract method 0xed1018ad.
//
// Solidity: function mintTransfers(bytes ) view returns(bool isExecuted)
func (_Bridge *BridgeCallerSession) MintTransfers(arg0 []byte) (bool, error) {
	return _Bridge.Contract.MintTransfers(&_Bridge.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeSession) Owner() (common.Address, error) {
	return _Bridge.Contract.Owner(&_Bridge.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bridge *BridgeCallerSession) Owner() (common.Address, error) {
	return _Bridge.Contract.Owner(&_Bridge.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Bridge *BridgeCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Bridge *BridgeSession) Paused() (bool, error) {
	return _Bridge.Contract.Paused(&_Bridge.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Bridge *BridgeCallerSession) Paused() (bool, error) {
	return _Bridge.Contract.Paused(&_Bridge.CallOpts)
}

// ServiceFee is a free data retrieval call binding the contract method 0x8abdf5aa.
//
// Solidity: function serviceFee() view returns(uint256)
func (_Bridge *BridgeCaller) ServiceFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "serviceFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ServiceFee is a free data retrieval call binding the contract method 0x8abdf5aa.
//
// Solidity: function serviceFee() view returns(uint256)
func (_Bridge *BridgeSession) ServiceFee() (*big.Int, error) {
	return _Bridge.Contract.ServiceFee(&_Bridge.CallOpts)
}

// ServiceFee is a free data retrieval call binding the contract method 0x8abdf5aa.
//
// Solidity: function serviceFee() view returns(uint256)
func (_Bridge *BridgeCallerSession) ServiceFee() (*big.Int, error) {
	return _Bridge.Contract.ServiceFee(&_Bridge.CallOpts)
}

// TotalFeesAccrued is a free data retrieval call binding the contract method 0x73f38ee0.
//
// Solidity: function totalFeesAccrued() view returns(uint256)
func (_Bridge *BridgeCaller) TotalFeesAccrued(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "totalFeesAccrued")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalFeesAccrued is a free data retrieval call binding the contract method 0x73f38ee0.
//
// Solidity: function totalFeesAccrued() view returns(uint256)
func (_Bridge *BridgeSession) TotalFeesAccrued() (*big.Int, error) {
	return _Bridge.Contract.TotalFeesAccrued(&_Bridge.CallOpts)
}

// TotalFeesAccrued is a free data retrieval call binding the contract method 0x73f38ee0.
//
// Solidity: function totalFeesAccrued() view returns(uint256)
func (_Bridge *BridgeCallerSession) TotalFeesAccrued() (*big.Int, error) {
	return _Bridge.Contract.TotalFeesAccrued(&_Bridge.CallOpts)
}

// WhbarToken is a free data retrieval call binding the contract method 0xe5d12855.
//
// Solidity: function whbarToken() view returns(address)
func (_Bridge *BridgeCaller) WhbarToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bridge.contract.Call(opts, &out, "whbarToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WhbarToken is a free data retrieval call binding the contract method 0xe5d12855.
//
// Solidity: function whbarToken() view returns(address)
func (_Bridge *BridgeSession) WhbarToken() (common.Address, error) {
	return _Bridge.Contract.WhbarToken(&_Bridge.CallOpts)
}

// WhbarToken is a free data retrieval call binding the contract method 0xe5d12855.
//
// Solidity: function whbarToken() view returns(address)
func (_Bridge *BridgeCallerSession) WhbarToken() (common.Address, error) {
	return _Bridge.Contract.WhbarToken(&_Bridge.CallOpts)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 amount, bytes receiverAddress) returns()
func (_Bridge *BridgeTransactor) Burn(opts *bind.TransactOpts, amount *big.Int, receiverAddress []byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "burn", amount, receiverAddress)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 amount, bytes receiverAddress) returns()
func (_Bridge *BridgeSession) Burn(amount *big.Int, receiverAddress []byte) (*types.Transaction, error) {
	return _Bridge.Contract.Burn(&_Bridge.TransactOpts, amount, receiverAddress)
}

// Burn is a paid mutator transaction binding the contract method 0xfe9d9303.
//
// Solidity: function burn(uint256 amount, bytes receiverAddress) returns()
func (_Bridge *BridgeTransactorSession) Burn(amount *big.Int, receiverAddress []byte) (*types.Transaction, error) {
	return _Bridge.Contract.Burn(&_Bridge.TransactOpts, amount, receiverAddress)
}

// Deprecate is a paid mutator transaction binding the contract method 0x0fcc0c28.
//
// Solidity: function deprecate() returns()
func (_Bridge *BridgeTransactor) Deprecate(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "deprecate")
}

// Deprecate is a paid mutator transaction binding the contract method 0x0fcc0c28.
//
// Solidity: function deprecate() returns()
func (_Bridge *BridgeSession) Deprecate() (*types.Transaction, error) {
	return _Bridge.Contract.Deprecate(&_Bridge.TransactOpts)
}

// Deprecate is a paid mutator transaction binding the contract method 0x0fcc0c28.
//
// Solidity: function deprecate() returns()
func (_Bridge *BridgeTransactorSession) Deprecate() (*types.Transaction, error) {
	return _Bridge.Contract.Deprecate(&_Bridge.TransactOpts)
}

// Mint is a paid mutator transaction binding the contract method 0xa70040fb.
//
// Solidity: function mint(bytes transactionId, address receiver, uint256 amount, uint256 txCost, bytes[] signatures) returns()
func (_Bridge *BridgeTransactor) Mint(opts *bind.TransactOpts, transactionId []byte, receiver common.Address, amount *big.Int, txCost *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "mint", transactionId, receiver, amount, txCost, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0xa70040fb.
//
// Solidity: function mint(bytes transactionId, address receiver, uint256 amount, uint256 txCost, bytes[] signatures) returns()
func (_Bridge *BridgeSession) Mint(transactionId []byte, receiver common.Address, amount *big.Int, txCost *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.Mint(&_Bridge.TransactOpts, transactionId, receiver, amount, txCost, signatures)
}

// Mint is a paid mutator transaction binding the contract method 0xa70040fb.
//
// Solidity: function mint(bytes transactionId, address receiver, uint256 amount, uint256 txCost, bytes[] signatures) returns()
func (_Bridge *BridgeTransactorSession) Mint(transactionId []byte, receiver common.Address, amount *big.Int, txCost *big.Int, signatures [][]byte) (*types.Transaction, error) {
	return _Bridge.Contract.Mint(&_Bridge.TransactOpts, transactionId, receiver, amount, txCost, signatures)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceOwnership(&_Bridge.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bridge *BridgeTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bridge.Contract.RenounceOwnership(&_Bridge.TransactOpts)
}

// SetCustodian is a paid mutator transaction binding the contract method 0x1f49d1f1.
//
// Solidity: function setCustodian(address account, bool isOperator) returns()
func (_Bridge *BridgeTransactor) SetCustodian(opts *bind.TransactOpts, account common.Address, isOperator bool) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "setCustodian", account, isOperator)
}

// SetCustodian is a paid mutator transaction binding the contract method 0x1f49d1f1.
//
// Solidity: function setCustodian(address account, bool isOperator) returns()
func (_Bridge *BridgeSession) SetCustodian(account common.Address, isOperator bool) (*types.Transaction, error) {
	return _Bridge.Contract.SetCustodian(&_Bridge.TransactOpts, account, isOperator)
}

// SetCustodian is a paid mutator transaction binding the contract method 0x1f49d1f1.
//
// Solidity: function setCustodian(address account, bool isOperator) returns()
func (_Bridge *BridgeTransactorSession) SetCustodian(account common.Address, isOperator bool) (*types.Transaction, error) {
	return _Bridge.Contract.SetCustodian(&_Bridge.TransactOpts, account, isOperator)
}

// SetServiceFee is a paid mutator transaction binding the contract method 0x5cdf76f8.
//
// Solidity: function setServiceFee(uint256 _serviceFee) returns()
func (_Bridge *BridgeTransactor) SetServiceFee(opts *bind.TransactOpts, _serviceFee *big.Int) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "setServiceFee", _serviceFee)
}

// SetServiceFee is a paid mutator transaction binding the contract method 0x5cdf76f8.
//
// Solidity: function setServiceFee(uint256 _serviceFee) returns()
func (_Bridge *BridgeSession) SetServiceFee(_serviceFee *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.SetServiceFee(&_Bridge.TransactOpts, _serviceFee)
}

// SetServiceFee is a paid mutator transaction binding the contract method 0x5cdf76f8.
//
// Solidity: function setServiceFee(uint256 _serviceFee) returns()
func (_Bridge *BridgeTransactorSession) SetServiceFee(_serviceFee *big.Int) (*types.Transaction, error) {
	return _Bridge.Contract.SetServiceFee(&_Bridge.TransactOpts, _serviceFee)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferOwnership(&_Bridge.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bridge *BridgeTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bridge.Contract.TransferOwnership(&_Bridge.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_Bridge *BridgeTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bridge.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_Bridge *BridgeSession) Withdraw() (*types.Transaction, error) {
	return _Bridge.Contract.Withdraw(&_Bridge.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_Bridge *BridgeTransactorSession) Withdraw() (*types.Transaction, error) {
	return _Bridge.Contract.Withdraw(&_Bridge.TransactOpts)
}

// BridgeBurnIterator is returned from FilterBurn and is used to iterate over the raw logs and unpacked data for Burn events raised by the Bridge contract.
type BridgeBurnIterator struct {
	Event *BridgeBurn // Event containing the contract specifics and raw log

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
func (it *BridgeBurnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeBurn)
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
		it.Event = new(BridgeBurn)
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
func (it *BridgeBurnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeBurnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeBurn represents a Burn event raised by the Bridge contract.
type BridgeBurn struct {
	Account         common.Address
	Amount          *big.Int
	ReceiverAddress common.Hash
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0x8d38f5a0c1764ff1cca876ce8fe136163fddfce925659e6ad05437cfff6fd392.
//
// Solidity: event Burn(address indexed account, uint256 amount, bytes indexed receiverAddress)
func (_Bridge *BridgeFilterer) FilterBurn(opts *bind.FilterOpts, account []common.Address, receiverAddress [][]byte) (*BridgeBurnIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	var receiverAddressRule []interface{}
	for _, receiverAddressItem := range receiverAddress {
		receiverAddressRule = append(receiverAddressRule, receiverAddressItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Burn", accountRule, receiverAddressRule)
	if err != nil {
		return nil, err
	}
	return &BridgeBurnIterator{contract: _Bridge.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0x8d38f5a0c1764ff1cca876ce8fe136163fddfce925659e6ad05437cfff6fd392.
//
// Solidity: event Burn(address indexed account, uint256 amount, bytes indexed receiverAddress)
func (_Bridge *BridgeFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *BridgeBurn, account []common.Address, receiverAddress [][]byte) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	var receiverAddressRule []interface{}
	for _, receiverAddressItem := range receiverAddress {
		receiverAddressRule = append(receiverAddressRule, receiverAddressItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Burn", accountRule, receiverAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeBurn)
				if err := _Bridge.contract.UnpackLog(event, "Burn", log); err != nil {
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
// Solidity: event Burn(address indexed account, uint256 amount, bytes indexed receiverAddress)
func (_Bridge *BridgeFilterer) ParseBurn(log types.Log) (*BridgeBurn, error) {
	event := new(BridgeBurn)
	if err := _Bridge.contract.UnpackLog(event, "Burn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeCustodianSetIterator is returned from FilterCustodianSet and is used to iterate over the raw logs and unpacked data for CustodianSet events raised by the Bridge contract.
type BridgeCustodianSetIterator struct {
	Event *BridgeCustodianSet // Event containing the contract specifics and raw log

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
func (it *BridgeCustodianSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeCustodianSet)
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
		it.Event = new(BridgeCustodianSet)
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
func (it *BridgeCustodianSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeCustodianSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeCustodianSet represents a CustodianSet event raised by the Bridge contract.
type BridgeCustodianSet struct {
	Operator common.Address
	Status   bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterCustodianSet is a free log retrieval operation binding the contract event 0x8ba69cfc061bc5317ea26989000c52d230a731bb6a8ba331663b306aa0df20d8.
//
// Solidity: event CustodianSet(address operator, bool status)
func (_Bridge *BridgeFilterer) FilterCustodianSet(opts *bind.FilterOpts) (*BridgeCustodianSetIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "CustodianSet")
	if err != nil {
		return nil, err
	}
	return &BridgeCustodianSetIterator{contract: _Bridge.contract, event: "CustodianSet", logs: logs, sub: sub}, nil
}

// WatchCustodianSet is a free log subscription operation binding the contract event 0x8ba69cfc061bc5317ea26989000c52d230a731bb6a8ba331663b306aa0df20d8.
//
// Solidity: event CustodianSet(address operator, bool status)
func (_Bridge *BridgeFilterer) WatchCustodianSet(opts *bind.WatchOpts, sink chan<- *BridgeCustodianSet) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "CustodianSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeCustodianSet)
				if err := _Bridge.contract.UnpackLog(event, "CustodianSet", log); err != nil {
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

// ParseCustodianSet is a log parse operation binding the contract event 0x8ba69cfc061bc5317ea26989000c52d230a731bb6a8ba331663b306aa0df20d8.
//
// Solidity: event CustodianSet(address operator, bool status)
func (_Bridge *BridgeFilterer) ParseCustodianSet(log types.Log) (*BridgeCustodianSet, error) {
	event := new(BridgeCustodianSet)
	if err := _Bridge.contract.UnpackLog(event, "CustodianSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeDeprecateIterator is returned from FilterDeprecate and is used to iterate over the raw logs and unpacked data for Deprecate events raised by the Bridge contract.
type BridgeDeprecateIterator struct {
	Event *BridgeDeprecate // Event containing the contract specifics and raw log

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
func (it *BridgeDeprecateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeDeprecate)
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
		it.Event = new(BridgeDeprecate)
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
func (it *BridgeDeprecateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeDeprecateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeDeprecate represents a Deprecate event raised by the Bridge contract.
type BridgeDeprecate struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDeprecate is a free log retrieval operation binding the contract event 0x33a356c68d152c1cdfdce73f7e4644d382dcbe94098932def08c281a37cdf6bd.
//
// Solidity: event Deprecate(address account, uint256 amount)
func (_Bridge *BridgeFilterer) FilterDeprecate(opts *bind.FilterOpts) (*BridgeDeprecateIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Deprecate")
	if err != nil {
		return nil, err
	}
	return &BridgeDeprecateIterator{contract: _Bridge.contract, event: "Deprecate", logs: logs, sub: sub}, nil
}

// WatchDeprecate is a free log subscription operation binding the contract event 0x33a356c68d152c1cdfdce73f7e4644d382dcbe94098932def08c281a37cdf6bd.
//
// Solidity: event Deprecate(address account, uint256 amount)
func (_Bridge *BridgeFilterer) WatchDeprecate(opts *bind.WatchOpts, sink chan<- *BridgeDeprecate) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Deprecate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeDeprecate)
				if err := _Bridge.contract.UnpackLog(event, "Deprecate", log); err != nil {
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
func (_Bridge *BridgeFilterer) ParseDeprecate(log types.Log) (*BridgeDeprecate, error) {
	event := new(BridgeDeprecate)
	if err := _Bridge.contract.UnpackLog(event, "Deprecate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeMintIterator is returned from FilterMint and is used to iterate over the raw logs and unpacked data for Mint events raised by the Bridge contract.
type BridgeMintIterator struct {
	Event *BridgeMint // Event containing the contract specifics and raw log

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
func (it *BridgeMintIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeMint)
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
		it.Event = new(BridgeMint)
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
func (it *BridgeMintIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeMintIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeMint represents a Mint event raised by the Bridge contract.
type BridgeMint struct {
	Account       common.Address
	Amount        *big.Int
	TxCost        *big.Int
	TransactionId common.Hash
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0x6f22b1e1e34475d2453ef65c35f74a9b1e2183b5038d895d7f439cc9121a8c82.
//
// Solidity: event Mint(address indexed account, uint256 amount, uint256 txCost, bytes indexed transactionId)
func (_Bridge *BridgeFilterer) FilterMint(opts *bind.FilterOpts, account []common.Address, transactionId [][]byte) (*BridgeMintIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Mint", accountRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &BridgeMintIterator{contract: _Bridge.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0x6f22b1e1e34475d2453ef65c35f74a9b1e2183b5038d895d7f439cc9121a8c82.
//
// Solidity: event Mint(address indexed account, uint256 amount, uint256 txCost, bytes indexed transactionId)
func (_Bridge *BridgeFilterer) WatchMint(opts *bind.WatchOpts, sink chan<- *BridgeMint, account []common.Address, transactionId [][]byte) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Mint", accountRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeMint)
				if err := _Bridge.contract.UnpackLog(event, "Mint", log); err != nil {
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

// ParseMint is a log parse operation binding the contract event 0x6f22b1e1e34475d2453ef65c35f74a9b1e2183b5038d895d7f439cc9121a8c82.
//
// Solidity: event Mint(address indexed account, uint256 amount, uint256 txCost, bytes indexed transactionId)
func (_Bridge *BridgeFilterer) ParseMint(log types.Log) (*BridgeMint, error) {
	event := new(BridgeMint)
	if err := _Bridge.contract.UnpackLog(event, "Mint", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Bridge contract.
type BridgeOwnershipTransferredIterator struct {
	Event *BridgeOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BridgeOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeOwnershipTransferred)
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
		it.Event = new(BridgeOwnershipTransferred)
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
func (it *BridgeOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeOwnershipTransferred represents a OwnershipTransferred event raised by the Bridge contract.
type BridgeOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bridge *BridgeFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BridgeOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BridgeOwnershipTransferredIterator{contract: _Bridge.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bridge *BridgeFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BridgeOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeOwnershipTransferred)
				if err := _Bridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Bridge *BridgeFilterer) ParseOwnershipTransferred(log types.Log) (*BridgeOwnershipTransferred, error) {
	event := new(BridgeOwnershipTransferred)
	if err := _Bridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgePausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the Bridge contract.
type BridgePausedIterator struct {
	Event *BridgePaused // Event containing the contract specifics and raw log

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
func (it *BridgePausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgePaused)
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
		it.Event = new(BridgePaused)
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
func (it *BridgePausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgePausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgePaused represents a Paused event raised by the Bridge contract.
type BridgePaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Bridge *BridgeFilterer) FilterPaused(opts *bind.FilterOpts) (*BridgePausedIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &BridgePausedIterator{contract: _Bridge.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Bridge *BridgeFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *BridgePaused) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgePaused)
				if err := _Bridge.contract.UnpackLog(event, "Paused", log); err != nil {
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

// ParsePaused is a log parse operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Bridge *BridgeFilterer) ParsePaused(log types.Log) (*BridgePaused, error) {
	event := new(BridgePaused)
	if err := _Bridge.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeServiceFeeSetIterator is returned from FilterServiceFeeSet and is used to iterate over the raw logs and unpacked data for ServiceFeeSet events raised by the Bridge contract.
type BridgeServiceFeeSetIterator struct {
	Event *BridgeServiceFeeSet // Event containing the contract specifics and raw log

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
func (it *BridgeServiceFeeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeServiceFeeSet)
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
		it.Event = new(BridgeServiceFeeSet)
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
func (it *BridgeServiceFeeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeServiceFeeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeServiceFeeSet represents a ServiceFeeSet event raised by the Bridge contract.
type BridgeServiceFeeSet struct {
	Account       common.Address
	NewServiceFee *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterServiceFeeSet is a free log retrieval operation binding the contract event 0x1a1f706a54e5071e5d16465e3d926d20df26a148f666efe1bdbfbc65d4f41b5b.
//
// Solidity: event ServiceFeeSet(address account, uint256 newServiceFee)
func (_Bridge *BridgeFilterer) FilterServiceFeeSet(opts *bind.FilterOpts) (*BridgeServiceFeeSetIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "ServiceFeeSet")
	if err != nil {
		return nil, err
	}
	return &BridgeServiceFeeSetIterator{contract: _Bridge.contract, event: "ServiceFeeSet", logs: logs, sub: sub}, nil
}

// WatchServiceFeeSet is a free log subscription operation binding the contract event 0x1a1f706a54e5071e5d16465e3d926d20df26a148f666efe1bdbfbc65d4f41b5b.
//
// Solidity: event ServiceFeeSet(address account, uint256 newServiceFee)
func (_Bridge *BridgeFilterer) WatchServiceFeeSet(opts *bind.WatchOpts, sink chan<- *BridgeServiceFeeSet) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "ServiceFeeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeServiceFeeSet)
				if err := _Bridge.contract.UnpackLog(event, "ServiceFeeSet", log); err != nil {
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
func (_Bridge *BridgeFilterer) ParseServiceFeeSet(log types.Log) (*BridgeServiceFeeSet, error) {
	event := new(BridgeServiceFeeSet)
	if err := _Bridge.contract.UnpackLog(event, "ServiceFeeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the Bridge contract.
type BridgeUnpausedIterator struct {
	Event *BridgeUnpaused // Event containing the contract specifics and raw log

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
func (it *BridgeUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeUnpaused)
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
		it.Event = new(BridgeUnpaused)
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
func (it *BridgeUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeUnpaused represents a Unpaused event raised by the Bridge contract.
type BridgeUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Bridge *BridgeFilterer) FilterUnpaused(opts *bind.FilterOpts) (*BridgeUnpausedIterator, error) {

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &BridgeUnpausedIterator{contract: _Bridge.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Bridge *BridgeFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *BridgeUnpaused) (event.Subscription, error) {

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeUnpaused)
				if err := _Bridge.contract.UnpackLog(event, "Unpaused", log); err != nil {
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

// ParseUnpaused is a log parse operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Bridge *BridgeFilterer) ParseUnpaused(log types.Log) (*BridgeUnpaused, error) {
	event := new(BridgeUnpaused)
	if err := _Bridge.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the Bridge contract.
type BridgeWithdrawIterator struct {
	Event *BridgeWithdraw // Event containing the contract specifics and raw log

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
func (it *BridgeWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeWithdraw)
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
		it.Event = new(BridgeWithdraw)
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
func (it *BridgeWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeWithdraw represents a Withdraw event raised by the Bridge contract.
type BridgeWithdraw struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed account, uint256 amount)
func (_Bridge *BridgeFilterer) FilterWithdraw(opts *bind.FilterOpts, account []common.Address) (*BridgeWithdrawIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Bridge.contract.FilterLogs(opts, "Withdraw", accountRule)
	if err != nil {
		return nil, err
	}
	return &BridgeWithdrawIterator{contract: _Bridge.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed account, uint256 amount)
func (_Bridge *BridgeFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *BridgeWithdraw, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Bridge.contract.WatchLogs(opts, "Withdraw", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeWithdraw)
				if err := _Bridge.contract.UnpackLog(event, "Withdraw", log); err != nil {
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

// ParseWithdraw is a log parse operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed account, uint256 amount)
func (_Bridge *BridgeFilterer) ParseWithdraw(log types.Log) (*BridgeWithdraw, error) {
	event := new(BridgeWithdraw)
	if err := _Bridge.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
