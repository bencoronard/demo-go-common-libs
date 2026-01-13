package vault

import (
	"context"
	"log"

	vault "github.com/hashicorp/vault/api"
	"go.uber.org/fx"
)

type VaultClientManager interface {
	renewToken(ctx context.Context)
	manageTokenLifecycle() error
}

type vaultClientManager struct {
	vc Client
}

func (m *vaultClientManager) renewToken(ctx context.Context) {
	for {
		_, err := m.vc.authenticate(ctx)
		if err != nil {
			log.Fatalf("unable to authenticate to Vault: %v", err)
		}

		if err := m.manageTokenLifecycle(); err != nil {
			log.Fatalf("unable to start managing token lifecycle: %v", err)
		}
	}
}

func (m *vaultClientManager) manageTokenLifecycle() error {
	token, err := m.vc.client().Auth().Token().LookupSelf()
	if err != nil {
		return err
	}

	if !token.Auth.Renewable {
		log.Printf("Token is not renewable. Re-attempting login")
		return nil
	}

	watcher, err := m.vc.client().NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret:    token,
		Increment: 3600,
	})
	if err != nil {
		return err
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		case err := <-watcher.DoneCh():
			if err != nil {
				log.Printf("Failed to renew token: %v. Re-attempting login", err)
				return nil
			}

			log.Printf("Token can no longer be renewed. Re-attempting login.")
			return nil
		case renewal := <-watcher.RenewCh():
			log.Printf("Successfully renewed: %#v", renewal)
		}
	}
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
