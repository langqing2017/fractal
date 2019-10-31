// Copyright 2018 The go-fractal Authors
// This file is part of the go-fractal library.

// Package txexec implements all transaction executors.
package txexec

import (
	"math"
	"math/big"
	"reflect"
	"unsafe"

	"github.com/langqing2017/fractal/common"
	"github.com/langqing2017/fractal/common/hexutil"
	"github.com/langqing2017/fractal/core/nonces"
	"github.com/langqing2017/fractal/core/state"
	"github.com/langqing2017/fractal/core/types"
	"github.com/langqing2017/fractal/crypto"
	"github.com/langqing2017/fractal/params"
	"github.com/langqing2017/fractal/utils/log"
)

/*
The State Transitioning Model

A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all the necessary work to work out a valid new state root.

1) Nonce handling
2) Pre pay gas
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==
  4a) Attempt to run transaction data
  4b) If valid, use result as code for the new state object
== end ==
*/
type StateTransition struct {
	gp          *types.GasPool
	msg         Message
	gas         uint64
	gasPrice    *big.Int
	initialGas  uint64
	value       *big.Int
	data        []byte
	prevStateDb *state.StateDB
	state       *state.StateDB

	nonceSet         *nonces.NonceSet
	maxBitLength     uint64
	callbackParamKey uint64
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	To() *common.Address

	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, contractCreation bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if contractCreation {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
			return 0, ErrOutOfGas
		}
		gas += nz * params.TxDataNonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, ErrOutOfGas
		}
		gas += z * params.TxDataZeroGas
	}
	return gas, nil
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(prevStateDb *state.StateDB, statedb *state.StateDB, msg Message, gp *types.GasPool, nonceSet *nonces.NonceSet, maxBitLength uint64, callbackParamKey uint64) *StateTransition {
	return &StateTransition{
		gp:               gp,
		msg:              msg,
		gasPrice:         msg.GasPrice(),
		value:            msg.Value(),
		data:             msg.Data(),
		prevStateDb:      prevStateDb,
		state:            statedb,
		nonceSet:         nonceSet,
		maxBitLength:     maxBitLength,
		callbackParamKey: callbackParamKey,
	}
}

func (st *StateTransition) useGas(amount uint64) error {
	if st.gas < amount {
		return ErrOutOfGas
	}
	st.gas -= amount

	return nil
}

func (st *StateTransition) buyGas() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Gas()), st.gasPrice)
	if st.state.GetBalance(st.msg.From()).Cmp(mgval) < 0 {
		//log.Warn("gas check", "balance", st.state.GetBalance(st.msg.From()), "mgval", mgval, "gasPrice", st.gasPrice)
		return ErrInsufficientBalanceForGas
	}
	if err := st.gp.SubGas(st.msg.Gas()); err != nil {
		return err
	}
	st.gas += st.msg.Gas()

	st.initialGas = st.msg.Gas()
	st.state.SubBalance(st.msg.From(), mgval)
	return nil
}

func (st *StateTransition) preCheck() error {
	data := reflect.ValueOf(st.prevStateDb)
	if !data.IsNil() {
		newStartNonce := st.prevStateDb.GetNonce(st.msg.From())
		prev := nonces.NewNonceSet(st.nonceSet)
		err, changed := st.nonceSet.Reset(newStartNonce)
		if err != nil {
			log.Error("StateTransition preCheck error", "err", err)
			return err
		}
		if changed {
			st.state.MarkNonceSetJournal(st.msg.From(), prev)
		}
	}
	// Make sure this transaction's nonce is correct.
	if st.msg.CheckNonce() {
		searchResult := st.nonceSet.Search(st.msg.Nonce(), st.maxBitLength)
		if searchResult != nonces.NotContainedAndAllowed {
			log.Debug("preCheck error: NonceError", "searchResult", searchResult, "msgNonce", st.msg.Nonce(), "from", st.msg.From())
			switch searchResult {
			case nonces.NotAllowedTooNew:
				return ErrNonceNotAllowedTooNew
			case nonces.Contained:
				return ErrNonceNotAllowedContained
			case nonces.NotAllowedTooOld:
				return ErrNonceNotAllowedTooOld
			}
			return ErrNonceNotAllowed
		}
	}

	if err := st.buyGas(); err != nil {
		return err
	}

	if !CanTransfer(st.state, st.msg.From(), st.value) {
		return ErrInsufficientBalance
	}

	return nil
}

func (st *StateTransition) refundGas() {
	// Apply refund counter, capped to half of the used gas.
	refund := st.gasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return the remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(st.msg.From(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}

func (st *StateTransition) SimpleTransitionDb(coinbase common.Address) (ret []byte, usedGas uint64, err error) {
	if err = st.preCheck(); err != nil {
		return nil, 0, err
	}
	msg := st.msg
	if msg == nil || msg.To() == nil {
		return nil, 0, ErrSimpleTxHasNoTarget
	}
	// Pay intrinsic gas
	gas, err := IntrinsicGas(nil, false)
	if err != nil {
		return nil, 0, err
	}
	if err = st.useGas(gas); err != nil {
		return nil, 0, err
	}

	// Increment the nonce for the next transaction
	prev := nonces.NewNonceSet(st.nonceSet)
	err, changed := st.nonceSet.Add(msg.Nonce(), st.maxBitLength)
	if err != nil {
		return nil, 0, err
	}
	if changed {
		st.state.MarkNonceSetJournal(st.msg.From(), prev)
	}

	Transfer(st.state, msg.From(), *msg.To(), st.value)

	st.refundGas()
	st.state.AddBalance(coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))

	return ret, st.gasUsed(), err
}

func (st *StateTransition) WasmTransitionDb(coinbase common.Address) (ret []byte, usedGas uint64, wasmFailed bool, err error) {
	if err = st.preCheck(); err != nil {
		return nil, 0, false, err
	}
	msg := st.msg
	deployContract := msg.To() == nil

	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data, deployContract)
	if err != nil {
		return nil, 0, false, err
	}
	if err = st.useGas(gas); err != nil {
		return nil, 0, false, err
	}

	// Increment the nonce for the next transaction
	prev := nonces.NewNonceSet(st.nonceSet)
	err, changed := st.nonceSet.Add(msg.Nonce(), st.maxBitLength)
	if err != nil {
		return nil, 0, false, err
	}
	if changed {
		st.state.MarkNonceSetJournal(st.msg.From(), prev)
	}

	if deployContract {
		err = st.createWasmAccount()
		if err != nil {
			return nil, 0, false, err
		}
	} else {
		err = st.callWasm()
	}

	st.refundGas()
	st.state.AddBalance(coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))

	return ret, st.gasUsed(), err == ErrWasmExec, err
}

func (st *StateTransition) createWasmAccount() error {
	contractAddr := crypto.CreateAddress(st.msg.From(), st.msg.Nonce())
	st.state.CreateAccount(contractAddr)
	Transfer(st.state, st.msg.From(), contractAddr, st.value)

	createDataGas := uint64(len(st.data)) * params.TxGasContractCreateData
	if err := st.useGas(createDataGas); err != nil {
		return ErrCodeStoreOutOfGas
	}
	st.state.SetCode(contractAddr, st.data)
	st.state.SetContractOwner(contractAddr, st.msg.From())
	log.Info("deploy contract", "from", st.msg.From(), "nonce", st.msg.Nonce(), "contract", contractAddr, "code", hexutil.Encode(st.data))
	return nil
}

func (st *StateTransition) callWasm() error {
	Transfer(st.state, st.msg.From(), *st.msg.To(), st.value)

	code := st.state.GetCode(*st.msg.To())
	if len(st.data) > 0 && len(code) > 0 {
		from := st.msg.From()
		to := st.msg.To()
		owner := st.state.GetContractOwner(*st.msg.To())
		value := st.value.Uint64()
		result := CallWasmContract(unsafe.Pointer(&code[0]), len(code), unsafe.Pointer(&st.data[0]), len(st.data), unsafe.Pointer(&from[0]), unsafe.Pointer(&to[0]), unsafe.Pointer(&owner[0]), value, &st.gas, st.callbackParamKey)
		if result != 0 {
			log.Error("CallWasmContract returned with error", "result", result)
			return ErrWasmExec
		}
	}
	return nil
}
