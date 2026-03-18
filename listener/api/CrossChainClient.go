// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package api

import (
	"errors"
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
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// CrossChainClientMetaData contains all meta data concerning the CrossChainClient contract.
var CrossChainClientMetaData = &bind.MetaData{
	ABI: "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"verifyResults\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_certHash\",\"type\":\"string\"}],\"name\":\"requestVerification\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_reqId\",\"type\":\"bytes32\"},{\"name\":\"_isValid\",\"type\":\"bool\"}],\"name\":\"fulfillVerification\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"reqId\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"certHash\",\"type\":\"string\"}],\"name\":\"CertVerificationRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"reqId\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"isValid\",\"type\":\"bool\"}],\"name\":\"CertVerified\",\"type\":\"event\"}]",
	Bin: "0x608060405234801561001057600080fd5b506103f2806100206000396000f300608060405260043610610057576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680631dc9e3511461005c5780636cbff42a146100a5578063efd716341461012a575b600080fd5b34801561006857600080fd5b5061008b6004803603810190808035600019169060200190929190505050610167565b604051808215151515815260200191505060405180910390f35b3480156100b157600080fd5b5061010c600480360381019080803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290505050610187565b60405180826000191660001916815260200191505060405180910390f35b34801561013657600080fd5b506101656004803603810190808035600019169060200190929190803515159060200190929190505050610350565b005b60006020528060005260406000206000915054906101000a900460ff1681565b600080334284604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166c0100000000000000000000000002815260140183815260200182805190602001908083835b60208310151561020c57805182526020820191506020810190506020830392506101e7565b6001836020036101000a03801982511681845116808217855250505050505090500193505050506040516020818303038152906040526040518082805190602001908083835b6020831015156102775780518252602082019150602081019050602083039250610252565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020905080600019167fb5cd035a0b5922a2541f503987b2bba816bc9e8bbc4698761e28e95825ed3d31846040518080602001828103825283818151815260200191508051906020019080838360005b8381101561030d5780820151818401526020810190506102f2565b50505050905090810190601f16801561033a5780820380516001836020036101000a031916815260200191505b509250505060405180910390a280915050919050565b80600080846000191660001916815260200190815260200160002060006101000a81548160ff02191690831515021790555081600019167f9bff70b60a91d537efb39164175eb9e698ba49809b8ebc9720dc0c6b129211c882604051808215151515815260200191505060405180910390a250505600a165627a7a7230582001b9f0775965bfb6d9faee6551ec803e08ee04be5856142c535586d9884706ec0029",
}

// CrossChainClientABI is the input ABI used to generate the binding from.
// Deprecated: Use CrossChainClientMetaData.ABI instead.
var CrossChainClientABI = CrossChainClientMetaData.ABI

// CrossChainClientBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use CrossChainClientMetaData.Bin instead.
var CrossChainClientBin = CrossChainClientMetaData.Bin

// DeployCrossChainClient deploys a new Ethereum contract, binding an instance of CrossChainClient to it.
func DeployCrossChainClient(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *CrossChainClient, error) {
	parsed, err := CrossChainClientMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(CrossChainClientBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CrossChainClient{CrossChainClientCaller: CrossChainClientCaller{contract: contract}, CrossChainClientTransactor: CrossChainClientTransactor{contract: contract}, CrossChainClientFilterer: CrossChainClientFilterer{contract: contract}}, nil
}

// CrossChainClient is an auto generated Go binding around an Ethereum contract.
type CrossChainClient struct {
	CrossChainClientCaller     // Read-only binding to the contract
	CrossChainClientTransactor // Write-only binding to the contract
	CrossChainClientFilterer   // Log filterer for contract events
}

// CrossChainClientCaller is an auto generated read-only Go binding around an Ethereum contract.
type CrossChainClientCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CrossChainClientTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CrossChainClientTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CrossChainClientFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CrossChainClientFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CrossChainClientSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CrossChainClientSession struct {
	Contract     *CrossChainClient // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CrossChainClientCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CrossChainClientCallerSession struct {
	Contract *CrossChainClientCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// CrossChainClientTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CrossChainClientTransactorSession struct {
	Contract     *CrossChainClientTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// CrossChainClientRaw is an auto generated low-level Go binding around an Ethereum contract.
type CrossChainClientRaw struct {
	Contract *CrossChainClient // Generic contract binding to access the raw methods on
}

// CrossChainClientCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CrossChainClientCallerRaw struct {
	Contract *CrossChainClientCaller // Generic read-only contract binding to access the raw methods on
}

// CrossChainClientTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CrossChainClientTransactorRaw struct {
	Contract *CrossChainClientTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCrossChainClient creates a new instance of CrossChainClient, bound to a specific deployed contract.
func NewCrossChainClient(address common.Address, backend bind.ContractBackend) (*CrossChainClient, error) {
	contract, err := bindCrossChainClient(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CrossChainClient{CrossChainClientCaller: CrossChainClientCaller{contract: contract}, CrossChainClientTransactor: CrossChainClientTransactor{contract: contract}, CrossChainClientFilterer: CrossChainClientFilterer{contract: contract}}, nil
}

// NewCrossChainClientCaller creates a new read-only instance of CrossChainClient, bound to a specific deployed contract.
func NewCrossChainClientCaller(address common.Address, caller bind.ContractCaller) (*CrossChainClientCaller, error) {
	contract, err := bindCrossChainClient(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CrossChainClientCaller{contract: contract}, nil
}

// NewCrossChainClientTransactor creates a new write-only instance of CrossChainClient, bound to a specific deployed contract.
func NewCrossChainClientTransactor(address common.Address, transactor bind.ContractTransactor) (*CrossChainClientTransactor, error) {
	contract, err := bindCrossChainClient(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CrossChainClientTransactor{contract: contract}, nil
}

// NewCrossChainClientFilterer creates a new log filterer instance of CrossChainClient, bound to a specific deployed contract.
func NewCrossChainClientFilterer(address common.Address, filterer bind.ContractFilterer) (*CrossChainClientFilterer, error) {
	contract, err := bindCrossChainClient(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CrossChainClientFilterer{contract: contract}, nil
}

// bindCrossChainClient binds a generic wrapper to an already deployed contract.
func bindCrossChainClient(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CrossChainClientMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CrossChainClient *CrossChainClientRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CrossChainClient.Contract.CrossChainClientCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CrossChainClient *CrossChainClientRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CrossChainClient.Contract.CrossChainClientTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CrossChainClient *CrossChainClientRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CrossChainClient.Contract.CrossChainClientTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CrossChainClient *CrossChainClientCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CrossChainClient.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CrossChainClient *CrossChainClientTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CrossChainClient.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CrossChainClient *CrossChainClientTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CrossChainClient.Contract.contract.Transact(opts, method, params...)
}

// VerifyResults is a free data retrieval call binding the contract method 0x1dc9e351.
//
// Solidity: function verifyResults(bytes32 ) view returns(bool)
func (_CrossChainClient *CrossChainClientCaller) VerifyResults(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _CrossChainClient.contract.Call(opts, &out, "verifyResults", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyResults is a free data retrieval call binding the contract method 0x1dc9e351.
//
// Solidity: function verifyResults(bytes32 ) view returns(bool)
func (_CrossChainClient *CrossChainClientSession) VerifyResults(arg0 [32]byte) (bool, error) {
	return _CrossChainClient.Contract.VerifyResults(&_CrossChainClient.CallOpts, arg0)
}

// VerifyResults is a free data retrieval call binding the contract method 0x1dc9e351.
//
// Solidity: function verifyResults(bytes32 ) view returns(bool)
func (_CrossChainClient *CrossChainClientCallerSession) VerifyResults(arg0 [32]byte) (bool, error) {
	return _CrossChainClient.Contract.VerifyResults(&_CrossChainClient.CallOpts, arg0)
}

// FulfillVerification is a paid mutator transaction binding the contract method 0xefd71634.
//
// Solidity: function fulfillVerification(bytes32 _reqId, bool _isValid) returns()
func (_CrossChainClient *CrossChainClientTransactor) FulfillVerification(opts *bind.TransactOpts, _reqId [32]byte, _isValid bool) (*types.Transaction, error) {
	return _CrossChainClient.contract.Transact(opts, "fulfillVerification", _reqId, _isValid)
}

// FulfillVerification is a paid mutator transaction binding the contract method 0xefd71634.
//
// Solidity: function fulfillVerification(bytes32 _reqId, bool _isValid) returns()
func (_CrossChainClient *CrossChainClientSession) FulfillVerification(_reqId [32]byte, _isValid bool) (*types.Transaction, error) {
	return _CrossChainClient.Contract.FulfillVerification(&_CrossChainClient.TransactOpts, _reqId, _isValid)
}

// FulfillVerification is a paid mutator transaction binding the contract method 0xefd71634.
//
// Solidity: function fulfillVerification(bytes32 _reqId, bool _isValid) returns()
func (_CrossChainClient *CrossChainClientTransactorSession) FulfillVerification(_reqId [32]byte, _isValid bool) (*types.Transaction, error) {
	return _CrossChainClient.Contract.FulfillVerification(&_CrossChainClient.TransactOpts, _reqId, _isValid)
}

// RequestVerification is a paid mutator transaction binding the contract method 0x6cbff42a.
//
// Solidity: function requestVerification(string _certHash) returns(bytes32)
func (_CrossChainClient *CrossChainClientTransactor) RequestVerification(opts *bind.TransactOpts, _certHash string) (*types.Transaction, error) {
	return _CrossChainClient.contract.Transact(opts, "requestVerification", _certHash)
}

// RequestVerification is a paid mutator transaction binding the contract method 0x6cbff42a.
//
// Solidity: function requestVerification(string _certHash) returns(bytes32)
func (_CrossChainClient *CrossChainClientSession) RequestVerification(_certHash string) (*types.Transaction, error) {
	return _CrossChainClient.Contract.RequestVerification(&_CrossChainClient.TransactOpts, _certHash)
}

// RequestVerification is a paid mutator transaction binding the contract method 0x6cbff42a.
//
// Solidity: function requestVerification(string _certHash) returns(bytes32)
func (_CrossChainClient *CrossChainClientTransactorSession) RequestVerification(_certHash string) (*types.Transaction, error) {
	return _CrossChainClient.Contract.RequestVerification(&_CrossChainClient.TransactOpts, _certHash)
}

// CrossChainClientCertVerificationRequestedIterator is returned from FilterCertVerificationRequested and is used to iterate over the raw logs and unpacked data for CertVerificationRequested events raised by the CrossChainClient contract.
type CrossChainClientCertVerificationRequestedIterator struct {
	Event *CrossChainClientCertVerificationRequested // Event containing the contract specifics and raw log

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
func (it *CrossChainClientCertVerificationRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CrossChainClientCertVerificationRequested)
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
		it.Event = new(CrossChainClientCertVerificationRequested)
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
func (it *CrossChainClientCertVerificationRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CrossChainClientCertVerificationRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CrossChainClientCertVerificationRequested represents a CertVerificationRequested event raised by the CrossChainClient contract.
type CrossChainClientCertVerificationRequested struct {
	ReqId    [32]byte
	CertHash string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterCertVerificationRequested is a free log retrieval operation binding the contract event 0xb5cd035a0b5922a2541f503987b2bba816bc9e8bbc4698761e28e95825ed3d31.
//
// Solidity: event CertVerificationRequested(bytes32 indexed reqId, string certHash)
func (_CrossChainClient *CrossChainClientFilterer) FilterCertVerificationRequested(opts *bind.FilterOpts, reqId [][32]byte) (*CrossChainClientCertVerificationRequestedIterator, error) {

	var reqIdRule []interface{}
	for _, reqIdItem := range reqId {
		reqIdRule = append(reqIdRule, reqIdItem)
	}

	logs, sub, err := _CrossChainClient.contract.FilterLogs(opts, "CertVerificationRequested", reqIdRule)
	if err != nil {
		return nil, err
	}
	return &CrossChainClientCertVerificationRequestedIterator{contract: _CrossChainClient.contract, event: "CertVerificationRequested", logs: logs, sub: sub}, nil
}

// WatchCertVerificationRequested is a free log subscription operation binding the contract event 0xb5cd035a0b5922a2541f503987b2bba816bc9e8bbc4698761e28e95825ed3d31.
//
// Solidity: event CertVerificationRequested(bytes32 indexed reqId, string certHash)
func (_CrossChainClient *CrossChainClientFilterer) WatchCertVerificationRequested(opts *bind.WatchOpts, sink chan<- *CrossChainClientCertVerificationRequested, reqId [][32]byte) (event.Subscription, error) {

	var reqIdRule []interface{}
	for _, reqIdItem := range reqId {
		reqIdRule = append(reqIdRule, reqIdItem)
	}

	logs, sub, err := _CrossChainClient.contract.WatchLogs(opts, "CertVerificationRequested", reqIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CrossChainClientCertVerificationRequested)
				if err := _CrossChainClient.contract.UnpackLog(event, "CertVerificationRequested", log); err != nil {
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

// ParseCertVerificationRequested is a log parse operation binding the contract event 0xb5cd035a0b5922a2541f503987b2bba816bc9e8bbc4698761e28e95825ed3d31.
//
// Solidity: event CertVerificationRequested(bytes32 indexed reqId, string certHash)
func (_CrossChainClient *CrossChainClientFilterer) ParseCertVerificationRequested(log types.Log) (*CrossChainClientCertVerificationRequested, error) {
	event := new(CrossChainClientCertVerificationRequested)
	if err := _CrossChainClient.contract.UnpackLog(event, "CertVerificationRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CrossChainClientCertVerifiedIterator is returned from FilterCertVerified and is used to iterate over the raw logs and unpacked data for CertVerified events raised by the CrossChainClient contract.
type CrossChainClientCertVerifiedIterator struct {
	Event *CrossChainClientCertVerified // Event containing the contract specifics and raw log

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
func (it *CrossChainClientCertVerifiedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CrossChainClientCertVerified)
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
		it.Event = new(CrossChainClientCertVerified)
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
func (it *CrossChainClientCertVerifiedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CrossChainClientCertVerifiedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CrossChainClientCertVerified represents a CertVerified event raised by the CrossChainClient contract.
type CrossChainClientCertVerified struct {
	ReqId   [32]byte
	IsValid bool
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterCertVerified is a free log retrieval operation binding the contract event 0x9bff70b60a91d537efb39164175eb9e698ba49809b8ebc9720dc0c6b129211c8.
//
// Solidity: event CertVerified(bytes32 indexed reqId, bool isValid)
func (_CrossChainClient *CrossChainClientFilterer) FilterCertVerified(opts *bind.FilterOpts, reqId [][32]byte) (*CrossChainClientCertVerifiedIterator, error) {

	var reqIdRule []interface{}
	for _, reqIdItem := range reqId {
		reqIdRule = append(reqIdRule, reqIdItem)
	}

	logs, sub, err := _CrossChainClient.contract.FilterLogs(opts, "CertVerified", reqIdRule)
	if err != nil {
		return nil, err
	}
	return &CrossChainClientCertVerifiedIterator{contract: _CrossChainClient.contract, event: "CertVerified", logs: logs, sub: sub}, nil
}

// WatchCertVerified is a free log subscription operation binding the contract event 0x9bff70b60a91d537efb39164175eb9e698ba49809b8ebc9720dc0c6b129211c8.
//
// Solidity: event CertVerified(bytes32 indexed reqId, bool isValid)
func (_CrossChainClient *CrossChainClientFilterer) WatchCertVerified(opts *bind.WatchOpts, sink chan<- *CrossChainClientCertVerified, reqId [][32]byte) (event.Subscription, error) {

	var reqIdRule []interface{}
	for _, reqIdItem := range reqId {
		reqIdRule = append(reqIdRule, reqIdItem)
	}

	logs, sub, err := _CrossChainClient.contract.WatchLogs(opts, "CertVerified", reqIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CrossChainClientCertVerified)
				if err := _CrossChainClient.contract.UnpackLog(event, "CertVerified", log); err != nil {
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

// ParseCertVerified is a log parse operation binding the contract event 0x9bff70b60a91d537efb39164175eb9e698ba49809b8ebc9720dc0c6b129211c8.
//
// Solidity: event CertVerified(bytes32 indexed reqId, bool isValid)
func (_CrossChainClient *CrossChainClientFilterer) ParseCertVerified(log types.Log) (*CrossChainClientCertVerified, error) {
	event := new(CrossChainClientCertVerified)
	if err := _CrossChainClient.contract.UnpackLog(event, "CertVerified", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
