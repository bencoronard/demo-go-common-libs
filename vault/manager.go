package vault

import (
	"context"
	"sync"

	"go.uber.org/fx"
)

type VaultClientManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type vaultClientManager struct {
	client Client
	mu     sync.Mutex
}

func (v *vaultClientManager) Start(ctx context.Context) error {
	panic("unimplemented")
}

func (v *vaultClientManager) Stop(ctx context.Context) error {
	panic("unimplemented")
}

func NewVaultTokenManager(lc fx.Lifecycle, vc Client) (VaultClientManager, error) {
	return &vaultClientManager{
		client: vc,
	}, nil
}
