package driver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/taikoxyz/taiko-client/pkg/rpc"
	"github.com/urfave/cli/v2"
)

const (
	// Time to wait before the next try, when receiving subscription errors.
	RetryDelay         = 10 * time.Second
	MaxReorgDepth      = 500
	ReorgRollbackDepth = 20
)

// Driver keeps the L2 node's local block chain in sync with the TaikoL1
// contract.
type Driver struct {
	rpc           *rpc.Client
	l2ChainSyncer *L2ChainSyncer
	state         *State

	l1HeadCh   chan *types.Header
	l1HeadSub  event.Subscription
	syncNotify chan struct{}

	ctx context.Context
	wg  sync.WaitGroup
}

// New initializes the given driver instance based on the command line flags.
func (d *Driver) InitFromCli(ctx context.Context, c *cli.Context) error {
	cfg, err := NewConfigFromCliContext(c)
	if err != nil {
		return err
	}

	return InitFromConfig(ctx, d, cfg)
}

// InitFromConfig initializes the driver instance based on the given configurations.
func InitFromConfig(ctx context.Context, d *Driver, cfg *Config) (err error) {
	log.Debug("Driver configurations", "config", cfg)

	d.l1HeadCh = make(chan *types.Header, 1024)
	d.wg = sync.WaitGroup{}
	d.syncNotify = make(chan struct{}, 1)
	d.ctx = ctx

	if d.rpc, err = rpc.NewClient(d.ctx, &rpc.ClientConfig{
		L1Endpoint:       cfg.L1Endpoint,
		L2Endpoint:       cfg.L2Endpoint,
		TaikoL1Address:   cfg.TaikoL1Address,
		TaikoL2Address:   cfg.TaikoL2Address,
		L2EngineEndpoint: cfg.L2EngineEndpoint,
		JwtSecret:        cfg.JwtSecret,
	}); err != nil {
		return err
	}

	if d.state, err = NewState(d.ctx, d.rpc); err != nil {
		return err
	}

	balance, err := d.rpc.L2.BalanceAt(d.ctx, crypto.PubkeyToAddress(cfg.ThrowawayBlocksBuilderPrivKey.PublicKey), nil)
	if err != nil {
		return fmt.Errorf("failed to get throwaway blocks builder balance: %w", err)
	}

	if balance.Cmp(common.Big0) == 0 {
		return fmt.Errorf("throwaway blocks builder has no fund")
	}

	peers, err := d.rpc.L2.PeerCount(d.ctx)
	if err != nil {
		return err
	}

	if cfg.P2PSyncVerifiedBlocks && peers == 0 {
		log.Warn("P2P syncing verified blocks enabled, but no connected peer found in L2 node")
	}

	if d.l2ChainSyncer, err = NewL2ChainSyncer(
		d.ctx,
		d.rpc,
		d.state,
		cfg.ThrowawayBlocksBuilderPrivKey,
		cfg.P2PSyncVerifiedBlocks,
	); err != nil {
		return err
	}

	d.l1HeadSub = d.state.SubL1HeadsFeed(d.l1HeadCh)

	return nil
}

// Start starts the driver instance.
func (d *Driver) Start() error {
	d.wg.Add(1)
	go d.eventLoop()

	return nil
}

// Close closes the driver instance.
func (d *Driver) Close() {
	d.state.Close()
	d.wg.Wait()
}

// eventLoop starts the main loop of L2 node's chain driver.
func (d *Driver) eventLoop() {
	defer d.wg.Done()
	exponentialBackoff := backoff.NewExponentialBackOff()

	// reqSync requests performing a synchronising operation, won't block
	// if we are already synchronising.
	reqSync := func() {
		select {
		case d.syncNotify <- struct{}{}:
		default:
		}
	}

	// doSyncWithBackoff performs a synchronising operation with a backoff strategy.
	doSyncWithBackoff := func() {
		if err := backoff.Retry(d.doSync, exponentialBackoff); err != nil {
			log.Error("Sync L2 node's block chain error", "error", err)
		}
	}

	// Call doSync() right away to catch up with the latest known L1 head.
	doSyncWithBackoff()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-d.syncNotify:
			doSyncWithBackoff()
		case <-d.l1HeadCh:
			reqSync()
		}
	}
}

// doSync fetches all `BlockProposed` events emitted from local
// L1 sync cursor to the L1 head, and then applies all corresponding
// L2 blocks into node's local block chain.
func (d *Driver) doSync() error {
	// Check whether the application is closing.
	if d.ctx.Err() != nil {
		log.Warn("Driver context error", "error", d.ctx.Err())
		return nil
	}

	l1Head := d.state.GetL1Head()

	if err := d.l2ChainSyncer.Sync(l1Head); err != nil {
		log.Error("Process new L1 blocks error", "error", err)
		return err
	}

	return nil
}

// ChainSyncer returns the driver's chain syncer.
func (d *Driver) ChainSyncer() *L2ChainSyncer {
	return d.l2ChainSyncer
}

// Name returns the application name.
func (d *Driver) Name() string {
	return "driver"
}
