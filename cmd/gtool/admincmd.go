package main

import (
	"fmt"
	"github.com/langqing2017/fractal/common"
	"github.com/langqing2017/fractal/common/hexutil"
	"github.com/langqing2017/fractal/keys"
	"github.com/langqing2017/fractal/p2p"
	"github.com/langqing2017/fractal/rpc/client"
	"github.com/langqing2017/fractal/utils/log"
	"gopkg.in/urfave/cli.v1"
)

var (
	adminCommand = cli.Command{
		Name:  "admin",
		Usage: "Manage Fractal Node",
		Flags: []cli.Flag{
			RpcFlag,
			AddressFlag,
		},
		Subcommands: []cli.Command{
			{
				Name:   "info",
				Usage:  "Show Fractal Node Info",
				Action: showInfo,
				Flags: []cli.Flag{
					RpcFlag,
				},
			},
			{
				Name:   "enode",
				Usage:  "Show Fractal Node Enode Address",
				Action: showEnode,
				Flags: []cli.Flag{
					RpcFlag,
				},
			},
			{
				Name:   "genminingkey",
				Usage:  "Generate Mining Key fro Current Address",
				Action: generateMiningKey,
				Flags: []cli.Flag{
					RpcFlag,
					AddressFlag,
				},
			},
		},
	}
)

func showInfo(ctx *cli.Context) error {
	initLogger(ctx)

	rpc := ctx.GlobalString(RpcFlag.Name)
	client, err := rpcclient.Dial(rpc)
	if err != nil {
		log.Error("connect to rpc error", "rpc", rpc)
		return err
	}

	var info *p2p.NodeInfo
	err = client.Call(&info, "net_nodeInfo")
	if err != nil {
		log.Error("get node info error", "err", err)
		return err
	}
	log.Info("get node info ok", "node", info)

	return nil
}

func showEnode(ctx *cli.Context) error {
	initLogger(ctx)

	rpc := ctx.GlobalString(RpcFlag.Name)
	client, err := rpcclient.Dial(rpc)
	if err != nil {
		log.Error("connect to rpc error", "rpc", rpc)
		return err
	}

	var info *p2p.NodeInfo
	err = client.Call(&info, "net_nodeInfo")
	if err != nil || info == nil {
		log.Error("get node info error", "err", err)
		return err
	}
	fmt.Println(info.Enode)

	return nil
}

func generateMiningKey(ctx *cli.Context) error {
	initLogger(ctx)

	addressString := ctx.GlobalString(AddressFlag.Name)
	address := common.HexToAddress(addressString)

	rpc := ctx.GlobalString(RpcFlag.Name)
	client, err := rpcclient.Dial(rpc)
	if err != nil {
		log.Error("connect to rpc error", "rpc", rpc)
		return err
	}

	var pubkey keys.MiningPubkey
	err = client.Call(&pubkey, "admin_generateMiningKey", address)
	if err != nil {
		log.Error("generate mining key error", "err", err)
		return err
	}
	log.Info("generate mining key ok", "pubkey", hexutil.Encode(pubkey[:]))

	return nil
}
