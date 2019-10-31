package packer

import (
	"github.com/langqing2017/fractal/core/types"
	"github.com/langqing2017/fractal/event"
)

type Packer interface {
	//
	InsertRemoteTxPackage(pkg *types.TxPackage) error

	// pack_service
	InsertTransactions(txs types.Transactions) []error
	StartPacking(packerIndex uint32)
	StopPacking()
	IsPacking() bool
	Subscribe(ch chan<- NewPackageEvent) event.Subscription
}

type NewPackageEvent struct {
	Pkgs []*types.TxPackage
}
