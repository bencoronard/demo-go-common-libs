package vault

import (
	"context"
	"sync"

	"go.uber.org/fx"
)

type VaultClientManager interface {
	renewToken(ctx context.Context) error
}

type vaultClientManager struct {
	vc Client
	mu sync.Mutex
}

func (m *vaultClientManager) renewToken(ctx context.Context) error {
	return m.vc.authenticate(ctx)
}

func NewVaultClientManager(lc fx.Lifecycle, vc Client) (VaultClientManager, error) {
	m := vaultClientManager{vc: vc}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return &m, nil
}
