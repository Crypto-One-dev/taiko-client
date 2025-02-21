package flags

import (
	"github.com/urfave/cli/v2"
)

// Flags used by driver.
var (
	L2NodeEngineEndpoint = cli.StringFlag{
		Name:     "l2.engine",
		Usage:    "Engine API RPC endpoint of a L2 ethereum node",
		Required: true,
		Category: driverCategory,
	}
	ThrowawayBlocksBuilderPrivKey = cli.StringFlag{
		Name: "l2.throwawayBlockBuilderPrivKey",
		Usage: "Private key of the L2 throwaway blocks builder," +
			"who will be the suggested fee recipient of L2 throwaway blocks",
		Required: true,
		Category: driverCategory,
	}
	JWTSecret = cli.StringFlag{
		Name:     "jwtSecret",
		Usage:    "Path to a JWT secret to use for authenticated RPC endpoints",
		Required: true,
		Category: driverCategory,
	}
)

// Optional flags used by driver.
var (
	P2PSyncVerifiedBlocks = cli.BoolFlag{
		Name:     "p2p.syncVerifiedBlocks",
		Usage:    "Try P2P syncing verified blocks between L2 nodes, will be helpful to bring a new node online quickly",
		Value:    false,
		Category: driverCategory,
	}
)

// All driver flags.
var DriverFlags = MergeFlags(CommonFlags, []cli.Flag{
	&L2NodeEngineEndpoint,
	&ThrowawayBlocksBuilderPrivKey,
	&JWTSecret,
	&P2PSyncVerifiedBlocks,
})
