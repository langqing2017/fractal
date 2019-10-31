package main

import "C"
import (
	"github.com/langqing2017/fractal/common"
	"github.com/langqing2017/fractal/common/hexutil"
	"github.com/langqing2017/fractal/core/types"
	"github.com/langqing2017/fractal/rpc/client"
	"github.com/langqing2017/fractal/utils/log"
	"gopkg.in/urfave/cli.v1"
)

var (
	blockCommand = cli.Command{
		Name:  "block",
		Usage: "Query Block",
		Flags: []cli.Flag{
			RpcFlag,
			BlockHeightFlag,
			BlockHashFlag,
		},
		Subcommands: []cli.Command{
			{
				Name:   "query",
				Usage:  "Query Block Detail",
				Action: queryBlock,
				Flags: []cli.Flag{
					RpcFlag,
					BlockHeightFlag,
					BlockHashFlag,
				},
			},
		},
	}
)

func queryBlock(ctx *cli.Context) error {
	initLogger(ctx)

	rpc := ctx.GlobalString(RpcFlag.Name)
	client, err := rpcclient.Dial(rpc)
	if err != nil {
		log.Error("connect to rpc error", "rpc", rpc)
		return err
	}

	var block *types.Block
	if ctx.GlobalIsSet(BlockHeightFlag.Name) {
		height := ctx.GlobalUint64(BlockHeightFlag.Name)
		err = client.Call(&block, "ftl_getBlockByHeight", hexutil.Uint64(height))
		if err != nil || block == nil {
			log.Error("get block error", "err", err)
			return err
		}
	}
	if ctx.GlobalIsSet(BlockHashFlag.Name) {
		hash := ctx.GlobalString(BlockHashFlag.Name)
		log.Info("get block", "hash", common.HexToHash(hash))
		err = client.Call(&block, "ftl_getBlock", common.HexToHash(hash))
		if err != nil || block == nil {
			log.Error("get block error", "err", err)
			return err
		}
	}
	log.Info("get block ok", "block", block)
	return nil
}
