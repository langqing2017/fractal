package api

import (
	"context"
	"github.com/langqing2017/fractal/chain"
	"github.com/langqing2017/fractal/common"
	"github.com/langqing2017/fractal/core/config"
	"github.com/langqing2017/fractal/core/pool"
	"github.com/langqing2017/fractal/core/types"
	"github.com/langqing2017/fractal/dbwrapper"
	"github.com/langqing2017/fractal/ftl/sync"
	"github.com/langqing2017/fractal/keys"
	"github.com/langqing2017/fractal/logbloom/bloomquery"
	"github.com/langqing2017/fractal/packer"
	"math/big"
)

type fractal interface {
	IsMining() bool
	StartMining() error
	StopMining()
	MiningKeyManager() *keys.MiningKeyManager
	Coinbase() common.Address

	Config() *config.Config
	Packer() packer.Packer
	BlockChain() *chain.BlockChain
	TxPool() pool.Pool
	Signer() types.Signer
	GasPrice() *big.Int
	GetPoolTransactions() types.Transactions

	FtlVersion() int

	Synchronizer() *sync.Synchronizer

	ChainDb() dbwrapper.Database
	GetBlock(ctx context.Context, fullHash common.Hash) *types.Block
	GetBlockStr(blockStr string) *types.Block
	GetReceipts(ctx context.Context, blockHash common.Hash) types.Receipts
	GetLogs(ctx context.Context, blockHash common.Hash) [][]*types.Log

	GetMainBranchBlock(height uint64) (*types.BlockHeader, error)
	BloomRequestsReceiver() chan chan *bloomquery.Retrieval
}
