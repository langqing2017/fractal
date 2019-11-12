package main

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/fractal-platform/fractal/cmd/utils"
	"github.com/fractal-platform/fractal/common"
	"github.com/fractal-platform/fractal/core/config"
	"github.com/fractal-platform/fractal/crypto"
	"github.com/fractal-platform/fractal/ftl/protocol"
	"github.com/fractal-platform/fractal/p2p"
	"github.com/fractal-platform/fractal/p2p/discover"
	"github.com/fractal-platform/fractal/utils/log"
	"gopkg.in/urfave/cli.v1"
	"os"
	"sort"
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit   = ""
	versionMeta = "unstable" // Version metadata to append to the version string
	// The app that holds all commands and flags.
	app = utils.NewApp(versionMeta, gitCommit, "the node finder command line interface")
)

func init() {
	// define help template
	cli.AppHelpTemplate = `{{.Name}} [options]

VERSION:
   {{.Version}}

{{if .Flags}}OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

	// Initialize the CLI app and start gftl
	app.Action = find_nodes
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2013-2019 The go-fractal Authors"
	app.Commands = []cli.Command{}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Before = func(ctx *cli.Context) error {
		// setup logger
		log.SetDefaultLogger(log.InitMultipleLog15Logger(log.LvlInfo, os.Stdout, os.Stdout))

		return nil
	}

	app.After = func(ctx *cli.Context) error {
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// NodeInfo represents a short summary of the Fractal sub-protocol metadata
// known about the host
type NodeInfo struct {
	Network uint64      `json:"network"` // Fractal network ID (1=Mainnet, 2=Testnet)
	Height  uint64      `json:"Height"`  // Height of the host's blockchain
	Genesis common.Hash `json:"genesis"` // SHA3 Hash of the host's genesis block
	Head    common.Hash `json:"head"`    // SHA3 Hash of the host's best owned block
}

func find_nodes(ctx *cli.Context) error {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Crit(fmt.Sprintf("Failed to generate node key: %v", err))
		return err
	}

	// bootnode for testnet
	bootnode, _ := discover.ParseNode("enode://5b736302b16b83e5ae102de228ffd376b4cb4748a136057ea84bbbd6d1026a18aa902168af2d15f47ff9c300414bf6999f6525f3b18e9225afb70c7b35dd22ed@161.189.2.180:60002")

	serverConfig := p2p.Config{}
	serverConfig.MaxPeers = 10000
	serverConfig.PrivateKey = key
	serverConfig.Name = "nodefinder"
	serverConfig.Logger = log.NewSubLogger()
	serverConfig.BootstrapNodes = append(serverConfig.BootstrapNodes, bootnode)
	serverConfig.DiscListenAddr = fmt.Sprintf(":%d", 41234)
	serverConfig.RwListenType = uint8(1)
	serverConfig.RwListenAddr = fmt.Sprintf(":%d", 41234)

	var nodes = mapset.NewSet()
	protocol := p2p.Protocol{
		Name:    protocol.ProtocolName,
		Version: protocol.ProtocolVersions[0],
		Length:  protocol.ProtocolLengths[0],
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			nodes.Add(p.ID())
			p.Log().Info("Fractal peer connected", "name", p.Name(), "ip", p.RemoteAddr(), "id", p.ID(), "count", nodes.Cardinality())
			return nil
		},
		NodeInfo: func() interface{} {
			return &NodeInfo{
				Network: config.DefaultTestnetConfig.ChainConfig.ChainID,
				Height:  0,
				Genesis: config.DefaultTestnetConfig.Genesis.ToBlock(nil).FullHash(),
				Head:    config.DefaultTestnetConfig.Genesis.ToBlock(nil).FullHash(),
			}
		},
		PeerInfo: func(id discover.NodeID) interface{} {
			return nil
		},
	}

	running := &p2p.Server{Config: serverConfig}
	running.Protocols = append(running.Protocols, protocol)
	log.Info("Starting peer-to-peer node", "instance", serverConfig.Name)
	if err := running.Start(); err != nil {
		log.Error("start p2p server failed", "err", err.Error())
		return err
	}

	select {}
	return nil
}

