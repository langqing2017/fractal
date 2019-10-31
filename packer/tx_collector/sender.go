// Copyright 2018 The go-fractal Authors
// This file is part of the go-fractal library.

// Package tx_collector contains the implementation of tx collector for packer.
package tx_collector

import (
	"bytes"
	"net"

	"github.com/langqing2017/fractal/core/types"
	"github.com/langqing2017/fractal/rlp"
	"github.com/langqing2017/fractal/utils/log"
)

type TxSender struct {
	srvAddr string
	conn    net.Conn
}

func NewTxSender(srvAddr string) *TxSender {
	return &TxSender{
		srvAddr: srvAddr,
	}
}

func (s *TxSender) Connect() error {
	conn, err := net.Dial("tcp", s.srvAddr)
	if err != nil {
		log.Error("")
		return err
	}
	s.conn = conn
	return nil
}

func (s *TxSender) SendTx(tx *types.Transaction) error {
	buf, err := rlp.EncodeToBytes(types.Transactions{tx})
	if err != nil {
		log.Error("TxSender Encode TX error", "err", err)
		return err
	}

	msg := &Msg{
		Code:    TxMsg,
		Size:    uint64(len(buf)),
		Payload: bytes.NewReader(buf),
	}
	err = WriteMsg(s.conn, msg)
	if err != nil {
		log.Error("TxSender WriteMsg error", "srv", s.srvAddr, "err", err)
	}
	return err
}
