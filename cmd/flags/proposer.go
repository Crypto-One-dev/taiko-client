package flags

import (
	"math/rand"

	"github.com/urfave/cli/v2"
)

// Required flags used by proposer.
var (
	L1ProposerPrivKey = cli.StringFlag{
		Name:     "l1.proposerPrivKey",
		Usage:    "Private key of the L1 proposer, who will send TaikoL1.proposeBlock transactions",
		Required: true,
		Category: proposerCategory,
	}
	L2SuggestedFeeRecipient = cli.StringFlag{
		Name:     "l2.suggestedFeeRecipient",
		Usage:    "Address of the proposed block's suggested fee recipient",
		Required: true,
		Category: proposerCategory,
	}
	ProposeInterval = cli.StringFlag{
		Name:     "proposeInterval",
		Usage:    "Time interval to propose L2 pending transactions",
		Required: true,
		Category: proposerCategory,
	}
)

// Optional flags used by proposer.
var (
	CommitSlot = cli.Uint64Flag{
		Name:     "commitSlot",
		Usage:    "The commit slot will be used by proposer, by default, a random number will be used",
		Value:    rand.Uint64(),
		Category: proposerCategory,
	}
	ShufflePoolContent = cli.BoolFlag{
		Name:     "shufflePoolContent",
		Usage:    "Perform a weighted shuffle when building the transactions list to propose",
		Value:    false,
		Category: proposerCategory,
	}
)

// Special flags for testing.
var (
	ProduceInvalidBlocks = cli.BoolFlag{
		Name:     "produceInvalidBlocks",
		Usage:    "Special flag for testnet testing, if activated, the proposer will start producing bad blocks",
		Hidden:   true,
		Category: proposerCategory,
	}
	ProduceInvalidBlocksInterval = cli.Uint64Flag{
		Name:     "produceInvalidBlocksInterval",
		Usage:    "Special flag for testnet testing, if activated, bad blocks will be produced every N valid blocks",
		Hidden:   true,
		Category: proposerCategory,
	}
)

// All proposer flags.
var ProposerFlags = MergeFlags(CommonFlags, []cli.Flag{
	&L1ProposerPrivKey,
	&L2SuggestedFeeRecipient,
	&ProposeInterval,
	&ProduceInvalidBlocks,
	&ProduceInvalidBlocksInterval,
	&ShufflePoolContent,
	&CommitSlot,
})
