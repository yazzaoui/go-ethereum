package backends

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/ethdb"
)

// InternalBackend implements the contract.Backend interface to interact with the
// protocol contracts. This is used internaly by the accountability module.
type InternalBackend struct {
	SimulatedBackend
	TxSender *func(signedTx *types.Transaction) error
}

func NewInternalBackend(blockchain *core.BlockChain, db ethdb.Database) bind.ContractBackend {
	backend := &InternalBackend{
		SimulatedBackend{
			database:   db,
			blockchain: blockchain,
			config:     blockchain.Config(),
		}, nil}

	filterBackend := &filterBackend{db, blockchain, &backend.SimulatedBackend}
	filterSystem := filters.NewFilterSystem(filterBackend, filters.Config{})
	backend.events = filters.NewEventSystem(filterSystem, false)
	backend.filterSystem = filterSystem
	return backend
}

// SuggestGasTipCap doesn't provide any tips with accountability.
func (b *InternalBackend) SuggestGasTipCap(_ context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

// PendingCodeAt is used for gas estimation but it really doesn't matter, we can use the current state
func (b *InternalBackend) PendingCodeAt(ctx context.Context, contract common.Address) ([]byte, error) {
	return b.CodeAt(ctx, contract, nil)
}

func (b *InternalBackend) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	// very hacky but that will avoid us further complications
	b.SimulatedBackend.pendingBlock = b.blockchain.GetBlockByHash(b.blockchain.CurrentBlock().Hash())
	b.SimulatedBackend.pendingState, _ = b.blockchain.State()
	return b.SimulatedBackend.EstimateGas(ctx, call)
}

func (b *InternalBackend) SendTransaction(_ context.Context, tx *types.Transaction) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return (*b.TxSender)(tx)
}
