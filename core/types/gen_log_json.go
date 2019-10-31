// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"
	"github.com/langqing2017/fractal/common"
	"github.com/langqing2017/fractal/common/hexutil"
)

var _ = (*logMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (l Log) MarshalJSON() ([]byte, error) {
	type Log struct {
		Address     common.Address `json:"address" gencodec:"required"`
		Topics      []common.Hash  `json:"topics" gencodec:"required"`
		Data        hexutil.Bytes  `json:"data" gencodec:"required"`
		BlockNumber hexutil.Uint64 `json:"block=Number" gencodec:"required"`
		TxHash      common.Hash    `json:"transactionHash" gencodec:"required"`
		PkgIndex    uint32         `json:"packageIndex" gencodec:"required"`
		TxIndex     uint32         `json:"transactionIndex" gencodec:"required"`
		Index       uint32         `json:"logIndex" gencodec:"required"`
	}
	var enc Log
	enc.Address = l.Address
	enc.Topics = l.Topics
	enc.Data = l.Data
	enc.BlockNumber = hexutil.Uint64(l.BlockNumber)
	enc.TxHash = l.TxHash
	enc.PkgIndex = l.PkgIndex
	enc.TxIndex = l.TxIndex
	enc.Index = l.Index
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (l *Log) UnmarshalJSON(input []byte) error {
	type Log struct {
		Address     *common.Address `json:"address" gencodec:"required"`
		Topics      []common.Hash   `json:"topics" gencodec:"required"`
		Data        *hexutil.Bytes  `json:"data" gencodec:"required"`
		BlockNumber *hexutil.Uint64 `json:"blockNumber" gencodec:"required"`
		TxHash      *common.Hash    `json:"transactionHash" gencodec:"required"`
		PkgIndex    *uint32         `json:"packageIndex" gencodec:"required"`
		TxIndex     *uint32         `json:"transactionIndex" gencodec:"required"`
		Index       *uint32         `json:"logIndex" gencodec:"required"`
	}
	var dec Log
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Address == nil {
		return errors.New("missing required field 'address' for Log")
	}
	l.Address = *dec.Address
	if dec.Topics == nil {
		return errors.New("missing required field 'topics' for Log")
	}
	l.Topics = dec.Topics
	if dec.Data == nil {
		return errors.New("missing required field 'data' for Log")
	}
	l.Data = *dec.Data
	if dec.BlockNumber == nil {
		return errors.New("missing required field 'blockNumber' for Log")
	}
	l.BlockNumber = uint64(*dec.BlockNumber)
	if dec.TxHash == nil {
		return errors.New("missing required field 'transactionHash' for Log")
	}
	l.TxHash = *dec.TxHash
	if dec.PkgIndex == nil {
		return errors.New("missing required field 'packageIndex' for Log")
	}
	l.PkgIndex = *dec.PkgIndex
	if dec.TxIndex == nil {
		return errors.New("missing required field 'transactionIndex' for Log")
	}
	l.TxIndex = *dec.TxIndex
	if dec.Index == nil {
		return errors.New("missing required field 'logIndex' for Log")
	}
	l.Index = *dec.Index
	return nil
}
